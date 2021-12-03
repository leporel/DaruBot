package logger

import (
	"DaruBot/pkg/errors"
	"fmt"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	writer := os.Stdout
	lg := New(writer, DebugLevel)

	lg.Info("test information")
	lg.Error("test error")
	lg.Debug("test debug")
	lg.Warn("test warning")

	lg = lg.WithPrefix("Module", "Core").(*logrusLogger)
	lg.Info("test information")
}

func TestNewError(t *testing.T) {
	writer := os.Stdout
	lg := New(writer, DebugLevel)

	err := errors.New("some error")

	lg.Error("Something go wrong", "err:", err)

	err = errors.WrapStack(err)

	lg.Error("Something go wrong", "err:", err)
}

type hookTest struct {
}

func (hk *hookTest) Fire(hd *HookData) error {
	var fields string

	for nm, f := range hd.Fields {
		fields = fmt.Sprintf("%s\n%s=%v", fields, nm, f)
	}

	fmt.Printf("[%s] [%s]: %s %s \n",
		hd.Time.Format("01.02 15:04:05"), hd.Level, hd.Message, fields)

	return nil
}

func TestHook(t *testing.T) {
	writer := os.Stdout
	lg := New(writer, DebugLevel).WithPrefix("pfx1", "awesome module")
	lg.AddHook(&hookTest{}, ErrorLevel, InfoLevel, DebugLevel)

	lg.Error("Something go wrong")
	lg.Info("Something happen")
}
