package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	l, err := logrus.ParseLevel(level)
	if err != nil {
		l = logrus.ErrorLevel
	}
	logger.SetLevel(l)
	logrus.SetLevel(l)
	return logger
}
