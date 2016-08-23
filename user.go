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

func (usr *User) GetInfo() (map[string]string, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cUser := C.CString(usr.Name)
	defer C.free(unsafe.Pointer(cUser))

	ccon := usr.Con.GetCcon()
	defer usr.Con.ReturnCcon(ccon)

	if status := C.gorods_get_user(cUser, ccon, &result, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Users Failed: %v", C.GoString(err)))
	}

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	//response := make(Users, 0)
	response := make(map[string]string)

	for _, userInfo := range slice {

		userAttributes := strings.Split(strings.Trim(C.GoString(userInfo), " \n"), "\n")

		for _, attr := range userAttributes {

			split := strings.Split(attr, ": ")

			attrName := split[0]
			attrVal := split[1]

			response[attrName] = attrVal

		}

	}

	C.gorods_free_string_result(&result)

	return response, nil
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
