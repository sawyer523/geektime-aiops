#!/bin/bash
# 更新包列表
apt-get update

# 安装必要的依赖
apt-get install -y apt-transport-https ca-certificates curl software-properties-common

# 添加 Docker 的官方 GPG 密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -

# 设置 Docker 的稳定版仓库
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# 再次更新包列表
apt-get update

# 安装 Docker
apt-get install -y docker-ce

# 启动 Docker 服务
systemctl start docker

# 设置 Docker 开机自启
systemctl enable docker

# 输出 Docker 版本，以验证安装
docker --version