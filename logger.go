package sdk

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var levelMap = map[string]logrus.Level{
	"DEBUG": logrus.DebugLevel,
	"INFO":  logrus.InfoLevel,
	"WARN":  logrus.WarnLevel,
	"ERROR": logrus.ErrorLevel,
}

// LogLevel 日志等级
var logLevel = logrus.ErrorLevel

func logInit() {
	var tmpLogLevel = viper.GetString("log.level")
	if tmpLogLevel != "" {
		if l, ok := levelMap[strings.ToUpper(tmpLogLevel)]; ok {
			logLevel = l
		}
	}
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logLevel)
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
