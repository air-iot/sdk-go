package logger

import (
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// Log 日志配置参数
type Log struct {
	Level         int    `json:"level" yaml:"level"`
	Format        string `json:"format" yaml:"format"`
	Output        string `json:"output" yaml:"output"`
	OutputFile    string `json:"outputFile" yaml:"outputFile"`
	MaxAge        int    `json:"maxAge" yaml:"maxAge"`               // 日志最大保存分钟
	RotationTime  int    `json:"rotationTime" yaml:"rotationTime"`   // 日志分割分钟
	RotationSize  int64  `json:"rotationSize" yaml:"rotationSize"`   // 日志分割文件大小
	RotationCount uint   `json:"rotationCount" yaml:"rotationCount"` // 日志保存个数
}

// NewLogger 创建日志模块
func NewLogger(c Log) (func(), error) {
	SetLevel(c.Level)
	SetFormatter(c.Format)

	// 设定日志输出
	//var file *os.File
	if c.Output != "" {
		switch c.Output {
		case "stdout":
			SetOutput(os.Stdout)
		case "stderr":
			SetOutput(os.Stderr)
		case "file":
			if name := c.OutputFile; name != "" {
				//_ = os.MkdirAll(filepath.Dir(name), 0777)
				//
				//f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
				//if err != nil {
				//	return nil, err
				//}
				//file = f
				writer, err := rotatelogs.New(
					name+".%Y%m%d%H%M",
					rotatelogs.WithLinkName(name), // 生成软链，指向最新日志文件
					rotatelogs.WithRotationSize(c.RotationSize),
					rotatelogs.WithRotationCount(c.RotationCount),
					rotatelogs.WithMaxAge(time.Duration(c.MaxAge)*time.Minute),             // 文件最大保存时间
					rotatelogs.WithRotationTime(time.Duration(c.RotationTime)*time.Minute), // 日志切割时间间隔
				)
				if err != nil {
					return nil, err
				}
				SetOutput(writer)
			}
		}
	}

	return func() {
		//if file != nil {
		//	file.Close()
		//}
	}, nil
}
