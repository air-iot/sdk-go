package flow

import "context"

type Request struct {
	ProjectId  string `json:"projectId,omitempty"`
	FlowId     string `json:"flowId,omitempty"`
	Job        string `json:"job,omitempty"`
	ElementId  string `json:"elementId,omitempty"`
	ElementJob string `json:"elementJob,omitempty"`
	Config     []byte `json:"config,omitempty"`
}

type DebugRequest struct {
	ProjectId string `json:"projectId,omitempty"`
	FlowId    string `json:"flowId,omitempty"`
	ElementId string `json:"elementId,omitempty"`
	Config    []byte `json:"config,omitempty"`
}

type DebugResult struct {
	Logs  []Syslog               `json:"logs"`
	Value map[string]interface{} `json:"value"`
}

type Syslog struct {
	Level string `json:"level"`
	Time  string `json:"time"`
	Msg   string `json:"msg"`
}

type Flow interface {
	// Handler
	// @description 执行流程插件
	// @param request 执行参数 {"projectId":"项目id","flowId":"流程id","job":"流程实例id","elementId":"节点id","elementJob":"节点的实例id","config":{}} config 节点配置
	// @return result "自定义返回的格式或者空"
	Handler(ctx context.Context, app App, request *Request) (result map[string]interface{}, err error)
	Debug(ctx context.Context, app App, request *DebugRequest) (result *DebugResult, err error)
}
