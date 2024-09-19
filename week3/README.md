# 作业
## [实现 Function Calling](./main.py)
* 定义 modify_config 函数，入参：service_name，key，value
* 定义 restart_service 函数，入参：service_name
* 定义 apply_manifest 函数，入参：resource_type，image

## 实践 Function Calling，观察以下输入是否能正确选择对应的函数
* 帮我修改 gateway 的配置，vendor 修改为 alipay
* 帮我重启 gateway 服务
* 帮我部署一个 deployment，镜像是 nginx

```shell
❯ python3 main.py
输入指令：帮我修改 gateway 的配置，vendor 修改为 alipay
调用 modify_config 函数
gateway vendor alipay
函数返回结果 None
None
❯ python3 main.py
输入指令：帮我重启 gateway 服务
调用 restart_service 函数
gateway
函数返回结果 None
None
❯ python3 main.py
输入指令：帮我部署一个 deployment，镜像是 nginx
调用 apply_manifest 函数
deployment nginx
函数返回结果 None
None
```