package driver

// Point 存储数据
type Point struct {
	Table      string            `json:"table"`      // 表id
	ID         string            `json:"id"`         // 设备编号
	CID        string            `json:"cid"`        // 子设备编号
	Fields     []Field           `json:"fields"`     // 数据点
	UnixTime   int64             `json:"time"`       // 数据采集时间 毫秒数
	FieldTypes map[string]string `json:"fieldTypes"` // 数据点类型
}

type Event struct {
	Table    string      `json:"table"`   // 表id
	ID       string      `json:"id"`      // 设备编号
	EventID  string      `json:"eventId"` // 事件ID
	UnixTime int64       `json:"time"`    // 数据采集时间 毫秒数
	Data     interface{} `json:"data"`    // 事件数据
}

type Log struct {
	SerialNo string `json:"serialNo"` // 流水号
	Status   string `json:"status"`   // 日志状态
	UnixTime int64  `json:"time"`     // 日志时间毫秒数
	Desc     string `json:"desc"`     // 描述
}

type TableData struct {
	TableID string                 `json:"table"` // 表id
	ID      string                 `json:"id"`    // 设备编号
	Data    map[string]interface{} `json:"data"`
}

// Field 字段
type Field struct {
	Tag   interface{} `json:"tag"`   // 数据点
	Value interface{} `json:"value"` // 数据采集值
}

type point struct {
	ID         string                 `json:"id"`
	CID        string                 `json:"cid"`    // 子设备编号
	Source     string                 `json:"source"` // 标识源类型
	Fields     map[string]interface{} `json:"fields"`
	UnixTime   int64                  `json:"time"`
	FieldTypes map[string]string      `json:"fieldTypes"` // 数据点类型
}

type Command struct {
	Table    string `json:"table"`
	Id       string `json:"id"`
	SerialNo string `json:"serialNo"`
	Command  []byte `json:"command"`
}

type BatchCommand struct {
	Table    string   `json:"table"`
	Ids      []string `json:"ids"`
	SerialNo string   `json:"serialNo"`
	Command  []byte   `json:"command"`
}

type grpcResult struct {
	Code   int         `json:"code"`
	Error  string      `json:"error"`
	Result interface{} `json:"result"`
}
