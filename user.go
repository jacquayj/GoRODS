/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
)

type User struct {
	Name   string
	Zone   string
	Groups Groups
	Con    *Connection
}

type Users []*User

func (usr *User) String() string {
	return fmt.Sprintf("%v#%v", usr.Name, usr.Zone)
}

func (usr *User) Info() string {
	return ""
}

func AddToGroup(userName string, zoneName string, groupName string, con *Connection) error {

	var (
		err *C.char
	)

	cUserName := C.CString(userName)
	cZoneName := C.CString(zoneName)
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_add_user_to_group(cUserName, cZoneName, cGroupName, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods AddToGroup %v Failed: %v", groupName, C.GoString(err)))
	}

	return nil
}

func RemoveFromGroup(userName string, zoneName string, groupName string, con *Connection) error {
	return newError(Fatal, fmt.Sprintf(""))
}

func (usr *User) AddToGroup(grp interface{}) error {

	switch grp.(type) {
	case string:
		return AddToGroup(usr.Name, usr.Zone, grp.(string), usr.Con)
	case *Group:
		return AddToGroup(usr.Name, usr.Zone, (grp.(*Group)).Name, usr.Con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods AddToGroup Failed: unknown type passed"))
}

func (usr *User) RemoveFromGroup(grp interface{}) error {
	switch grp.(type) {
	case string:
		return RemoveFromGroup(usr.Name, usr.Zone, grp.(string), usr.Con)
	case *Group:
		return RemoveFromGroup(usr.Name, usr.Zone, (grp.(*Group)).Name, usr.Con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods RemoveFromGroup Failed: unknown type passed"))
}
