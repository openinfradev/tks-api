package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type ClusterHandler struct {
	usecase usecase.IClusterUsecase
}

func NewClusterHandler(h usecase.IClusterUsecase) *ClusterHandler {
	return &ClusterHandler{
		usecase: h,
	}
}

// GetClusters godoc
// @Tags Clusters
// @Summary Get clusters
// @Description Get cluster list
// @Accept json
// @Produce json
// @Param organizationId query string false "organizationId"
// @Success 200 {object} domain.GetClustersResponse
// @Router /clusters [get]
// @Security     JWT
func (h *ClusterHandler) GetClusters(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	organizationId := urlParams.Get("organizationId")
	clusters, err := h.usecase.Fetch(organizationId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetClustersResponse
	out.Clusters = make([]domain.ClusterResponse, len(clusters))
	for i, cluster := range clusters {
		if err := domain.Map(cluster, &out.Clusters[i]); err != nil {
			log.Info(err)
			continue
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetCluster godoc
// @Tags Clusters
// @Summary Get cluster
// @Description Get cluster detail
// @Accept json
// @Produce json
// @Param clusterId path string true "clusterId"
// @Success 200 {object} domain.Cluster
// @Router /clusters/{clusterId} [get]
// @Security     JWT
func (h *ClusterHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, ok := vars["clusterId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "C_INVALID_CLUSTER_ID", ""))
		return
	}

	cluster, err := h.usecase.Get(domain.ClusterId(clusterId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetClusterResponse
	if err := domain.Map(cluster, &out.Cluster); err != nil {
		log.Info(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetClusterSiteValues godoc
// @Tags Clusters
// @Summary Get cluster site values for creating
// @Description Get cluster site values for creating
// @Accept json
// @Produce json
// @Param clusterId path string true "clusterId"
// @Success 200 {object} domain.ClusterSiteValuesResponse
// @Router /clusters/{clusterId}/site-values [get]
// @Security     JWT
func (h *ClusterHandler) GetClusterSiteValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, ok := vars["clusterId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "C_INVALID_CLUSTER_ID", ""))
		return
	}

	clusterSiteValues, err := h.usecase.GetClusterSiteValues(domain.ClusterId(clusterId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetClusterSiteValuesResponse
	out.ClusterSiteValues = clusterSiteValues

	ResponseJSON(w, http.StatusOK, out)
}

// GetCluster godoc
// @Tags Clusters
// @Summary Create cluster
// @Description Create cluster
// @Accept json
// @Produce json
// @Param body body domain.CreateClusterRequest true "create cluster request"
// @Success 200 {object} domain.CreateClusterResponse
// @Router /clusters [post]
// @Security     JWT
func (h *ClusterHandler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateClusterRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var dto domain.Cluster
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}

	if err = domain.Map(input, &dto.Conf); err != nil {
		log.Info(err)
	}

	// [TODO] set default value
	dto.Conf.SetDefault()
	log.Info(dto.Conf)

	//txHandle := r.Context().Value("txHandle").(*gorm.DB)
	clusterId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateClusterResponse
	out.ID = clusterId.String()

	ResponseJSON(w, http.StatusOK, out)
}

// DeleteCluster godoc
// @Tags Clusters
// @Summary Delete cluster
// @Description Delete cluster
// @Accept json
// @Produce json
// @Param clusterId path string true "clusterId"
// @Success 200 {object} domain.Cluster
// @Router /clusters/{clusterId} [delete]
// @Security     JWT
func (h *ClusterHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, ok := vars["clusterId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "C_INVALID_CLUSTER_ID", ""))
		return
	}

	err := h.usecase.Delete(domain.ClusterId(clusterId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

func (h *ClusterHandler) GetKubernetesInfo(w http.ResponseWriter, r *http.Request) {
	// GetKubernetesInfo godoc
	// @Tags Clusters
	// @Summary Get kubernetes info
	// @Description Get kubernetes info for cluster
	// @Accept json
	// @Produce json
	// @Param clusterId path string true "clusterId"
	// @Success 200 {object} ClusterKubeInfo
	// @Router /clusters/{clusterId}/kubeInfo [get]
	/*
		vars := mux.Vars(r)
		clusterId, ok := vars["clusterId"]
		if !ok {
			log.Error("Failed to get clusterId")
		}

		clientset_user, err := h.GetClientFromClusterId(clusterId)
		if err != nil {
			log.Error("Failed to get clientset for clusterId ", clusterId)
			InternalServerError(w)
			return
		}

		var out struct {
			KubeInfo ClusterKubeInfo `json:"kubeInfo"`
		}
		out.KubeInfo.Updated = time.Now()

		pods, err := clientset_user.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			out.KubeInfo.Pods = len(pods.Items)
		} else {
			log.Error("Failed to get pods. err : ", err)
		}

		nodes, err := clientset_user.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			out.KubeInfo.Nodes = len(nodes.Items)
		} else {
			log.Error("Failed to get nodes. err : ", err)
		}

		services, err := clientset_user.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			out.KubeInfo.Services = len(services.Items)
		} else {
			log.Error("Failed to get services. err : ", err)

		}

		namespaces, err := clientset_user.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			out.KubeInfo.Namespaces = len(namespaces.Items)
		} else {
			log.Error("Failed to get namespaces. err : ", err)

		}

		version, err := helper.GetKubernetesVserionByClusterId(clusterId)
		if err == nil {
			out.KubeInfo.Version = version
		} else {
			log.Error("Failed to get kubernetes version. err : ", err)
		}

		ResponseJSON(w, http.StatusOK, out)
	*/
}

func (h *ClusterHandler) GetClusterApplications(w http.ResponseWriter, r *http.Request) {
	// GetClusterApplications godoc
	// @Tags Clusters
	// @Summary Get application list
	// @Description Get application list by clusterId
	// @Accept json
	// @Produce json
	// @Param clusterId path string false "clusterId"
	// @Success 200 {object} []ApplicationJson
	// @Router /clusters/{clusterId}/applications [get]
	/*
		vars := mux.Vars(r)
		clusterId, ok := vars["clusterId"]
		if !ok {
			log.Error("Failed to get clusterId")
			ErrorJSON(w, "invalid clusterId", http.StatusBadRequest)
		}

		var applications = []*pb.AppGroup{}
		res, err := appInfoClient.GetAppGroupsByClusterID(context.TODO(), &pb.IDRequest{ID: clusterId})
		if err != nil {
			log.Error("Failed to get appgroups err : ", err)
			InternalServerError(w)
			return
		}
		applications = res.GetAppGroups()

		var out struct {
			Applications []ApplicationJson `json:"applications"`
		}
		out.Applications = make([]ApplicationJson, 0)
		for _, appGroup := range applications {
			outApplication := ApplicationJson{}
			reflectApplication(&outApplication, appGroup)
			out.Applications = append(out.Applications, outApplication)
		}

		ResponseJSON(w, http.StatusOK, out)
	*/
}

func (h *ClusterHandler) GetClusterApplicationsKubeInfo(w http.ResponseWriter, r *http.Request) {
	/*
		vars := mux.Vars(r)
		clusterId, ok := vars["clusterId"]
		if !ok {
			log.Error("Failed to get clusterId")
			ErrorJSON(w, "invalid clusterId", http.StatusBadRequest)
		}

		var out struct {
			ApplicationKubeInfos []KubeInfoJson `json:"applicationKubeInfos"`
		}
		out.ApplicationKubeInfos = make([]KubeInfoJson, 0)

		res, err := appInfoClient.GetAppGroupsByClusterID(context.TODO(), &pb.IDRequest{Id: clusterId})
		if err != nil {
			log.Error("Failed to get appgroups err : ", err)
			InternalServerError(w)
			return
		}
		for _, appGroup := range res.GetAppGroups() {
			namespace := ""
			if appGroup.GetType() == pb.AppGroupType_LMA || appGroup.GetType() == pb.AppGroupType_LMA_EFK {
				namespace = "lma"
			} else if appGroup.GetType() == pb.AppGroupType_SERVICE_MESH {
				namespace = "istio-system"
			} else {
				ErrorJSON(w, fmt.Sprintf("invalid appgroup : %s", appGroup.GetType()), http.StatusBadRequest)
				continue
			}

			clientset, err := h.GetClientFromClusterId(appGroup.GetClusterId())
			if err != nil {
				log.Error("Failed to get clientset for clusterId", appGroup.GetClusterId())
				continue
			}

			pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Failed to get pods. ", err)
				continue
			}

			services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Failed to get service. ", err)
				continue
			}

			outKubeInfo := KubeInfoJson{}
			outKubeInfo.ApplicationId = appGroup.GetAppGroupId()
			outKubeInfo.Pods = make([]PodJson, 0)
			outKubeInfo.Services = make([]ServiceJson, 0)

			for _, pod := range pods.Items {
				outPod := PodJson{}
				outPod.Id = string(pod.UID)
				outPod.Name = pod.Name
				outPod.Namespace = pod.Namespace
				outPod.Status = string(pod.Status.Phase)
				outPod.Message = pod.Status.Message

				if pod.Status.StartTime != nil {
					outPod.Started = time.Unix(pod.Status.StartTime.Unix(), 0)
				}
				outKubeInfo.Pods = append(outKubeInfo.Pods, outPod)
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

				outKubeInfo.Services = append(outKubeInfo.Services, outService)

			}
			out.ApplicationKubeInfos = append(out.ApplicationKubeInfos, outKubeInfo)
		}

		ResponseJSON(w, http.StatusOK, out)
	*/
}

func (h *ClusterHandler) GetClusterKubeConfig(w http.ResponseWriter, r *http.Request) {
	// GetClusterKubeConfig godoc
	// @Tags Clusters
	// @Summary Get kubernetes kubeconfig
	// @Description Get kubernetes kubeconfig for cluster
	// @Accept json
	// @Produce json
	// @Param clusterId path string true "clusterId"
	// @Success 200 {object} object
	// @Router /clusters/{clusterId}/kubeconfig [get]
	/*
		vars := mux.Vars(r)
		clusterId, ok := vars["clusterId"]
		if !ok {
			log.Error("Failed to get clusterId")
			ErrorJSON(w, "invalid clusterId", http.StatusBadRequest)
			return
		}

		organizationId := r.Header.Get("OrganizationId")

		kubeconfig, err := helper.GetKubeConfig(clusterId)
		if err != nil {
			ErrorJSON(w, "failed to get admin cluster", http.StatusBadRequest)
			return
		}

		var out struct {
			Kubeconfig string `json:"kubeconfig"`
		}

		out.Kubeconfig = string(kubeconfig[:])

		h.AddHistory(r, organizationId, "cluster", fmt.Sprintf("클러스터 [%s]의 kubeconfig를 다운로드 하였습니다.", clusterId))

		ResponseJSON(w, http.StatusOK, out)
	*/
}

func (h *ClusterHandler) GetClusterKubeResources(w http.ResponseWriter, r *http.Request) {
	// GetClusterKubeResources godoc
	// @Tags Clusters
	// @Summary Get kubernetes resources
	// @Description Get kubernetes resources
	// @Accept json
	// @Produce json
	// @Param clusterId path string true "clusterId"
	// @Success 200 {object} ClusterJson
	// @Router /clusters/{clusterId}/kube-resources [get]
	/*
		vars := mux.Vars(r)
		clusterId, ok := vars["clusterId"]
		if !ok {
			log.Error("Failed to get clusterId")
			ErrorJSON(w, "Invalid clusterId", http.StatusBadRequest)
		}

		clientset, err := h.GetClientFromClusterId(clusterId)
		if err != nil {
			log.Error("Failed to get clientset for clusterId", clusterId)
			InternalServerError(w)
			return
		}

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to get pods", err)
			InternalServerError(w)
			return
		}

		services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to get services", err)
			InternalServerError(w)
			return
		}

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to get namespaces", err)
			InternalServerError(w)
			return
		}

		events, err := clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to get events", err)
			InternalServerError(w)
			return
		}

		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to get events", err)
			InternalServerError(w)
			return
		}

		var out struct {
			Pods       []PodJson       `json:"pods"`
			Services   []ServiceJson   `json:"services"`
			Namespaces []NamespaceJson `json:"namespaces"`
			Events     []EventJson     `json:"events"`
			Nodes      []NodeJson      `json:"nodes"`
		}
		out.Pods = make([]PodJson, 0)
		out.Services = make([]ServiceJson, 0)
		out.Namespaces = make([]NamespaceJson, 0)
		out.Events = make([]EventJson, 0)
		out.Nodes = make([]NodeJson, 0)

		for _, pod := range pods.Items {
			outPod := PodJson{}
			outPod.Id = string(pod.UID)
			outPod.Name = pod.Name
			outPod.Namespace = pod.Namespace
			outPod.Status = string(pod.Status.Phase)
			outPod.Message = pod.Status.Message
			if pod.Status.StartTime != nil {
				outPod.Started = time.Unix(pod.Status.StartTime.Unix(), 0)
			}

			out.Pods = append(out.Pods, outPod)
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

			out.Services = append(out.Services, outService)
		}

		for _, namespace := range namespaces.Items {

			outNamespace := NamespaceJson{}
			outNamespace.Id = string(namespace.UID)
			outNamespace.Name = namespace.Name
			outNamespace.Status = string(namespace.Status.Phase)

			outNamespace.Labels = make([]string, 0)
			for k, v := range namespace.Labels {
				if k == "kubernetes.io/metadata.name" || k == "name" {
					continue
				}
				outNamespace.Labels = append(outNamespace.Labels, "["+k+":"+v+"]")
			}

			out.Namespaces = append(out.Namespaces, outNamespace)
		}

		for i, item := range events.Items {
			if i >= 100 {
				break
			}
			outEvent := EventJson{}
			outEvent.Id = string(item.UID)
			outEvent.Namespace = item.Namespace
			outEvent.Type = item.Type
			outEvent.Reason = item.Reason
			outEvent.Message = item.Message
			outEvent.Updated = time.Unix(item.CreationTimestamp.Unix(), 0)

			out.Events = append(out.Events, outEvent)
		}

		for i, item := range nodes.Items {
			if i >= 100 {
				break
			}

			outNode := NodeJson{}

			for _, contidion := range item.Status.Conditions {
				if contidion.Type == "Ready" {
					if contidion.Status == "True" {
						outNode.Status = "Ready"
					} else {
						outNode.Status = "NotReady"
					}
				}
			}
			outNode.Id = string(item.UID)
			outNode.Name = string(item.Name)
			outNode.InstanceType = string(item.Labels["node.kubernetes.io/instance-type"])

			_, exist := item.Labels["node-role.kubernetes.io/control-plane"]
			if exist {
				outNode.Role = "ControlPlane"
			} else {
				_, exist := item.Labels["taco-lma"]
				if exist {
					outNode.Role = "TKS Node"
				} else {
					outNode.Role = "USER Node"
				}
			}
			outNode.Updated = time.Unix(item.CreationTimestamp.Unix(), 0)

			out.Nodes = append(out.Nodes, outNode)
		}

		ResponseJSON(w, http.StatusOK, out)
	*/
}

func (h *ClusterHandler) SetIstioLabel(w http.ResponseWriter, r *http.Request) {
	// SetIstioLabel godoc
	// @Tags Clusters
	// @Summary Set Istio label to namespace
	// @Description Set istio label to namespace on kubernetes
	// @Accept json
	// @Produce json
	// @Param clusterId path string true "clusterId"
	// @Success 200 {object} object
	// @Router /clusters/{clusterId}/kube-resources/{namespace}/istio-label [post]
	/*
		vars := mux.Vars(r)
		clusterId, ok := vars["clusterId"]
		if !ok {
			log.Error("Failed to get clusterId")
			ErrorJSON(w, "invalid clusterId", http.StatusBadRequest)
			return
		}
		namespace, ok := vars["namespace"]
		if !ok {
			log.Error("Failed to get namespace")
			ErrorJSON(w, "invalid namespace", http.StatusBadRequest)
			return
		}

		var input struct {
			Value string `json:"value"`
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

		log.Info(input)

		if input.Value != "enabled" && input.Value != "disabled" {
			ErrorJSON(w, "invalid value", http.StatusBadRequest)
			return
		}

		organizationId := r.Header.Get("OrganizationId")

		clientset, err := h.GetClientFromClusterId(clusterId)
		if err != nil {
			log.Error("Failed to get clientset for clusterId", clusterId)
			InternalServerError(w)
			return
		}

		labelPatch := fmt.Sprintf(`[{"op":"add","path":"/metadata/labels/%s","value":"%s" }]`, "istio-injection", input.Value)
		_, err = clientset.CoreV1().Namespaces().Patch(context.TODO(), namespace, types.JSONPatchType, []byte(labelPatch), metav1.PatchOptions{})

		h.AddHistory(r, organizationId, "cluster", fmt.Sprintf("클러스터 [%s]의 namespace[%s] 에 서비스메쉬 레이블 설정을 하였습니다.", clusterId, namespace))

		var out struct {
		}

		ResponseJSON(w, http.StatusOK, out)
	*/
}
