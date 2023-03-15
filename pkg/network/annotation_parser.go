package network

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	// AnnotationPrefix Standard Prefix for all annotations
	AnnotationPrefix = "openstack-k8s-operators.org"
	// NetworksSuffix Networks Suffix
	NetworksSuffix = "networks"
	// IPSetsSuffix IPSets Suffix
	IPSetsSuffix = "ipsets"
)

// ParseJSONAnnotation Parses json annotation
func ParseJSONAnnotation(annotation string, value interface{}, annotations map[string]string) (bool, error) {
	raw := ""
	exists, matchedKey := ParseStringAnnotation(annotation, &raw, annotations)
	if !exists {
		return false, nil
	}
	if err := json.Unmarshal([]byte(raw), value); err != nil {
		return true, errors.Wrapf(err, "failed to parse json annotation, %v: %v", matchedKey, raw)
	}
	return true, nil
}

// ParseStringMapAnnotation Parses string map annotation
func ParseStringMapAnnotation(annotation string, value *map[string]string, annotations map[string]string) (bool, error) {
	raw := ""
	exists, matchedKey := ParseStringAnnotation(annotation, &raw, annotations)
	if !exists {
		return false, nil
	}
	rawKVPairs := splitString(raw, ",")
	keyValues := make(map[string]string)
	for _, kvPair := range rawKVPairs {
		parts := strings.SplitN(kvPair, "=", 2)
		if len(parts) != 2 {
			return false, errors.Errorf("failed to parse stringMap annotation, %v: %v", matchedKey, raw)
		}
		key := parts[0]
		value := parts[1]
		if len(key) == 0 {
			return false, errors.Errorf("failed to parse stringMap annotation, %v: %v", matchedKey, raw)
		}
		keyValues[key] = value
	}
	if value != nil {
		*value = keyValues
	}
	return true, nil
}

// ParseStringAnnotation Parses string annotation
func ParseStringAnnotation(annotation string, value *string, annotations map[string]string) (bool, string) {
	key := BuildAnnotationKey(annotation)
	if raw, ok := annotations[key]; ok {
		*value = raw
		return true, key
	}
	return false, ""
}

// BuildAnnotationKey returns list of full annotation keys based on suffix and parse options
func BuildAnnotationKey(suffix string) string {
	return fmt.Sprintf("%v/%v", AnnotationPrefix, suffix)
}

func splitString(separatedString string, separator string) []string {
	var result []string
	parts := strings.Split(separatedString, separator)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		result = append(result, part)
	}
	return result
}
