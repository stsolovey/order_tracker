package logger

import (
	"github.com/sirupsen/logrus"
)

func New(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	if level == "" {
		level = "info"
	}

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logger.Warnf("Invalid log level: '%s'. Setting to INFO.", level)

		parsedLevel = logrus.InfoLevel
	}

	logger.SetLevel(parsedLevel)

	return logger
}
