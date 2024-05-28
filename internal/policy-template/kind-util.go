package policytemplate

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/openinfradev/tks-api/pkg/domain"
)

var KindToApiGroup = map[string]string{
	"Pod":                            "",
	"Node":                           "",
	"Namespace":                      "",
	"Service":                        "",
	"Secret":                         "",
	"ConfigMap":                      "",
	"PersistentVolume":               "",
	"PersistentVolumeClaim":          "",
	"ReplicationController":          "", //deprecated
	"ServiceAccount":                 "",
	"LimitRange":                     "",
	"ResourceQuota":                  "",
	"Deployment":                     "apps",
	"ReplicaSet":                     "apps",
	"StatefulSet":                    "apps",
	"DaemonSet":                      "apps",
	"HorizontalPodAutoscaler":        "autoscaling",
	"VerticalPodAutoscaler":          "autoscaling", // 확인 안됨
	"Job":                            "batch",
	"CronJob":                        "batch",
	"Ingress":                        "networking.k8s.io",
	"NetworkPolicy":                  "networking.k8s.io",
	"StorageClass":                   "storage.k8s.io",
	"VolumeAttachment":               "storage.k8s.io",
	"Role":                           "rbac.authorization.k8s.io",
	"RoleBinding":                    "rbac.authorization.k8s.io",
	"ClusterRole":                    "rbac.authorization.k8s.io",
	"ClusterRoleBinding":             "rbac.authorization.k8s.io",
	"ValidatingWebhookConfiguration": "admissionregistration.k8s.io",
	"MutatingWebhookConfiguration":   "admissionregistration.k8s.io",
	"CustomResourceDefinition":       "apiextensions.k8s.io",
	"Certificate":                    "cert-manager.io",
	"Issuer":                         "cert-manager.io",
	"Lease":                          "coordination.k8s.io",
	"Lock":                           "coordination.k8s.io", // 안 나옴
	"EndpointSlice":                  "discovery.k8s.io",
	"Event":                          "events.k8s.io",
	"FlowSchema":                     "flowcontrol.apiserver.k8s.io",
	"PriorityLevelConfiguration":     "flowcontrol.apiserver.k8s.io",
	"ManagedNamespacedResource":      "meta.k8s.io",
	"PriorityClass":                  "scheduling.k8s.io",
	"PodSecurityPolicy":              "policy",
	"PodDisruptionBudget":            "policy",
}

func CheckAndNormalizeKinds(kinds []string) ([]domain.Kinds, error) {
	if kinds == nil {
		return nil, nil
	}

	var result = []domain.Kinds{}
	var invalidKinds = []string{}
	var normalizedMap = map[string]domain.Kinds{}

	for _, kind := range kinds {
		if apiGroup, ok := KindToApiGroup[kind]; ok {
			if ai, ok := normalizedMap[apiGroup]; ok {
				if !slices.Contains(ai.Kinds, kind) {
					ai.Kinds = append(ai.Kinds, kind)
					normalizedMap[apiGroup] = ai
				}
			} else {
				normalizedMap[apiGroup] = domain.Kinds{
					APIGroups: []string{apiGroup},
					Kinds:     []string{kind},
				}
			}
		} else {
			invalidKinds = append(invalidKinds, kind)
		}
	}

	if len(invalidKinds) > 0 {
		return nil, fmt.Errorf("invalid kinds: %v", invalidKinds)
	}

	for _, nornormalized := range normalizedMap {
		result = append(result, nornormalized)
	}

	return result, nil
}

func CheckAndConvertToSyncData(kinds []string) (*[][]domain.CompactGVKEquivalenceSet, error) {
	invalid_kinds := []string{}
	results := []domain.CompactGVKEquivalenceSet{}

	for _, kind := range kinds {
		gvk, ok := KindMap[kind]

		if ok {
			results = append(results, domain.CompactGVKEquivalenceSet{
				Groups: []string{gvk.Group}, Versions: []string{gvk.Version}, Kinds: []string{gvk.Kind},
			})
		} else {
			invalid_kinds = append(invalid_kinds, kind)
		}
	}

	if len(invalid_kinds) > 0 {
		return nil, fmt.Errorf("invalid kinds %v", invalid_kinds)
	}

	return &[][]domain.CompactGVKEquivalenceSet{
		results,
	}, nil
}

func MarshalSyncData(syncData *[][]domain.CompactGVKEquivalenceSet) (string, error) {
	result, err := json.MarshalIndent(syncData, "", "  ")

	if err != nil {
		return "", err
	}

	return string(result), nil
}

func ParseAndCheckSyncData(syncjson string) (*[][]domain.CompactGVKEquivalenceSet, error) {
	result := [][]domain.CompactGVKEquivalenceSet{}
	err := json.Unmarshal([]byte(syncjson), &result)

	if err != nil {
		return nil, err
	}

	invalid_kinds := []string{}

	for _, sets := range result {
		for _, set := range sets {
			for _, kind := range set.Kinds {
				if _, ok := KindMap[kind]; !ok {
					invalid_kinds = append(invalid_kinds, kind)
				}
			}
		}
	}

	if len(invalid_kinds) > 0 {
		return nil, fmt.Errorf("invalid kinds %v", invalid_kinds)
	}

	return &result, nil
}
