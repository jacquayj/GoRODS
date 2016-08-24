/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"strconv"
	"time"
)

func cTimeToTime(cTime *C.char) time.Time {
	unixStamp, _ := strconv.ParseInt(C.GoString(cTime), 10, 64)
	return time.Unix(unixStamp, 0)
}

func TimeStringToTime(ts string) time.Time {
	unixStamp, _ := strconv.ParseInt(ts, 10, 64)
	return time.Unix(unixStamp, 0)
}

func getTypeString(t int) string {
	switch t {
	case DataObjType:
		return "d"
	case CollectionType:
		return "C"
	case ResourceType:
		return "R"
	case UserType:
		return "u"
	default:
		panic(newError(Fatal, "unrecognized meta type constant"))
	}
}

func isString(obj interface{}) bool {
	switch obj.(type) {
	case string:
		return true
	default:
	}

	return false
}
