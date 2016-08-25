/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
)

type AccessObject interface {
	GetName() string
	GetZone() string
}

type ACL struct {
	AccessObject AccessObject
	AccessLevel  int
	Type         int
}

type ACLs []*ACL

func (acl ACL) GetTypeString(typ int) string {
	switch typ {
	case UserType:
		return "u"
	case GroupType:
		return "g"
	case UnknownType:
		return "?"
	case Null:
		return "null"
	case Read:
		return "read"
	case Write:
		return "write"
	case Own:
		return "own"
	default:
		return ""
	}
}

func (acl ACL) String() string {
	typeString := acl.GetTypeString(acl.Type)

	return fmt.Sprintf("%v:%v#%v:%v", typeString, acl.AccessObject.GetName(), acl.AccessObject.GetZone(), acl.GetTypeString(acl.AccessLevel))
}
