package sdk

type Driver interface {
	// 返回驱动的schema
	Start(*DG, []byte) error
	Reload(*DG, []byte) error
	Run(*DG, string, []byte) error
}
