package http

/*
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-common/pkg/argowf"
	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
)

type StackJson = struct {
	Id                string         `json:"id"`
	Cluster           domain.Cluster `json:"cluster"`
	Status            string         `json:"status"`
	StatusDescription string         `json:"statusDescription"`
}

// GetStacks godoc
// @Tags Stacks
// @Summary Get tks stacks
// @Description Get tks stack list
// @Accept json
// @Produce json
// @Param projectId query string false "projectId"
// @Success 200 {object} []StackJson
// @Router /stacks [get]
func (h *APIHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	projectId := urlParams.Get("projectId")
	includeDeleted := urlParams.Get("includeDeleted")

	contractIds := make([]string, 0)

	if projectId != "" {
		contractIds = append(contractIds, projectId)
	} else {
		// [TODO] 전체 클러스터 조회를 위해 contractId 별로 loop 를 돈다.
		// tks-contract 에 전체 클러스터 조회 API 추가시 해당 부분 삭제한다.
		contracts, err := contractClient.GetContracts(context.TODO(), &pb.GetContractsRequest{})
		if err != nil {
			log.Error("Failed to get contracts err : ", err)
			ErrorJSON(w, fmt.Sprintf("Failed to get contract err : %s", err), http.StatusBadRequest)
		}
		for _, contract := range contracts.GetContracts() {
			contractIds = append(contractIds, contract.GetContractId())
		}
	}

	var out struct {
		Stacks []StackJson `json:"stacks"`
	}
	out.Stacks = make([]StackJson, 0)

	for _, contractId := range contractIds {
		clusters, err := clusterInfoClient.GetClusters(context.TODO(), &pb.GetClustersRequest{ContractId: contractId, CspId: ""})
		if err != nil {
			log.Error("Failed to get clusters err : ", err)
		}

		for _, cluster := range clusters.GetClusters() {

			if includeDeleted != "true" {
				if cluster.GetStatus() == pb.ClusterStatus_DELETED {
					continue
				}
			}

			resApplications, err := appInfoClient.GetAppGroupsByClusterID(context.TODO(), &pb.IDRequest{Id: cluster.GetId()})
			if err != nil {
				log.Error("Failed to get appgroups err : ", err)
				continue
			}
			applications := resApplications.GetAppGroups()

			status, statusDesc := getStackStatus(cluster, applications)

			outStack := StackJson{}
			outStack.Id = cluster.GetId()
			reflectCluster(&outStack.Cluster, cluster)
			parsedUuid, err := uuid.Parse(cluster.GetCreator())
			if err == nil {
				var user repository.User
				err = h.Repository.GetUserById(&user, parsedUuid)
				outStack.Cluster.Creator = user.Name
			}

			outStack.Status = status
			outStack.StatusDescription = statusDesc
			out.Stacks = append(out.Stacks, outStack)
		}
	}

	ResponseJSON(w, out, http.StatusOK)
}

// GetStack godoc
// @Tags Stacks
// @Summary Get tks service
// @Description Get tks service detail
// @Accept json
// @Produce json
// @Param clusterId path string true "clusterId"
// @Success 200 {object} StackJson
// @Router /stacks/{clusterId} [get]
func (h *APIHandler) GetStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, ok := vars["clusterId"]
	if !ok {
		log.Error("Failed to get clusterId")
	}

	resClusters, err := clusterInfoClient.GetCluster(context.TODO(), &pb.GetClusterRequest{ClusterId: clusterId})
	if err != nil {
		log.Error("Failed to get cluster err : ", err)
	}
	cluster := resClusters.GetCluster()

	resApplications, err := appInfoClient.GetAppGroupsByClusterID(context.TODO(), &pb.IDRequest{Id: clusterId})
	if err != nil {
		log.Error("Failed to get appgroups err : ", err)
		InternalServerError(w)
		return
	}
	applications := resApplications.GetAppGroups()

	status, statusDesc := getStackStatus(cluster, applications)

	var out struct {
		Stack StackJson `json:"stack"`
	}

	outCluster := ClusterJson{}
	reflectCluster(&outCluster, cluster)
	parsedUuid, err := uuid.Parse(cluster.GetCreator())
	if err == nil {
		var user repository.User
		err = h.Repository.GetUserById(&user, parsedUuid)
		outCluster.Creator = user.Name
	}

	out.Stack = StackJson{
		Id:                cluster.GetId(),
		Cluster:           outCluster,
		Status:            status,
		StatusDescription: statusDesc,
	}

	ResponseJSON(w, out, http.StatusOK)
}

// CreateCluster godoc
// @Tags Stacks
// @Summary Create tks service
// @Description Create tks service
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} object
// @Router /stacks [post]
func (h *APIHandler) CreateStack(w http.ResponseWriter, r *http.Request) {
	userId, _ := h.GetSession(r)

	var input struct {
		ProjectId       string `json:"projectId"`
		InfraProvider   string `json:"infraProvider"`
		TemplateId      string `json:"templateId"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		NumberOfAz      string `json:"numberOfAz"`
		MachineType     string `json:"machineType"`
		Region          string `json:"region"`
		MachineReplicas string `json:"machineReplicas"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Check quota
	quota := 0
	contracts, err := contractClient.GetContracts(context.TODO(), &pb.GetContractsRequest{})
	if err != nil {
		log.Error("Failed to get contracts err : ", err)
		ErrorJSON(w, fmt.Sprintf("Failed to get contract err : %s", err), http.StatusBadRequest)
	}
	for _, contract := range contracts.GetContracts() {
		clusters, err := clusterInfoClient.GetClusters(context.TODO(), &pb.GetClustersRequest{ContractId: contract.GetContractId(), CspId: ""})
		if err != nil {
			log.Error("Failed to get clusters err : ", err)
			continue
		}
		for _, cluster := range clusters.GetClusters() {
			if cluster.GetStatus() != pb.ClusterStatus_ERROR && cluster.GetStatus() != pb.ClusterStatus_DELETED && cluster.GetCreator() == userId {
				quota += 1
			}
		}
	}

	var MAX_STACK_QUOTA = 2
	if quota >= MAX_STACK_QUOTA {
		ErrorJSON(w, "Exceeded stack quota", http.StatusBadRequest)
		return
	}

	// Get csp_id
	res, err := cspInfoClient.GetCSPIDsByContractID(context.TODO(), &pb.IDRequest{Id: input.ProjectId})
	if err != nil {
		log.Error("Failed to get csp info err : ", err)
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}
	cspIds := res.GetIds()
	if len(cspIds) < 1 {
		log.Error("CSPIds must have values. cspIds : ", cspIds)
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	workflow := ""
	if input.TemplateId == "aws-reference" {
		workflow = "tks-stack-create-aws"
	} else if input.TemplateId == "aws-msa-reference" {
		workflow = "tks-stack-create-aws-msa"
	} else {
		log.Error("Invalid templateId  : ", input.TemplateId)
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	{
		nameSpace := "argo"
		opts := argowf.SubmitOptions{}
		opts.Parameters = []string{
			fmt.Sprintf("tks_info_url=%s:%d", viper.GetString("info-address"), viper.GetInt("info-port")),
			fmt.Sprintf("tks_contract_url=%s:%d", viper.GetString("contract-address"), viper.GetInt("contract-port")),
			fmt.Sprintf("tks_cluster_lcm_url=%s:%d", viper.GetString("lcm-address"), viper.GetInt("lcm-port")),
			"cluster_name=" + input.Name,
			"contract_id=" + input.ProjectId,
			"csp_id=" + cspIds[0],
			"creator=" + userId,
			"description=" + input.TemplateId + "-" + input.Description,
				//"machine_type=" + input.MachineType,
				//"num_of_az=" + input.NumberOfAz,
				//"machine_replicas=" + input.MachineReplicas,
		}

		workflowId, err := argowfClient.SumbitWorkflowFromWftpl(workflow, nameSpace, opts)
		if err != nil {
			log.Error(err)
			InternalServerError(w)
			return
		}
		log.Info("Submitted workflow: ", workflowId)

		// wait & get clusterId ( max 1min 	)
		cnt := 0
		for range time.Tick(2 * time.Second) {
			if cnt >= 60 { // max wait 60sec
				break
			}

			workflow, err := argowfClient.GetWorkflow("argo", workflowId)
			if err != nil {
				log.Error(err)
				InternalServerError(w)
				break
			}

			if workflow.Status.Phase != "Running" {
				log.Error(err)
				InternalServerError(w)
				break
			}

			if workflow.Status.Progress == "1/2" { // start creating cluster
				time.Sleep(time.Second * 5) // Buffer
				break
			}
			cnt += 1
		}
	}

	h.AddHistory(r, input.ProjectId, "stack", fmt.Sprintf("스택 [%s]을 생성하였습니다.", input.Name))

	var out struct{}
	ResponseJSON(w, out, http.StatusOK)
}

// DeleteStack godoc
// @Tags Stacks
// @Summary Delete tks service
// @Description Delete tks service
// @Accept json
// @Produce json
// @Param clusterId path string true "clusterId"
// @Success 200 {string} ClusterId
// @Router /stacks/{clusterId} [delete]
func (h *APIHandler) DeleteStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, ok := vars["clusterId"]
	if !ok {
		log.Error("Failed to get clusterId")
		ErrorJSON(w, "Invalid clusterId", http.StatusBadRequest)
	}

	var input struct {
		ProjectId  string `json:"projectId"`
		TemplateId string `json:"templateId"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	workflow := ""
	if input.TemplateId == "aws-reference" {
		workflow = "tks-stack-delete-aws"
	} else if input.TemplateId == "aws-msa-reference" {
		workflow = "tks-stack-delete-aws-msa"
	} else {
		log.Error("Invalid templateId  : ", input.TemplateId)
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	{
		nameSpace := "argo"
		opts := argowf.SubmitOptions{}
		opts.Parameters = []string{
			fmt.Sprintf("tks_info_url=%s:%d", viper.GetString("info-address"), viper.GetInt("info-port")),
			fmt.Sprintf("tks_contract_url=%s:%d", viper.GetString("contract-address"), viper.GetInt("contract-port")),
			fmt.Sprintf("tks_cluster_lcm_url=%s:%d", viper.GetString("lcm-address"), viper.GetInt("lcm-port")),
			"contract_id=" + input.ProjectId,
			"cluster_id=" + clusterId,
		}

		workflowId, err := argowfClient.SumbitWorkflowFromWftpl(workflow, nameSpace, opts)
		if err != nil {
			log.Error(err)
			InternalServerError(w)
			return
		}
		log.Info("Submitted workflow: ", workflowId)

		// wait & get clusterId ( max 1min 	)
		cnt := 0
		for range time.Tick(2 * time.Second) {
			if cnt >= 60 { // max wait 60sec
				break
			}

			workflow, err := argowfClient.GetWorkflow("argo", workflowId)
			if err != nil {
				log.Error(err)
				InternalServerError(w)
				break
			}

			if workflow.Status.Phase != "Running" {
				log.Error(err)
				InternalServerError(w)
				break
			}

			if workflow.Status.Progress == "1/2" { // start deleting service
				time.Sleep(time.Second * 10) // Buffer
				break
			}
			cnt += 1
		}
	}

	var out struct {
		ClusterId string `json:"clusterId"`
	}
	out.ClusterId = ""

	h.AddHistory(r, input.ProjectId, "stack", fmt.Sprintf("스택 [%s]을 삭제하였습니다.", clusterId))

	ResponseJSON(w, out, http.StatusOK)
}

func getStackStatus(cluster *pb.Cluster, applications []*pb.AppGroup) (string, string) {
	for _, application := range applications {
		if application.Status == pb.AppGroupStatus_APP_GROUP_INSTALLING {
			return "APPLICATION_INSTALLING", application.StatusDesc
		}
		if application.Status == pb.AppGroupStatus_APP_GROUP_DELETING {
			return "APPLICATION_DELETING", application.StatusDesc
		}
		if application.Status == pb.AppGroupStatus_APP_GROUP_ERROR {
			return "APPLICATION_ERROR", application.StatusDesc
		}
	}

	if cluster.Status == pb.ClusterStatus_INSTALLING {
		return "CLUSTER_INSTALLING", cluster.StatusDesc
	}
	if cluster.Status == pb.ClusterStatus_DELETING {
		return "CLUSTER_DELETING", cluster.StatusDesc
	}
	if cluster.Status == pb.ClusterStatus_DELETED {
		return "CLUSTER_DELETED", cluster.StatusDesc
	}
	if cluster.Status == pb.ClusterStatus_ERROR {
		return "CLUSTER_ERROR", cluster.StatusDesc
	}

	// workflow 중간 중간 비는 status 처리...
	if strings.Contains(cluster.Description, "aws-reference") {
		if len(applications) != 1 {
			return "APPLICATION_INSTALLING", "(0/0)"
		}
	} else if strings.Contains(cluster.Description, "aws-msa-reference") {
		if len(applications) != 2 {
			return "APPLICATION_INSTALLING", "(0/0)"
		}
	}

	return "RUNNING", cluster.StatusDesc

}
*/
