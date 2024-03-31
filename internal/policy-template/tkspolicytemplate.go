package policytemplate

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/openinfradev/tks-api/internal/kubernetes"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var TKSPolicyTemplateGVR = schema.GroupVersionResource{
	Group: "tkspolicy.openinfradev.github.io", Version: "v1",
	Resource: "tkspolicytemplates",
}

type Anything struct {
	Value interface{} `json:"-"`
}

type CRD struct {
	Spec CRDSpec `json:"spec,omitempty"`
}

type CRDSpec struct {
	Names      Names       `json:"names,omitempty"`
	Validation *Validation `json:"validation,omitempty"`
}

type Names struct {
	Kind       string   `json:"kind,omitempty"`
	ShortNames []string `json:"shortNames,omitempty"`
}

type Validation struct {
	OpenAPIV3Schema *apiextensionsv1.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
	LegacySchema    *bool                            `json:"legacySchema,omitempty"` // *bool allows for "unset" state which we need to apply appropriate defaults
}

type Target struct {
	Target string   `json:"target,omitempty"`
	Rego   string   `json:"rego,omitempty"`
	Libs   []string `json:"libs,omitempty"`
	Code   []Code   `json:"code,omitempty"`
}

type Code struct {
	Engine string    `json:"engine"`
	Source *Anything `json:"source"`
}

type TKSPolicyTemplateSpec struct {
	CRD      CRD      `json:"crd,omitempty"`
	Targets  []Target `json:"targets,omitempty"`
	Clusters []string `json:"clusters,omitempty"`
	ToLatest []string `json:"toLatest,omitempty"`
	Version  string   `json:"version"`
}

type TemplateStatus struct {
	ConstraintTemplateStatus string `json:"status" enums:"ready,applying,deleting,error"`
	Reason                   string `json:"reason"`
	LastUpdate               string `json:"lastUpdate"`
	Version                  string `json:"version"`
}

type TKSPolicyTemplateStatus struct {
	TemplateStatus map[string]*TemplateStatus `json:"clusterStatus"`
	LastUpdate     string                     `json:"lastUpdate"`
	UpdateQueue    map[string]bool            `json:"updateQueue,omitempty"`
}

// TKSPolicyTemplate is the Schema for the tkspolicytemplates API
type TKSPolicyTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TKSPolicyTemplateSpec   `json:"spec,omitempty"`
	Status TKSPolicyTemplateStatus `json:"status,omitempty"`
}

func (tksPolicyTemplate *TKSPolicyTemplate) JSON() (string, error) {
	result, err := json.MarshalIndent(tksPolicyTemplate, "", "  ")

	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (tksPolicyTemplate *TKSPolicyTemplate) YAML() (string, error) {
	target := map[string]interface{}{}

	jsonStr, err := tksPolicyTemplate.JSON()

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

func (tksPolicyTemplate *TKSPolicyTemplate) ToUnstructured() (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tksPolicyTemplate)

	if err != nil {
		return nil, err
	}

	tksPolicyTemplateUnstructured := &unstructured.Unstructured{
		Object: obj,
	}

	return tksPolicyTemplateUnstructured, nil
}

func ApplyTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, tksPolicyTemplate *TKSPolicyTemplate) error {
	if syncToKubernetes() {
		dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

		if err != nil {
			return err
		}

		policyTemplate, err := GetTksPolicyTemplateCR(ctx, primaryClusterId, strings.ToLower(tksPolicyTemplate.Kind))

		if err != nil {
			if errors.IsNotFound(err) {
				tksPolicyTemplateUnstructured, err := tksPolicyTemplate.ToUnstructured()

				if err != nil {
					return err
				}

				_, err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
					Create(ctx, tksPolicyTemplateUnstructured, metav1.CreateOptions{})
				return err
			} else {
				return err
			}
		}

		policyTemplate.Spec = tksPolicyTemplate.Spec
		tksPolicyTemplateUnstructured, err := policyTemplate.ToUnstructured()

		if err != nil {
			return err
		}

		_, err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
			Update(ctx, tksPolicyTemplateUnstructured, metav1.UpdateOptions{})

		return err
	}
	return nil
}

func DeleteTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, name string) error {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		return err
	}

	err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Delete(ctx, name, metav1.DeleteOptions{})

	return err
}

func ExistsTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, name string) (bool, error) {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		return false, err
	}

	result, err := dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
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

func GetTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, name string) (*TKSPolicyTemplate, error) {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		return nil, err
	}

	result, err := dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	// Unstructured를 바로 TKSPolicyTemplate으로 컨버팅할 수 없기 때문에 json으로 변환
	jsonBytes, err := json.Marshal(result.Object)

	if err != nil {
		return nil, err
	}

	var tksPolicyTemplate TKSPolicyTemplate
	err = json.Unmarshal(jsonBytes, &tksPolicyTemplate)

	if err != nil {
		return nil, err
	}

	return &tksPolicyTemplate, nil
}

func UpdateTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, tksPolicyTemplate *TKSPolicyTemplate) error {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		return err
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tksPolicyTemplate)

	if err != nil {
		return err
	}

	tksPolicyUnstructured := &unstructured.Unstructured{
		Object: obj,
	}

	_, err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Update(ctx, tksPolicyUnstructured, metav1.UpdateOptions{})

	return err
}
