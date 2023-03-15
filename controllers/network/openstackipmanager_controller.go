/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	networkv1beta1 "github.com/openstack-k8s-operators/infra-operator/apis/network/v1beta1"
	network "github.com/openstack-k8s-operators/infra-operator/pkg/network"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
)

// OpenStackIPManagerReconciler reconciles a OpenStackIPManager object
type OpenStackIPManagerReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Scheme  *runtime.Scheme
	Log     logr.Logger
}

//+kubebuilder:rbac:groups=network.openstack.org,resources=openstackipmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.openstack.org,resources=iprequesters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.openstack.org,resources=openstackipmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.openstack.org,resources=openstackipmanagers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenStackIPManager object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *OpenStackIPManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	_ = log.FromContext(ctx)
	r.Log.Info("Reconciling IPManager")
	instance := &networkv1beta1.OpenStackIPManager{}
	name := client.ObjectKey{
		Namespace: req.Namespace, Name: "ipmanager"}
	err := r.Get(ctx, name, instance)
	if err != nil {
		r.Log.Error(err, "Unable to retrieve IPManager CR")
		return ctrl.Result{}, err
	}

	h, err := helper.NewHelper(
		instance,
		r.Client,
		r.Kclient,
		r.Scheme,
		r.Log)

	if err != nil {
		// helper might be nil, so can't use util.LogErrorForObject since it requires helper as first arg
		r.Log.Error(err, fmt.Sprintf("unable to acquire helper for OpenStackIPManager %s", instance.Name))
		return ctrl.Result{}, err
	}

	if instance.Status.Conditions == nil {
		instance.Status.Conditions = condition.Conditions{}
		cl := condition.CreateList(
			condition.FalseCondition(networkv1beta1.OpenStackIPManagerConditionUpdateInProgress, "", "", condition.ReadyInitMessage))
		instance.Status.Conditions.Init(&cl)
	}

	defer func() {
		if err = h.SetAfter(instance); err != nil {
			_err = err
			return
		}
		changes := h.GetChanges()
		if changes["spec"] {
			instance.Status.Conditions.MarkTrue(networkv1beta1.OpenStackIPManagerConditionUpdateInProgress, networkv1beta1.IPManagerInProgressMessage)
			patch := client.MergeFrom(h.GetBeforeObject())
			err = h.GetClient().Patch(ctx, instance, patch)
		} else {
			err = h.PatchInstance(ctx, instance)
		}
		if err != nil {
			_err = err
			return
		}
	}()

	if req.Name == "ipmanager" {
		r.Log.Info(fmt.Sprintf("@@@@@@ Current CR Version: %v", instance.GetResourceVersion()))
		r.Log.Info(fmt.Sprintf("@@@@@@ Instance In-Progress: %v", instance.Status.Conditions.Get(networkv1beta1.OpenStackIPManagerConditionUpdateInProgress)))
		if IsInProgress(instance) {
			instance.Status.Conditions.MarkFalse(networkv1beta1.OpenStackIPManagerConditionUpdateInProgress, "", "", networkv1beta1.IPManagerCompleteMessage)
		}
		instance.Status.Networks = instance.Spec.Networks
		return ctrl.Result{}, nil
	}

	if IsInProgress(instance) {
		// Another update in progress retry
		r.Log.Info("Another Update In-Progress")
		return ctrl.Result{}, fmt.Errorf("Another Update In Progress")
	}

	// Get All IPRequestors
	requestor := &networkv1beta1.IPRequester{}
	reqName := client.ObjectKey{
		Namespace: req.Namespace, Name: req.Name}

	err = r.Get(ctx, reqName, requestor)
	if err != nil {
		r.Log.Error(err, "Unable to retrieve IPRequestor CR")
		return ctrl.Result{}, err
	}
	reqHelper, err := helper.NewHelper(
		requestor,
		r.Client,
		r.Kclient,
		r.Scheme,
		r.Log)

	if err != nil {
		// helper might be nil, so can't use util.LogErrorForObject since it requires helper as first arg
		r.Log.Error(err, fmt.Sprintf("unable to acquire helper for IPRequestor %s", requestor.Name))
		return ctrl.Result{}, err
	}
	netValues := []networkv1beta1.Network{}
	annotations := requestor.GetAnnotations()
	exists, err := r.ParseAnnotationNetworks(network.NetworksSuffix, &netValues, annotations)
	if exists {
		ipValues := map[string]networkv1beta1.IPReservation{}
		exists, _ := r.ParseAnnotationIPReservation(network.IPSetsSuffix, &ipValues, annotations)
		if !exists {
			r.Log.Info("Reservation does not exist adding...")
			// Reserve and Build the IPReservation Annotation
			for _, n := range netValues {
				aip := network.AssignIPDetails{}
				netN := instance.Status.Networks[n.Name]
				for i, subnet := range netN.Subnets {
					if n.SubnetName == subnet.Name {
						if instance.Spec.Networks[n.Name].Subnets[i].Reservations == nil {
							instance.Spec.Networks[n.Name].Subnets[i].Reservations = make(map[string]networkv1beta1.IPReservation)
						}
						_, cidr, _ := net.ParseCIDR(subnet.Cidr)
						aip.IPNet = *cidr
						aip.RangeStart = net.ParseIP(subnet.AllocationRange.AllocationStart)
						aip.RangeEnd = net.ParseIP(subnet.AllocationRange.AllocationEnd)
						aip.Reservelist = maps.Values(subnet.Reservations)
						reservation, _, _ := network.AssignIP(aip, net.ParseIP(n.FixedIP))
						instance.Spec.Networks[n.Name].Subnets[i].Reservations[req.Name] = reservation
						r.Log.Info(fmt.Sprintf("##### Current CR Version: %v", instance.GetResourceVersion()))
						ipValues[n.Name] = reservation
						break
					}
				}
			}
			updatedAnnotation, _ := json.Marshal(ipValues)
			SetMetaDataAnnotation(requestor, network.IPSetsSuffix, string(updatedAnnotation))
			err = reqHelper.PatchInstance(ctx, requestor)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenStackIPManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1beta1.OpenStackIPManager{}).
		Watches(&source.Kind{Type: &networkv1beta1.IPRequester{}},
			&handler.EnqueueRequestForObject{}, builder.WithPredicates(predicate.AnnotationChangedPredicate{})).
		Complete(r)
}

