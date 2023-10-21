package entity

// Point 存储数据
type Point struct {
	Table      string            `json:"table"`      // 表id
	ID         string            `json:"id"`         // 设备编号
	CID        string            `json:"cid"`        // 子设备编号
	Fields     []Field           `json:"fields"`     // 数据点
	UnixTime   int64             `json:"time"`       // 数据采集时间 毫秒数
	FieldTypes map[string]string `json:"fieldTypes"` // 数据点类型
}

type WritePoint struct {
	ID         string                 `json:"id"`
	CID        string                 `json:"cid"`    // 子设备编号
	Source     string                 `json:"source"` // 标识源类型
	Fields     map[string]interface{} `json:"fields"`
	UnixTime   int64                  `json:"time"`
	FieldTypes map[string]string      `json:"fieldTypes"` // 数据点类型
}

// Field 字段
type Field struct {
	Tag   Tag         `json:"tag"`   // 数据点
	Value interface{} `json:"value"` // 数据采集值
}
