package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var levelMap = map[string]logrus.Level{
	"DEBUG": logrus.DebugLevel,
	"INFO":  logrus.InfoLevel,
	"WARN":  logrus.WarnLevel,
	"ERROR": logrus.ErrorLevel,
}

func NewLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logLevel, ok := levelMap[strings.ToUpper(level)]
	if !ok {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	return logger
}
