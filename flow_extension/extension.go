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
	Schema(app App) (string, error)
	Run(app App, input []byte) (map[string]interface{}, error)
}
