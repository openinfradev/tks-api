package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	"github.com/spf13/viper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clientcmd "k8s.io/client-go/tools/clientcmd"

	"github.com/openinfradev/tks-api/pkg/log"
)

func getAdminConfig() (*rest.Config, error) {
	kubeconfigPath := viper.GetString("kubeconfig-path")
	if kubeconfigPath == "" {
		log.Info("Use in-cluster config")
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Error("Failed to load incluster kubeconfig")
			return nil, err
		}
		return config, nil
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", viper.GetString("kubeconfig-path"))
		if err != nil {
			log.Error("Failed to local kubeconfig")
			return nil, err
		}
		return config, nil
	}
}

func GetClientAdminCluster() (*kubernetes.Clientset, error) {
	config, err := getAdminConfig()
	if err != nil {
		log.Error("Failed to load kubeconfig")
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func GetAwsSecret() (awsAccessKeyId string, awsSecretAccessKey string, err error) {
	clientset, err := GetClientAdminCluster()
	if err != nil {
		return "", "", err
	}

	secrets, err := clientset.CoreV1().Secrets("argo").Get(context.TODO(), "awsconfig-secret", metav1.GetOptions{})
	if err != nil {
		log.Error(err)
		return "", "", err
	}

	strCredentials := string(secrets.Data["credentials"][:])
	arr := strings.Split(strCredentials, "\n")
	if len(arr) < 3 {
		return "", "", err
	}

	fmt.Sscanf(arr[1], "aws_access_key_id = %s", &awsAccessKeyId)
	fmt.Sscanf(arr[2], "aws_secret_access_key = %s", &awsSecretAccessKey)

	return
}

func GetAwsAccountIdSecret() (awsAccountId string, err error) {
	clientset, err := GetClientAdminCluster()
	if err != nil {
		return "", err
	}

	secrets, err := clientset.CoreV1().Secrets("argo").Get(context.TODO(), "tks-aws-user", metav1.GetOptions{})
	if err != nil {
		log.Error(err)
		return "", err
	}

	awsAccountId = string(secrets.Data["account_id"][:])
	return
}

func GetKubeConfig(clusterId string) ([]byte, error) {
	clientset, err := GetClientAdminCluster()
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-tks-user-kubeconfig", metav1.GetOptions{})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return secrets.Data["value"], nil
}

func GetClientFromClusterId(clusterId string) (*kubernetes.Clientset, error) {
	clientset, err := GetClientAdminCluster()
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-tks-kubeconfig", metav1.GetOptions{})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	config_user, err := clientcmd.RESTConfigFromKubeConfig(secrets.Data["value"])
	if err != nil {
		log.Error(err)
		return nil, err
	}
	clientset_user, err := kubernetes.NewForConfig(config_user)
	if err != nil {
		return nil, err
	}

	return clientset_user, nil
}

func GetKubernetesVserionByClusterId(clusterId string) (string, error) {
	clientset, err := GetClientAdminCluster()
	if err != nil {
		return "", err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-tks-kubeconfig", metav1.GetOptions{})
	if err != nil {
		log.Error(err)
		return "", err
	}

	config_user, err := clientcmd.RESTConfigFromKubeConfig(secrets.Data["value"])
	if err != nil {
		log.Error(err)
		return "", err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config_user)
	if err != nil {
		log.Error(err)
		return "", err
	}

	information, err := discoveryClient.ServerVersion()
	if err != nil {
		log.Error("Error while fetching server version information", err)
		return "", err
	}

	return information.GitVersion, nil
}

func GetKubernetesVserion() (string, error) {
	config, err := getAdminConfig()
	if err != nil {
		log.Error("Failed to load kubeconfig")
		return "", err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Error(err)
		return "", err
	}

	information, err := discoveryClient.ServerVersion()
	if err != nil {
		log.Error("Error while fetching server version information", err)
		return "", err
	}

	return information.GitVersion, nil
}

func MergeKubeconfigsWithSingleUser(kubeconfigs []string) (string, error) {
	type kubeConfigType struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Clusters   []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server                   string `yaml:"server"`
				CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster   string `yaml:"cluster"`
				User      string `yaml:"user"`
				Namespace string `yaml:"namespace,omitempty"`
			} `yaml:"context"`
		} `yaml:"contexts"`

		Users []interface{} `yaml:"users,omitempty"`
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	defer encoder.Close()

	encoder.SetIndent(2)

	var config kubeConfigType
	var combindConfig kubeConfigType
	for _, kc := range kubeconfigs {
		err := yaml.Unmarshal([]byte(kc), &config)
		if err != nil {
			return "", err
		}
		combindConfig.APIVersion = config.APIVersion
		combindConfig.Kind = config.Kind
		combindConfig.Clusters = append(combindConfig.Clusters, config.Clusters...)
		combindConfig.Contexts = append(combindConfig.Contexts, config.Contexts...)
		combindConfig.Users = config.Users
	}

	err := encoder.Encode(combindConfig)
	//modContents, err := yaml.Marshal(combindConfig)

	// write the kubeconfig to a file
	err = os.WriteFile("combind-kubeconfig", buf.Bytes(), 0644)

	return buf.String(), err
}
