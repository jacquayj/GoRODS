/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"time"
)

// AccessObject is an interface for Users and Groups, used within ACL slices to denote the access level of a DataObj or Collection
type AccessObject interface {
	Name() string
	Zone() *Zone
	Comment() (string, error)
	CreateTime() (time.Time, error)
	ModifyTime() (time.Time, error)
	Id() (int, error)
	Type() (int, error)
	Con() *Connection
}

// ACL is used to describe the access level that a particular AccessObject (User/Group) has on a DataObj or Collection
type ACL struct {
	AccessObject AccessObject
	AccessLevel  int
	Type         int
}

// ACLs is a slice of ACL pointers
type ACLs []*ACL

// User is a shortcut to cast the AccessObject as it's underlying data structure type (*User)
func (acl *ACL) User() *User {
	if acl.Type == UserType || acl.Type == AdminType || acl.Type == GroupAdminType {
		return acl.AccessObject.(*User)
	}

	return nil
}

// Group is a shortcut to cast the AccessObject as it's underlying data structure type (*Group)
func (acl *ACL) Group() *Group {
	if acl.Type == GroupType {
		return acl.AccessObject.(*Group)
	}

	return nil
}

// String returns a formatted string describing the ACL struct
// example: g:designers#tempZone:read
func (acl *ACL) String() string {
	typeString := getTypeString(acl.Type)

	return fmt.Sprintf("%v:%v#%v:%v", typeString, acl.AccessObject.Name(), acl.AccessObject.Zone().Name(), getTypeString(acl.AccessLevel))
}
