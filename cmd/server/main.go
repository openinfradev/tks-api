package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/openinfradev/tks-api/internal/aws/ses"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "github.com/openinfradev/tks-api/api/swagger"
	"github.com/openinfradev/tks-api/internal/database"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/route"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
)

func init() {
	flag.String("external-address", "http://tks-api.tks.svc:9110", "service address")
	flag.Int("port", 8080, "service port")
	flag.String("web-root", "../../web", "path of root path for web")
	flag.String("argo-address", "http://localhost", "service address for argoworkflow")
	flag.Int("argo-port", 0, "service port for argoworkflow")
	flag.String("dbhost", "localhost", "host of postgreSQL")
	flag.String("dbname", "tks", "name of releation")
	flag.String("dbport", "5432", "port of postgreSQL")
	flag.String("dbuser", "postgres", "postgreSQL user")
	flag.String("dbpassword", "password", "password for postgreSQL user")
	flag.String("kubeconfig-path", "/Users/1110640/.kube/config_dev", "path of kubeconfig. used development only!")
	flag.String("jwt-secret", "tks-api-secret", "secret value of jwt")
	flag.String("git-base-url", "https://github.com", "git base url")
	flag.String("git-account", "decapod10", "git account of admin cluster")
	flag.String("revision", "main", "revision")
	flag.Bool("migrate-db", true, "If the values is true, enable db migration. recommend only development")

	// app-serve-apps
	flag.String("image-registry-url", "harbor-dev.taco-cat.xyz/appserving", "URL of image registry")
	flag.String("harbor-pw-secret", "harbor-core", "name of harbor password secret")
	flag.String("git-repository-url", "github.com/openinfradev", "URL of git repository")

	// keycloak
	flag.String("keycloak-address", "https://keycloak-kyuho.taco-cat.xyz/auth", "URL of keycloak")
	flag.String("keycloak-admin", "admin", "user of keycloak")
	flag.String("keycloak-password", "admin", "password of keycloak")
	flag.String("keycloak-client-secret", keycloak.DefaultClientSecret, "realm of keycloak")

	// aws ses
	flag.String("aws-region", "ap-northeast-2", "region of aws ses")
	flag.String("aws-access-key-id", "", "access key id of aws ses")
	flag.String("aws-secret-access-key", "", "access key of aws ses")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Error(err)
	}

}

// @title tks-api service
// @version 1.0
// @description This is backend api service for tks platform

// @contact.name taekyu.kang@sk.com
// @contact.url
// @contact.email taekyu.kang@sk.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securitydefinitions.apikey  JWT
// @in                          header
// @name                        Authorization

// @host tks-api-dev.taco-cat.xyz
// @BasePath /api/1.0/
func main() {
	log.Info("*** Arguments *** ")
	for i, s := range viper.AllSettings() {
		log.Info(fmt.Sprintf("%s : %v", i, s))
	}
	log.Info("****************** ")

	// For web service
	asset := route.NewAssetHandler(viper.GetString("web-root"))

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("cannot connect gormDB")
	}

	// Initialize external client
	var argoClient argowf.ArgoClient
	if viper.GetString("argo-address") == "" || viper.GetInt("argo-port") == 0 {
		argoClient, err = argowf.NewMock()
		if err != nil {
			log.Fatal("failed to create argowf client : ", err)
		}
	} else {
		argoClient, err = argowf.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
		if err != nil {
			log.Fatal("failed to create argowf client : ", err)
		}
	}

	keycloak := keycloak.New(&keycloak.Config{
		Address:       viper.GetString("keycloak-address"),
		AdminId:       viper.GetString("keycloak-admin"),
		AdminPassword: viper.GetString("keycloak-password"),
		ClientSecret:  viper.GetString("keycloak-client-secret"),
	})

	err = keycloak.InitializeKeycloak()
	if err != nil {
		log.Fatal("failed to initialize keycloak : ", err)
	}
	err = ses.Initialize()
	if err != nil {
		log.Fatal("failed to initialize ses : ", err)
	}

	route := route.SetupRouter(db, argoClient, asset, keycloak)

	log.Info("Starting server on ", viper.GetInt("port"))
	err = http.ListenAndServe("0.0.0.0:"+strconv.Itoa(viper.GetInt("port")), route)
	if err != nil {
		log.Fatal(err)
	}
}
