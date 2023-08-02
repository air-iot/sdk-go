package algorithm

type RunConfig struct {
	ProjectID string                 `json:"projectID"`
	Function  string                 `json:"function"`
	Input     map[string]interface{} `json:"input"`
}

type grpcResult struct {
	Code   int         `json:"code"`
	Error  string      `json:"error"`
	Result interface{} `json:"result"`
}
