package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/spf13/viper"

	rbacV1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clientcmd "k8s.io/client-go/tools/clientcmd"

	"github.com/openinfradev/tks-api/pkg/log"
)

func getAdminConfig(ctx context.Context) (*rest.Config, error) {
	kubeconfigPath := viper.GetString("kubeconfig-path")
	if kubeconfigPath == "" {
		log.Info(ctx, "Use in-cluster config")
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Error(ctx, "Failed to load incluster kubeconfig")
			return nil, err
		}
		return config, nil
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", viper.GetString("kubeconfig-path"))
		if err != nil {
			log.Error(ctx, "Failed to local kubeconfig")
			return nil, err
		}
		return config, nil
	}
}

func GetClientAdminCluster(ctx context.Context) (*kubernetes.Clientset, error) {
	config, err := getAdminConfig(ctx)
	if err != nil {
		log.Error(ctx, "Failed to load kubeconfig")
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// 쿠버네티스 기본 타입 이외의 타입(예: 정책 템플릿, 정책 등)을 처리하기 위한 dynamic client 생성
func GetDynamicClientAdminCluster(ctx context.Context) (*dynamic.DynamicClient, error) {
	config, err := getAdminConfig(ctx)
	if err != nil {
		log.Error(ctx, "Failed to load kubeconfig")
		return nil, err
	}

	// create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return dynamicClient, nil
}

func GetAwsSecret(ctx context.Context) (awsAccessKeyId string, awsSecretAccessKey string, err error) {
	clientset, err := GetClientAdminCluster(ctx)
	if err != nil {
		return "", "", err
	}

	secrets, err := clientset.CoreV1().Secrets("argo").Get(context.TODO(), "awsconfig-secret", metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, err)
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

func GetAwsAccountIdSecret(ctx context.Context) (awsAccountId string, err error) {
	clientset, err := GetClientAdminCluster(ctx)
	if err != nil {
		return "", err
	}

	secrets, err := clientset.CoreV1().Secrets("argo").Get(context.TODO(), "tks-aws-user", metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}

	awsAccountId = string(secrets.Data["account_id"][:])
	return
}

func GetKubeConfig(ctx context.Context, clusterId string) ([]byte, error) {
	clientset, err := GetClientAdminCluster(ctx)
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-tks-user-kubeconfig", metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, err)
		return nil, err
	}

	return secrets.Data["value"], nil
}

func GetClientFromClusterId(ctx context.Context, clusterId string) (*kubernetes.Clientset, error) {
	clientset, err := GetClientAdminCluster(ctx)
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-tks-kubeconfig", metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, err)
		return nil, err
	}

	config_user, err := clientcmd.RESTConfigFromKubeConfig(secrets.Data["value"])
	if err != nil {
		log.Error(ctx, err)
		return nil, err
	}
	clientset_user, err := kubernetes.NewForConfig(config_user)
	if err != nil {
		return nil, err
	}

	return clientset_user, nil
}

func GetKubernetesVserionByClusterId(ctx context.Context, clusterId string) (string, error) {
	clientset, err := GetClientAdminCluster(ctx)
	if err != nil {
		return "", err
	}

	secrets, err := clientset.CoreV1().Secrets(clusterId).Get(context.TODO(), clusterId+"-tks-kubeconfig", metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}

	config_user, err := clientcmd.RESTConfigFromKubeConfig(secrets.Data["value"])
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config_user)
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}

	information, err := discoveryClient.ServerVersion()
	if err != nil {
		log.Error(ctx, "Error while fetching server version information", err)
		return "", err
	}

	return information.GitVersion, nil
}

func GetKubernetesVserion(ctx context.Context) (string, error) {
	config, err := getAdminConfig(ctx)
	if err != nil {
		log.Error(ctx, "Failed to load kubeconfig")
		return "", err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}

	information, err := discoveryClient.ServerVersion()
	if err != nil {
		log.Error(ctx, "Error while fetching server version information", err)
		return "", err
	}

	return information.GitVersion, nil
}

func GetResourceApiVersion(ctx context.Context, kubeconfig []byte, kind string) (string, error) {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	apiResourceList, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}

	for _, apiResource := range apiResourceList {
		for _, resource := range apiResource.APIResources {
			if resource.Kind == kind {
				return resource.Version, nil
			}
		}
	}

	return "", nil
}

func EnsureClusterRole(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	// generate clusterrole object
	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		obj := getClusterRole(role, projectName+"-"+role)

		if _, err := clientset.RbacV1().ClusterRoles().Get(context.Background(), projectName+"-"+role, metav1.GetOptions{}); err != nil {
			_, err = clientset.RbacV1().ClusterRoles().Create(context.Background(), obj, metav1.CreateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		} else {
			_, err = clientset.RbacV1().ClusterRoles().Update(context.Background(), obj, metav1.UpdateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		}
	}

	return nil
}

func RemoveClusterRole(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	// remove clusterrole object
	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		if err := clientset.RbacV1().ClusterRoles().Delete(context.Background(), projectName+"-"+role, metav1.DeleteOptions{}); err != nil {
			log.Error(ctx, err)
		}
	}

	return nil
}

func EnsureClusterRoleBinding(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		obj := generateClusterRoleToClusterRoleBinding(role+"@"+projectName, projectName+"-"+role, projectName+"-"+role)
		if _, err = clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), projectName+"-"+role, metav1.GetOptions{}); err != nil {
			_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), obj, metav1.CreateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		} else {
			_, err = clientset.RbacV1().ClusterRoleBindings().Update(context.Background(), obj, metav1.UpdateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		}
	}

	return nil
}

