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

type ACL struct {
	Name       string
	Zone       string
	DataAccess string
	ACLType    string
}

type ACLs []*ACL

func (acl ACL) String() string {
	typeString := ""

	if acl.ACLType == "group" {
		typeString = "g:"
	} else if acl.ACLType == "user" {
		typeString = "u:"
	}

	return fmt.Sprintf("%v%v#%v:%v", typeString, acl.Name, acl.Zone, acl.DataAccess)
}
