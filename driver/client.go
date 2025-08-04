package driver

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/air-iot/api-client-go/v4/apicontext"
	"github.com/air-iot/api-client-go/v4/config"
	"github.com/air-iot/errors"
	"github.com/air-iot/json"
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	api "github.com/air-iot/api-client-go/v4/api"
	pb "github.com/air-iot/api-client-go/v4/driver"
	"github.com/air-iot/logger"
	dGrpc "github.com/air-iot/sdk-go/v4/driver/grpc"
)

type Client struct {
	lock sync.RWMutex

	conn        *grpc.ClientConn
	cli         pb.DriverServiceClient
	instructCli pb.DriverInstructServiceClient
	app         App
	driver      Driver
	clean       func()
	cacheConfig *cacheConfig
	streamCount int32
}

const totalStream = 8
const STREAM_HEARTBEAT = "heartbeat"

func (c *Client) Start(app App, driver Driver) *Client {
	c.app = app
	c.driver = driver
	c.streamCount = 0
	c.start()
	return c
}

type cacheConfig struct {
	lock sync.RWMutex
	data map[string]map[string]struct{}
}

func NewCacheConfig() *cacheConfig {
	return &cacheConfig{
		data: make(map[string]map[string]struct{}),
	}
}

func (c *cacheConfig) clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	clear(c.data)
}

func (c *cacheConfig) set(table, id string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	m, ok := c.data[id]
	if !ok {
		m = make(map[string]struct{})
	}
	m[table] = struct{}{}
	c.data[id] = m
}

func (c *cacheConfig) del(table string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for id, m := range c.data {
		for k := range m {
			if k == table {
				delete(m, table)
			}
		}
		if len(m) == 0 {
			delete(c.data, id)
		}
	}
}

func (c *cacheConfig) delDevice(table, device string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	tables, ok := c.data[device]
	if !ok {
		return
	}

	if _, ok := tables[table]; ok {
		delete(tables, table)
	}

	if len(tables) == 0 {
		delete(c.data, device)
	}
}

func (c *cacheConfig) get(id string) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	m, ok := c.data[id]
	if !ok {
		return "", errors.New("未找到设备所在的表")
	}
	if len(m) > 1 {
		return "", errors.New("设备所在的表不唯一")
	}
	for k := range m {
		return k, nil
	}
	return "", errors.New("未找到设备所在的表")
}

func (c *Client) start() {
	ctx := logger.NewModuleContext(context.Background(), entity.MODULE_STARTDRIVER)
	if Cfg.GroupID != "" {
		ctx = logger.NewGroupContext(ctx, Cfg.GroupID)
	}
	ctx, cancel := context.WithCancel(ctx)
	c.clean = func() {
		cancel()
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				waitTime := Cfg.DriverGrpc.WaitTime
				if err := c.run(ctx); err != nil {
					logger.WithContext(ctx).Errorln(err)
				}
				time.Sleep(waitTime)
			}
		}
	}()

}

func (c *Client) Stop() {
	ctx := logger.NewModuleContext(context.Background(), entity.MODULE_STARTDRIVER)
	if Cfg.GroupID != "" {
		ctx = logger.NewGroupContext(ctx, Cfg.GroupID)
	}
	logger.WithContext(ctx).Infof("停止驱动管理连接")
	if c.clean != nil {
		c.clean()
	}
	c.close(ctx)
}

func (c *Client) run(ctx context.Context) error {
	if err := c.connDriver(ctx); err != nil {
		return err
	}
	defer c.close(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sessionId := primitive.NewObjectID().Hex()
	c.startSteam(ctx, sessionId)
	c.healthCheck(ctx)
	return nil
}

func (c *Client) close(ctx context.Context) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logger.WithContext(ctx).Errorf("关闭grpc连接. %v", err)
		} else {
			c.conn = nil
		}
	}
}

