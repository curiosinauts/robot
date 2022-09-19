package main

import (
	"fmt"

	"github.com/fatih/color"
)

// MessageLogger logger specialized for slack messages
type MessageLogger struct {
	debug  bool
	isatty bool
}

// Log logs message
func (ml *MessageLogger) Log(subject, msg string) {
	if ml.isatty {
		titlef := color.New(color.FgGreen).Add(color.Bold).SprintFunc()
		fmt.Printf(logFormat, titlef(subject), msg)
		return
	}

	fmt.Printf(logFormat, subject, msg)
}

// SetDebug updates flag for enabling logging
func (ml *MessageLogger) SetDebug(debug bool) {
	ml.debug = debug
}

var logFormat = "%-25s: %s\n"

// Debug debug
func (ml *MessageLogger) Debug(subject, msg string) {
	if ml.debug {
		if ml.isatty {
			titlef := color.New(color.FgBlue).Add(color.Bold).SprintFunc()
			fmt.Printf(logFormat, titlef(subject), msg)
			return
		}

		fmt.Printf(logFormat, subject, msg)
	}
}

// Error error
func (ml *MessageLogger) Error(subject, msg string) {
	if ml.debug {
		if ml.isatty {
			titlef := color.New(color.FgRed).Add(color.Bold).SprintFunc()
			fmt.Printf(logFormat, titlef(subject), msg)
			return
		}

		fmt.Printf(logFormat, subject, msg)
	}
}
