/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

// Log level constants
const (
	Info = iota
	Warn
	Fatal
)

// GoRodsError stores information about errors
type GoRodsError struct {
	LogLevel  int
	Message   string
	IRODSCode string
	Time      time.Time
}

// Error returns error string, alias of String(). Sample output:
//
// 	2016-04-22 10:02:30.802355258 -0400 EDT: Fatal - iRODS Connect Failed: rcConnect failed
func (err *GoRodsError) Error() string {
	return err.String()
}

// String returns error string. Sample output:
//
// 	2016-04-22 10:02:30.802355258 -0400 EDT: Fatal - iRODS Connect Failed: rcConnect failed
func (err *GoRodsError) String() string {
	return fmt.Sprintf("%v: %v - %v%v", err.Time, err.lookupError(err.LogLevel), err.Message, err.IRODSCode)
}

func (err *GoRodsError) lookupError(code int) string {
	var constLookup = map[int]string{
		Info:  "Info",
		Warn:  "Warn",
		Fatal: "Fatal",
	}

	return constLookup[code]
}

func newError(logLevel int, status C.int, message string) *GoRodsError {
	var (
		errStr    *C.char
		subErrStr *C.char
	)

	err := new(GoRodsError)

	err.LogLevel = logLevel
	err.Message = message
	err.Time = time.Now()

	if status != -1 {
		defer C.free(unsafe.Pointer(errStr))
		defer C.free(unsafe.Pointer(subErrStr))

		errStr = C.rodsErrorName(status, &subErrStr)

		err.IRODSCode = " " + C.GoString(errStr) + " " + C.GoString(subErrStr)
	}

	return err
}
