package sdk

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// Init Init("", "INFO")
func InitLog(logPath string, tmpLevel logrus.Level) {

	var writer io.Writer
	var err error

	if logPath == "" {
		writer = os.Stdout
	} else {
		os.Mkdir(logPath, 0777)

		writer, err = rotatelogs.New(
			path.Join(logPath, "%Y%m%d.log"),
			rotatelogs.WithMaxAge(time.Hour*time.Duration(365*24*2)), // 文件最大保存时间
			rotatelogs.WithRotationTime(time.Hour*time.Duration(24)), // 日志切割时间间隔
		)

		if err != nil {
			panic(err)
		}
	}

	//lfHook := lfshook.NewHook(writerMap, &logrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logrus.SetOutput(writer)
	logrus.SetLevel(tmpLevel)
}

// Logger 日志
type Logger struct {
	*logrus.Logger
}

// New 创建logger
func New(writer io.Writer, tmpLevel logrus.Level) *Logger {

	return &Logger{Logger: &logrus.Logger{
		Out:       writer,
		Formatter: &logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"},
		Level:     tmpLevel,
	}}
}

// Debugln 调试输出
func (p *Logger) Debugln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		p.Logger.Debugln(args...)
	} else {
		p.Logger.WithFields(fields).Debugln(args...)
	}
}

// Debugf 调试输出
func (p *Logger) Debugf(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		p.Logger.Debugf(format, args...)
	} else {
		p.Logger.WithFields(fields).Debugf(format, args...)
	}
}

// Infoln 信息输出
func (p *Logger) Infoln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		p.Logger.Infoln(args...)
	} else {
		p.Logger.WithFields(fields).Infoln(args...)
	}
}

// Infof 信息输出
func (p *Logger) Infof(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		p.Logger.Infof(format, args...)
	} else {
		p.Logger.WithFields(fields).Infof(format, args...)
	}
}

// Warnln 告警输出
func (p *Logger) Warnln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		p.Logger.Warnln(args...)
	} else {
		p.Logger.WithFields(fields).Warnln(args...)
	}
}

// Warnf 告警输出
func (p *Logger) Warnf(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		p.Logger.Warnf(format, args...)
	} else {
		p.Logger.WithFields(fields).Warnf(format, args...)
	}
}

// Errorln 错误输出
func (p *Logger) Errorln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		p.Logger.Errorln(args...)
	} else {
		p.Logger.WithFields(fields).Errorln(args...)
	}
}

// Errorf 错误输出
func (p *Logger) Errorf(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		p.Logger.Errorf(format, args...)
	} else {
		p.Logger.WithFields(fields).Errorf(format, args...)
	}
}

// Debugln 调试输出
func Debugln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		logrus.Debugln(args...)
	} else {
		logrus.WithFields(fields).Debugln(args...)
	}
}

// Debugf 调试输出
func Debugf(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		logrus.Debugf(format, args...)
	} else {
		logrus.WithFields(fields).Debugf(format, args...)
	}
}

// Infoln 信息输出
func Infoln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		logrus.Infoln(args...)
	} else {
		logrus.WithFields(fields).Infoln(args...)
	}
}

// Infof 信息输出
func Infof(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		logrus.Infof(format, args...)
	} else {
		logrus.WithFields(fields).Infof(format, args...)
	}
}

// Warnln 告警输出
func Warnln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		logrus.Warnln(args...)
	} else {
		logrus.WithFields(fields).Warnln(args...)
	}
}

// Warnf 告警输出
func Warnf(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		logrus.Warnf(format, args...)
	} else {
		logrus.WithFields(fields).Warnf(format, args...)
	}
}

// Errorln 错误输出
func Errorln(fields map[string]interface{}, args ...interface{}) {
	if fields == nil {
		logrus.Errorln(args...)
	} else {
		logrus.WithFields(fields).Errorln(args...)
	}
}

// Errorf 错误输出
func Errorf(fields map[string]interface{}, format string, args ...interface{}) {
	if fields == nil {
		logrus.Errorf(format, args...)
	} else {
		logrus.WithFields(fields).Errorf(format, args...)
	}
}
