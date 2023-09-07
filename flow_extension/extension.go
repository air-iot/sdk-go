package flow_extionsion

type Request struct {
	ProjectId  string `json:"projectId,omitempty"`
	FlowId     string `json:"flowId,omitempty"`
	Job        string `json:"job,omitempty"`
	ElementId  string `json:"elementId,omitempty"`
	ElementJob string `json:"elementJob,omitempty"`
	Config     []byte `json:"config,omitempty"`
}

type Extension interface {
	// Schema
	// @description 查询schema
	// @return schema "驱动配置schema"
	Schema(app App) (schema string, err error)

	// Run
	// @description 执行算法服务
	// @param input 执行参数 {} input 执行参数,应与输出的schema格式相同
	// @return result "自定义返回的格式,应与输出的schema格式相同"
	Run(app App, input []byte) (result map[string]interface{}, err error)
}
