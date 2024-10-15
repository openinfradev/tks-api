package policytemplate

import (
	"encoding/json"
	"strings"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PartOfKey                 = "app.kubernetes.io/part-of"
	PartOfVal                 = "tks-policy-operator"
	TksLabelPrefix            = "tks/"
	PolicyIDLabel             = TksLabelPrefix + "policy-id"
	TemplateIDLabel           = TksLabelPrefix + "policy-template-id"
	RequireSyncDataAnnotation = "metadata.gatekeeper.sh/requires-sync-data"
)

func PolicyToTksPolicyCR(policy *model.Policy) *TKSPolicy {
	if policy == nil {
		return nil
	}

	var params *apiextensionsv1.JSON = &apiextensionsv1.JSON{Raw: []byte("{}")}

	var jsonResult map[string]interface{}

	err := json.Unmarshal([]byte(policy.Parameters), &jsonResult)

	if err == nil && len(jsonResult) > 0 {
		jsonParams := apiextensionsv1.JSON{Raw: []byte(policy.Parameters)}
		params = &jsonParams
	}

	labels := map[string]string{}
	labels[PartOfKey] = PartOfVal
	labels[PolicyIDLabel] = policy.ID.String()
	labels[TemplateIDLabel] = policy.TemplateId.String()

	if policy.MatchYaml != nil {
		var match domain.Match

		err := yaml.Unmarshal([]byte(*policy.MatchYaml), &match)

		if err != nil {
			policy.Match = &match
		}
	}

	targetClusterIds := make([]string, 0)

	if policy.TargetClusterIds != nil {
		targetClusterIds = policy.TargetClusterIds
	}

	return &TKSPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tkspolicy.openinfradev.github.io/v1",
			Kind:       "TKSPolicy",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:   policy.PolicyResourceName,
			Labels: labels,
		},

		Spec: TKSPolicySpec{
			EnforcementAction: policy.EnforcementAction,
			Clusters:          targetClusterIds,
			Template:          policy.PolicyTemplate.Kind,
			Match:             policy.Match,
			Parameters:        params,
		},
	}
}

func PolicyTemplateToTksPolicyTemplateCR(policyTemplate *model.PolicyTemplate) *TKSPolicyTemplate {
	if policyTemplate == nil {
		return nil
	}

	labels := map[string]string{}
	labels[PartOfKey] = PartOfVal
	labels[TemplateIDLabel] = policyTemplate.ID.String()

	annotations := map[string]string{}

	if policyTemplate.SyncJson != nil {
		annotations[RequireSyncDataAnnotation] = "\"" + *policyTemplate.SyncJson + "\""
	} else if policyTemplate.SyncKinds != nil {
		sync, err := CheckAndConvertToSyncData(*policyTemplate.SyncKinds)

		if err == nil {
			syncVal, err := MarshalSyncData(sync)
			if err == nil {
				annotations[RequireSyncDataAnnotation] = "\"" + syncVal + "\""
			}
		}
	}

	var validation *Validation = nil

	if len(policyTemplate.ParametersSchema) > 0 {
		validation = &Validation{
			OpenAPIV3Schema: ParamDefsToJSONSchemaProeprties(policyTemplate.ParametersSchema, false),
		}
	}

	return &TKSPolicyTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tkspolicy.openinfradev.github.io/v1",
			Kind:       "TKSPolicyTemplate",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:        strings.ToLower(policyTemplate.Kind),
			Labels:      labels,
			Annotations: annotations,
		},

		Spec: TKSPolicyTemplateSpec{
			CRD: CRD{
				Spec: CRDSpec{
					Names: Names{
						Kind: policyTemplate.Kind,
					},
					Validation: validation,
				},
			},
			Targets: []Target{{
				Target: "admission.k8s.gatekeeper.sh",
				Rego:   stripCarriageReturn(AddTksGuardToRego(policyTemplate.Rego)),
				Libs:   stripCarriageReturns(policyTemplate.Libs),
			}},
			Version: policyTemplate.Version,
		},
	}
}

func stripCarriageReturn(str string) string {
	return strings.ReplaceAll(str, "\r", "")
}

func stripCarriageReturns(strs []string) []string {
	if strs == nil {
		return nil
	}

	result := make([]string, len(strs))

	for i, str := range strs {
		result[i] = stripCarriageReturn(str)
	}

	return result

}

func syncToKubernetes() bool {
	return true
	// return os.Getenv("SYNC_POLICY_TO_K8S") != ""
}
