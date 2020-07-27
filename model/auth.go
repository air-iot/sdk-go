package model

type Auth struct {
	Expires int64  `json:"expires"`
	Token   string `json:"token"`
}
