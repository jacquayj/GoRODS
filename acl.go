/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"time"
)

type AccessObject interface {
	GetName() string
	GetZone() string
	GetComment() string
	GetCreateTime() time.Time
	GetModifyTime() time.Time
	GetId() int
	GetType() int
	GetCon() *Connection
}

type ACL struct {
	AccessObject AccessObject
	AccessLevel  int
	Type         int
}

type ACLs []*ACL

func (acl *ACL) User() *User {
	if acl.Type == UserType {
		return acl.AccessObject.(*User)
	}

	return nil
}

func (acl *ACL) Group() *Group {
	if acl.Type == GroupType {
		return acl.AccessObject.(*Group)
	}

	return nil
}

func (acl *ACL) String() string {
	typeString := getTypeString(acl.Type)

	return fmt.Sprintf("%v:%v#%v:%v", typeString, acl.AccessObject.GetName(), acl.AccessObject.GetZone(), getTypeString(acl.AccessLevel))
}
