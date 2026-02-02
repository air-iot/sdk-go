package entity

// DeviceStatus 设备状态
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"  // 在线
	DeviceStatusOffline DeviceStatus = "offline" // 掉线
)

// DeviceStatusInfo 设备状态信息
type DeviceStatusInfo struct {
	Table    string       `json:"table"`    // 表ID
	ID       string       `json:"id"`       // 设备ID
	Status   DeviceStatus `json:"status"`   // 设备状态
	LastSeen int64        `json:"lastSeen"` // 最后一次上数时间（毫秒时间戳）
}

// NetworkSettings 网络配置
type NetworkSettings struct {
	Timeout int `json:"timeout"` // 超时时间（秒），超过此时间没有上数则判定为掉线
}