func (c *Client) connDriver(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	ctx, cancel := context.WithTimeout(ctx, Cfg.DriverGrpc.Timeout)
	defer cancel()
	logger.WithContext(ctx).Infof("连接driver: 配置=%+v", Cfg.DriverGrpc)
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", Cfg.DriverGrpc.Host, Cfg.DriverGrpc.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(Cfg.DriverGrpc.Limit*1024*1024), grpc.MaxCallSendMsgSize(Cfg.DriverGrpc.Limit*1024*1024)),
	)
	if err != nil {
		return fmt.Errorf("grpc.Dial error: %w", err)
	}
	c.conn = conn
	c.cli = pb.NewDriverServiceClient(conn)
	c.instructCli = pb.NewDriverInstructServiceClient(conn)
	return nil
}

func (c *Client) healthCheck(ctx context.Context) {
	logger.WithContext(ctx).Infof("健康检查: 启动")
	nextTime := time.Now().Local().Add(Cfg.DriverGrpc.WaitTime * time.Duration(Cfg.DriverGrpc.Health.Retry))
	for {
		select {
		case <-ctx.Done():
			logger.WithContext(ctx).Infof("健康检查: 停止")
			return
		default:
			waitTime := Cfg.DriverGrpc.WaitTime
			ctx1 := logger.NewModuleContext(ctx, entity.MODULE_HEALTHCHECK)
			if Cfg.GroupID != "" {
				ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
			}
			newLogger := logger.WithContext(ctx1)
			newLogger.Debugf("健康检查: 开始")
			retry := Cfg.DriverGrpc.Health.Retry
			state := false
			for retry >= 0 {
				healthRes, err := c.healthRequest(ctx)
				if err != nil {
					errCtx := logger.NewErrorContext(ctx1, err)
					logger.WithContext(errCtx).Errorf("健康检查: 健康检查第 %d 次错误", Cfg.DriverGrpc.Health.Retry-retry+1)
					state = true
					time.Sleep(waitTime)
				} else {
					state = false
					if healthRes.GetStatus() == pb.HealthCheckResponse_SERVING {
						newLogger.Debugf("健康检查: 正常")
						if healthRes.Errors != nil && len(healthRes.Errors) > 0 {
							for _, e := range healthRes.Errors {
								newLogger.Errorf("健康检查: code=%s,错误=%s", e.Code.String(), e.Message)
								if e.Code == pb.Error_Start {
									state = true
								}
							}
						}
					} else if healthRes.GetStatus() == pb.HealthCheckResponse_SERVICE_UNKNOWN {
						newLogger.Errorf("健康检查: 服务端未找到本驱动服务")
						state = true
					}
					break
				}
				retry--
			}

			if state {
				return
			} else if time.Now().Local().After(nextTime) {
				nextTime = time.Now().Local().Add(time.Duration(Cfg.DriverGrpc.Health.Retry) * waitTime)
				getV := atomic.LoadInt32(&c.streamCount)
				newLogger.Debugf("健康检查: 找到流数量=%d", getV)
				if getV < totalStream {
					newLogger.Errorf("健康检查: 找到流数量不匹配,应为=%d,实际为=%d", totalStream, getV)
					return
				}
			}
			time.Sleep(waitTime)
		}
	}

}

func (c *Client) healthRequest(ctx context.Context) (*pb.HealthCheckResponse, error) {
	reqCtx, reqCancel := context.WithTimeout(ctx, Cfg.DriverGrpc.Health.RequestTime)
	defer reqCancel()
	healthRes, err := c.cli.HealthCheck(reqCtx, &pb.HealthCheckRequest{Service: Cfg.ServiceID, ProjectId: Cfg.Project, DriverId: Cfg.Driver.ID})
	return healthRes, err
}

