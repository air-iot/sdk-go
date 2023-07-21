package entity

type Log struct {
	SerialNo string `json:"serialNo"` // 流水号
	Status   string `json:"status"`   // 日志状态
	UnixTime int64  `json:"time"`     // 日志时间毫秒数
	Desc     string `json:"desc"`     // 描述
}
