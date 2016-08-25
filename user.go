/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type User struct {
	Name       string
	Zone       string // need to convert to zone obj
	CreateTime time.Time
	ModifyTime time.Time
	Id         int
	Type       string
	Info       string
	Comment    string

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

	// if err := usr.init(); err != nil {
	// 	return nil, err
	// }

	return usr, nil
}

func (usr *User) GetName() string {
	return usr.Name
}

func (usr *User) GetZone() string {
	return usr.Zone
}

func (usr *User) GetGroups() (Groups, error) {
	if err := usr.init(); err != nil {
		return nil, err
	}
	return usr.Groups, nil
}

func (usr *User) init() error {
	if !usr.Init {
		if err := usr.RefreshInfo(); err != nil {
			return err
		}
		if err := usr.RefreshGroups(); err != nil {
			return err
		}
		usr.Init = true
	}

	return nil
}

func (usr *User) RefreshInfo() error {

	// create_ts:01471441907
	// modify_ts:01471441907
	// user_id:10011
	// user_name:john
	// user_type_name:rodsuser
	// zone_name:tempZone
	// user_info:
	// r_comment:

	if infoMap, err := usr.GetInfo(); err == nil {
		usr.Comment = infoMap["r_comment"]
		usr.CreateTime = TimeStringToTime(infoMap["create_ts"])
		usr.ModifyTime = TimeStringToTime(infoMap["modify_ts"])
		usr.Id, _ = strconv.Atoi(infoMap["user_id"])
		usr.Type = infoMap["user_type_name"]
		//usr.Zone = infoMap["zone_name"]
		usr.Info = infoMap["user_info"]
	} else {
		return err
	}

	return nil
}

func (usr *User) RefreshGroups() error {

	if grps, err := usr.FetchGroups(); err == nil {
		usr.Groups = grps
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

func (usr *User) FetchGroups() (Groups, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	cName := C.CString(usr.Name)
	defer C.free(unsafe.Pointer(cName))

	result.size = C.int(0)

	ccon := usr.Con.GetCcon()

	if status := C.gorods_get_user_groups(ccon, cName, &result, &err); status != 0 {
		usr.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Groups Failed: %v", C.GoString(err)))
	}

	usr.Con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	if grps, err := usr.Con.GetGroups(); err == nil {
		response := make(Groups, 0)

		for _, groupName := range slice {

			gName := C.GoString(groupName)

			if gName != usr.Name {
				grp := grps.FindByName(gName)

				if grp != nil {
					response = append(response, grp)
				} else {
					return nil, newError(Fatal, fmt.Sprintf("iRods FetchGroups Failed: Group in response not found in cache"))
				}
			}
		}

		return response, nil
	} else {
		return nil, err
	}

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

	if status := C.gorods_get_user(cUser, ccon, &result, &err); status != 0 {
		usr.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Users Failed: %v", C.GoString(err)))
	}

	usr.Con.ReturnCcon(ccon)

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

func DeleteUser(userName string, zoneName string, con *Connection) error {
	var (
		err *C.char
	)

	cZoneName := C.CString(zoneName)
	cUserName := C.CString(userName)
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cUserName))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_delete_user(cUserName, cZoneName, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods DeleteUser %v Failed: %v", userName, C.GoString(err)))
	}

	return nil
}

func CreateUser(userName string, zoneName string, typ int, con *Connection) error {
	var (
		err   *C.char
		cType *C.char
	)

	switch typ {
	case AdminType:
		cType = C.CString("rodsadmin")
	case UserType:
		cType = C.CString("rodsuser")
	case GroupAdminType:
		cType = C.CString("groupadmin")
	default:
		return newError(Fatal, fmt.Sprintf("iRods CreateUser Failed: Unknown user type passed"))
	}

	cZoneName := C.CString(zoneName)
	cUserName := C.CString(userName)
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cType))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_create_user(cUserName, cZoneName, cType, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods CreateUser %v Failed: %v", userName, C.GoString(err)))
	}

	return nil
}
