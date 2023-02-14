package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-common/pkg/log"
	"github.com/openinfradev/tks-proto/tks_pb"
	pb "github.com/openinfradev/tks-proto/tks_pb"
	"github.com/spf13/viper"
)

type AppServeAppJson = struct {
	Id              string    `json:"id"`
	ProjectId       string    `json:"projectId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Type            string    `json:"type"`
	EndpointUrl     string    `json:"endpointUrl"`
	TargetClusterId string    `json:"targetClusterId"`
	Status          string    `json:"status"`
	Creator         string    `json:"creator"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AppServeAppTaskJson = struct {
	Id             string    `json:"id"`
	AppServeAppId  string    `json:"appServeAppId"`
	Version        string    `json:"version"`
	Status         string    `json:"status"`
	Output         string    `json:"output"`
	ArtifactUrl    string    `json:"artifactUrl"`
	ImageUrl       string    `json:"imageUrl"`
	ExecutablePath string    `json:"executablePath"`
	Profile        string    `json:"profile"`
	Port           string    `json:"port"`
	HelmRevision   int32     `json:"helmRevision"`
	ResourceSpec   string    `json:"resourceSpec"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (h *APIHandler) GetAppServeApps(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	projectId := urlParams.Get("projectId")
	if projectId == "" {
		ErrorJSON(w, "Invalid projectId", http.StatusOK)
		return
	}

	res, err := asaClient.GetAppServeApps(context.TODO(), &tks_pb.GetAppServeAppsRequest{
		ContractId: projectId,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get app-serve-apps. Err: %s", err)
		log.Error(errMsg)
	}

	log.Info(res.GetAppServeApps())
	clusters, err := clusterInfoClient.GetClusters(context.TODO(), &tks_pb.GetClustersRequest{ContractId: projectId, CspId: ""})
	if err != nil {
		log.Error("Failed to get clusters err : ", err)
	}

	var out struct {
		Apps []AppServeAppJson `json:"apps"`
	}
	out.Apps = make([]AppServeAppJson, 0)

	for _, app := range res.GetAppServeApps() {
		for _, cluster := range clusters.GetClusters() {

			if app.GetTargetClusterId() == cluster.GetId() &&
				cluster.GetStatus() == pb.ClusterStatus_RUNNING {

				outApp := AppServeAppJson{}
				reflectAppServeApp(&outApp, app)

				out.Apps = append(out.Apps, outApp)
			}
		}
	}

	ResponseJSON(w, out, http.StatusOK)
}

func (h *APIHandler) GetAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	asaId, ok := vars["asaId"]
	if !ok {
		log.Error("Failed to get asaId")
	}

	res, err := asaClient.GetAppServeApp(context.TODO(), &tks_pb.GetAppServeAppRequest{
		AppServeAppId: asaId,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get app-serve-app. Err: %s", err)
		log.Error(errMsg)
	}

	log.Info(res.GetAppServeAppCombined())

	var out struct {
		App   AppServeAppJson       `json:"app"`
		Tasks []AppServeAppTaskJson `json:"tasks"`
	}
	reflectAppServeApp(&out.App, res.GetAppServeAppCombined().GetAppServeApp())

	out.Tasks = make([]AppServeAppTaskJson, 0)
	reflectAppServeAppTasks(&out.Tasks, res.GetAppServeAppCombined().GetTasks())

	ResponseJSON(w, out, http.StatusOK)
}

func (h *APIHandler) CreateAppServeApp(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectId       string `json:"projectId"`
		ContractId      string `json:"contractId"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		Type            string `json:"type"`
		Version         string `json:"version"`
		ArtifactUrl     string `json:"artifactUrl"`
		ImageUrl        string `json:"imageUrl"`
		Port            string `json:"port"`
		Profile         string `json:"Profile"`
		TargetClusterId string `json:"targetClusterId"`
		ExecutablePath  string `json:"executablePath"`
		ResourceSpec    string `json:"resourceSpec"`
		Creator         string `json:"creator"`
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

	var requestApp struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Type            string `json:"type"`
		Version         string `json:"version"`
		ArtifactUrl     string `json:"artifact_url"`
		ImageUrl        string `json:"image_url"`
		Port            string `json:"port"`
		ContractId      string `json:"contract_id"`
		Profile         string `json:"profile"`
		TargetClusterId string `json:"target_cluster_id"`
		ExecutablePath  string `json:"executable_path"`
		ResourceSpec    string `json:"resource_spec"`
	}

	requestApp.Name = input.Name
	requestApp.Type = input.Type
	requestApp.Version = input.Version
	requestApp.ArtifactUrl = input.ArtifactUrl
	requestApp.ImageUrl = input.ImageUrl
	requestApp.Port = input.Port
	requestApp.ContractId = input.ProjectId
	requestApp.Profile = input.Profile
	requestApp.TargetClusterId = input.TargetClusterId
	requestApp.ExecutablePath = input.ExecutablePath
	requestApp.ResourceSpec = input.ResourceSpec

	reqBodyBytes, err := json.Marshal(requestApp)
	if err != nil {
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	buff := bytes.NewBuffer(reqBodyBytes)

	log.Info(input)

	url := "http://" + viper.GetString("asa-address") + ":" + strconv.Itoa(viper.GetInt("asa-port"))
	res, err := http.Post(url+"/apps", "application/json", buff)
	if err != nil || res == nil {
		log.Error("error from app-serve-lcm err: ", err)
	}
	if res.StatusCode != 200 {
		log.Error("error from app-serve-lcm return code: ", res.StatusCode)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("error closing http body")
		}
	}()

	body, err = ioutil.ReadAll(res.Body)
	log.Info(body)

	if err != nil {
		log.Error("Create application error : ", err)
		h.AddHistory(r, input.ProjectId, "application", fmt.Sprintf("어플리케이션 [%s] 생성에 실패하였습니다.", input.Name))
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.AddHistory(r, input.ProjectId, "appserve", fmt.Sprintf("어플리케이션 [%s]을 생성하였습니다.", input.Name))

	var out struct {
		apps []AppServeAppJson `json:"apps"`
	}

	time.Sleep(time.Second * 6)
	ResponseJSON(w, out, http.StatusOK)
}

func (h *APIHandler) UpdateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	asaId, ok := vars["asaId"]
	if !ok || asaId == "" {
		ErrorJSON(w, "invalid asaId", http.StatusBadRequest)
		return
	}

	var input struct {
		ProjectId       string `json:"projectId"`
		ContractId      string `json:"contractId"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		Type            string `json:"type"`
		Version         string `json:"version"`
		ArtifactUrl     string `json:"artifactUrl"`
		ImageUrl        string `json:"imageUrl"`
		Port            string `json:"port"`
		Profile         string `json:"Profile"`
		TargetClusterId string `json:"targetClusterId"`
		ExecutablePath  string `json:"executablePath"`
		ResourceSpec    string `json:"resourceSpec"`
		Creator         string `json:"creator"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	var requestApp struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Type            string `json:"type"`
		Version         string `json:"version"`
		ArtifactUrl     string `json:"artifact_url"`
		ImageUrl        string `json:"image_url"`
		Port            string `json:"port"`
		ContractId      string `json:"contract_id"`
		Profile         string `json:"profile"`
		TargetClusterId string `json:"target_cluster_id"`
		ExecutablePath  string `json:"executable_path"`
		ResourceSpec    string `json:"resource_spec"`
	}

	requestApp.ID = asaId
	requestApp.Name = input.Name
	requestApp.Type = input.Type
	requestApp.Version = input.Version
	requestApp.ArtifactUrl = input.ArtifactUrl
	requestApp.ImageUrl = input.ImageUrl
	requestApp.Port = input.Port
	requestApp.ContractId = input.ProjectId
	requestApp.Profile = input.Profile
	requestApp.TargetClusterId = input.TargetClusterId
	requestApp.ExecutablePath = input.ExecutablePath
	requestApp.ResourceSpec = input.ResourceSpec

	reqBodyBytes, err := json.Marshal(requestApp)
	if err != nil {
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}
	buff := bytes.NewBuffer(reqBodyBytes)

	log.Info(input)

	url := "http://" + viper.GetString("asa-address") + ":" + strconv.Itoa(viper.GetInt("asa-port"))
	req, err := http.NewRequest("PUT", url+"/apps/"+asaId, buff)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error("Failed to create http request err: ", err)
		InternalServerError(w)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Error("Failed to create http request err: ", err)
		InternalServerError(w)
		return
	}

	if res.StatusCode != 200 {
		log.Error("error from app-serve-lcm return code: ", res.StatusCode)
		InternalServerError(w)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("error closing http body")
		}
	}()

	body, err = ioutil.ReadAll(res.Body)
	log.Info(body)

	if err != nil {
		log.Error("Update application error : ", err)
		h.AddHistory(r, input.ProjectId, "application", fmt.Sprintf("어플리케이션 [%s] 업데이트에 실패하였습니다.", input.Name))
		InternalServerError(w)
		return
	}

	h.AddHistory(r, input.ProjectId, "appserve", fmt.Sprintf("어플리케이션 [%s]을 업데이트 하였습니다.", input.Name))

	var out struct {
		apps []AppServeAppJson `json:"apps"`
	}

	time.Sleep(time.Second * 6)
	ResponseJSON(w, out, http.StatusOK)
}

