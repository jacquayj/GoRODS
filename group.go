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

func (grp Group) String() string {
	return fmt.Sprintf("%v", grp.Name)
}

func (grp *Group) GetUsers() (Users, error) {

	var (
		result C.goRodsGroupResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cGroupName := C.CString(grp.Name)
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := grp.Con.GetCcon()
	defer grp.Con.ReturnCcon(ccon)

	if status := C.gorods_get_group(ccon, &result, cGroupName, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Group %v Failed: %v", grp.Name, C.GoString(err)))
	}

	unsafeArr := unsafe.Pointer(result.grpArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(Users, 0)

	for _, userNames := range slice {

		usrFrags := strings.Split(C.GoString(userNames), "#")

		response = append(response, &User{
			Name: usrFrags[0],
			Zone: usrFrags[1],
		})

	}

	C.gorods_free_group_result(&result)

	return response, nil

}
