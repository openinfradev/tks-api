package policytemplate

import (
	"context"
	"encoding/json"

	"github.com/openinfradev/tks-api/internal/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var TKSClusterGVR = schema.GroupVersionResource{
	Group: "tkspolicy.openinfradev.github.io", Version: "v1",
	Resource: "tksclusters",
}

type TKSClusterSpec struct {
	ClusterName  string            `json:"clusterName" validate:"required"`
	Context      string            `json:"context" validate:"required"`
	TemplatesRef TemplateReference `json:"templateref,omitempty"`
}

type TKSProxy struct {
	ControllerManager *DeploymentInfo `json:"controllerManager,omitempty"`
	Audit             *DeploymentInfo `json:"audit,omitempty"`
}

type DeploymentInfo struct {
	Image         string   `json:"image,omitempty"`
	Args          []string `json:"args,omitempty"`
	TotalReplicas int      `json:"totalReplicas,omitempty"`
	NumReplicas   int      `json:"numReplicas,omitempty"`
}

type TemplateReference struct {
	Policies  map[string]string `json:"polices,omitempty"`
	Templates map[string]string `json:"templates,omitempty"`
}

type TKSClusterStatus struct {
	Status              string            `json:"status" enums:"running,deleting,error"`
	Error               string            `json:"error,omitempty"`
	TKSProxy            TKSProxy          `json:"tksproxy,omitempty"`
	LastStatusCheckTime int64             `json:"laststatuschecktime,omitempty"`
	Policies            map[string]string `json:"polices,omitempty"`
	Templates           map[string]string `json:"templates,omitempty"`
	LastUpdate          string            `json:"lastUpdate"`
	UpdateQueue         map[string]bool   `json:"updateQueue,omitempty"`
}

type TKSCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TKSClusterSpec   `json:"spec,omitempty"`
	Status TKSClusterStatus `json:"status,omitempty"`
}

type TKSClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TKSCluster `json:"items"`
}

func GetTksClusterCR(ctx context.Context, primaryClusterId string, name string) (*TKSCluster, error) {
	dynamicClient, err := kubernetes.GetDynamicClientAdminCluster(ctx)

	if err != nil {
		return nil, err
	}

	result, err := dynamicClient.Resource(TKSClusterGVR).Namespace(primaryClusterId).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	// Unstructured를 바로 TKSPolicyTemplate으로 컨버팅할 수 없기 때문에 json으로 변환
	jsonBytes, err := json.Marshal(result.Object)

	if err != nil {
		return nil, err
	}

	var tksCluster TKSCluster
	err = json.Unmarshal(jsonBytes, &tksCluster)

	if err != nil {
		return nil, err
	}

	return &tksCluster, nil
}
