FROM alpine:latest

ADD main /app/
ADD config.yaml /app/etc/config.yaml
WORKDIR /app
ENTRYPOINT ["./main"]