1. 为[这段 Golang 代码](https://gist.github.com/abhishekkr/3beebbc1db54b3b54914#file-tcp_server-go) 写一个多阶段构建的 [Dockerfile](./go/Dockerfile)
```shell
❯ docker build  -t tcp-server . --load
[+] Building 2.3s (15/15) FINISHED                                             docker-container:desktop-linux
 => [internal] load build definition from Dockerfile                                                     0.0s
 => => transferring dockerfile: 638B                                                                     0.0s
 => [internal] load metadata for public.ecr.aws/docker/library/alpine:3.20                               1.9s
 => [internal] load metadata for public.ecr.aws/docker/library/golang:1.22-alpine                        1.9s
 => [internal] load .dockerignore                                                                        0.0s
 => => transferring context: 2B                                                                          0.0s
 => [builder 1/4] FROM public.ecr.aws/docker/library/golang:1.22-alpine@sha256:48eab5e3505d8c8b42a06fe5  0.0s
 => => resolve public.ecr.aws/docker/library/golang:1.22-alpine@sha256:48eab5e3505d8c8b42a06fe5f1cf4c34  0.0s
 => [stage-1 1/4] FROM public.ecr.aws/docker/library/alpine:3.20@sha256:beefdbd8a1da6d2915566fde36db9db  0.0s
 => => resolve public.ecr.aws/docker/library/alpine:3.20@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb7  0.0s
 => [internal] load build context                                                                        0.0s
 => => transferring context: 139B                                                                        0.0s
 => CACHED [stage-1 2/4] RUN apk -U add --no-cache ca-certificates          curl         busybox-extras  0.0s
 => CACHED [stage-1 3/4] WORKDIR /app                                                                    0.0s
 => CACHED [builder 2/4] WORKDIR /src                                                                    0.0s
 => CACHED [builder 3/4] COPY . /src                                                                     0.0s
 => CACHED [builder 4/4] RUN CGO_ENABLED=0 go build -o bin/tcp-server tcp_server.go                      0.0s
 => CACHED [stage-1 4/4] COPY --chown=9097 --from=builder /src/bin /app                                  0.0s
 => exporting to oci image format                                                                        0.4s
 => => exporting layers                                                                                  0.2s
 => => exporting manifest sha256:b882755c05f0db84d8961c19219c43757b80557c057d1ce52f99f9daf4408754        0.0s
 => => exporting config sha256:246044c792d416a323ded34dd2baa81e1958215afe637dd46b8ed098126ccb56          0.0s
 => => sending tarball                                                                                   0.2s
 => importing to docker                                                                                  0.0s

View build details: docker-desktop://dashboard/build/desktop-linux/desktop-linux0/kzv4xmg7a2qw4hx8yiobty7nl

What's next:
    View a summary of image vulnerabilities and recommendations → docker scout quickview
❯ docker run --rm tcp-server
Listening on localhost:3333
```

2. 为 Helm Demo 增加 [Pre-install Hooks](./helm-demo/templates/pre-install-job.yaml)（Job 类型），并打印一段内容。
```shell
❯ helm install vote . -n vote-helm --create-namespace
NAME: vote
LAST DEPLOYED: Thu Sep 12 11:58:33 2024
NAMESPACE: vote-helm
STATUS: deployed
REVISION: 1
TEST SUITE: None

❯ k -n vote-helm get job
NAME                   STATUS     COMPLETIONS   DURATION   AGE
vote-pre-install-job   Complete   1/1           4s         5m19s
```