package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-common/pkg/log"
)

type AppGroupHandler struct {
	usecase usecase.IAppGroupUsecase
}

func NewAppGroupHandler(h usecase.IAppGroupUsecase) *AppGroupHandler {
	return &AppGroupHandler{
		usecase: h,
	}
}

// CreateappGroup godoc
// @Tags appGroups
// @Summary Install appGroup
// @Description Install appGroup
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} appGroupId
// @Router /app-groups [post]
func (h *AppGroupHandler) CreateAppGroup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ClusterId   string `json:"clusterId"`
		Type        string `json:"type"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "Invalid json", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if input.Type != "LMA" && input.Type != "SERVICE_MESH" && input.Type != "LMA_EFK" {
		ErrorJSON(w, "Invalid application type", http.StatusBadRequest)
		return
	}

	appGroupId, err := h.usecase.Create(input.ClusterId, input.Name, input.Type, "", input.Description)
	if err != nil {
		log.Error("Failed to create appGroup err : ", err)
		InternalServerError(w)
		return
	}

	var out struct {
		AppGroupId string `json:"appGroupId"`
	}
	out.AppGroupId = appGroupId

	ResponseJSON(w, out, http.StatusOK)
}

// GetAppGroups godoc
// @Tags AppGroups
// @Summary Get appGroup list
// @Description Get appGroup list by giving params
// @Accept json
// @Produce json
// @Param clusterId query string false "clusterId"
// @Success 200 {object} []domain.AppGroup
// @Router /app-groups [get]
func (h *AppGroupHandler) GetAppGroups(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	clusterId := urlParams.Get("clusterId")
	if clusterId == "" {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	appGroups, err := h.usecase.Fetch(clusterId)
	if err != nil {
		ErrorJSON(w, "Failed to get appGroups", http.StatusBadRequest)
		return
	}

	var out struct {
		AppGroups []domain.AppGroup `json:"appGroups"`
	}
	out.AppGroups = appGroups

	ResponseJSON(w, out, http.StatusOK)

}

// GetAppGroup godoc
// @Tags AppGroups
// @Summary Get appGroup detail
// @Description Get appGroup detail by appGroupId
// @Accept json
// @Produce json
// @Param appGroupId path string true "appGroupId"
// @Success 200 {object} []domain.AppGroup
// @Router /app-groups/{appGroupId} [get]
func (h *AppGroupHandler) GetAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	appGroup, err := h.usecase.Get(appGroupId)
	if err != nil {
		InternalServerError(w)
		return
	}

	var out struct {
		AppGroup domain.AppGroup `json:"appGroup"`
	}
	out.AppGroup = appGroup

	ResponseJSON(w, out, http.StatusOK)
}

// DeleteAppGroup godoc
// @Tags AppGroups
// @Summary Uninstall appGroup
// @Description Uninstall appGroup
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} object
// @Router /app-groups [delete]
func (h *AppGroupHandler) DeleteAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	err := h.usecase.Delete(appGroupId)
	if err != nil {
		log.Error("Failed to create appGroup err : ", err)
		InternalServerError(w)
		return
	}

	ResponseJSON(w, nil, http.StatusOK)
}

// GetApplicationKubeInfo godoc
// @Tags Applications
// @Summary Get application detail
// @Description Get application kubernetes information by applicationId
// @Accept json
// @Produce json
// @Param applicationId path string true "applicationId"
// @Success 200 {object} KubeInfoJson
// @Router /applications/{applicationId}/kubeInfo [get]
func (h *APIHandler) GetApplicationKubeInfo(w http.ResponseWriter, r *http.Request) {
	/*
		vars := mux.Vars(r)
		applicationId, ok := vars["applicationId"]
		if !ok {
			log.Error("Failed to get applicationId")
			ErrorJSON(w, fmt.Sprintf("Failed to get applicationId : %s", applicationId), http.StatusBadRequest)
			return
		}

		res, err := appInfoClient.GetAppGroup(context.TODO(), &pb.GetAppGroupRequest{AppGroupId: applicationId})
		if err != nil {
			log.Error("Failed to get appgroup err : ", err)
			InternalServerError(w)
			return
		}

		namespace := ""
		if res.GetAppGroup().GetType() == pb.AppGroupType_LMA || res.GetAppGroup().GetType() == pb.AppGroupType_LMA_EFK {
			namespace = "lma"
		} else if res.GetAppGroup().GetType() == pb.AppGroupType_SERVICE_MESH {
			namespace = "istio-system"
		} else {
			ErrorJSON(w, fmt.Sprintf("invalid appgroup : %s", res.GetAppGroup().GetType()), http.StatusBadRequest)
			return
		}

		clientset, err := h.GetClientFromClusterId(res.GetAppGroup().GetClusterId())
		if err != nil {
			log.Error("Failed to get clientset for clusterId", res.GetAppGroup().GetClusterId())
			InternalServerError(w)
			return
		}

		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			InternalServerError(w)
			return
		}

		services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			InternalServerError(w)
			return
		}

		var out struct {
			KubeInfo KubeInfoJson `json:"kubeInfo"`
		}
		out.KubeInfo.ApplicationId = res.GetAppGroup().GetAppGroupId()
		out.KubeInfo.Pods = make([]PodJson, 0)
		out.KubeInfo.Services = make([]ServiceJson, 0)

		for _, pod := range pods.Items {
			outPod := PodJson{}
			outPod.Id = string(pod.UID)
			outPod.Name = pod.Name
			outPod.Namespace = pod.Namespace
			outPod.Status = string(pod.Status.Phase)
			outPod.Message = pod.Status.Message
			outPod.Started = time.Unix(pod.Status.StartTime.Unix(), 0)

			out.KubeInfo.Pods = append(out.KubeInfo.Pods, outPod)
		}

		for _, service := range services.Items {
			outService := ServiceJson{}
			outService.Id = string(service.UID)
			outService.Name = service.Name
			outService.Namespace = service.Namespace
			outService.Type = string(service.Spec.Type)
			outService.ClusterIp = service.Spec.ClusterIP
			outService.ExternalIps = service.Spec.ExternalIPs

			outService.LoadBalancerIps = make([]string, 0)
			for _, lb := range service.Status.LoadBalancer.Ingress {
				if len(lb.IP) > 0 {
					outService.LoadBalancerIps = append(outService.LoadBalancerIps, lb.IP)
				} else if len(lb.Hostname) > 0 {
					outService.LoadBalancerIps = append(outService.LoadBalancerIps, lb.Hostname)
				}
			}

			outService.Ports = make([]string, 0)
			for _, port := range service.Spec.Ports {
				targetPort := port.TargetPort.StrVal
				if port.TargetPort.Type == 0 {
					targetPort = strconv.Itoa(int(port.TargetPort.IntVal))
				}
				portStr := strconv.Itoa(int(port.Port)) + ":" + targetPort + "/" + string(port.Protocol)
				outService.Ports = append(outService.Ports, portStr)
			}

			outService.Created = time.Unix(service.CreationTimestamp.Unix(), 0)

			out.KubeInfo.Services = append(out.KubeInfo.Services, outService)
		}

		ResponseJSON(w, out, http.StatusOK)
	*/
}
