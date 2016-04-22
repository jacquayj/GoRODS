/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a GoLang wrapper of the iRods C API (iRods client library). 
// gorods uses cgo to call iRods client functions.
package gorods

import (
 	"fmt"
	//"errors"
	"time"
)

const (
	Info = iota
	Warn
	Fatal
)

type GoRodsError struct {
	LogLevel int
	Message string
	Time time.Time
}

func (err *GoRodsError) String() string {
	return fmt.Sprintf("%v: %v - %v", err.Time, err.lookupError(err.LogLevel), err.Message)
}

func (err *GoRodsError) lookupError(code int) string {
	var constLookup = map[int]string {
		Info: "Info",
		Warn: "Warn",
		Fatal: "Fatal",
	}

	return constLookup[code]
}

func newError(logLevel int, message string) *GoRodsError {
	err := new(GoRodsError)

	err.LogLevel = logLevel
	err.Message = message
	err.Time = time.Now()

	return err
}