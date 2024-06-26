#!/bin/bash
set -ex

TMPDIR=${TMPDIR:-"/tmp/k8s-webhook-server/serving-certs"}
SKIP_CERT=${SKIP_CERT:-false}
CRC_IP=${CRC_IP:-$(/sbin/ip -o -4 addr list crc | awk '{print $4}' | cut -d/ -f1)}
FIREWALL_ZONE=${FIREWALL_ZONE:-"libvirt"}

#Open 9443
sudo firewall-cmd --zone=${FIREWALL_ZONE} --add-port=9443/tcp
sudo firewall-cmd --runtime-to-permanent

# Generate the certs and the ca bundle
if [ "$SKIP_CERT" = false ] ; then
    mkdir -p ${TMPDIR}
    rm -rf ${TMPDIR}/* || true

    openssl req -newkey rsa:2048 -days 3650 -nodes -x509 \
    -subj "/CN=${HOSTNAME}" \
    -addext "subjectAltName = IP:${CRC_IP}" \
    -keyout ${TMPDIR}/tls.key \
    -out ${TMPDIR}/tls.crt

    cat ${TMPDIR}/tls.crt ${TMPDIR}/tls.key | base64 -w 0 > ${TMPDIR}/bundle.pem

fi

CA_BUNDLE=`cat ${TMPDIR}/bundle.pem`

# Patch the webhook(s)
cat >> ${TMPDIR}/patch_webhook_configurations.yaml <<EOF_CAT
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: vmemcached.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/validate-memcached-openstack-org-v1beta1-memcached
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: vmemcached.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - memcached.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - memcacheds
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mmemcached.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/mutate-memcached-openstack-org-v1beta1-memcached
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: mmemcached.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - memcached.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - memcacheds
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: vdnsmasq.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/validate-network-openstack-org-v1beta1-dnsmasq
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: vdnsmasq.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - dnsmasqs
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mdnsmasq.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/mutate-network-openstack-org-v1beta1-dnsmasq
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: mdnsmasq.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - dnsmasqs
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: vnetconfig.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/validate-network-openstack-org-v1beta1-netconfig
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: vnetconfig.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - netconfigs
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mnetconfig.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/mutate-network-openstack-org-v1beta1-netconfig
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: mnetconfig.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - netconfigs
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: vreservation.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/validate-network-openstack-org-v1beta1-reservation
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: vreservation.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - reservations
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mreservation.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/mutate-network-openstack-org-v1beta1-reservation
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: mreservation.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - reservations
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: vipset.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/validate-network-openstack-org-v1beta1-ipset
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: vipset.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - ipsets
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mipset.kb.io
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: ${CA_BUNDLE}
    url: https://${CRC_IP}:9443/mutate-network-openstack-org-v1beta1-ipset
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: mipset.kb.io
  objectSelector: {}
  rules:
  - apiGroups:
    - network.openstack.org
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - ipsets
    scope: '*'
  sideEffects: None
  timeoutSeconds: 10
EOF_CAT

oc apply -n openstack -f ${TMPDIR}/patch_webhook_configurations.yaml