func (h *APIHandler) DeleteAppServeApp(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	asaId, ok := vars["asaId"]
	if !ok || asaId == "" {
		log.Error("Failed to get asaId")
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	res, err := asaClient.GetAppServeApp(context.TODO(), &tks_pb.GetAppServeAppRequest{
		AppServeAppId: asaId,
	})
	if err != nil {
		log.Error("Failed to get app-serve-app. Err: %s", err)
		InternalServerError(w)
		return
	}

	url := "http://" + viper.GetString("asa-address") + ":" + strconv.Itoa(viper.GetInt("asa-port"))
	req, err := http.NewRequest("DELETE", url+"/apps/"+asaId, nil)
	if err != nil {
		log.Error("Failed to create http request: ", err)
		InternalServerError(w)
		return
	}

	client := &http.Client{}
	resHttp, err := client.Do(req)
	if err != nil || resHttp == nil {
		log.Error("error from app-serve-lcm err: ", err)
		InternalServerError(w)
		return
	}
	if resHttp.StatusCode != 200 {
		log.Error("error from app-serve-lcm return code: ", resHttp.StatusCode)
		InternalServerError(w)
		return
	}
	defer func() {
		if err := resHttp.Body.Close(); err != nil {
			log.Error("error closing http body")
		}
	}()

	if err != nil {
		log.Error("Delete application error : ", err)
		h.AddHistory(r, res.GetAppServeAppCombined().GetAppServeApp().GetContractId(), "appserve", fmt.Sprintf("어플리케이션 [%s] 삭제에 실패하였습니다.", res.GetAppServeAppCombined().GetAppServeApp().GetName()))
		InternalServerError(w)
		return
	}

	h.AddHistory(r, res.GetAppServeAppCombined().GetAppServeApp().GetContractId(), "appserve", fmt.Sprintf("어플리케이션 [%s]을 삭제하였습니다.", res.GetAppServeAppCombined().GetAppServeApp().GetName()))

	var out struct {
	}

	time.Sleep(time.Second * 6)
	ResponseJSON(w, out, http.StatusOK)
}

func reflectAppServeApp(out *AppServeAppJson, asa *tks_pb.AppServeApp) {
	out.Id = asa.GetId()
	out.Name = asa.GetName()
	out.ProjectId = asa.GetContractId()
	out.Type = asa.GetType()
	out.EndpointUrl = asa.GetEndpointUrl()
	out.TargetClusterId = asa.GetTargetClusterId()
	out.Status = asa.GetStatus()
	out.CreatedAt = asa.GetCreatedAt().AsTime()
	out.UpdatedAt = asa.GetUpdatedAt().AsTime()
}

func reflectAppServeAppTasks(out *[]AppServeAppTaskJson, tasks []*tks_pb.AppServeAppTask) {

	if last := len(tasks) - 1; last >= 0 {
		for i, task := last, tasks[0]; i >= 0; i-- {
			task = tasks[i]
			outTask := AppServeAppTaskJson{}

			outTask.Id = task.GetId()
			outTask.Version = task.GetVersion()
			outTask.Status = task.GetStatus()
			outTask.Output = task.GetOutput()
			outTask.ArtifactUrl = task.GetArtifactUrl()
			outTask.ImageUrl = task.GetImageUrl()
			outTask.ExecutablePath = task.GetExecutablePath()
			outTask.Profile = task.GetProfile()
			outTask.Port = task.GetPort()
			outTask.ResourceSpec = task.GetResourceSpec()
			outTask.HelmRevision = task.GetHelmRevision()
			outTask.CreatedAt = task.GetCreatedAt().AsTime()
			outTask.UpdatedAt = task.GetUpdatedAt().AsTime()

			*out = append(*out, outTask)
		}

	}

}