func (c *Client) WriteEvent(ctx context.Context, event entity.Event) error {
	if event.Table == "" || event.ID == "" || event.EventID == "" {
		return fmt.Errorf("表、设备或事件ID为空")
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
		return fmt.Errorf("设备ID为空")
	}
	res, err := c.cli.FindTableData(ctx, &pb.TableDataRequest{
		ProjectId:   Cfg.Project,
		DriverId:    Cfg.Driver.ID,
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
		return fmt.Errorf("解析请求结果错误: %v", err)
	}
	return nil
}

func (c *Client) GetCommands(ctx context.Context, table, id string, result interface{}) error {
	if table == "" {
		return fmt.Errorf("表ID为空")
	}
	if id == "" {
		return fmt.Errorf("设备ID为空")
	}
	res, err := c.instructCli.GetCommands(apicontext.GetGrpcContext(ctx, map[string]string{config.XRequestProject: Cfg.Project}), &pb.RequestCommand{
		TableId:     table,
		TableDataId: id,
	})
	if err != nil {
		return err
	}
	if !res.GetStatus() {
		return fmt.Errorf(res.GetInfo())
	}
	if err := json.Unmarshal(res.GetResult(), result); err != nil {
		return fmt.Errorf("解析请求结果错误: %v", err)
	}
	return nil
}

func (c *Client) UpdateCommand(ctx context.Context, id string, data entity.DriverInstruct) error {
	if id == "" {
		return fmt.Errorf("ID为空")
	}
	marshalB, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("数据错误: %v", err)
	}
	res, err := c.instructCli.Update(
		apicontext.GetGrpcContext(ctx, map[string]string{config.XRequestProject: Cfg.Project}), &api.UpdateRequest{
			Id:   id,
			Data: marshalB,
		})
	if err != nil {
		return err
	}
	if !res.GetStatus() {
		return fmt.Errorf(res.GetInfo())
	}

	return nil
}

func (c *Client) RunLog(ctx context.Context, l entity.Log) error {
	if l.SerialNo == "" {
		return fmt.Errorf("流水号为空")
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
		return fmt.Errorf("表或记录id为空")
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

func (c *Client) startSteam(ctx context.Context, sessionId string) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("schema: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_SCHEMA)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("schema: 启动stream")
				if err := c.SchemaStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("schema: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("start: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_START)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("start: 启动stream")
				if err := c.StartStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("start: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("执行指令: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_RUN)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("执行指令: 启动stream")
				if err := c.RunStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("执行指令: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("写数据点: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_WRITETAG)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("写数据点: 启动stream")
				if err := c.WriteTagStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("写数据点: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("批量执行指令: stream创建错误")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_BATCHRUN)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("批量执行指令: 启动stream")
				if err := c.BatchRunStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("批量执行指令: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("调试: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_DEBUG)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("调试: 启动stream")
				if err := c.DebugStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("调试: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("httpProxy: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_HTTPPROXY)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("httpProxy: 启动stream")
				if err := c.HttpProxyStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("httpProxy: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.WithContext(ctx).Infof("配置更新: 通过上下文关闭stream检查")
				return
			default:
				newCtx := context.WithoutCancel(ctx)
				if Cfg.GroupID != "" {
					newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
				}
				newCtx = logger.NewModuleContext(newCtx, entity.MODULE_CONFIGUPDATE)
				newLogger := logger.WithContext(newCtx)
				newLogger.Infof("配置更新: 启动stream")
				if err := c.ConfigUpdateStream(newCtx, sessionId); err != nil {
					errCtx := logger.NewErrorContext(newCtx, err)
					logger.WithContext(errCtx).Errorf("配置更新: stream创建错误")
				}
				time.Sleep(Cfg.DriverGrpc.WaitTime)
			}
		}
	}()
}

func (c *Client) SchemaStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.SchemaStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("schema: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("schema: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.SchemaRequest) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("schema stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewModuleContext(newCtx, entity.MODULE_SCHEMA)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			logger.WithContext(newCtx).Debugf("schema: 接收到查询请求")
			schema, err := c.driver.Schema(newCtx, c.app, res.GetLocale())
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("schema: 配置返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) StartStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.StartStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		//hCancel()
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("start: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("start: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		if res.GetRequest() == STREAM_HEARTBEAT {
			continue
		}
		ctx1 := logger.NewModuleContext(context.Background(), entity.MODULE_START)
		logger.WithContext(ctx1).Debugf("start: 接收到开始请求")
		var cfg entity.Instance
		if err := json.Unmarshal(res.Config, &cfg); err != nil {
			startRes := new(entity.GrpcResult)
			startRes.Error = err.Error()
			startRes.Code = 400
			bts, _ := json.Marshal(startRes)
			if err := stream.Send(&pb.StartResult{
				Request: res.Request,
				Message: bts,
			}); err != nil {
				errCtx := logger.NewErrorContext(ctx1, err)
				logger.WithContext(errCtx).Errorf("start: 解析配置的错误返回到驱动管理错误")
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
		if cfg.GroupId != "" {
			Cfg.GroupID = cfg.GroupId
		}
		if Cfg.GroupID != "" {
			ctx1 = logger.NewGroupContext(ctx1, Cfg.GroupID)
		}
		c.updateDriverCache(cfg)
		run := func(res *pb.StartRequest) {
			newCtx, cancel := context.WithTimeout(ctx1, Cfg.DriverGrpc.Timeout)
			defer cancel()
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					startRes := new(entity.GrpcResult)
					startRes.Error = errStr
					startRes.Code = 400
					bts, _ := json.Marshal(startRes)
					if err := stream.Send(&pb.StartResult{
						Request: res.Request,
						Message: bts,
					}); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("start: 启动驱动结果返回到驱动管理错误")
					}
				}
			}()
			startRes := new(entity.GrpcResult)
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("start: 启动驱动结果返回到驱动管理错误")
			}
		}
		run(res)
	}
}

