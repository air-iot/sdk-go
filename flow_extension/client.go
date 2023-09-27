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
				newLogger := logger.WithContext(logger.NewModuleContext(context.Background(), MODULE_HEALTHCHECK))
				newLogger.Infof("健康检查开始")
				retry := 3
				state := false
				for retry >= 0 {
					healthRes, err := c.cli.HealthCheck(ctx, &pb.ExtensionHealthCheckRequest{Id: Cfg.Extension.Id})
					if err != nil {
						newLogger.Errorf("健康检查重试错误,%v", err)
						state = true
						time.Sleep(time.Second * time.Duration(wait))
					} else {
						state = false
						if healthRes.GetStatus() == pb.ExtensionHealthCheckResponse_SERVING {
							if healthRes.Errors != nil && len(healthRes.Errors) > 0 {
								for _, e := range healthRes.Errors {
									newLogger.Errorf("执行 %s, 错误为 %s", e.Code.String(), e.Message)
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
							newLogger.Errorf("grpc close error: %v", err)
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
				ctx1 := context.Background()
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, MODULE_SCHEMA))
				newLogger.Infof("启动schema stream")
				if err := c.Schema(context.Background()); err != nil {
					newLogger.Errorf("schema stream错误,%v", err)
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
				ctx1 := context.Background()
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, MODULE_RUN))
				newLogger.Infof("启动run stream")
				if err := c.Run(context.Background()); err != nil {
					newLogger.Errorf("run stream错误,%v", err)
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
		go func(res *pb.ExtensionSchemaRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Extension.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, MODULE_SCHEMA)
			result, err := c.extension.Schema(ctx1, c.app)
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
				logger.WithContext(ctx1).Errorf("配置(schema)返回到流程扩展节点错误,%v", err)
			}
		}(res)
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
		go func(res *pb.ExtensionRunRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Extension.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, MODULE_RUN)
			result, err := c.extension.Run(ctx1, c.app, res.GetData())
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
				logger.WithContext(ctx1).Errorf("执行(run)结果返回到流程扩展节点错误,%v", err)
			}
		}(res)
	}
}
