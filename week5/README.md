# 作业

1. [使用 Informer + RateLimitingQueue 监听 Pod 事件](./listener/main.go)
```shell
go run main.go
Add/Delete for Pod cert-manager-7b9875fbcc-mz666, Namespace: cert-manager
Add/Delete for Pod cert-manager-cainjector-948d47c6-xkq8n, Namespace: cert-manager
Add/Delete for Pod cert-manager-webhook-78bd84d46b-98vbd, Namespace: cert-manager
Add/Delete for Pod nginx-676b6c5bbc-7gdwx, Namespace: default
Add/Delete for Pod cilium-7srnh, Namespace: kube-system
Add/Delete for Pod cilium-envoy-26jjd, Namespace: kube-system
Add/Delete for Pod cilium-envoy-m6gf5, Namespace: kube-system
Add/Delete for Pod cilium-envoy-rtlkv, Namespace: kube-system
Add/Delete for Pod cilium-operator-8ffbc64b8-fkwnb, Namespace: kube-system
Add/Delete for Pod cilium-s2zms, Namespace: kube-system
Add/Delete for Pod cilium-wh57c, Namespace: kube-system
Add/Delete for Pod coredns-7c65d6cfc9-hwhct, Namespace: kube-system
Add/Delete for Pod coredns-7c65d6cfc9-mrwl9, Namespace: kube-system
Add/Delete for Pod etcd-k8s-master, Namespace: kube-system
Add/Delete for Pod kube-apiserver-k8s-master, Namespace: kube-system
Add/Delete for Pod kube-controller-manager-k8s-master, Namespace: kube-system
Add/Delete for Pod kube-proxy-psg2h, Namespace: kube-system
Add/Delete for Pod kube-proxy-xgzfh, Namespace: kube-system
Add/Delete for Pod kube-proxy-zn69p, Namespace: kube-system
Add/Delete for Pod kube-scheduler-k8s-master, Namespace: kube-system
Add/Delete for Pod metrics-server-587b667b55-vfmhx, Namespace: kube-system
^Csignal: interrupt

```

2. [建一个新的自定义 CRD（Group：aiops.geektime.com, Version: v1alpha1, Kind: AIOps），并使用 dynamicClient 获取该资源](./crd/main.go)

```shell
❯ go run main.go get AIOps
Name: my-resource-instance, Namespace: default, UID: 13111cbe-1002-4faa-9ad8-0769e56e6df7

```