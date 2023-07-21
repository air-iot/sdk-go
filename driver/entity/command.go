package entity

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
