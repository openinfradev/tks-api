package policytemplate

import (
	"strings"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PolicyToTksPolicyCR(policy *model.Policy) *TKSPolicy {
	if policy == nil {
		return nil
	}

	jsonParams := apiextensionsv1.JSON{Raw: []byte(policy.Parameters)}

	labels := map[string]string{}
	labels["app.kubernetes.io/part-of"] = "tks-policy-operator"

	if policy.MatchYaml != nil {
		var match domain.Match

		err := yaml.Unmarshal([]byte(*policy.MatchYaml), &match)

		if err != nil {
			policy.Match = &match
		}
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
			Clusters: policy.TargetClusterIds,
			Template: policy.PolicyTemplate.Kind,
			Match:    policy.Match,
			Params:   &jsonParams,
		},
	}
}

func PolicyTemplateToTksPolicyTemplateCR(policyTemplate *model.PolicyTemplate) *TKSPolicyTemplate {
	if policyTemplate == nil {
		return nil
	}

	labels := map[string]string{}
	labels["app.kubernetes.io/part-of"] = "tks-policy-operator"

	return &TKSPolicyTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tkspolicy.openinfradev.github.io/v1",
			Kind:       "TKSPolicyTemplate",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:   strings.ToLower(policyTemplate.Kind),
			Labels: labels,
		},

		Spec: TKSPolicyTemplateSpec{
			CRD: CRD{
				Spec: CRDSpec{
					Names: Names{
						Kind: policyTemplate.Kind,
					},
					Validation: &Validation{
						OpenAPIV3Schema: ParamDefsToJSONSchemaProeprties(policyTemplate.ParametersSchema),
					},
				},
			},
			Targets: []Target{{
				Target: "admission.k8s.gatekeeper.sh",
				Rego:   policyTemplate.Rego,
				Libs:   policyTemplate.Libs,
			}},
			Version: policyTemplate.Version,
		},
	}
}

func syncToKubernetes() bool {
	return true
	// return os.Getenv("SYNC_POLICY_TO_K8S") != ""
}
