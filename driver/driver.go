package driver

type Driver interface {
	Start(App, []byte) error
	Run(app App, cmd *Command) (interface{}, error)
	BatchRun(app App, cmd *BatchCommand) (interface{}, error)
	WriteTag(app App, cmd *Command) (interface{}, error)
	Debug(App, []byte) (interface{}, error)
	Stop(App) error
	Schema(App) (string, error)
}
