package tools

import (
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"fmt"
	"runtime/debug"
)

// Recover is method to use with defer statement
func Recover(lg logger.Logger) {
	if r := recover(); r != nil {
		if lg != nil {
			switch t := r.(type) {
			case error:
				if !errors.HasStack(t) {
					r = errors.WrapStack(t)
				}
			}
			lg.Error("Recover panic", r)
		} else {
			var stack string

			switch t := r.(type) {
			case error:
				if !errors.HasStack(t) {
					stack = errors.GetStack(t)
				} else {
					stack = string(debug.Stack())
				}
			default:
				stack = string(debug.Stack())
			}

			fmt.Printf("Recover panic: %v stack: %v", r, stack)
		}
	}
}
