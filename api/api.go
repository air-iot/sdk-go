package api

import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Client interface {
	// model
	FindModelQuery(query, result interface{}) error
	FindModelById(id string, result interface{}) error
	SaveModel(data, result interface{}) error
	DelModelById(id string, result interface{}) error
	UpdateModelById(id string, data, result interface{}) error
	ReplaceModelById(id string, data, result interface{}) error

	// ext
	FindExtQuery(collection string, query, result interface{}) error
	FindExtById(collection, id string, result interface{}) error
	SaveExt(collection string, data, result interface{}) error
	SaveManyExt(collection string, data, result interface{}) error
	DelExtById(collection, id string, result interface{}) error
	UpdateExtById(collection, id string, data, result interface{}) error
	ReplaceExtById(collection, id string, data, result interface{}) error
}
