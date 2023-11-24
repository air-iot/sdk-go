package entity

import "time"

type WarnProcessed string

type WarnStatus string

const (
	PROCESSED   WarnProcessed = "已处理"
	UNPROCESSED WarnProcessed = "未处理"
)

const (
	CONFIRMED   WarnStatus = "已确认"
	UNCONFIRMED WarnStatus = "未确认"
)

type Warn struct {
	ID          string        `json:"id"`
	TableId     string        `json:"tableId"`
	TableDataId string        `json:"tableDataId"`
	Level       string        `json:"level"`
	Ruleid      string        `json:"ruleid"`
	Fields      []WarnTag     `json:"fields"`
	WarningType []string      `json:"type"`
	Processed   WarnProcessed `json:"processed"`
	Time        *time.Time    `json:"time"`
	Alert       bool          `json:"alert"`
	Status      WarnStatus    `json:"status"`
	Handle      bool          `json:"handle"`
	Desc        string        `json:"desc"`
}

type WarnSend struct {
	ID          string        `json:"id"`
	Table       Table         `json:"table"`
	TableData   TableData     `json:"tableData"`
	Level       string        `json:"level"`
	Ruleid      string        `json:"ruleid"`
	Fields      []WarnTag     `json:"fields"`
	WarningType []string      `json:"type"`
	Processed   WarnProcessed `json:"processed"`
	Time        string        `json:"time"` //time.RFC3339
	Alert       bool          `json:"alert"`
	Status      WarnStatus    `json:"status"`
	Handle      bool          `json:"handle"`
	Desc        string        `json:"desc"`
}

type WarnTag struct {
	Tag
	Value interface{} `json:"value"`
}

type WarnRecovery struct {
	ID   []string         `json:"id"`
	Data WarnRecoveryData `json:"data"`
}

type WarnRecoveryData struct {
	Time   *time.Time `json:"recoveryTime"`
	Fields []WarnTag  `json:"recoveryFields"`
}

type WarnRecoverySend struct {
	ID   []string             `json:"id"`
	Data WarnRecoveryDataSend `json:"data"`
}

type WarnRecoveryDataSend struct {
	Time   string    `json:"recoveryTime"` //time.RFC3339
	Fields []WarnTag `json:"recoveryFields"`
}
