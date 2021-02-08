package logger

import (
	"DaruBot/internal/config"
	"DaruBot/pkg/logger"
	"github.com/op/go-logging"
	"github.com/sirupsen/logrus"
	"os"
)

func NewLogger(c config.Configurations) logger.Logger {
	w := os.Stdout

	if c.Logger.FileOutput {
		// TODO logs to file
		//w = io.MultiWriter(os.Stdout, fileOut)
	}

	level := logger.DebugLevel
	if !c.IsDebug() {
		level = logger.InfoLevel
	}

	lg := logger.New(w, level)

	return lg
}

type goLogging struct {
	lvl logging.Level
	lg  logger.Logger
}

func (g *goLogging) Log(level logging.Level, callDepth int, record *logging.Record) error {
	//if level > g.lvl {
	//	return nil
	//}

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

func (g *goLogging) GetLevel(module string) logging.Level {
	return g.lvl
}

func (g *goLogging) SetLevel(Level logging.Level, module string) {
	//g.lvl = Level
}

func (g *goLogging) IsEnabledFor(Level logging.Level, module string) bool {
	return Level <= g.lvl
}

func ConvertToGoLogging(logger logger.Logger, Level logging.Level) logging.LeveledBackend {
	return &goLogging{
		lvl: Level,
		lg:  logger,
	}
}
