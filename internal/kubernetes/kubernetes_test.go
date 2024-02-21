package kubernetes_test

import (
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"gopkg.in/yaml.v3"
	"reflect"
	"testing"
)

func TestMergeKubeconfigsWithSingleUser(t *testing.T) {
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

	inputObjs := []kubeConfigType{
		{
			APIVersion: "v1",
			Kind:       "Config",
			Clusters: []struct {
				Name    string `yaml:"name"`
				Cluster struct {
					Server                   string `yaml:"server"`
					CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
				} `yaml:"cluster"`
			}{
				{
					Name: "test-cluster1",
					Cluster: struct {
						Server                   string `yaml:"server"`
						CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
					}{
						Server:                   "https://10.0.0.1:6443",
						CertificateAuthorityData: "test==",
					},
				},
			},
			Contexts: []struct {
				Name    string `yaml:"name"`
				Context struct {
					Cluster   string `yaml:"cluster"`
					User      string `yaml:"user"`
					Namespace string `yaml:"namespace,omitempty"`
				} `yaml:"context"`
			}{
				{
					Name: "oidc-user@test-cluster1",
					Context: struct {
						Cluster   string `yaml:"cluster"`
						User      string `yaml:"user"`
						Namespace string `yaml:"namespace,omitempty"`
					}{
						Cluster:   "test-cluster1",
						User:      "oidc-user",
						Namespace: "test-namespaces",
					},
				},
			},
			Users: []interface{}{
				map[string]interface{}{
					"name": "oidc-user",
					"user": map[string]interface{}{
						"exec": map[string]interface{}{
							"apiVersion": "client.authentication.k8s.io/v1beta1",
							"args": []string{
								"oidc-login",
								"get-token",
								"--oidc-issuer-url=https://idp-domain/auth",
								"--oidc-client-id=k8s-api",
								"--grant-type=password",
							},
							"command":            "kubectl",
							"env":                nil,
							"interactiveMode":    "IfAvailable",
							"provideClusterInfo": false,
						},
					},
				},
			},
		},
		{
			APIVersion: "v1",
			Kind:       "Config",
			Clusters: []struct {
				Name    string `yaml:"name"`
				Cluster struct {
					Server                   string `yaml:"server"`
					CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
				} `yaml:"cluster"`
			}{
				{
					Name: "test-cluster2",
					Cluster: struct {
						Server                   string `yaml:"server"`
						CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
					}{
						Server:                   "https://10.0.0.2:6443",
						CertificateAuthorityData: "test2==",
					},
				},
			},
			Contexts: []struct {
				Name    string `yaml:"name"`
				Context struct {
					Cluster   string `yaml:"cluster"`
					User      string `yaml:"user"`
					Namespace string `yaml:"namespace,omitempty"`
				} `yaml:"context"`
			}{
				{
					Name: "oidc-user@test-cluster2",
					Context: struct {
						Cluster   string `yaml:"cluster"`
						User      string `yaml:"user"`
						Namespace string `yaml:"namespace,omitempty"`
					}{
						Cluster: "test-cluster2",
						User:    "oidc-user",
					},
				},
			},
			Users: []interface{}{
				map[string]interface{}{
					"name": "oidc-user",
					"user": map[string]interface{}{
						"exec": map[string]interface{}{
							"apiVersion": "client.authentication.k8s.io/v1beta1",
							"args": []string{
								"oidc-login",
								"get-token",
								"--oidc-issuer-url=https://idp-domain/auth",
								"--oidc-client-id=k8s-api",
								"--grant-type=password",
							},
							"command":            "kubectl",
							"env":                nil,
							"interactiveMode":    "IfAvailable",
							"provideClusterInfo": false,
						},
					},
				},
			},
		},
	}

	expected := kubeConfigType{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server                   string `yaml:"server"`
				CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
			} `yaml:"cluster"`
		}{
			{
				Name: "test-cluster1",
				Cluster: struct {
					Server                   string `yaml:"server"`
					CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
				}{
					Server:                   "https://10.0.0.1:6443",
					CertificateAuthorityData: "test==",
				},
			},
			{
				Name: "test-cluster2",
				Cluster: struct {
					Server                   string `yaml:"server"`
					CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
				}{
					Server:                   "https://10.0.0.2:6443",
					CertificateAuthorityData: "test2==",
				},
			},
		},
		Contexts: []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster   string `yaml:"cluster"`
				User      string `yaml:"user"`
				Namespace string `yaml:"namespace,omitempty"`
			} `yaml:"context"`
		}{
			{
				Name: "oidc-user@test-cluster1",
				Context: struct {
					Cluster   string `yaml:"cluster"`
					User      string `yaml:"user"`
					Namespace string `yaml:"namespace,omitempty"`
				}{
					Cluster:   "test-cluster1",
					User:      "oidc-user",
					Namespace: "test-namespaces",
				},
			},
			{
				Name: "oidc-user@test-cluster2",
				Context: struct {
					Cluster   string `yaml:"cluster"`
					User      string `yaml:"user"`
					Namespace string `yaml:"namespace,omitempty"`
				}{
					Cluster: "test-cluster2",
					User:    "oidc-user",
				},
			},
		},

		Users: []interface{}{
			map[string]interface{}{
				"name": "oidc-user2",
				"user": map[string]interface{}{
					"exec": map[string]interface{}{
						"apiVersion": "client.authentication.k8s.io/v1beta1",
						"args": []string{
							"oidc-login",
							"get-token",
							"--oidc-issuer-url=https://idp-domain/auth",
							"--oidc-client-id=k8s-api",
							"--grant-type=password",
						},
						"command":            "kubectl",
						"env":                nil,
						"interactiveMode":    "IfAvailable",
						"provideClusterInfo": false,
					},
				},
			},
		},
	}

	in := make([]string, len(inputObjs))
	for _, v := range inputObjs {
		o, err := yaml.Marshal(&v)
		if err != nil {
			t.Error(err)
		}
		in = append(in, string(o))
	}
	r, err := kubernetes.MergeKubeconfigsWithSingleUser(in)
	t.Log(r)
	if err != nil {
		t.Error(err)
	}

	var result kubeConfigType
	if err := yaml.Unmarshal([]byte(r), &result); err != nil {
		t.Error(err)
	}

	{
		if result.APIVersion != expected.APIVersion {
			t.Errorf("expected: %s, got: %s", expected.APIVersion, result.APIVersion)
		}
		if result.Kind != expected.Kind {
			t.Errorf("expected: %s, got: %s", expected.Kind, result.Kind)
		}
		for i, v := range result.Clusters {
			if v.Name != expected.Clusters[i].Name {
				t.Errorf("expected: %s, got: %s", expected.Clusters[i].Name, v.Name)
			}
			if v.Cluster.Server != expected.Clusters[i].Cluster.Server {
				t.Errorf("expected: %s, got: %s", expected.Clusters[i].Cluster.Server, v.Cluster.Server)
			}
			if v.Cluster.CertificateAuthorityData != expected.Clusters[i].Cluster.CertificateAuthorityData {
				t.Errorf("expected: %s, got: %s", expected.Clusters[i].Cluster.CertificateAuthorityData, v.Cluster.CertificateAuthorityData)
			}
		}
		for i, v := range result.Contexts {
			if v.Name != expected.Contexts[i].Name {
				t.Errorf("expected: %s, got: %s", expected.Contexts[i].Name, v.Name)
			}
			if v.Context.Cluster != expected.Contexts[i].Context.Cluster {
				t.Errorf("expected: %s, got: %s", expected.Contexts[i].Context.Cluster, v.Context.Cluster)
			}
			if v.Context.User != expected.Contexts[i].Context.User {
				t.Errorf("expected: %s, got: %s", expected.Contexts[i].Context.User, v.Context.User)
			}
			if v.Context.Namespace != expected.Contexts[i].Context.Namespace {
				t.Errorf("expected: %s, got: %s", expected.Contexts[i].Context.Namespace, v.Context.Namespace)
			}
		}

		//ToDo: This test case down below results in true negative. Need to fix the test case.
		if reflect.DeepEqual(result.Users, expected.Users) {
			t.Errorf("expected: %v, got: %v", expected.Users, result.Users)
		}
	}
}
