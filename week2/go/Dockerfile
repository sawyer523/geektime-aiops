FROM --platform=$TARGETPLATFORM public.ecr.aws/docker/library/golang:1.22-alpine AS builder

WORKDIR /src

COPY . /src

RUN CGO_ENABLED=0 go build -o bin/tcp-server tcp_server.go

FROM --platform=$TARGETPLATFORM public.ecr.aws/docker/library/alpine:3.20

RUN apk -U add --no-cache ca-certificates  \
        curl \
        busybox-extras \
        busybox && \
        addgroup -S 9097 && \
        adduser -S 9097 -G 9097 -u 9097

USER 9097
WORKDIR /app
COPY --chown=9097 --from=builder /src/bin /app

VOLUME /data/conf

EXPOSE 3333
ENTRYPOINT ["./tcp-server"]
