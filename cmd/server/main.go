package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "github.com/openinfradev/tks-api/api/swagger"
	"github.com/openinfradev/tks-api/internal/database"
	"github.com/openinfradev/tks-api/internal/route"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
)

func init() {
	flag.Int("port", 8080, "service port")
	flag.String("web-root", "../../web", "path of root path for web")
	flag.String("argo-address", "http://localhost", "service address for argoworkflow")
	flag.Int("argo-port", 2746, "service port for argoworkflow")
	flag.String("dbhost", "localhost", "host of postgreSQL")
	flag.String("dbport", "5432", "port of postgreSQL")
	flag.String("dbuser", "postgres", "postgreSQL user")
	flag.String("dbpassword", "password", "password for postgreSQL user")
	flag.String("kubeconfig-path", "/Users/1110640/.kube/config", "path of kubeconfig. used development only!")
	flag.String("jwt-secret", "tks-api-secret", "secret value of jwt")
	flag.String("git-base-url", "https://github.com", "git base url")
	flag.String("git-account", "demo-decapod10", "git account of admin cluster")

	// app-serve-apps
	flag.String("image-registry-url", "harbor-dev.taco-cat.xyz/appserving", "URL of image registry")
	flag.String("harbor-pw-secret", "harbor-core", "name of harbor password secret")
	flag.String("git-repository-url", "github.com/openinfradev", "URL of git repository")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

// @title tks-api service
// @version 1.0
// @description This is backend api service for tks

// @contact.name taekyu.kang@sk.com
// @contact.url
// @contact.email taekyu.kang@sk.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
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
	argoClient, err := argowf.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
	if err != nil {
		log.Fatal("failed to create argowf client : ", err)
	}

	route := route.SetupRouter(db, argoClient, asset)

	log.Info("Starting server on ", viper.GetInt("port"))
	err = http.ListenAndServe("0.0.0.0:"+strconv.Itoa(viper.GetInt("port")), route)
	if err != nil {
		log.Fatal(err)
	}
}
