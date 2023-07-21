package driver

import (
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"net/http"
)

type Driver interface {
	Start(App, []byte) error
	Run(app App, cmd *entity.Command) (interface{}, error)
	BatchRun(app App, cmd *entity.BatchCommand) (interface{}, error)
	WriteTag(app App, cmd *entity.Command) (interface{}, error)
	Debug(App, []byte) (interface{}, error)
	HttpProxy(app App, t string, header http.Header, data []byte) (interface{}, error)
	Stop(App) error
	Schema(App) (string, error)
}
