package driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/air-iot/json"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/air-iot/api-client-go/v4/driver"
	"github.com/air-iot/logger"
	dGrpc "github.com/air-iot/sdk-go/v4/driver/grpc"
)

const (
	wait = 5
)

type Client struct {
	conn             *grpc.ClientConn
	cli              pb.DriverServiceClient
	app              App
	driver           Driver
	cleanStream      func()
	cleanHealthCheck func()
	cacheConfig      sync.Map
	cacheConfigNum   sync.Map
}

func (c *Client) Start(app App, driver Driver) *Client {
	c.app = app
	c.driver = driver
	if err := c.connDriver(); err != nil {
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

func (c *Client) connDriver() error {
	logger.Infof("连接driver: %+v", C.DriverGrpc)
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("%s:%d", C.DriverGrpc.Host, C.DriverGrpc.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc.Dial error: %s", err)
	}
	cli := pb.NewDriverServiceClient(conn)
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
				logger.Infof("健康检查")
				retry := 3
				state := false
				for retry >= 0 {
					_, err := c.cli.HealthCheck(ctx, &pb.HealthCheckRequest{Service: C.ServiceID})
					if err != nil {
						logger.Errorf("健康检查重试错误,%s", err.Error())
						state = true
						time.Sleep(time.Second * time.Duration(wait))
					} else {
						state = false
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
					if err := c.connDriver(); err != nil {
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

func (c *Client) WriteEvent(ctx context.Context, event Event) error {
	if event.Table == "" || event.ID == "" || event.EventID == "" {
		return errors.New("表、资产或事件ID为空")
	}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	res, err := c.cli.Event(ctx, &pb.Request{
		Project: C.Project,
		Data:    b,
	})
	if err != nil {
		return err
	}
	if !res.GetStatus() {
		return fmt.Errorf(res.GetInfo())
	}
	return nil
}

func (c *Client) RunLog(ctx context.Context, l Log) error {
	if l.SerialNo == "" {
		return errors.New("流水号为空")
	}
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	res, err := c.cli.CommandLog(ctx, &pb.Request{
		Project: C.Project,
		Data:    b,
	})
	if err != nil {
		return err
	}
	if !res.GetStatus() {
		return fmt.Errorf(res.GetInfo())
	}
	return nil
}

func (c *Client) UpdateTableData(ctx context.Context, l TableData, result interface{}) error {
	if l.TableID == "" || l.ID == "" {
		return errors.New("表或记录id为空")
	}
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	res, err := c.cli.UpdateTableData(ctx, &pb.Request{
		Project: C.Project,
		Data:    b,
	})
	if err != nil {
		return err
	}
	if !res.GetStatus() {
		return fmt.Errorf(res.GetInfo())
	}
	if err := json.Unmarshal(res.GetResult(), result); err != nil {
		return err
	}
	return nil
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
				time.Sleep(time.Second * time.Duration(wait))
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("start stream break")
				return
			default:
				logger.Infof("start stream")
				if err := c.StartStream(context.Background()); err != nil {
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
				if err := c.RunStream(context.Background()); err != nil {
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
				logger.Infof("writeTag stream break")
				return
			default:
				logger.Infof("writeTag stream")
				if err := c.WriteTagStream(context.Background()); err != nil {
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
				logger.Infof("batchRun stream break")
				return
			default:
				logger.Infof("batchRun stream")
				if err := c.BatchRunStream(context.Background()); err != nil {
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
				logger.Infof("debug stream break")
				return
			default:
				logger.Infof("debug stream")
				if err := c.DebugStream(context.Background()); err != nil {
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

func (c *Client) SchemaStream(ctx context.Context) error {
	stream, err := c.cli.SchemaStream(dGrpc.GetGrpcContext(ctx, C.ServiceID, C.Project, C.Driver.ID, C.Driver.Name))
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
		schema, err := c.driver.Schema(c.app)
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

func (c *Client) StartStream(ctx context.Context) error {
	stream, err := c.cli.StartStream(dGrpc.GetGrpcContext(ctx, C.ServiceID, C.Project, C.Driver.ID, C.Driver.Name))
	if err != nil {
		return fmt.Errorf("start stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("start stream close err,%s", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("start stream err, %s", err)
		}
		startRes := new(grpcResult)
		var cfg Instance
		if err := json.Unmarshal(res.Config, &cfg); err != nil {
			startRes.Error = err.Error()
			startRes.Code = 400
			bts, _ := json.Marshal(startRes)
			if err := stream.Send(&pb.StartResult{
				Request: res.Request,
				Message: bts,
			}); err != nil {
				logger.Errorf("start stream 发送错误,%s", err)
			}
			continue
		}
		if cfg.Tables != nil {
			for _, t := range cfg.Tables {
				if t.Devices == nil {
					continue
				}
				for _, device := range t.Devices {
					devM, ok := c.cacheConfigNum.Load(device.Id)
					var devI map[string]interface{}
					if ok {
						devI, _ = devM.(map[string]interface{})
					} else {
						devI = map[string]interface{}{}
					}
					devI[t.Id] = struct{}{}
					c.cacheConfigNum.Store(device.Id, devI)
					c.cacheConfig.Store(device.Id, t.Id)
				}
			}
		}
		if err := c.driver.Start(c.app, res.Config); err != nil {
			startRes.Error = err.Error()
			startRes.Code = 400
		} else {
			startRes.Code = 200
		}
		bts, _ := json.Marshal(startRes)
		if err := stream.Send(&pb.StartResult{
			Request: res.Request,
			Message: bts,
		}); err != nil {
			logger.Errorf("start stream 发送错误,%s", err)
		}
	}
}

func (c *Client) RunStream(ctx context.Context) error {
	stream, err := c.cli.RunStream(dGrpc.GetGrpcContext(ctx, C.ServiceID, C.Project, C.Driver.ID, C.Driver.Name))
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
			return fmt.Errorf("run stream err, %s", err)
		}
		gr := new(grpcResult)
		runRes, err := c.driver.Run(c.app, &Command{
			Table:    res.TableId,
			Id:       res.Id,
			SerialNo: res.SerialNo,
			Command:  res.Command,
		})
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
	}
}

func (c *Client) WriteTagStream(ctx context.Context) error {
	stream, err := c.cli.WriteTagStream(dGrpc.GetGrpcContext(ctx, C.ServiceID, C.Project, C.Driver.ID, C.Driver.Name))
	if err != nil {
		return fmt.Errorf("writeTag stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("writeTag stream close err,%s", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("writeTag stream err, %s", err)
		}
		gr := new(grpcResult)
		runRes, err := c.driver.WriteTag(c.app, &Command{
			Table:    res.TableId,
			Id:       res.Id,
			SerialNo: res.SerialNo,
			Command:  res.Command,
		})
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
			logger.Errorf("writeTag stream 发送错误,%s", err)
		}
	}
}

func (c *Client) BatchRunStream(ctx context.Context) error {
	stream, err := c.cli.BatchRunStream(dGrpc.GetGrpcContext(ctx, C.ServiceID, C.Project, C.Driver.ID, C.Driver.Name))
	if err != nil {
		return fmt.Errorf("batchRun stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("batchRun stream close err,%s", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("batchRun stream err, %s", err)
		}
		gr := new(grpcResult)
		runRes, err := c.driver.BatchRun(c.app, &BatchCommand{
			Table:    res.TableId,
			Ids:      res.Id,
			SerialNo: res.SerialNo,
			Command:  res.Command,
		})
		if err != nil {
			gr.Error = err.Error()
			gr.Code = 400
		} else {
			gr.Result = runRes
			gr.Code = 200
		}
		bts, _ := json.Marshal(gr)
		if err := stream.Send(&pb.BatchRunResult{
			Request: res.Request,
			Message: bts,
		}); err != nil {
			logger.Errorf("batchRun stream 发送错误,%s", err)
		}
	}
}

func (c *Client) DebugStream(ctx context.Context) error {
	stream, err := c.cli.DebugStream(dGrpc.GetGrpcContext(ctx, C.ServiceID, C.Project, C.Driver.ID, C.Driver.Name))
	if err != nil {
		return fmt.Errorf("debug stream err,%s", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Infof("debug stream close err,%s", err)
		}
	}()
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("debug stream err, %s", err)
		}
		gr := new(grpcResult)
		runRes, err := c.driver.Debug(c.app, res.Data)
		if err != nil {
			gr.Error = err.Error()
			gr.Code = 400
		} else {
			gr.Result = runRes
			gr.Code = 200
		}
		bts, _ := json.Marshal(gr)
		if err := stream.Send(&pb.Debug{
			Request: res.Request,
			Data:    bts,
		}); err != nil {
			logger.Errorf("debug stream 发送错误,%s", err)
		}
	}
}
