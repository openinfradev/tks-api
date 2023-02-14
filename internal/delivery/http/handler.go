package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	gcache "github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-common/pkg/argowf"
	"github.com/openinfradev/tks-common/pkg/grpc_client"
	"github.com/openinfradev/tks-common/pkg/log"
	pb "github.com/openinfradev/tks-proto/tks_pb"
)

var (
	argowfClient argowf.Client

	contractClient    pb.ContractServiceClient
	cspInfoClient     pb.CspInfoServiceClient
	clusterInfoClient pb.ClusterInfoServiceClient
	appInfoClient     pb.AppInfoServiceClient
	asaClient         pb.AppServeAppServiceClient
	lcmClient         pb.ClusterLcmServiceClient
)

type APIHandler struct {
	Repository *repository.Repository
	Cache      *gcache.Cache
}

func InitializeGrpcClient() {
	var err error
	if _, contractClient, err = grpc_client.CreateContractClient(viper.GetString("contract-address"), viper.GetInt("contract-port"), false, ""); err != nil {
		log.Fatal("failed to create contract client : ", err)
	}

	if _, cspInfoClient, err = grpc_client.CreateCspInfoClient(viper.GetString("info-address"), viper.GetInt("info-port"), false, ""); err != nil {
		log.Fatal("failed to create cspinfo client : ", err)
	}

	if _, clusterInfoClient, err = grpc_client.CreateClusterInfoClient(viper.GetString("info-address"), viper.GetInt("info-port"), false, ""); err != nil {
		log.Fatal("failed to create cluster client : ", err)
	}

	if _, appInfoClient, err = grpc_client.CreateAppInfoClient(viper.GetString("info-address"), viper.GetInt("info-port"), false, ""); err != nil {
		log.Fatal("failed to create appinfo client : ", err)
	}

	if _, lcmClient, err = grpc_client.CreateLcmClient(viper.GetString("lcm-address"), viper.GetInt("lcm-port"), false, ""); err != nil {
		log.Fatal("failed to create lcm client : ", err)
	}

	if _, asaClient, err = grpc_client.CreateAppServeAppClient(viper.GetString("info-address"), viper.GetInt("info-port"), false, ""); err != nil {
		log.Fatal("failed to create asa client : ", err)
	}
}

func InitializeHttpClient() {
	_client, err := argowf.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
	if err != nil {
		log.Fatal("failed to create argowf client : ", err)
	}
	argowfClient = _client
}

func ErrorJSON(w http.ResponseWriter, message string, code int) {
	var out struct {
		Message string `json:"message"`
		Code    int    `json:"status_code"`
	}
	out.Message = message
	out.Code = code

	log.Error(fmt.Sprintf("[API_RESPONSE_ERROR] [%s]", message))
	ResponseJSON(w, out, code)
}

func InternalServerError(w http.ResponseWriter) {
	ErrorJSON(w, "internal server error", http.StatusInternalServerError)
}

func ResponseJSON(w http.ResponseWriter, data interface{}, code int) {
	//time.Sleep(time.Second * 1) // for test
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	log.Info(fmt.Sprintf("[API_RESPONSE] [%s]", data))
	json.NewEncoder(w).Encode(data)
}

func GetSession(r *http.Request) (string, string) {
	return r.Header.Get("Id"), r.Header.Get("AccountId")
}

func (h *APIHandler) GetClientFromClusterId(clusterId string) (*kubernetes.Clientset, error) {
	const prefix = "CACHE_KEY_KUBE_CLIENT_"
	value, found := h.Cache.Get(prefix + clusterId)
	if found {
		return value.(*kubernetes.Clientset), nil
	}
	client, err := helper.GetClientFromClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	h.Cache.Set(prefix+clusterId, client, gcache.DefaultExpiration)
	return client, nil
}

func (h *APIHandler) GetKubernetesVserion() (string, error) {
	const prefix = "CACHE_KEY_KUBE_VERSION_"
	value, found := h.Cache.Get(prefix)
	if found {
		return value.(string), nil
	}
	version, err := helper.GetKubernetesVserion()
	if err != nil {
		return "", err
	}

	h.Cache.Set(prefix, version, gcache.DefaultExpiration)
	return version, nil
}

func (h *APIHandler) GetSession(r *http.Request) (string, string) {
	return r.Header.Get("Id"), r.Header.Get("AccountId")
}

func (h *APIHandler) AddHistory(r *http.Request, projectId string, historyType string, description string) error {
	/*
		userId, _ := h.GetSession(r)

		err := h.Repository.AddHistory(userId, projectId, historyType, description)
		if err != nil {
			log.Error(err)
			return err
		}
	*/

	return nil
}
