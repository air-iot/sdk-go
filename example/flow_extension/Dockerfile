FROM alpine:latest

ADD go-flow-ext-demo /app/
ADD etc/config.docker.yaml /app/etc/config.yaml
WORKDIR /app
ENTRYPOINT ["./go-flow-ext-demo"]