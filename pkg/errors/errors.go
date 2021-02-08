package errors

import (
	"errors"
	"fmt"
	"runtime/debug"

	werr "github.com/pkg/errors"
)

/*
-------------------------------
	Wrapper
-------------------------------
*/

// WrapStack add StackTrace and wrap error with message if it passed
// If error have StackTrace, new stack trace will be ignoring
func WrapStack(err error, msgs ...interface{}) error {
	if err != nil {
		msg := ""
		if !HasStack(err) {
			err = werr.WithStack(err)
		}
		for _, m := range msgs {
			msg = fmt.Sprintf("%+v; ", m)
		}
		err = werr.WithMessage(err, msg)
	}
	return err
}

// WrapMessage wrap error with message if it passed
// without StackTrace
func WrapMessage(err error, msgs ...interface{}) error {
	if err != nil {
		msg := ""
		for _, m := range msgs {
			msg = fmt.Sprintf("%+v; ", m)
		}
		err = werr.WithMessage(err, msg)
	}
	return err
}

/*
-------------------------------
	Wrapper tools
-------------------------------
*/

// Cause Return first error, anyway, error was wrapped or not
func Cause(err error) error {
	return werr.Cause(err)
}

//  IsWrapped If error from standard pkg was wrapped
func IsWrapped(err error) bool {
	if _, ok := err.(interface{ Cause() error }); ok {
		return true
	}
	return false
}

// HasStack Return true if StackTrace exist
func HasStack(err error) bool {
	type stackTracer interface {
		StackTrace() werr.StackTrace
	}
	_, ok := err.(stackTracer)
	return ok
}

// UnWrapStack Return StackTrace if it has exist
func UnWrapStack(err error) string {
	type stackTracer interface {
		StackTrace() werr.StackTrace
	}
	e, ok := err.(stackTracer)
	if ok {
		st := e.StackTrace()
		if len(st) > 0 {
			return fmt.Sprintf("%+v", st)
		}
	}

	return ""
}

// GetStack always return stack
// if there is not wrapped stack then return runtime/debug
func GetStack(err error) string {
	if stack := UnWrapStack(err); stack != "" {
		return stack
	}

	return string(debug.Stack())
}

// New returns an error with the supplied message.
func New(msg string) error {
	return errors.New(msg)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error. Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return werr.Errorf(format, args...)
}
