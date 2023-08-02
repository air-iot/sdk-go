package json

import (
	"fmt"
	"github.com/jinzhu/copier"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

// 定义JSON操作
var (
	json          = jsoniter.ConfigCompatibleWithStandardLibrary
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)

// MarshalToString JSON编码为字符串
func MarshalToString(v interface{}) string {
	s, err := jsoniter.MarshalToString(v)
	if err != nil {
		return ""
	}
	return s
}

// CopyByJson json 转换
func CopyByJson(dst, src interface{}) error {
	if dst == nil {
		return fmt.Errorf("dst cannot be nil")
	}
	if src == nil {
		return fmt.Errorf("src cannot be nil")
	}
	bytes, err := Marshal(src)
	if err != nil {
		return fmt.Errorf("unable to marshal src: %s", err.Error())
	}

	if err := Unmarshal(bytes, dst); err != nil {
		return fmt.Errorf("unable to unmarshal into dst: %s", err.Error())
	}
	return nil
}

// MapToStruct map 转 struct
func MapToStruct(dst, src interface{}) error {
	return mapstructure.Decode(src, dst)
}

// Copy struct复制
func Copy(dst, src interface{}) error {
	return copier.Copy(dst, src)
}

// MapCopy map 转 map
func MapCopy(src map[string]interface{}) (dst map[string]interface{}) {
	dst = make(map[string]interface{})
	for k, v := range src {
		dst[k] = v
	}
	return
}
