package logger

import (
	"DaruBot/pkg/errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Level uint32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Log(logrus.Level, string)

	AddHook(hooks []Hook)

	WithPrefix(k string, v interface{}) Logger
}

type logrusLogger struct {
	log *logrus.Entry
}

func (l *logrusLogger) Log(lvl logrus.Level, msg string) {
	l.log.Log(lvl, msg)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.log.Infoln(args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.log.Warnln(args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.log.Debugln(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.log.Warnf(format, args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	var hasStack bool // Always print stack when call Error
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			if errors.HasStack(err) {
				args = append(args, errors.GetStack(err))
				hasStack = true
			}
			break
		}
	}

	if !hasStack {
		args = append(args, errors.GetStack(nil))
	}

	l.log.Errorln(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	var hasStack bool // Always print stack when call Error
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			if errors.HasStack(err) {
				args = append(args, errors.GetStack(err))
				hasStack = true
			}
			break
		}
	}

	if !hasStack {
		args = append(args, errors.GetStack(nil))
	}

	l.log.Errorf(format, args...)
}

func (l *logrusLogger) WithPrefix(k string, v interface{}) Logger {
	return &logrusLogger{
		log: l.log.WithField(k, v),
	}
}

func (l logrusLogger) AddHook(hooks []Hook) {
	// TODO wrap hooks to logrus.Hook with level Error and Warning
	//	l.log.Logger.AddHook()
}

func New(writer io.Writer, level Level) *logrusLogger {
	log := logrus.New()

	log.SetOutput(writer)
	// log.SetReportCaller(true) // comment until logrus dont have ability set level of pkg, other way we print stack trace every time when call log.Error()
	log.SetNoLock()
	log.SetLevel(logrus.Level(level))

	var formatter logrus.Formatter
	formatter = &logrus.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, "/")
			funcname := s[len(s)-1]

			return fmt.Sprintf("%s()", funcname), fmt.Sprintf(" %s:%d", filepath.Base(f.File), f.Line)
		},
	}

	log.SetFormatter(formatter)

	entry := logrus.NewEntry(log)
	return &logrusLogger{
		log: entry,
	}
}

type HookData struct {
	Time    time.Time
	Level   Level
	Caller  *runtime.Frame
	Message string
	err     string
}

type Hook interface {
	Fire(*HookData) error
}
