package policytemplate

import (
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

func CheckAndNormalizeKinds(kinds []domain.Kinds) ([]domain.Kinds, error) {
	if kinds == nil {
		return nil, nil
	}

	var result = []domain.Kinds{}
	var invalidKinds = []string{}
	var normalizedMap = map[string]domain.Kinds{}

	for _, kind := range kinds {
		for _, kinditem := range kind.Kinds {
			if apiGroup, ok := KindToApiGroup[kinditem]; ok {
				if ai, ok := normalizedMap[apiGroup]; ok {
					if !slices.Contains(ai.Kinds, kinditem) {
						ai.Kinds = append(ai.Kinds, kinditem)
						normalizedMap[apiGroup] = ai
					}
				} else {
					normalizedMap[apiGroup] = domain.Kinds{
						APIGroups: []string{apiGroup},
						Kinds:     []string{kinditem},
					}
				}
			} else {
				invalidKinds = append(invalidKinds, kinditem)
			}
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
