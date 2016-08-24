/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

type Group struct {
	Name  string
	Users Users
	Con   *Connection
}

type Groups []*Group

func (grp *Group) String() string {
	return fmt.Sprintf("%v", grp.Name)
}

func (grp *Group) GetUsers() (Users, error) {

	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cGroupName := C.CString(grp.Name)
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := grp.Con.GetCcon()

	if status := C.gorods_get_group(ccon, &result, cGroupName, &err); status != 0 {
		grp.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Group %v Failed: %v", grp.Name, C.GoString(err)))
	}

	grp.Con.ReturnCcon(ccon)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	// ensure users are loaded
	if len(grp.Con.Users) == 0 {
		grp.Con.RefreshUsers()
	}

	response := make(Users, 0)

	for _, userNames := range slice {

		usrFrags := strings.Split(C.GoString(userNames), "#")

		if usr := grp.Con.Users.FindByName(usrFrags[0]); usr != nil {
			response = append(response, usr)
		} else {
			return nil, newError(Fatal, fmt.Sprintf("iRods GetUsers Failed: User in response not found in cache"))
		}

	}

	C.gorods_free_string_result(&result)

	return response, nil

}

func (grp *Group) AddUser(usr interface{}) error {

	switch usr.(type) {
	case string:
		// Need to lookup user by string in cache for zone info

		// ensure users are loaded
		if len(grp.Con.Users) == 0 {
			grp.Con.RefreshUsers()
		}

		usrName := usr.(string)

		if existingUsr := grp.Con.Users.FindByName(usrName); existingUsr != nil {
			zoneName := existingUsr.Zone
			return AddToGroup(usrName, zoneName, grp.Name, grp.Con)
		} else {
			return newError(Fatal, fmt.Sprintf("iRods AddUser Failed: can't find iRODS user by string"))
		}

	case *User:
		aUsr := usr.(*User)
		return AddToGroup(aUsr.Name, aUsr.Zone, grp.Name, aUsr.Con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods AddUser Failed: unknown type passed"))
}

// func (grp *Group) RemoveUser(usr interface{}) error {
// 	switch grp.(type) {
// 	case string:
// 		return RemoveFromGroup(usr.Name, usr.Zone, grp.(string), usr.Con)
// 	case *Group:
// 		return RemoveFromGroup(usr.Name, usr.Zone, (grp.(*Group)).Name, usr.Con)
// 	default:
// 	}

// 	return newError(Fatal, fmt.Sprintf("iRods RemoveFromGroup Failed: unknown type passed"))
// }

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
	// Implement me!

	return newError(Fatal, fmt.Sprintf(""))
}
