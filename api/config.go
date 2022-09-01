package api

type Config struct {
	Schema      string      `json:"schema" yaml:"schema"`
	Host        string      `json:"host" yaml:"host"`
	Port        uint        `json:"port" yaml:"port"`
	Credentials Credentials `json:"credentials" yaml:"credentials"`
}

type Credentials struct {
	AK string `json:"ak" yaml:"ak"`
	SK string `json:"sk" yaml:"sk"`
}
