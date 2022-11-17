package api

import (
	"io"
	"io/ioutil"
)

type AuthToken struct {
	ExpiresAt int64  `json:"expiresAt"`
	Token     string `json:"token"`
}

// 采集数据
type (
	RealTimeData struct {
		TagId string      `json:"tagId"`
		Uid   string      `json:"uid"`
		Time  int64       `json:"time"`
		Value interface{} `json:"value"`
	}

	QueryData struct {
		Results []Results `json:"results"`
	}

	Series struct {
		Name    string          `json:"name"`
		Columns []string        `json:"columns"`
		Values  [][]interface{} `json:"values"`
	}

	Results struct {
		Series []Series `json:"series"`
	}
)

// MediaFileInfo 媒体库中文件信息
type MediaFileInfo struct {
	// 文件名
	Name string `json:"name"`
	// 文件大小
	Size int64 `json:"size"`
	// 文件流
	reader io.ReadCloser
}

// Read 读取数据到缓冲区
func (f *MediaFileInfo) Read(buf []byte) (int, error) {
	return f.reader.Read(buf)
}

// ReadAll 取取全部数据
func (f *MediaFileInfo) ReadAll() ([]byte, error) {
	if f.reader == nil {
		return nil, io.EOF
	}
	data, err := ioutil.ReadAll(f.reader)
	f.reader = nil
	return data, err
}

func (f *MediaFileInfo) Close() error {
	if f.reader == nil {
		return io.EOF
	}

	err := f.reader.Close()
	f.reader = nil
	return err
}
