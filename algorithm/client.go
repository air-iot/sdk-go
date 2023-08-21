package algorithm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/air-iot/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/air-iot/api-client-go/v4/algorithm"
	"github.com/air-iot/logger"
)

type Client struct {
	conn             *grpc.ClientConn
	cli              pb.AlgorithmServiceClient
	app              App
	algorithmService Service
	clean            func()
	cacheConfig      sync.Map
	cacheConfigNum   sync.Map
	//healthTime        time.Time
}

func (c *Client) Start(app App, algorithmService Service) *Client {
	c.app = app
	c.algorithmService = algorithmService
	err := algorithmService.Start(app)
	if err != nil {
		panic(err.Error())
	}
	c.start()
	return c
}

func (c *Client) start() {
	if err := c.connAlgorithm(); err != nil {
		logger.Errorln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.clean = func() {
		cancel()
	}
	c.healthCheck(ctx)
	c.startSteam(ctx)
}

func (c *Client) Stop() {
	if c.clean != nil {
		c.clean()
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logger.Errorf("grpc close error: %s", err.Error())
		}
	}
}

func (c *Client) restart() {
	logger.Infof("重启算法管理连接")
	c.Stop()
	c.start()
}

func (c *Client) connAlgorithm() error {
	logger.Infof("连接算法管理: %+v", C.AlgorithmGrpc)
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("%s:%d", C.AlgorithmGrpc.Host, C.AlgorithmGrpc.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc.Dial error: %s", err)
	}
	cli := pb.NewAlgorithmServiceClient(conn)
	c.conn = conn
	c.cli = cli
	return nil
}

func (c *Client) healthCheck(ctx context.Context) {
	//time.Sleep(3 * time.Second)
	logger.Debugf("健康检查启动")
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("健康检查停止")
				return
			default:
				logger.Tracef("健康检查开始")
				retry := C.AlgorithmGrpc.Health.Retry
				state := false
				for retry >= 0 {
					healthRes, err := c.healthRequest(ctx)
					if err != nil {
						logger.Errorf("健康检查错误,%s", err.Error())
						state = true
						time.Sleep(time.Second * time.Duration(C.AlgorithmGrpc.WaitTime))
					} else {
						state = false
						if healthRes.GetStatus() == pb.HealthCheckResponse_SERVING {
							logger.Tracef("健康检查正常")
							if healthRes.Errors != nil && len(healthRes.Errors) > 0 {
								for _, e := range healthRes.Errors {
									logger.Errorf("执行 %s, 错误为%s", e.Code.String(), e.Message)
								}
							}
						} else if healthRes.GetStatus() == pb.HealthCheckResponse_SERVICE_UNKNOWN {
							logger.Errorf("健康检查异常,服务端未找到本算法服务")
							state = true
						}
						break
					}
					retry--
				}
				if state {
					c.restart()
				}
				time.Sleep(time.Second * time.Duration(C.AlgorithmGrpc.WaitTime))
			}
		}
	}()
}

func (c *Client) healthRequest(ctx context.Context) (*pb.HealthCheckResponse, error) {
	reqCtx, reqCancel := context.WithTimeout(ctx, time.Second*time.Duration(C.AlgorithmGrpc.Health.RequestTime))
	defer reqCancel()
	healthRes, err := c.cli.HealthCheck(reqCtx, &pb.HealthCheckRequest{Service: C.ServiceID})
	return healthRes, err
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
				if err := c.SchemaStream(context.Background()); err != nil {
					logger.Errorln(err)
				}
				time.Sleep(time.Second * time.Duration(C.AlgorithmGrpc.WaitTime))
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
				if err := c.RunStream(context.Background()); err != nil {
					logger.Errorln(err)
				}
				time.Sleep(time.Second * time.Duration(C.AlgorithmGrpc.WaitTime))
			}
		}
	}()
}

func (c *Client) SchemaStream(ctx context.Context) error {
	stream, err := c.cli.SchemaStream(GetGrpcContext(ctx, C.ServiceID, C.Algorithm.ID, C.Algorithm.Name))
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
			return fmt.Errorf("schema stream err, %s", err)
		}
		schema, err := c.algorithmService.Schema(c.app)
		schemaRes := new(grpcResult)
		if err != nil {
			schemaRes.Error = err.Error()
			schemaRes.Code = 400
		} else {
			schemaRes.Result = schema
			schemaRes.Code = 200
		}
		bts, _ := json.Marshal(schemaRes)
		if err := stream.Send(&pb.SchemaResult{
			Request: res.Request,
			Message: bts,
		}); err != nil {
			logger.Errorf("schema stream 发送错误,%s", err)
		}
	}
}

func (c *Client) RunStream(ctx context.Context) error {
	stream, err := c.cli.RunStream(GetGrpcContext(ctx, C.ServiceID, C.Algorithm.ID, C.Algorithm.Name))
	if err != nil {
		return fmt.Errorf("run stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("run stream close err,%s", err)
		}
	}()
	for {
		res0, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("run stream err, %s", err)
		}
		go func(res *pb.RunRequest) {
			gr := new(grpcResult)
			runRes, err := c.algorithmService.Run(c.app, res.Data)
			if err != nil {
				gr.Error = err.Error()
				gr.Code = 400
			} else {
				gr.Result = runRes
				gr.Code = 200
			}
			bts, _ := json.Marshal(gr)
			if err := stream.Send(&pb.RunResult{
				Request: res.Request,
				Message: bts,
			}); err != nil {
				logger.Errorf("run stream 发送错误,%s", err)
			}
		}(res0)
	}
}
