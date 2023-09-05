package flow_extionsion

import (
	"context"
	"fmt"
	"time"

	"github.com/air-iot/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/air-iot/api-client-go/v4/engine"
	"github.com/air-iot/logger"
)

const (
	wait = 5
)

type Client struct {
	conn             *grpc.ClientConn
	cli              pb.ExtensionServiceClient
	app              App
	extension        Extension
	cleanStream      func()
	cleanHealthCheck func()
}

func (c *Client) Start(app App, extension Extension) *Client {
	c.app = app
	c.extension = extension
	if err := c.connFlow(); err != nil {
		logger.Errorln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctxHealth, cancelHealth := context.WithCancel(context.Background())
	c.cleanStream = func() {
		cancel()
	}
	c.cleanHealthCheck = func() {
		cancelHealth()
	}
	c.healthCheck(ctxHealth)
	c.startSteam(ctx)
	return c
}

func (c *Client) connFlow() error {
	logger.Infof("连接flow: %+v", Cfg.FlowEngine)
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("%s:%d", Cfg.FlowEngine.Host, Cfg.FlowEngine.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc.Dial error: %s", err)
	}
	c.conn = conn
	c.cli = pb.NewExtensionServiceClient(conn)
	return nil
}

func (c *Client) healthCheck(ctx context.Context) {
	logger.Infof("健康检查开始")
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("健康检查结束")
				return
			default:
				logger.Infof("健康检查")
				retry := 3
				state := false
				for retry >= 0 {
					healthRes, err := c.cli.HealthCheck(ctx, &pb.ExtensionHealthCheckRequest{Id: Cfg.Extension.Id})
					if err != nil {
						logger.Errorf("健康检查重试错误,%s", err.Error())
						state = true
						time.Sleep(time.Second * time.Duration(wait))
					} else {
						state = false
						if healthRes.GetStatus() == pb.ExtensionHealthCheckResponse_SERVING {
							if healthRes.Errors != nil && len(healthRes.Errors) > 0 {
								for _, e := range healthRes.Errors {
									logger.Errorf("执行 %s, 错误为%s", e.Code.String(), e.Message)
								}
							}
						}
						break
					}
					retry--
				}
				if state {
					c.cleanStream()
					if c.conn != nil {
						if err := c.conn.Close(); err != nil {
							logger.Errorf("grpc close error: %s", err.Error())
						}
					}
					if err := c.connFlow(); err != nil {
						logger.Errorln(err)
					}
					ctx1, cancel1 := context.WithCancel(context.Background())
					c.startSteam(ctx1)
					c.cleanStream = func() {
						cancel1()
					}
				}
				time.Sleep(time.Second * time.Duration(wait))
			}
		}
	}()
}

func (c *Client) startSteam(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("schema stream break")
				return
			default:
				logger.Infof("schema stream")
				if err := c.Schema(context.Background()); err != nil {
					logger.Errorln(err)
				}
				time.Sleep(time.Second * time.Duration(wait))
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("run stream break")
				return
			default:
				logger.Infof("run stream")
				if err := c.Run(context.Background()); err != nil {
					logger.Errorln(err)
				}
				time.Sleep(time.Second * time.Duration(wait))
			}
		}
	}()
}

func (c *Client) Stop() {
	c.cleanStream()
	c.cleanHealthCheck()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logger.Errorf("grpc close error: %s", err.Error())
		}
	}
}

func (c *Client) Schema(ctx context.Context) error {
	stream, err := c.cli.SchemaStream(GetGrpcContext(ctx, Cfg.Extension.Id, Cfg.Extension.Name))
	if err != nil {
		return fmt.Errorf("schema stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("schema stream close err,%s", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream err, %s", err)
		}
		result, err := c.extension.Schema(c.app)
		gr := &pb.ExtensionResult{
			Request: res.GetRequest(),
		}
		if err != nil {
			gr.Status = false
			gr.Info = err.Error()
		} else {
			gr.Status = true
			gr.Result = []byte(result)
		}
		if err := stream.Send(gr); err != nil {
			logger.Errorf("schema stream 发送错误,%s", err)
		}
	}
}

func (c *Client) Run(ctx context.Context) error {
	stream, err := c.cli.RunStream(GetGrpcContext(ctx, Cfg.Extension.Id, Cfg.Extension.Name))
	if err != nil {
		return fmt.Errorf("run stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("run stream close err,%s", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream err, %s", err)
		}
		result, err := c.extension.Run(c.app, res.GetData())
		gr := &pb.ExtensionResult{
			Request: res.GetRequest(),
		}
		if err != nil {
			gr.Status = false
			gr.Info = err.Error()
		} else {
			gr.Status = true
		}
		if result == nil {
			result = map[string]interface{}{}
		}
		b, _ := json.Marshal(result)
		gr.Result = b
		if err := stream.Send(gr); err != nil {
			logger.Errorf("run stream 发送错误,%s", err)
		}
	}
}
