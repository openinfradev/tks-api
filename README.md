# TKS-API backend

### Run

```
$ go build -o server cmd/server/main.go
$ ./server
```

### Generate swagger files

```
$ make docs
```

### Dev run

```
# swagger build & build & run
$ make dev_run
```

### Configuration kubernetes config

kubernetes client 설정은 2가지 방식이 가능하다.
아래와 같이 kubeconfig path 를 parameter 로 넘기는 방식 ( 주로 local debugging 목적 )

```
$ ./sever --kubeconfig /kube/config ...
```

incluster 의 serviceAccount 를 활용한 방식 ( recommended ).
단 이 방식을 사용하기 위해서는 반드시 클러스터에서 tks::default serviceAccount에 대한 clusterrolebinding 설정이 필요하다.

```
in-cluster. 기본 view 권한 설정 ( 여기에 nodes, secrets 등은 빠져있음 )
$ kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:tks
```

아래는 node 정보를 가져오기 위한 추가 rbac 설정

```
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: tks-api-test
  namespace: tks
rules:
# Just an example, feel free to change it
- apiGroups: [""]
  resources: ["nodes", "secrets","services","pods","namespaces","events"]
  verbs: ["get", "watch", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tks-api-test
  namespace: tks
subjects:
- kind: ServiceAccount
  name: default
  namespace: tks
roleRef:
  kind: ClusterRole
  name: tks-api-test
  apiGroup: rbac.authorization.k8s.io
```
