package kubernetes_test

import (
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"testing"
)

func TestMergeKubeconfigsWithSingleUser(t *testing.T) {
	kubeconfigs := []string{
		`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: test
    server: https://10.10.202.33:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: oidc-user
  name: oidc-user@test-cluster
kind: Config
preferences: {}
users:
- name: oidc-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://tks-console-stg.skbroadband.com/auth/realms/o25pwnjp0
      - --oidc-client-id=c03fk0ox4-k8s-api
      - --grant-type=password
      command: kubectl
      env: null
      interactiveMode: IfAvailable
      provideClusterInfo: false`,

		`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: test
    server: https://10.10.202.33:6443
  name: test-cluster2
contexts:
- context:
    cluster: test-cluster2
    user: oidc-user
  name: oidc-user@test-cluster2
kind: Config
preferences: {}
users:
- name: oidc-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://tks-console-stg.skbroadband.com/auth/realms/o25pwnjp0
      - --oidc-client-id=c03fk0ox4-k8s-api
      - --grant-type=password
      command: kubectl
      env: null
      interactiveMode: IfAvailable
      provideClusterInfo: false `,
	}

	out, err := kubernetes.MergeKubeconfigsWithSingleUser(kubeconfigs)
	if err != nil {
		t.Error(err)
	}
	t.Log(out)
}
