/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

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
	name        string
	zone        *Zone
	createTime  time.Time
	modifyTime  time.Time
	id          int
	typ         int
	info        string
	comment     string
	parentSlice *Users

	hasInit bool

	groups Groups
	con    *Connection
}

type Users []*User

// initUser
func initUser(name string, zone *Zone, con *Connection) (*User, error) {

	usr := new(User)

	usr.name = name
	usr.zone = zone
	usr.con = con

	return usr, nil
}

func (usr *User) Remove() bool {
	for n, p := range *usr.parentSlice {
		if p.name == usr.name {
			usr.parentSlice.Remove(n)
			return true
		}
	}

	return false
}

func (usr *User) Name() string {
	return usr.name
}

func (usr *User) Zone() *Zone {
	return usr.zone
}

func (usr *User) Comment() (string, error) {
	if err := usr.init(); err != nil {
		return usr.comment, err
	}
	return usr.comment, nil
}

func (usr *User) Info() (string, error) {
	if err := usr.init(); err != nil {
		return usr.info, err
	}
	return usr.info, nil
}

func (usr *User) CreateTime() (time.Time, error) {
	if err := usr.init(); err != nil {
		return usr.createTime, err
	}
	return usr.createTime, nil
}

func (usr *User) ModifyTime() (time.Time, error) {
	if err := usr.init(); err != nil {
		return usr.modifyTime, err
	}
	return usr.modifyTime, nil
}

func (usr *User) Id() (int, error) {
	if err := usr.init(); err != nil {
		return usr.id, err
	}
	return usr.id, nil
}

func (usr *User) Type() (int, error) {
	if err := usr.init(); err != nil {
		return usr.typ, err
	}
	return usr.typ, nil
}

func (usr *User) Con() *Connection {
	return usr.con
}

func (usr *User) Groups() (Groups, error) {
	if err := usr.init(); err != nil {
		return nil, err
	}
	return usr.groups, nil
}

func (usr *User) Delete() error {
	if err := deleteUser(usr.Name(), usr.Zone(), usr.con); err != nil {
		return err
	}

	if err := usr.con.RefreshUsers(); err != nil {
		return err
	}

	return nil
}

func (usr *User) init() error {
	if !usr.hasInit {
		if err := usr.RefreshInfo(); err != nil {
			return err
		}
		if err := usr.RefreshGroups(); err != nil {
			return err
		}
		usr.hasInit = true
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

	typeMap := map[string]int{
		"rodsuser":   UserType,
		"rodsadmin":  AdminType,
		"groupadmin": GroupAdminType,
	}

	if infoMap, err := usr.FetchInfo(); err == nil {
		usr.comment = infoMap["r_comment"]
		usr.createTime = timeStringToTime(infoMap["create_ts"])
		usr.modifyTime = timeStringToTime(infoMap["modify_ts"])
		usr.id, _ = strconv.Atoi(infoMap["user_id"])
		usr.typ = typeMap[infoMap["user_type_name"]]
		usr.info = infoMap["user_info"]

		if zones, err := usr.con.GetZones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], usr.con); zne != nil {
				usr.zone = zne
			} else {
				return newError(Fatal, fmt.Sprintf("iRods Refresh User Info Failed: Unable to locate zone in cache"))
			}
		}
	} else {
		return err
	}

	return nil
}

func (usr *User) RefreshGroups() error {

	if grps, err := usr.FetchGroups(); err == nil {
		usr.groups = grps
	} else {
		return err
	}

	return nil
}

func (usrs Users) FindByName(name string, con *Connection) *User {
	for _, usr := range usrs {
		if usr.name == name {
			return usr
		}
	}

	zne, err := con.GetLocalZone()
	if err != nil {
		return nil
	}

	usr, _ := initUser(name, zne, con)

	return usr
}

func (usrs *Users) Remove(index int) {
	*usrs = append((*usrs)[:index], (*usrs)[index+1:]...)
}

func (usr *User) String() string {
	usr.init()
	return fmt.Sprintf("%v:%v#%v", getTypeString(usr.typ), usr.name, usr.zone)
}

func (usr *User) FetchGroups() (Groups, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	cName := C.CString(usr.name)
	defer C.free(unsafe.Pointer(cName))

	result.size = C.int(0)

	ccon := usr.con.GetCcon()

	if status := C.gorods_get_user_groups(ccon, cName, &result, &err); status != 0 {
		usr.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Groups Failed: %v", C.GoString(err)))
	}

	usr.con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	if grps, err := usr.con.GetGroups(); err == nil {
		response := make(Groups, 0)

		for _, groupName := range slice {

			gName := C.GoString(groupName)

			if gName != usr.name {
				grp := grps.FindByName(gName, usr.con)

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

func (usr *User) FetchInfo() (map[string]string, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cUser := C.CString(usr.name)
	defer C.free(unsafe.Pointer(cUser))

	ccon := usr.con.GetCcon()

	if status := C.gorods_get_user(cUser, ccon, &result, &err); status != 0 {
		usr.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Users Failed: %v", C.GoString(err)))
	}

	usr.con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

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

	return response, nil
}

func (usr *User) AddToGroup(grp interface{}) error {

	switch grp.(type) {
	case string:
		return addToGroup(usr.name, usr.zone, grp.(string), usr.con)
	case *Group:
		return addToGroup(usr.name, usr.zone, (grp.(*Group)).name, usr.con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods AddToGroup Failed: unknown type passed"))
}

func (usr *User) RemoveFromGroup(grp interface{}) error {
	switch grp.(type) {
	case string:
		return removeFromGroup(usr.name, usr.zone, grp.(string), usr.con)
	case *Group:
		return removeFromGroup(usr.name, usr.zone, (grp.(*Group)).name, usr.con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods RemoveFromGroup Failed: unknown type passed"))
}

func (usr *User) ChangePassword(newPass string) error {
	var (
		err *C.char
	)

	cUserName := C.CString(usr.Name())
	cNewPass := C.CString(newPass)
	cMyPass := C.CString(usr.Con().Options.Password)
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cNewPass))
	defer C.free(unsafe.Pointer(cMyPass))

	ccon := usr.Con().GetCcon()
	defer usr.Con().ReturnCcon(ccon)

	if status := C.gorods_change_user_password(cUserName, cNewPass, cMyPass, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods ChangePassword Failed: %v", C.GoString(err)))
	}

	return nil
}

func deleteUser(userName string, zone *Zone, con *Connection) error {
	var (
		err *C.char
	)

	cZoneName := C.CString(zone.Name())
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

func createUser(userName string, zoneName string, typ int, con *Connection) error {
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
