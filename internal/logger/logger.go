package logger

import (
	"DaruBot/config"
	"DaruBot/pkg/logger"
	"github.com/op/go-logging"
	"github.com/sirupsen/logrus"
	"os"
)

func NewLogger(c config.Configurations) logger.Logger {
	w := os.Stdout

	if c.Logger.FileOutput {
		// TODO Make logs to file
		//w = io.MultiWriter(os.Stdout, fileOut)
	}

	level := logrus.DebugLevel
	if !c.IsDebug() {
		level = logrus.InfoLevel
	}

	lg := logger.New(w, level)

	return lg
}

type Hook interface {
	Fire(string) error
}

type goLogging struct {
	lg logger.Logger
}

func (g *goLogging) Log(level logging.Level, callDepth int, record *logging.Record) error {
	var lvl logrus.Level = logrus.InfoLevel

	switch level {
	case logging.CRITICAL:
		lvl = logrus.ErrorLevel
	case logging.ERROR:
		lvl = logrus.ErrorLevel
	case logging.WARNING:
		lvl = logrus.WarnLevel
	case logging.NOTICE:
		lvl = logrus.InfoLevel
	case logging.INFO:
		lvl = logrus.InfoLevel
	case logging.DEBUG:
		lvl = logrus.DebugLevel
	}

	g.lg.Log(lvl, record.Message())

	return nil
}

func ConvertToGoLogging(logger logger.Logger) logging.Backend {
	return &goLogging{
		lg: logger,
	}
}
