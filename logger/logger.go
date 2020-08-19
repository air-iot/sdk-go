package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

func NewLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	l, err := logrus.ParseLevel(level)
	if err != nil {
		l = logrus.ErrorLevel
	}
	logger.SetLevel(l)
	return logger
}
