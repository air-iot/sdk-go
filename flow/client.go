package flow

import (
	"context"
	"fmt"
	"github.com/air-iot/json"
	"time"

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
	cli              pb.PluginServiceClient
	app              App
	flow             Flow
	cleanStream      func()
	cleanHealthCheck func()
}

func (c *Client) Start(app App, flow Flow) *Client {
	c.app = app
	c.flow = flow
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
	cli := pb.NewPluginServiceClient(conn)
	c.conn = conn
	c.cli = cli
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
					healthRes, err := c.cli.HealthCheck(ctx, &pb.HealthCheckRequest{Name: Cfg.Flow.Name})
					if err != nil {
						newLogger.Errorf("健康检查重试错误,%v", err)
						state = true
						time.Sleep(time.Second * time.Duration(wait))
					} else {
						state = false
						if healthRes.GetStatus() == pb.HealthCheckResponse_SERVING {
							if healthRes.Errors != nil && len(healthRes.Errors) > 0 {
								for _, e := range healthRes.Errors {
									newLogger.Errorf("执行 %s, 错误为%s", e.Code.String(), e.Message)
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
							newLogger.Errorf("grpc close error: %s", err.Error())
						}
					}
					if err := c.connFlow(); err != nil {
						newLogger.Errorln(err)
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
				logger.Infof("stream break")
				return
			default:
				ctx1 := context.Background()
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, MODULE_HANDLER))
				newLogger.Infof("启动start stream")
				if err := c.Handler(ctx1); err != nil {
					newLogger.Errorf("start stream错误,%v", err)
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
			logger.Errorf("grpc close error: %v", err)
		}
	}
}

func (c *Client) Handler(ctx context.Context) error {
	stream, err := c.cli.Register(GetGrpcContext(ctx, Cfg.Flow.Name, Cfg.Flow.Mode))
	if err != nil {
		return fmt.Errorf("stream err,%w", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("stream close err,%v", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream err, %s", err)
		}
		go func(res *pb.FlowRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Flow.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, MODULE_HANDLER)
			result, err := c.flow.Handler(ctx1, c.app, &Request{
				ProjectId:  res.ProjectId,
				FlowId:     res.FlowId,
				Job:        res.Job,
				ElementId:  res.ElementId,
				ElementJob: res.ElementJob,
				Config:     res.Config,
			})
			gr := &pb.FlowResponse{
				ElementJob: res.ElementJob,
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
				logger.WithContext(ctx1).Errorf("执行(handler)结果返回到流程引擎错误,%v", err)
			}
		}(res)
	}
}
