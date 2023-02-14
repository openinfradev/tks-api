package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-common/pkg/log"
	"github.com/openinfradev/tks-proto/tks_pb"
	pb "github.com/openinfradev/tks-proto/tks_pb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Overview = struct {
	Projects int `json:"projects"`
	Clusters int `json:"clusters"`
	Nodes    int `json:"nodes"`
	Users    int `json:"users"`
}

type KubeEvent = struct {
	ClusterId string    `json:"clusterId"`
	Id        string    `json:"id"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Updated   time.Time `json:"updated"`
}

type KubePod = struct {
	ClusterId string    `json:"clusterId"`
	Id        string    `json:"id"`
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Started   time.Time `json:"started"`
}

// GetOverview godoc
// @Tags Dashboard
// @Summary Get overview for dashboard
// @Description Get user based overview
// @Accept json
// @Produce json
// @Success 200 {object} Overview
// @Router /dashboard [get]
func (h *APIHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	_, accountId := h.GetSession(r)

	projectsCnt := 0
	clustersCnt := 0
	nodesCnt := 0
	usersCnt := 0

	var projectIds []string
	err := h.Repository.GetProjectIdsByUser(&projectIds, accountId)
	if err != nil {
		ErrorJSON(w, "not found project by user", http.StatusBadRequest)
		return
	}

	resContracts, err := contractClient.GetContracts(context.TODO(), &pb.GetContractsRequest{})
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Failed to get contract err : %s", err), http.StatusBadRequest)
		return
	}
	for _, contract := range resContracts.GetContracts() {
		if !helper.Contains(projectIds, contract.GetContractId()) {
			continue
		}
		projectsCnt += 1
	}

	for _, projectId := range projectIds {
		resClusters, err := clusterInfoClient.GetClusters(context.TODO(), &pb.GetClustersRequest{ContractId: projectId, CspId: ""})
		if err != nil {
			log.Error(err)
			continue
		}

		for _, cluster := range resClusters.GetClusters() {
			if cluster.GetStatus() != pb.ClusterStatus_RUNNING {
				continue
			}

			clientset_user, err := h.GetClientFromClusterId(cluster.GetId())
			if err != nil {
				log.Error("Failed to get clientset for clusterId ", cluster.GetId())
				continue
			}

			nodes, err := clientset_user.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Failed to get nodes. err : ", err)
				continue
			}
			nodesCnt += len(nodes.Items)
			clustersCnt += 1
		}

		var users []repository.User
		err = h.Repository.GetUsersInProject(&users, projectId)
		if err != nil {
			log.Error(err)
		}
		usersCnt = len(users)
	}

	var out struct {
		Overview Overview `json:"overview"`
	}

	out.Overview.Projects = projectsCnt
	out.Overview.Clusters = clustersCnt
	out.Overview.Nodes = nodesCnt
	out.Overview.Users = usersCnt

	ResponseJSON(w, out, http.StatusOK)
}

// GetAdminKubernetesInfo godoc
// @Tags Dashboard
// @Summary Get kubernetes overview
// @Description Get kubernetes overview for user
// @Accept json
// @Produce json
// @Success 200 {object} domain.ClusterKubeInfo
// @Router /dashboard/kube-info [get]
func (h *APIHandler) GetAdminKubernetesInfo(w http.ResponseWriter, r *http.Request) {
	clientset, err := helper.GetClientAdminCluster()
	if err != nil {
		log.Error("Failed to get clientset err : ", err)
		InternalServerError(w)
		return
	}

	var out struct {
		KubeInfo domain.ClusterKubeInfo `json:"kubeInfo"`
	}
	out.KubeInfo.Updated = time.Now()

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		out.KubeInfo.Pods = len(pods.Items)
	} else {
		log.Error("Failed to get pods. err : ", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		out.KubeInfo.Nodes = len(nodes.Items)
	} else {
		log.Error("Failed to get nodes. err : ", err)
	}

	services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		out.KubeInfo.Services = len(services.Items)
	} else {
		log.Error("Failed to get services. err : ", err)

	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		out.KubeInfo.Namespaces = len(namespaces.Items)
	} else {
		log.Error("Failed to get namespaces. err : ", err)

	}

	version, err := h.GetKubernetesVserion()
	if err == nil {
		out.KubeInfo.Version = version
	} else {
		log.Error("Failed to get kubernetes version. err : ", err)
	}

	ResponseJSON(w, out, http.StatusOK)
}

// GetAdnormalKubernetesEvents godoc
// @Tags Dashboard
// @Summary Get kubernetes events
// @Description Get kubernetes events for cluster
// @Accept json
// @Produce json
// @Success 200 {object} []KubeEvent
// @Router /dashboard/kubeEvent [get]
func (h *APIHandler) GetAdnormalKubernetesEvents(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	filter := urlParams.Get("filter")
	amount := urlParams.Get("amount")

	_, accountId := h.GetSession(r)

	var projectIds []string
	err := h.Repository.GetProjectIdsByUser(&projectIds, accountId)
	if err != nil {
		ErrorJSON(w, "not found project by user", http.StatusBadRequest)
		return
	}

	var out struct {
		Events []KubeEvent `json:"events"`
	}
	out.Events = make([]KubeEvent, 0)

	resContracts, err := contractClient.GetContracts(context.TODO(), &pb.GetContractsRequest{})
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Failed to get contract err : %s", err), http.StatusBadRequest)
		return
	}
	for _, contract := range resContracts.GetContracts() {
		if !helper.Contains(projectIds, contract.GetContractId()) {
			continue
		}

		clusters, err := clusterInfoClient.GetClusters(context.TODO(), &tks_pb.GetClustersRequest{ContractId: contract.GetContractId(), CspId: ""})
		if err != nil {
			log.Error("Failed to get clusters err : ", err)
		}

		for _, cluster := range clusters.GetClusters() {
			if cluster.GetStatus() != pb.ClusterStatus_RUNNING {
				continue
			}

			clientset_user, err := h.GetClientFromClusterId(cluster.Id)
			if err != nil {
				log.Error("Failed to get clientset for clusterId ", cluster.Id)
				continue
			}

			events, err := clientset_user.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Failed to get events", err)
				continue
			}

			for _, item := range events.Items {
				if filter == "type" {
					if item.Type == "Normal" {
						continue
					}
				}

				if amount != "full" && len(out.Events) >= 10 {
					break
				}
				outEvent := KubeEvent{}
				outEvent.ClusterId = string(cluster.Id)
				outEvent.Id = string(item.UID)
				outEvent.Namespace = item.Namespace
				outEvent.Type = item.Type
				outEvent.Reason = item.Reason
				outEvent.Message = item.Message
				outEvent.Updated = time.Unix(item.CreationTimestamp.Unix(), 0)

				out.Events = append(out.Events, outEvent)
			}
		}
	}

	ResponseJSON(w, out, http.StatusOK)
}

// GetAdnormalKubernetesPods godoc
// @Tags Dashboard
// @Summary Get kubernetes pods
// @Description Get kubernetes pods for cluster
// @Accept json
// @Produce json
// @Success 200 {object} []PodJson
// @Router /dashboard/kubePods [get]
func (h *APIHandler) GetAdnormalKubernetesPods(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	filter := urlParams.Get("filter")

	_, accountId := h.GetSession(r)

	var projectIds []string
	err := h.Repository.GetProjectIdsByUser(&projectIds, accountId)
	if err != nil {
		ErrorJSON(w, "not found project by user", http.StatusBadRequest)
		return
	}

	var out struct {
		Pods []KubePod `json:"pods"`
	}
	out.Pods = make([]KubePod, 0)

	resContracts, err := contractClient.GetContracts(context.TODO(), &pb.GetContractsRequest{})
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Failed to get contract err : %s", err), http.StatusBadRequest)
		return
	}
	for _, contract := range resContracts.GetContracts() {
		if !helper.Contains(projectIds, contract.GetContractId()) {
			continue
		}

		clusters, err := clusterInfoClient.GetClusters(context.TODO(), &tks_pb.GetClustersRequest{ContractId: contract.GetContractId(), CspId: ""})
		if err != nil {
			log.Error("Failed to get clusters err : ", err)
		}

		for _, cluster := range clusters.GetClusters() {
			if cluster.GetStatus() != pb.ClusterStatus_RUNNING {
				continue
			}

			clientset_user, err := h.GetClientFromClusterId(cluster.Id)
			if err != nil {
				log.Error("Failed to get clientset for clusterId ", cluster.Id)
				continue
			}

			pods, err := clientset_user.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Failed to get events", err)
				continue
			}

			for _, pod := range pods.Items {
				if filter == "type" {
					if pod.Status.Phase == "Running" || pod.Status.Phase == "Succeeded" {
						continue
					}
				}

				if len(out.Pods) >= 10 {
					break
				}
				outPod := KubePod{}
				outPod.Id = string(pod.UID)
				outPod.ClusterId = string(cluster.Id)
				outPod.Name = pod.Name
				outPod.Namespace = pod.Namespace
				outPod.Status = string(pod.Status.Phase)
				outPod.Message = pod.Status.Message
				if pod.Status.StartTime != nil {
					outPod.Started = time.Unix(pod.Status.StartTime.Unix(), 0)
				}

				out.Pods = append(out.Pods, outPod)
			}
		}
	}

	ResponseJSON(w, out, http.StatusOK)

}
