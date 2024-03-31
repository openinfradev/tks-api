package policytemplate

import (
	"context"
	"encoding/json"

	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var TKSPolicyGVR = schema.GroupVersionResource{
	Group: "tkspolicy.openinfradev.github.io", Version: "v1",
	Resource: "tkspolicies",
}

type TKSPolicySpec struct {
	Clusters []string `json:"clusters"`
	Template string   `json:"template" validate:"required"`

	// map[string]interface를 쓸 수 없어서 apiextensionsv1.JSON 사용
	Params *apiextensionsv1.JSON `json:"params,omitempty"`
	Match  *domain.Match         `json:"match,omitempty"`
}

type PolicyStatus struct {
	ConstraintStatus string `json:"status" enums:"ready,applying,deleting,error"`
	Reason           string `json:"reason"`
	LastUpdate       string `json:"lastUpdate"`
	TemplateVersion  string `json:"templateVersion"`
}

// TKSPolicyStatus defines the observed state of TKSPolicy
type TKSPolicyStatus struct {
	Clusters    map[string]PolicyStatus `json:"clusters"`
	LastUpdate  string                  `json:"lastUpdate"`
	UpdateQueue map[string]bool         `json:"updateQueue,omitempty"`
}

type TKSPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TKSPolicySpec   `json:"spec,omitempty"`
	Status TKSPolicyStatus `json:"status,omitempty"`
}

func (tksPolicy *TKSPolicy) JSON() (string, error) {
	result, err := json.MarshalIndent(tksPolicy, "", "  ")

	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (tksPolicy *TKSPolicy) YAML() (string, error) {
	target := map[string]interface{}{}

	jsonStr, err := tksPolicy.JSON()

	if err != nil {
		return "", err
	}

	err = json.Unmarshal([]byte(jsonStr), &target)

	if err != nil {
		return "", err
	}

	result, err := yaml.Marshal(&target)

	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (tksPolicy *TKSPolicy) ToUnstructured() (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tksPolicy)

	if err != nil {
		return nil, err
	}

	tksPolicyUnstructured := &unstructured.Unstructured{
		Object: obj,
	}

	return tksPolicyUnstructured, nil
}

func ApplyTksPolicyCR(ctx context.Context, primaryClusterId string, tksPolicy *TKSPolicy) error {
	if syncToKubernetes() {
		dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

		if err != nil {
			return err
		}

		policy, err := GetTksPolicyCR(ctx, primaryClusterId, tksPolicy.Name)

		if err != nil {
			if errors.IsNotFound(err) {
				tksPolicyUnstructured, err := tksPolicy.ToUnstructured()

				if err != nil {
					return err
				}

				_, err = dynamicClient.Resource(TKSPolicyGVR).Namespace(primaryClusterId).
					Create(ctx, tksPolicyUnstructured, metav1.CreateOptions{})
				return err
			} else {
				return err
			}
		}

		policy.Spec = tksPolicy.Spec
		tksPolicyUnstructured, err := policy.ToUnstructured()

		if err != nil {
			return err
		}

		_, err = dynamicClient.Resource(TKSPolicyGVR).Namespace(primaryClusterId).
			Update(ctx, tksPolicyUnstructured, metav1.UpdateOptions{})

		return err
	}
	return nil
}

func DeleteTksPolicyCR(ctx context.Context, primaryClusterId string, name string) error {
	if syncToKubernetes() {
		dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

		if err != nil {
			return err
		}

		err = dynamicClient.Resource(TKSPolicyGVR).Namespace(primaryClusterId).
			Delete(ctx, name, metav1.DeleteOptions{})

		return err
	}
	return nil
}

func GetTksPolicyCR(ctx context.Context, primaryClusterId string, name string) (*TKSPolicy, error) {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		return nil, err
	}

	result, err := dynamicClient.Resource(TKSPolicyGVR).Namespace(primaryClusterId).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	// Unstructured를 바로 TKSPolicyTemplate으로 컨버팅할 수 없기 때문에 json으로 변환
	jsonBytes, err := json.Marshal(result.Object)

	if err != nil {
		return nil, err
	}

	var tksPolicy TKSPolicy
	err = json.Unmarshal(jsonBytes, &tksPolicy)

	if err != nil {
		return nil, err
	}

	return &tksPolicy, nil
}

func ExistsTksPolicyCR(ctx context.Context, primaryClusterId string, name string) (bool, error) {
	if syncToKubernetes() {
		dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

		if err != nil {
			return false, err
		}

		result, err := dynamicClient.Resource(TKSPolicyGVR).Namespace(primaryClusterId).
			Get(ctx, name, metav1.GetOptions{})

		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			} else {
				return false, err
			}
		}

		return result != nil, nil
	}
	return true, nil
}
