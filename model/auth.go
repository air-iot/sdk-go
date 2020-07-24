package model

type Auth struct {
	Expires int    `json:"expires"`
	Token   string `json:"token"`
}
