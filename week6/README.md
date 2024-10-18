# 作业
1. [完善 chatgpt.go，实现 deleteResource 方法，能其能以对话的方式删除 K8s 资源。](./cmd/chatgpt.go)
```shell
❯ k get po
NAME                     READY   STATUS    RESTARTS   AGE
nginx-676b6c5bbc-7gdwx   1/1     Running   0          6d6h


❯ ./k8scopilot ask chatgpt
我是 K8s Copilot，有什么可以帮助你：
> 帮我删除 default ns 下的 deploy nginx
你确定要删除 default/Deployment/nginx? (y/N): y
Deleted deployment: nginx

❯ k get po
No resources found in default namespace.
```