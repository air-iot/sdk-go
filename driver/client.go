package driver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/air-iot/json"
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/air-iot/api-client-go/v4/driver"
	"github.com/air-iot/logger"
	dGrpc "github.com/air-iot/sdk-go/v4/driver/grpc"
)

type Client struct {
	lock           sync.RWMutex
	conn           *grpc.ClientConn
	cli            pb.DriverServiceClient
	app            App
	driver         Driver
	clean          func()
	cacheConfig    sync.Map
	cacheConfigNum sync.Map
	streamCount    int32
	//healthTime        time.Time
}

const totalStream = 7

func (c *Client) Start(app App, driver Driver) *Client {
	c.app = app
	c.driver = driver
	c.streamCount = 0
	c.restart()
	return c
}

func (c *Client) start() {
	if err := c.connDriver(); err != nil {
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
	logger.Infof("重启驱动管理连接")
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Stop()
	c.start()
}

func (c *Client) connDriver() error {
	logger.Infof("连接driver: %+v", Cfg.DriverGrpc)
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("%s:%d", Cfg.DriverGrpc.Host, Cfg.DriverGrpc.Port),
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
	logger.Infof("健康检查启动")
	go func() {
		waitTime := time.Second * time.Duration(Cfg.DriverGrpc.WaitTime)
		nextTime := time.Now().Local().Add(time.Duration(Cfg.DriverGrpc.Health.Retry) * waitTime)
		for {
			select {
			case <-ctx.Done():
				logger.Infof("健康检查停止")
				return
			default:
				ctx1 := logger.NewModuleContext(ctx, entity.MODULE_HEALTHCHECK)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(ctx1)
				newLogger.Debugf("健康检查开始")
				retry := Cfg.DriverGrpc.Health.Retry
				state := false

				for retry >= 0 {
					healthRes, err := c.healthRequest(ctx)
					if err != nil {
						newLogger.Errorf("健康检查错误,%s", err.Error())
						state = true
						time.Sleep(waitTime)
					} else {
						state = false
						if healthRes.GetStatus() == pb.HealthCheckResponse_SERVING {
							newLogger.Debugf("健康检查正常")
							if healthRes.Errors != nil && len(healthRes.Errors) > 0 {
								for _, e := range healthRes.Errors {
									newLogger.Errorf("执行 %s, 错误为%s", e.Code.String(), e.Message)
								}
							}
						} else if healthRes.GetStatus() == pb.HealthCheckResponse_SERVICE_UNKNOWN {
							newLogger.Errorf("健康检查异常,服务端未找到本驱动服务")
							state = true
						}
						break
					}
					retry--
				}

				if state {
					c.restart()
				} else if time.Now().Local().After(nextTime) {
					nextTime = time.Now().Local().Add(time.Duration(Cfg.DriverGrpc.Health.Retry) * waitTime)
					getV := atomic.LoadInt32(&c.streamCount)
					newLogger.Debugf("健康检查,找到流数量为:%d", getV)
					if getV < totalStream {
						newLogger.Errorf("健康检查异常,找到流数量不匹配,应为:%d,实际为:%d", totalStream, getV)
						c.restart()
					}
				}
				time.Sleep(waitTime)
			}
		}
	}()
}

func (c *Client) healthRequest(ctx context.Context) (*pb.HealthCheckResponse, error) {
	reqCtx, reqCancel := context.WithTimeout(ctx, time.Second*time.Duration(Cfg.DriverGrpc.Health.RequestTime))
	defer reqCancel()
	healthRes, err := c.cli.HealthCheck(reqCtx, &pb.HealthCheckRequest{Service: Cfg.ServiceID})
	return healthRes, err
}

func (c *Client) WriteEvent(ctx context.Context, event entity.Event) error {
	if event.Table == "" || event.ID == "" || event.EventID == "" {
		return errors.New("表、设备或事件ID为空")
	}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	res, err := c.cli.Event(ctx, &pb.Request{
		Project: Cfg.Project,
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

func (c *Client) FindDevice(ctx context.Context, table, id string, ret interface{}) error {
	if id == "" {
		return errors.New("设备ID为空")
	}
	res, err := c.cli.FindTableData(ctx, &pb.TableDataRequest{
		Service:     Cfg.ServiceID,
		TableId:     table,
		TableDataId: id,
	})
	if err != nil {
		return err
	}
	if !res.GetStatus() {
		return fmt.Errorf(res.GetInfo())
	}
	if err := json.Unmarshal(res.GetResult(), ret); err != nil {
		return fmt.Errorf("解析请求结果错误, %v", err)
	}
	return nil
}

func (c *Client) RunLog(ctx context.Context, l entity.Log) error {
	if l.SerialNo == "" {
		return errors.New("流水号为空")
	}
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	res, err := c.cli.CommandLog(ctx, &pb.Request{
		Project: Cfg.Project,
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

func (c *Client) UpdateTableData(ctx context.Context, l entity.TableData, result interface{}) error {
	if l.TableID == "" || l.ID == "" {
		return errors.New("表或记录id为空")
	}
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	res, err := c.cli.UpdateTableData(ctx, &pb.Request{
		Project: Cfg.Project,
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
				logger.Infof("关闭驱动schema stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_SCHEMA))
				newLogger.Infof("启动驱动schema stream")
				if err := c.SchemaStream(ctx1); err != nil {
					newLogger.Errorf("驱动schema stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("关闭驱动start stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_START))
				newLogger.Infof("启动驱动start stream")
				if err := c.StartStream(ctx1); err != nil {
					newLogger.Errorf("驱动start stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("关闭驱动run stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_RUN))
				newLogger.Infof("启动驱动run stream")
				if err := c.RunStream(ctx1); err != nil {
					newLogger.Errorf("驱动run stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("关闭驱动writeTag stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_WRITETAG))
				newLogger.Infof("启动驱动writeTag stream")
				if err := c.WriteTagStream(ctx1); err != nil {
					newLogger.Errorf("驱动writeTag stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("关闭驱动batchRun stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_BATCHRUN))
				newLogger.Infof("启动驱动batchRun stream")
				if err := c.BatchRunStream(ctx1); err != nil {
					newLogger.Errorf("驱动batchRun stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("关闭驱动debug stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_DEBUG))
				newLogger.Infof("启动驱动debug stream")
				if err := c.DebugStream(ctx1); err != nil {
					newLogger.Errorf("驱动debug stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("关闭驱动http proxy stream")
				return
			default:
				ctx1 := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
				}
				newLogger := logger.WithContext(logger.NewModuleContext(ctx1, entity.MODULE_HTTPPROXY))
				newLogger.Infof("启动驱动http proxy stream")
				if err := c.HttpProxyStream(ctx1); err != nil {
					newLogger.Errorf("驱动http proxy stream错误,%v", err)
				}
				time.Sleep(time.Second * time.Duration(Cfg.DriverGrpc.WaitTime))
			}
		}
	}()
}

func (c *Client) SchemaStream(ctx context.Context) error {
	stream, err := c.cli.SchemaStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("schema stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Errorf("schema stream close err,%v", err)
		}
	}()
	logger.Infof("schema stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("schema stream err, %w", err)
		}
		go func(res *pb.SchemaRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, entity.MODULE_SCHEMA)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			schema, err := c.driver.Schema(ctx1, c.app)
			schemaRes := new(entity.GrpcResult)
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
				logger.WithContext(ctx1).Errorf("驱动配置(schema)返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}

func (c *Client) StartStream(ctx context.Context) error {
	stream, err := c.cli.StartStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("start stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Infof("start stream close err,%v", err)
		}
	}()
	logger.Infof("start stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("start stream err, %w", err)
		}
		ctx1 := logger.NewModuleContext(context.Background(), entity.MODULE_START)
		newLogger := logger.WithContext(ctx1)
		startRes := new(entity.GrpcResult)
		var cfg entity.Instance
		if err := json.Unmarshal(res.Config, &cfg); err != nil {
			startRes.Error = err.Error()
			startRes.Code = 400
			bts, _ := json.Marshal(startRes)
			if err := stream.Send(&pb.StartResult{
				Request: res.Request,
				Message: bts,
			}); err != nil {
				newLogger.Errorf("启动驱动(start)时(解析配置的错误)返回到驱动管理错误,%v", err)
			}
			continue
		}
		if cfg.Debug != nil {
			if *cfg.Debug {
				logger.SetLevel(logger.DebugLevel)
			} else {
				logger.SetLevel(logger.InfoLevel)
			}
		}
		c.cacheConfigNum = sync.Map{}
		c.cacheConfig = sync.Map{}
		if cfg.GroupId != "" {
			Cfg.GroupID = cfg.GroupId
		}
		if Cfg.GroupID != "" {
			ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
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
		go func(res *pb.StartRequest) {
			newCtx, cancel := context.WithTimeout(ctx1, time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			if err := c.driver.Start(newCtx, c.app, res.Config); err != nil {
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
				newLogger.Errorf("启动驱动(start)结果返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}

func (c *Client) RunStream(ctx context.Context) error {
	stream, err := c.cli.RunStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("run stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Infof("run stream close err,%v", err)
		}
	}()
	logger.Infof("run stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("run stream err, %w", err)
		}
		go func(res *pb.RunRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, entity.MODULE_RUN)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			gr := new(entity.GrpcResult)
			runRes, err := c.driver.Run(ctx1, c.app, &entity.Command{
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
				logger.WithContext(ctx1).Errorf("执行指令(run)结果返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}

func (c *Client) WriteTagStream(ctx context.Context) error {
	stream, err := c.cli.WriteTagStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("writeTag stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Infof("writeTag stream close err,%v", err)
		}
	}()
	logger.Infof("writeTag stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("writeTag stream err, %w", err)
		}
		go func(res *pb.RunRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, entity.MODULE_WRITETAG)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			gr := new(entity.GrpcResult)
			runRes, err := c.driver.WriteTag(ctx1, c.app, &entity.Command{
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
				logger.WithContext(ctx1).Errorf("写数据点(writeTag)结果返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}

func (c *Client) BatchRunStream(ctx context.Context) error {
	stream, err := c.cli.BatchRunStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("batchRun stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Infof("batchRun stream close err,%v", err)
		}
	}()
	logger.Infof("batchRun stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("batchRun stream err, %w", err)
		}
		go func(res *pb.BatchRunRequest) {
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, entity.MODULE_BATCHRUN)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			gr := new(entity.GrpcResult)
			runRes, err := c.driver.BatchRun(ctx1, c.app, &entity.BatchCommand{
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
				logger.WithContext(ctx1).Errorf("批量执行指令(batchRun)结果返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}

func (c *Client) DebugStream(ctx context.Context) error {
	stream, err := c.cli.DebugStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("debug stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Infof("debug stream close err,vs", err)
		}
	}()
	logger.Infof("debug stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("debug stream err, %s", err)
		}
		go func(res *pb.Debug) {
			gr := new(entity.GrpcResult)
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, entity.MODULE_DEBUG)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			runRes, err := c.driver.Debug(ctx1, c.app, res.Data)
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
				logger.WithContext(ctx1).Errorf("调试(debug)结果返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}

func (c *Client) HttpProxyStream(ctx context.Context) error {
	stream, err := c.cli.HttpProxyStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name))
	if err != nil {
		return fmt.Errorf("http proxy stream err,%w", err)
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			logger.Infof("http proxy stream close err,%v", err)
		}
	}()
	logger.Infof("http proxy stream conn success")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("http proxy stream err, %s", err)
		}
		go func(res *pb.HttpProxyRequest) {
			gr := new(entity.GrpcResult)
			var header http.Header
			ctx1, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(Cfg.Driver.Timeout))
			defer cancel()
			ctx1 = logger.NewModuleContext(ctx1, entity.MODULE_HTTPPROXY)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			if res.GetHeaders() != nil {
				if err := json.Unmarshal(res.GetHeaders(), &header); err != nil {
					gr.Error = fmt.Sprintf("http proxy stream err, %v", err)
					gr.Code = 400
				} else {
					runRes, err := c.driver.HttpProxy(ctx1, c.app, res.GetType(), header, res.GetData())
					if err != nil {
						gr.Error = err.Error()
						gr.Code = 400
					} else {
						gr.Result = runRes
						gr.Code = 200
					}
				}
			} else {
				runRes, err := c.driver.HttpProxy(ctx1, c.app, res.GetType(), header, res.GetData())
				if err != nil {
					gr.Error = err.Error()
					gr.Code = 400
				} else {
					gr.Result = runRes
					gr.Code = 200
				}
			}
			bts, _ := json.Marshal(gr)
			if err := stream.Send(&pb.HttpProxyResult{
				Request: res.Request,
				Data:    bts,
			}); err != nil {
				logger.WithContext(ctx1).Errorf("http代理(httpProxy)请求结果返回到驱动管理错误,%v", err)
			}
		}(res)
	}
}
