package entity

type Event struct {
	Table    string      `json:"table"`   // 表id
	ID       string      `json:"id"`      // 设备编号
	EventID  string      `json:"eventId"` // 事件ID
	UnixTime int64       `json:"time"`    // 数据采集时间 毫秒数
	Data     interface{} `json:"data"`    // 事件数据
}
