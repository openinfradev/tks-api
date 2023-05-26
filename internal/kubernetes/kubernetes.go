package kubernetes

import (
	"context"
	"fmt"

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

func GetKubeConfig(clusterId string) ([]byte, error) {
	clientset, err := GetClientAdminCluster()
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-user-kubeconfig", metav1.GetOptions{})
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

	fmt.Printf("%#v\n", information)

	fmt.Println("major", information.Major)
	fmt.Println("minor", information.Minor)
	fmt.Println("platform", information.Platform)

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

	fmt.Printf("%#v\n", information)

	fmt.Println("major", information.Major)
	fmt.Println("minor", information.Minor)
	fmt.Println("platform", information.Platform)

	return information.GitVersion, nil
}
