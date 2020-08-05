package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(level)
	return logger
}
