package usecase

import (
	"github.com/spf13/viper"

	"github.com/openinfradev/tks-common/pkg/argowf"
	"github.com/openinfradev/tks-common/pkg/log"
)

var (
	argowfClient argowf.Client
)

func InitializeHttpClient() {
	_client, err := argowf.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
	if err != nil {
		log.Fatal("failed to create argowf client : ", err)
	}
	argowfClient = _client
}
