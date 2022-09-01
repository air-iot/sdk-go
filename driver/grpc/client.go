package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type Config struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

func GetGrpcContext(ctx context.Context, serviceId, projectId string) context.Context {
	md := metadata.New(map[string]string{"serviceId": serviceId, "projectId": projectId})
	// 发送 metadata
	// 创建带有meta的context
	return metadata.NewOutgoingContext(ctx, md)
}