func (c *Client) RunStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.RunStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("执行指令: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("执行指令: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.RunRequest) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("执行指令stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewTDMContext(newCtx, res.TableId, res.Id, entity.MODULE_RUN)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			logger.WithContext(newCtx).Debugf("执行指令: 设备表=%s,设备=%s,指令=%s", res.TableId, res.Id, res.Command)
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					gr := new(entity.GrpcResult)
					gr.Error = errStr
					gr.Code = 400
					bts, _ := json.Marshal(gr)
					if err := stream.Send(&pb.RunResult{
						Request: res.Request,
						Message: bts,
					}); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("执行指令: 执行指令结果返回到驱动管理错误")
					}
				}
			}()
			gr := new(entity.GrpcResult)
			runRes, err := c.driver.Run(newCtx, c.app, &entity.Command{
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("执行指令: 执行指令结果返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) WriteTagStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.WriteTagStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("写数据点: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("写数据点: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.RunRequest) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("写数据点stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewTDMContext(newCtx, res.TableId, res.Id, entity.MODULE_WRITETAG)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			logger.WithContext(newCtx).Debugf("写数据点: 设备表=%s,设备=%s,指令=%s", res.TableId, res.Id, res.Command)
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					gr := new(entity.GrpcResult)
					gr.Error = errStr
					gr.Code = 400
					bts, _ := json.Marshal(gr)
					if err := stream.Send(&pb.RunResult{
						Request: res.Request,
						Message: bts,
					}); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("写数据点: 写数据点执行结果返回到驱动管理错误")
					}
				}
			}()
			gr := new(entity.GrpcResult)
			runRes, err := c.driver.WriteTag(newCtx, c.app, &entity.Command{
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("写数据点: 写数据点执行结果返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) BatchRunStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.BatchRunStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("批量执行指令: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("批量执行指令: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.BatchRunRequest) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("批量执行指令stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewModuleContext(newCtx, entity.MODULE_BATCHRUN)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			newCtx = logger.NewTableContext(newCtx, res.TableId)
			logger.WithContext(newCtx).Debugf("批量执行指令: 设备表=%s,设备=%+v,指令=%s", res.TableId, res.Id, res.Command)
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					gr := new(entity.GrpcResult)
					gr.Error = errStr
					gr.Code = 400
					bts, _ := json.Marshal(gr)
					if err := stream.Send(&pb.BatchRunResult{
						Request: res.Request,
						Message: bts,
					}); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("批量执行指令: 批量执行指令结果返回到驱动管理错误")
					}
				}
			}()
			gr := new(entity.GrpcResult)
			runRes, err := c.driver.BatchRun(newCtx, c.app, &entity.BatchCommand{
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("批量执行指令: 批量执行指令结果返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) DebugStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.DebugStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		//hCancel()
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("调试: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("调试: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.Debug) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("调试stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewModuleContext(newCtx, entity.MODULE_DEBUG)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			logger.WithContext(newCtx).Debugf("调试: 请求数据=%s", res.Data)
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					gr := new(entity.GrpcResult)
					gr.Error = errStr
					gr.Code = 400
					bts, _ := json.Marshal(gr)
					if err := stream.Send(&pb.Debug{
						Request: res.Request,
						Data:    bts,
					}); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("调试: 调试结果返回到驱动管理错误")
					}
				}
			}()
			runRes, err := c.driver.Debug(newCtx, c.app, res.Data)
			gr := new(entity.GrpcResult)
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("调试: 调试结果返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) HttpProxyStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.HttpProxyStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		//hCancel()
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("httpProxy: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("httpProxy: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.HttpProxyRequest) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("httpProxy stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			var header http.Header
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewModuleContext(newCtx, entity.MODULE_HTTPPROXY)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			logger.WithContext(newCtx).Debugf("httpProxy: type=%s,header=%s,请求数据=%s", res.Type, res.Headers, res.Data)
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					gr := new(entity.GrpcResult)
					gr.Error = errStr
					gr.Code = 400
					bts, _ := json.Marshal(gr)
					if err := stream.Send(&pb.HttpProxyResult{
						Request: res.Request,
						Data:    bts,
					}); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("httpProxy: 请求结果返回到驱动管理错误")
					}
				}
			}()
			gr := new(entity.GrpcResult)
			if res.GetHeaders() != nil {
				if err := json.Unmarshal(res.GetHeaders(), &header); err != nil {
					gr.Error = fmt.Sprintf("httpProxy流错误:%v", err)
					gr.Code = 400
				} else {
					runRes, err := c.driver.HttpProxy(newCtx, c.app, res.GetType(), header, res.GetData())
					if err != nil {
						gr.Error = err.Error()
						gr.Code = 400
					} else {
						gr.Result = runRes
						gr.Code = 200
					}
				}
			} else {
				runRes, err := c.driver.HttpProxy(newCtx, c.app, res.GetType(), header, res.GetData())
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
				errCtx := logger.NewErrorContext(newCtx, err)
				logger.WithContext(errCtx).Errorf("httpProxy: 请求结果返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) ConfigUpdateStream(ctx context.Context, sessionId string) error {
	stream, err := c.cli.ConfigUpdateStream(dGrpc.GetGrpcContext(ctx, Cfg.ServiceID, Cfg.Project, Cfg.Driver.ID, Cfg.Driver.Name, sessionId))
	if err != nil {
		return err
	}
	defer func() {
		//hCancel()
		atomic.AddInt32(&c.streamCount, -1)
		if err := stream.CloseSend(); err != nil {
			errCtx := logger.NewErrorContext(ctx, err)
			logger.WithContext(errCtx).Errorf("httpProxy: stream关闭错误")
		}
	}()
	logger.WithContext(ctx).Infof("配置更新: stream连接成功")
	atomic.AddInt32(&c.streamCount, 1)
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}
		go func(res *pb.ConfigUpdateRequest) {
			if res.GetRequest() == STREAM_HEARTBEAT {
				//logger.WithContext(hCtx).Debugf("httpProxy stream收到心跳响应包输入到管道")
				//ch <- struct{}{}
				return
			}
			newCtx, cancel := context.WithTimeout(context.Background(), Cfg.DriverGrpc.Timeout)
			defer cancel()
			newCtx = logger.NewModuleContext(newCtx, entity.MODULE_CONFIGUPDATE)
			if Cfg.GroupID != "" {
				newCtx = logger.NewGroupContext(newCtx, Cfg.GroupID)
			}
			logger.WithContext(newCtx).Debugf("配置更新: type=%s,请求数据=%s", res.OpsType, res.Data)
			defer func() {
				if errR := recover(); errR != nil {
					var errStr string
					switch v := errR.(type) {
					case error:
						errStr = v.Error()
						logger.Errorf("%+v", errors.WithStack(v))
					default:
						errStr = fmt.Sprintf("%v", v)
						logger.Errorln(v)
					}
					gr := new(pb.ConfigUpdateResponse)
					gr.Request = res.Request
					gr.Status = false
					gr.Detail = errStr
					if err := stream.Send(gr); err != nil {
						errCtx := logger.NewErrorContext(newCtx, err)
						logger.WithContext(errCtx).Errorf("配置更新: 请求结果返回到驱动管理错误")
					}
				}
			}()
			gr := new(pb.ConfigUpdateResponse)
			gr.Request = res.Request
			gr.Status = true
			switch res.OpsType {
			case pb.ConfigUpdateRequest_EDIT_DRIVER:
				var cfg entity.Instance
				if err := json.Unmarshal(res.GetEditDriver().GetDriver(), &cfg); err != nil {
					gr.Detail = err.Error()
					gr.Status = false
					if err := stream.Send(gr); err != nil {
						logger.WithContext(logger.NewErrorContext(newCtx, err)).Errorf("配置更新: 解析实例配置错误")
					}
					return
				}
				c.updateDriverCache(cfg)
			case pb.ConfigUpdateRequest_ADD_TABLE:
				var t entity.TableCfg
				tmp := res.GetAddTable()
				if err := json.Unmarshal(tmp.GetTable(), &t); err != nil {
					gr.Detail = err.Error()
					gr.Status = false
					if err := stream.Send(gr); err != nil {
						logger.WithContext(logger.NewErrorContext(newCtx, err)).Errorf("配置更新: 添加表解析配置错误")
					}
					return
				}
				t.Id = tmp.GetTableId()
				c.updateTableCache(t)
			case pb.ConfigUpdateRequest_EDIT_TABLE:
				var t entity.TableCfg
				tmp := res.GetEditTable()
				if err := json.Unmarshal(tmp.GetTable(), &t); err != nil {
					gr.Detail = err.Error()
					gr.Status = false
					if err := stream.Send(gr); err != nil {
						logger.WithContext(logger.NewErrorContext(newCtx, err)).Errorf("配置更新: 更新表解析配置错误")
					}
					return
				}
				t.Id = tmp.GetTableId()
				c.updateTableCache(t)
			case pb.ConfigUpdateRequest_DEL_TABLE:
				t := entity.TableCfg{
					Id:      res.GetDelTable().GetTableId(),
					Devices: []entity.Device{},
				}
				c.updateTableCache(t)
			case pb.ConfigUpdateRequest_ADD_DEVICE:
				c.cacheConfig.set(res.GetAddDeviceData().GetTableId(), res.GetAddDeviceData().GetTableDataId())
			case pb.ConfigUpdateRequest_DEL_DEVICE:
				//c.cacheConfigNum.Delete(res.GetDelDeviceData().GetTableDataId())
				c.cacheConfig.delDevice(res.GetDelDeviceData().GetTableId(), res.GetDelDeviceData().GetTableDataId())
			}
			err := c.driver.ConfigUpdate(newCtx, c.app, res)
			if err != nil {
				gr.Detail = err.Error()
				gr.Status = false
			}
			if err := stream.Send(gr); err != nil {
				logger.WithContext(logger.NewErrorContext(newCtx, err)).Errorf("配置更新: 请求结果返回到驱动管理错误")
			}
		}(res)
	}
}

func (c *Client) updateDriverCache(cfg entity.Instance) {
	c.cacheConfig.clear()
	if cfg.Tables != nil {
		for _, t := range cfg.Tables {
			if t.Devices == nil {
				continue
			}
			for _, device := range t.Devices {
				c.cacheConfig.set(t.Id, device.Id)
			}
		}
	}
}

func (c *Client) updateTableCache(t entity.TableCfg) {
	c.cacheConfig.del(t.Id)
	for _, device := range t.Devices {
		c.cacheConfig.set(t.Id, device.Id)
	}
}
