package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/openinfradev/tks-api/api/swagger"
	"github.com/openinfradev/tks-api/internal/database"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/mail"
	"github.com/openinfradev/tks-api/internal/route"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
)

func parseCmd() {
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
	flag.String("kubeconfig-path", "", "path of kubeconfig. used development only!")
	flag.String("jwt-secret", "tks-api-secret", "secret value of jwt")
	flag.String("git-base-url", "https://github.com", "git base url")
	flag.String("git-account", "decapod10", "git account of admin cluster")
	flag.String("external-gitea-url", "http://ip-10-0-76-86.ap-northeast-2.compute.internal:30303", "gitea url for byoh agent download")
	flag.String("revision", "main", "revision")
	flag.String("aws-secret", "awsconfig-secret", "aws secret")
	flag.Int("migrate-db", 1, "If the values is true, enable db migration. recommend only development")

	// console
	flag.String("console-address", "https://tks-console-dev.taco-cat.xyz", "service address for console")

	// app-serve-apps
	flag.String("image-registry-url", "harbor.taco-cat.xyz/appserving", "URL of image registry")
	flag.String("harbor-pw-secret", "harbor-core", "name of harbor password secret")
	flag.String("git-repository-url", "github.com/openinfradev", "URL of git repository")

	// keycloak
	flag.String("keycloak-address", "https://keycloak-kyuho.taco-cat.xyz/auth", "URL of keycloak")
	flag.String("keycloak-admin", "admin", "user of keycloak")
	flag.String("keycloak-password", "admin", "password of keycloak")
	flag.String("keycloak-client-secret", keycloak.DefaultClientSecret, "realm of keycloak")

	flag.String("mail-provider", "aws", "mail provider")
	// mail (smtp)
	flag.String("smtp-host", "", "smtp hosts")
	flag.Int("smtp-port", 0, "smtp port")
	flag.String("smtp-username", "", "smtp username")
	flag.String("smtp-password", "", "smtp password")
	flag.String("smtp-from-email", "", "smtp from email")
	// mail (aws ses)
	flag.String("aws-region", "ap-northeast-2", "region of aws ses")
	flag.String("aws-access-key-id", "", "access key id of aws ses")
	flag.String("aws-secret-access-key", "", "access key of aws ses")

	// alerts
	flag.String("alert-slack", "", "slack url for LMA alert")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Error(context.Background(), err)
	}

	address := viper.GetString("external-address")
	arr := strings.Split(address, "//")
	if len(arr) >= 2 {
		address = arr[1]
	}

	swagger.SwaggerInfo.Host = address
}

//	@title			tks-api service
//	@version		1.0
//	@description	This is backend api service for tks platform

//	@contact.name	taekyu.kang@sk.com
//	@contact.url
//	@contact.email	taekyu.kang@sk.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@securitydefinitions.apikey	JWT
//	@in							header
//	@name						Authorization

// @host		tks-api-dev.taco-cat.xyz
// @BasePath	/api/1.0/
func main() {
	parseCmd()

	ctx := context.Background()
	log.Info(ctx, "*** Arguments *** ")
	for i, s := range viper.AllSettings() {
		log.Info(ctx, fmt.Sprintf("%s : %v", i, s))
	}
	log.Info(ctx, "****************** ")

	// For web service
	asset := route.NewAssetHandler(viper.GetString("web-root"))

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(ctx, "cannot connect gormDB")
	}

	// Ensure default rows in database
	err = database.EnsureDefaultRows(db)
	if err != nil {
		log.Fatal(ctx, "cannot Initializing Default Rows in Database: ", err)
	}

	// Initialize external client
	var argoClient argowf.ArgoClient
	if viper.GetString("argo-address") == "" || viper.GetInt("argo-port") == 0 {
		argoClient, err = argowf.NewMock()
		if err != nil {
			log.Fatal(ctx, "failed to create argowf client : ", err)
		}
	} else {
		argoClient, err = argowf.New(viper.GetString("argo-address"), viper.GetInt("argo-port"), false, "")
		if err != nil {
			log.Fatal(ctx, "failed to create argowf client : ", err)
		}
	}

	keycloak := keycloak.New(&keycloak.Config{
		Address:       viper.GetString("keycloak-address"),
		AdminId:       viper.GetString("keycloak-admin"),
		AdminPassword: viper.GetString("keycloak-password"),
		ClientSecret:  viper.GetString("keycloak-client-secret"),
	})

	err = keycloak.InitializeKeycloak(ctx)
	if err != nil {
		log.Fatal(ctx, "failed to initialize keycloak : ", err)
	}
	err = mail.Initialize(ctx)
	if err != nil {
		log.Fatal(ctx, "failed to initialize ses : ", err)
	}

	route := route.SetupRouter(db, argoClient, keycloak, asset)

	log.Info(ctx, "Starting server on ", viper.GetInt("port"))
	err = http.ListenAndServe("0.0.0.0:"+strconv.Itoa(viper.GetInt("port")), route)
	if err != nil {
		log.Fatal(ctx, err)
	}
}
