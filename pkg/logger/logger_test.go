package logger

import (
	"DaruBot/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	writer := os.Stdout
	lg := New(writer, logrus.DebugLevel)

	lg.Info("test information")
	lg.Error("test error")
	lg.Debug("test debug")
	lg.Warn("test warning")

	lg = lg.WithPrefix("Module", "Core").(*logrusLogger)
	lg.Info("test information")
}

func TestNewError(t *testing.T) {
	writer := os.Stdout
	lg := New(writer, logrus.DebugLevel)

	err := errors.New("some error")

	lg.Error("Something go wrong", "err:", err)

	err = errors.WrapStack(err)

	lg.Error("Something go wrong", "err:", err)
}
