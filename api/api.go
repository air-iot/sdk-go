package api

import (
	"io"
	"net/url"
)

type Client interface {
	Host() string
	AK() string
	SK() string
	Get(url url.URL, result interface{}) error
	Post(url url.URL, data, result interface{}) error
	Delete(url url.URL, result interface{}) error
	Put(url url.URL, data, result interface{}) error
	Patch(url url.URL, data, result interface{}) error

	// model
	FindModelQuery(query, result interface{}) error
	FindModelById(id string, result interface{}) error
	SaveModel(data, result interface{}) error
	DelModelById(id string, result interface{}) error
	UpdateModelById(id string, data, result interface{}) error
	ReplaceModelById(id string, data, result interface{}) error

	// node
	FindNodeQuery(query, result interface{}) error
	FindNodeById(id string, result interface{}) error
	SaveNode(data, result interface{}) error
	DelNodeById(id string, result interface{}) error
	UpdateNodeById(id string, data, result interface{}) error
	ReplaceNodeById(id string, data, result interface{}) error

	// user
	FindUserQuery(query, result interface{}) error
	FindUserById(id string, result interface{}) error
	SaveUser(data, result interface{}) error
	DelUserById(id string, result interface{}) error
	UpdateUserById(id string, data, result interface{}) error
	ReplaceUserById(id string, data, result interface{}) error

	// handler
	FindHandlerQuery(query, result interface{}) error
	FindHandlerById(id string, result interface{}) error
	SaveHandler(data, result interface{}) error
	DelHandlerById(id string, result interface{}) error
	UpdateHandlerById(id string, data, result interface{}) error
	ReplaceHandlerById(id string, data, result interface{}) error

	// ext
	FindExtQuery(collection string, query, result interface{}) error
	FindExtById(collection, id string, result interface{}) error
	SaveExt(collection string, data, result interface{}) error
	SaveManyExt(collection string, data, result interface{}) error
	DelExtById(collection, id string, result interface{}) error
	UpdateExtById(collection, id string, data, result interface{}) error
	ReplaceExtById(collection, id string, data, result interface{}) error

	// event
	FindEventQuery(query, result interface{}) error
	FindEventById(id string, result interface{}) error
	SaveEvent(data, result interface{}) error
	DelEventById(id string, result interface{}) error
	UpdateEventById(id string, data, result interface{}) error
	ReplaceEventById(id string, data, result interface{}) error

	// setting
	FindSettingQuery(query, result interface{}) error
	FindSettingById(id string, result interface{}) error
	SaveSetting(data, result interface{}) error
	DelSettingById(id string, result interface{}) error
	UpdateSettingById(id string, data, result interface{}) error
	ReplaceSettingById(id string, data, result interface{}) error

	// table
	FindTableQuery(query, result interface{}) error
	FindTableById(id string, result interface{}) error
	SaveTable(data, result interface{}) error
	DelTableById(id string, result interface{}) error
	UpdateTableById(id string, data, result interface{}) error
	ReplaceTableById(id string, data, result interface{}) error

	// data
	GetLatest(query interface{}) (result []RealTimeData, err error)
	PostLatest(data interface{}) (result []RealTimeData, err error)
	GetQuery(query interface{}) (result *QueryData, err error)
	PostQuery(query interface{}) (result *QueryData, err error)

	// 报警
	FindWarnQuery(archive bool, query, result interface{}) error
	FindWarnById(id string, archive bool, result interface{}) error
	SaveWarn(data, archive bool, result interface{}) error
	DelWarnById(id string, archive bool, result interface{}) error
	UpdateWarnById(id string, archive bool, data, result interface{}) error
	ReplaceWarnById(id string, archive bool, data, result interface{}) error

	// driver
	ChangeCommand(id string, data, result interface{}) error
	DriverConfig(projectId, driverId, serviceId string) ([]byte, error)

	// GetMediaFile 获取媒体库中的指定 path 文件
	//
	// 注: 返回值中的 MediaFileInfo.Data 需要手动关闭
	GetMediaFile(path string) (*MediaFileInfo, error)

	// UploadMediaFile 向媒体库上传文件
	// 上传成功后返回该文件的 url, 否则返回错误信息
	//
	// filename: 文件名称
	// catalog: 目录
	// action: 当文件存在时的处理方式. cover: 覆盖, rename: 文件名自动加1
	UploadMediaFile(filename, catalog, action string, reader io.ReadCloser) (string, error)

	// DeleteMediaFile 删除指定路径的文件
	//
	// path: 文件路径, 例如: /chenpc/media_upload_file.txt
	// completeDelete: 是否彻底删除文件
	DeleteMediaFile(path string, completeDelete bool) error
}
