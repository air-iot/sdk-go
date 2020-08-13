FROM alpine:latest

ADD main /app/
ADD config.yaml /app/etc/config.yaml
WORKDIR /app
EXPOSE 9000
ENTRYPOINT ["./main"]