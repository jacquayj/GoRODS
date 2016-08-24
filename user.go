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
	Name string
	Zone string // need to convert to zone obj

	Init bool

	Groups Groups
	Con    *Connection
}

type Users []*User

// initUser
func initUser(name string, zone string, con *Connection) (*User, error) {

	usr := new(User)

	usr.Name = name
	usr.Zone = zone
	usr.Con = con

	if err := usr.init(); err != nil {
		return nil, err
	}

	return usr, nil
}

func (usr *User) init() error {
	if !usr.Init {
		if err := usr.RefreshInfo(); err != nil {
			return err
		}
		// Need to implement getgroups
		// if err := usr.RefreshGroups(); err != nil {
		// 	return err
		// }
		usr.Init = true
	}

	return nil
}

func (usr *User) RefreshInfo() error {
	// r_comment:
	// create_ts:01471444167
	// modify_ts:01471444167
	// user_id:10019
	// user_name:designers
	// user_type_name:rodsgroup
	// zone_name:tempZone
	// user_info:
	if _, err := usr.GetInfo(); err == nil {
		// usr.Comment = infoMap["r_comment"]
		// usr.CreateTime = TimeStringToTime(infoMap["create_ts"])
		// usr.ModifyTime = TimeStringToTime(infoMap["modify_ts"])
		// usr.Id, _ = strconv.Atoi(infoMap["user_id"])
		// usr.Type = infoMap["user_type_name"]
		// //usr.Zone = infoMap["zone_name"]
		// usr.Info = infoMap["user_info"]
	} else {
		return err
	}

	return nil
}

func (usrs Users) FindByName(name string) *User {
	for _, usr := range usrs {
		if usr.Name == name {
			return usr
		}
	}
	return nil
}

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
