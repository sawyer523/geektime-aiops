# 作业
1. 尝试增强实战二，增加 configmap 字段，实现一并生成 ConfigMap。
```shell
❯ make manifests
/aiops/homework/week7/cronhpa/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
Downloading sigs.k8s.io/kustomize/kustomize/v5@v5.4.3
go: downloading sigs.k8s.io/kustomize/kustomize/v5 v5.4.3
go: downloading sigs.k8s.io/kustomize/kyaml v0.17.2
go: downloading sigs.k8s.io/kustomize/api v0.17.3
go: downloading sigs.k8s.io/kustomize/cmd/config v0.14.2
go: downloading k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00
/aiops/homework/week7/cronhpa/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/cronhpas.autoscaling.aiops.com created
```
```shell
❯ k create deployment nginx --image nginx
deployment.apps/nginx created
```
```shell
❯ make run
/aiops/homework/week7/cronhpa/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/aiops/homework/week7/cronhpa/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./cmd/main.go
2024-10-22T10:55:20+08:00       INFO    setup   starting manager
2024-10-22T10:55:20+08:00       INFO    starting server {"name": "health probe", "addr": "[::]:8081"}
2024-10-22T10:55:20+08:00       INFO    Starting EventSource    {"controller": "cronhpa", "controllerGroup": "autoscaling.aiops.com", "controllerKind": "CronHPA", "source": "kind source: *v1.CronHPA"}
2024-10-22T10:55:20+08:00       INFO    Starting Controller     {"controller": "cronhpa", "controllerGroup": "autoscaling.aiops.com", "controllerKind": "CronHPA"}
2024-10-22T10:55:20+08:00       INFO    Starting workers        {"controller": "cronhpa", "controllerGroup": "autoscaling.aiops.com", "controllerKind": "CronHPA", "worker count": 1}

```
```shell
❯ k get deploy
NAME    READY   UP-TO-DATE   AVAILABLE   AGE
nginx   1/1     1            1           2m48s

❯ k get cm
NAME               DATA   AGE
kube-root-ca.crt   1      37d
```
```shell
❯ k apply -f config/samples/autoscaling_v1_cronhpa.yaml
cronhpa.autoscaling.aiops.com/nginx created
❯ k get po
NAME                     READY   STATUS              RESTARTS   AGE
nginx-676b6c5bbc-f84xz   0/1     ContainerCreating   0          8s
nginx-676b6c5bbc-jzfs8   0/1     ContainerCreating   0          8s
nginx-676b6c5bbc-nztwj   1/1     Running             0          9m35s

❯ k get cm
NAME               DATA   AGE
kube-root-ca.crt   1      37d
nginx              1      33s

```


2. 尝试增强实战一，增加 configmap 字段，实现一并生成 ConfigMap。

```shell
❯ make manifests
/aiops/homework/week7/application/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

❯ make install
/aiops/homework/week7/application/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
Downloading sigs.k8s.io/kustomize/kustomize/v5@v5.4.3
/aiops/homework/week7/application/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/applicaitons.application.aiops.com created

 make run
/aiops/homework/week7/application/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/aiops/homework/week7/application/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./cmd/main.go
2024-10-22T15:28:00+08:00       INFO    setup   starting manager
2024-10-22T15:28:00+08:00       INFO    starting server {"name": "health probe", "addr": "[::]:8081"}
2024-10-22T15:28:00+08:00       INFO    Starting EventSource    {"controller": "applicaiton", "controllerGroup": "application.aiops.com", "controllerKind": "Applicaiton", "source": "kind source: *v1.Applicaiton"}
2024-10-22T15:28:00+08:00       INFO    Starting Controller     {"controller": "applicaiton", "controllerGroup": "application.aiops.com", "controllerKind": "Applicaiton"}
2024-10-22T15:28:00+08:00       INFO    Starting workers        {"controller": "applicaiton", "controllerGroup": "application.aiops.com", "controllerKind": "Applicaiton", "worker count": 1}

❯ k get po
No resources found in default namespace.
❯ k get cm
NAME               DATA   AGE
kube-root-ca.crt   1      37d


❯ k apply -f config/samples/application_v1_applicaiton.yaml
applicaiton.application.aiops.com/nginx created

❯ k get po
NAME                     READY   STATUS    RESTARTS   AGE
nginx-6d4785d688-5vlzn   1/1     Running   0          11s

❯ k get svc
NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP   37d
nginx        ClusterIP   10.101.193.222   <none>        80/TCP    20s

❯ k get ing
NAME    CLASS   HOSTS                   ADDRESS   PORTS   AGE
nginx   nginx   application.aiops.com             80      31s

❯ k get cm nginx -o yaml
apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: "2024-10-22T07:31:23Z"
  name: nginx
  namespace: default
  ownerReferences:
  - apiVersion: application.aiops.com/v1
    kind: Applicaiton
    name: nginx
    uid: 50542eaa-1a2f-47a8-8697-999878f74f21
  resourceVersion: "949971"
  uid: 42309a33-86e9-41ab-a9c5-ae3f682e7be2


```