// ParseAnnotationIPReservation - Parses IPResevation annotation
func (r *OpenStackIPManagerReconciler) ParseAnnotationIPReservation(annotation string,
	values *map[string]networkv1beta1.IPReservation, annotations map[string]string) (bool, error) {
	raw := ""
	exists, matchedKey := network.ParseStringAnnotation(annotation, &raw, annotations)
	if !exists {
		return false, nil
	}
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return true, fmt.Errorf("failed to parse json annotation, %v: %v", matchedKey, raw)
	}
	return true, nil
}

// ParseAnnotationNetworks - Parses Networks annotation
func (r *OpenStackIPManagerReconciler) ParseAnnotationNetworks(annotation string, values *[]networkv1beta1.Network, annotations map[string]string) (bool, error) {
	raw := ""
	exists, matchedKey := network.ParseStringAnnotation(annotation, &raw, annotations)
	if !exists {
		return false, nil
	}

	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return true, fmt.Errorf("failed to parse json annotation, %v: %v", matchedKey, raw)
	}
	return true, nil

}

// SetMetaDataAnnotation sets the annotation on the given object.
// If the given Object did not yet have annotations, they are initialized.
func SetMetaDataAnnotation(meta metav1.Object, key, value string) {
	annotations := meta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[network.BuildAnnotationKey(key)] = value
	meta.SetAnnotations(annotations)
}

// IsInProgress InProgess
func IsInProgress(instance *networkv1beta1.OpenStackIPManager) bool {
	updateInprogressCond := instance.Status.Conditions.Get(networkv1beta1.OpenStackIPManagerConditionUpdateInProgress)
	return updateInprogressCond != nil && updateInprogressCond.Status == corev1.ConditionTrue

}
