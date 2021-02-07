package utils

import (
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"fmt"
	"runtime/debug"
)

// Recover is method to use with defer statement
func Recover(lg logger.Logger) {
	if r := recover(); r != nil {

		var stack string

		switch t := r.(type) {
		case error:
			stack = errors.GetStack(t)
		default:
			stack = string(debug.Stack())
		}

		if lg != nil {
			lg.Error("Recover panic", r, "stack", stack)
		} else {
			fmt.Printf("Recover panic: %v stack: %v", r, stack)
		}
	}
}
