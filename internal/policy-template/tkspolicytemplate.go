package policytemplate

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
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
	Rego   string   `json:"rego,omitempty" yaml:"rego,omitempty"`
	Libs   []string `json:"libs,omitempty" yaml:"libs,omitempty"`
	Code   []Code   `json:"code,omitempty"`
}

type Code struct {
	Engine string    `json:"engine"`
	Source *Anything `json:"source"`
}

// ============== Copied Fron Operator Start ==============
// CRD, Target struct는 이 파일의 것을 사용

// TKSPolicyTemplateSpec defines the desired state of TKSPolicyTemplate
type TKSPolicyTemplateSpec struct {
	CRD      CRD      `json:"crd,omitempty"`
	Targets  []Target `json:"targets,omitempty"`
	Clusters []string `json:"clusters,omitempty"`
	Version  string   `json:"version"`
	ToLatest []string `json:"toLatest,omitempty"`
}

// TemplateStatus defines the constraints state of ConstraintTemplate on the cluster
type TemplateStatus struct {
	ConstraintTemplateStatus string `json:"constraintTemplateStatus" enums:"ready,applying,deleting,error"`
	Reason                   string `json:"reason,omitempty"`
	LastUpdate               string `json:"lastUpdate"`
	Version                  string `json:"version"`
}

// TKSPolicyTemplateStatus defines the observed state of TKSPolicyTemplate
type TKSPolicyTemplateStatus struct {
	TemplateStatus map[string]TemplateStatus `json:"templateStatus,omitempty"`
	LastUpdate     string                    `json:"lastUpdate"`
	UpdateQueue    map[string]bool           `json:"updateQueue,omitempty"`
}

// TKSPolicyTemplate is the Schema for the tkspolicytemplates API
type TKSPolicyTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TKSPolicyTemplateSpec   `json:"spec,omitempty"`
	Status TKSPolicyTemplateStatus `json:"status,omitempty"`
}

// TKSPolicyTemplateList contains a list of TKSPolicyTemplate
type TKSPolicyTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TKSPolicyTemplate `json:"items"`
}

// ============== Copied Fron Operator End ==============

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

func (tksPolicyTemplate *TKSPolicyTemplate) GetId() string {
	return tksPolicyTemplate.ObjectMeta.Labels[TemplateIDLabel]
}

func ApplyTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, tksPolicyTemplate *TKSPolicyTemplate) error {
	if syncToKubernetes() {
		dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
				err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)
			return err
		}

		policyTemplate, err := GetTksPolicyTemplateCR(ctx, primaryClusterId, strings.ToLower(tksPolicyTemplate.Kind))

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
				err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)
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
			log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
				err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)
			return err
		}

		_, err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
			Update(ctx, tksPolicyTemplateUnstructured, metav1.UpdateOptions{})

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
				err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)
		}

		return err
	}
	return nil
}

func DeleteTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, name string) error {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)
		return err
	}

	err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Delete(ctx, name, metav1.DeleteOptions{})

	// 삭제할 리소스가 존재하지 않았다면 성공으로 처리
	if errors.IsNotFound(err) {
		return nil
	}

	return err
}

func ExistsTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, name string) (bool, error) {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)
		return false, err
	}

	result, err := dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)

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
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)
		return nil, err
	}

	result, err := dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)
		return nil, err
	}

	// Unstructured를 바로 TKSPolicyTemplate으로 컨버팅할 수 없기 때문에 json으로 변환
	jsonBytes, err := json.Marshal(result.Object)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)
		return nil, err
	}

	var tksPolicyTemplate TKSPolicyTemplate
	err = json.Unmarshal(jsonBytes, &tksPolicyTemplate)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, name)
		return nil, err
	}

	return &tksPolicyTemplate, nil
}

func UpdateTksPolicyTemplateCR(ctx context.Context, primaryClusterId string, tksPolicyTemplate *TKSPolicyTemplate) error {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)

		return err
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tksPolicyTemplate)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)
		return err
	}

	tksPolicyUnstructured := &unstructured.Unstructured{
		Object: obj,
	}

	_, err = dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		Update(ctx, tksPolicyUnstructured, metav1.UpdateOptions{})

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
			err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)
		return err
	}

	return err
}

//func ListTksPolicyTemplateCR(ctx context.Context, primaryClusterId string) ([]TKSPolicyTemplate, error) {
//	if syncToKubernetes() {
//		dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)
//
//		if err != nil {
//			return nil, err
//		}
//
//		results, err := dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
//			List(ctx, metav1.ListOptions{})
//
//		if err != nil {
//			return nil, err
//		}
//
//		tkspolicytemplates := make([]TKSPolicyTemplate, len(results.Items))
//
//		for i, result := range results.Items {
//			jsonBytes, err := json.Marshal(result.Object)
//
//			if err != nil {
//				return nil, err
//			}
//
//			var tksPolicyTemplate TKSPolicyTemplate
//			err = json.Unmarshal(jsonBytes, &tksPolicyTemplate)
//
//			if err != nil {
//				return nil, err
//			}
//
//			tkspolicytemplates[i] = tksPolicyTemplate
//		}
//
//		return tkspolicytemplates, nil
//	}
//
//	tkspolicytemplates := make([]TKSPolicyTemplate, 0)
//	return tkspolicytemplates, nil
//}

func GetTksPolicyTemplateCRs(ctx context.Context, primaryClusterId string) (tksPolicyTemplates []TKSPolicyTemplate, err error) {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s'",
			err.Error(), err, primaryClusterId)

		return nil, err
	}

	resources, err := dynamicClient.Resource(TKSPolicyTemplateGVR).Namespace(primaryClusterId).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s'",
			err.Error(), err, primaryClusterId)

		return nil, err
	}

	var tksPolicyTemplate TKSPolicyTemplate
	for _, c := range resources.Items {
		if err = runtime.DefaultUnstructuredConverter.FromUnstructured(c.UnstructuredContent(), &tksPolicyTemplate); err != nil {
			log.Errorf(ctx, "error is :%s(%T), primaryClusterId='%s', policyTemplateName='%+v'",
				err.Error(), err, primaryClusterId, tksPolicyTemplate.Name)

			return nil, err
		}
		tksPolicyTemplates = append(tksPolicyTemplates, tksPolicyTemplate)
	}

	return tksPolicyTemplates, nil
}