func RemoveClusterRoleBinding(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		if err := clientset.RbacV1().ClusterRoleBindings().Delete(context.Background(), projectName+"-"+role, metav1.DeleteOptions{}); err != nil {
			log.Error(ctx, err)
		}
	}

	return nil
}

func EnsureRoleBinding(ctx context.Context, kubeconfig []byte, projectName string, namespace string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		obj := generateClusterRoleToRoleBinding(role+"@"+projectName, projectName+"-"+role, namespace, projectName+"-"+role)
		if _, err = clientset.RbacV1().RoleBindings(namespace).Get(context.Background(), projectName+"-"+role, metav1.GetOptions{}); err != nil {
			_, err = clientset.RbacV1().RoleBindings(namespace).Create(context.Background(), obj, metav1.CreateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		} else {
			_, err = clientset.RbacV1().RoleBindings(namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		}
	}

	return nil
}

func RemoveRoleBinding(ctx context.Context, kubeconfig []byte, projectName string, namespace string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		if err := clientset.RbacV1().RoleBindings(namespace).Delete(context.Background(), projectName+"-"+role, metav1.DeleteOptions{}); err != nil {
			log.Error(ctx, err)
		}
	}

	return nil
}

const (
	leaderRole = "leader"
	memberRole = "member"
	viewerRole = "viewer"
)

func getClusterRole(role, objName string) *rbacV1.ClusterRole {

	clusterRole := rbacV1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: objName,
		},
		Rules: []rbacV1.PolicyRule{
			{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
			},
		},
	}

	switch role {
	case leaderRole:
		clusterRole.Rules[0].Verbs = []string{"*"}
	case memberRole:
		clusterRole.Rules[0].Verbs = []string{"*"}
	case viewerRole:
		clusterRole.Rules[0].Verbs = []string{"get", "list", "watch"}
	default:
		return nil
	}

	return &clusterRole
}

func EnsureCommonClusterRole(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	obj := generateCommonClusterRole(projectName + "-common")
	if _, err = clientset.RbacV1().ClusterRoles().Get(context.Background(), projectName+"-common", metav1.GetOptions{}); err != nil {
		_, err = clientset.RbacV1().ClusterRoles().Create(context.Background(), obj, metav1.CreateOptions{})
		if err != nil {
			log.Error(ctx, err)
			return err
		}
	} else {
		_, err = clientset.RbacV1().ClusterRoles().Update(context.Background(), obj, metav1.UpdateOptions{})
		if err != nil {
			log.Error(ctx, err)
			return err
		}
	}

	return nil
}

func RemoveCommonClusterRole(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	if err := clientset.RbacV1().ClusterRoles().Delete(context.Background(), projectName+"-common", metav1.DeleteOptions{}); err != nil {
		log.Error(ctx, err)
		return err
	}

	return nil
}

func EnsureCommonClusterRoleBinding(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		obj := generateClusterRoleToClusterRoleBinding(role+"@"+projectName, projectName+"-common-"+role, projectName+"-common")
		if _, err = clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), projectName+"-common"+"-"+role, metav1.GetOptions{}); err != nil {
			_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), obj, metav1.CreateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		} else {
			_, err = clientset.RbacV1().ClusterRoleBindings().Update(context.Background(), obj, metav1.UpdateOptions{})
			if err != nil {
				log.Error(ctx, err)
				return err
			}
		}
	}

	return nil
}

func RemoveCommonClusterRoleBinding(ctx context.Context, kubeconfig []byte, projectName string) error {
	config_user, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(ctx, err)
		return err
	}

	clientset := kubernetes.NewForConfigOrDie(config_user)

	for _, role := range []string{leaderRole, memberRole, viewerRole} {
		if err := clientset.RbacV1().ClusterRoles().Delete(context.Background(), projectName+"-common-"+role, metav1.DeleteOptions{}); err != nil {
			log.Error(ctx, err)
			return err
		}
	}

	return nil
}
func generateCommonClusterRole(objName string) *rbacV1.ClusterRole {
	clusterRole := rbacV1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: objName,
		},
		Rules: []rbacV1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{""},
				Resources: []string{"namespaces", "nodes", "storageclasses", "persistentvolumes"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{"certificates.k8s.io"},
				Resources: []string{"certificatesigningrequests"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"clusterroles", "clusterrolebindings"},
			},
		},
	}

	return &clusterRole
}

func generateClusterRoleToClusterRoleBinding(groupName, objName, roleName string) *rbacV1.ClusterRoleBinding {
	clusterRoleBinding := rbacV1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: objName,
		},
		Subjects: []rbacV1.Subject{
			{
				Kind: "Group",
				Name: groupName,
			},
		},
		RoleRef: rbacV1.RoleRef{
			Kind: "ClusterRole",
			Name: roleName,
		},
	}

	return &clusterRoleBinding
}

func generateClusterRoleToRoleBinding(groupName, objName, roleName, namespace string) *rbacV1.RoleBinding {
	roleBinding := rbacV1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: namespace,
		},
		Subjects: []rbacV1.Subject{
			{
				Kind: "Group",
				Name: groupName,
			},
		},
		RoleRef: rbacV1.RoleRef{
			Kind: "ClusterRole",
			Name: roleName,
		},
	}

	return &roleBinding
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
	if err != nil {
		return "", err
	}
	//modContents, err := yaml.Marshal(combindConfig)

	// write the kubeconfig to a file
	err = os.WriteFile("combind-kubeconfig", buf.Bytes(), 0644)

	return buf.String(), err
}
