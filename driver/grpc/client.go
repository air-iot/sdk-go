package grpc

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/metadata"
)

type Config struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

func GetGrpcContext(ctx context.Context, serviceId, projectId, driverId, driverName string) context.Context {
	md := metadata.New(map[string]string{
		"serviceId":  hex.EncodeToString([]byte(serviceId)),
		"projectId":  hex.EncodeToString([]byte(projectId)),
		"driverId":   hex.EncodeToString([]byte(driverId)),
		"driverName": hex.EncodeToString([]byte(driverName))})
	// 发送 metadata
	// 创建带有meta的context
	return metadata.NewOutgoingContext(ctx, md)
}
