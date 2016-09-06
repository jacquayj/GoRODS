/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"time"
)

type AccessObject interface {
	GetName() string
	GetZone() *Zone
	GetComment() (string, error)
	GetCreateTime() (time.Time, error)
	GetModifyTime() (time.Time, error)
	GetId() (int, error)
	GetType() (int, error)
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

	return fmt.Sprintf("%v:%v#%v:%v", typeString, acl.AccessObject.GetName(), acl.AccessObject.GetZone().GetName(), getTypeString(acl.AccessLevel))
}
