package entity

type TableData struct {
	TableID string                 `json:"table"` // 表id
	ID      string                 `json:"id"`    // 设备编号
	Data    map[string]interface{} `json:"data"`
}
