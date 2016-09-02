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
	Name        string
	Zone        *Zone
	CreateTime  time.Time
	ModifyTime  time.Time
	Id          int
	Type        int
	Info        string
	Comment     string
	ParentSlice *Users

	Init bool

	Groups Groups
	Con    *Connection
}

type Users []*User

// initUser
func initUser(name string, zone *Zone, con *Connection) (*User, error) {

	usr := new(User)

	usr.Name = name
	usr.Zone = zone
	usr.Con = con

	return usr, nil
}

func (usr *User) Remove() bool {
	for n, p := range *usr.ParentSlice {
		if p.Name == usr.Name {
			usr.ParentSlice.Remove(n)
			return true
		}
	}

	return false
}

func (usr *User) GetName() string {
	return usr.Name
}

func (usr *User) GetZone() *Zone {
	return usr.Zone
}

func (usr *User) GetComment() string {
	usr.init()
	return usr.Comment
}

func (usr *User) GetCreateTime() (time.Time, error) {
	if err := usr.init(); err != nil {
		return usr.CreateTime, err
	}
	return usr.CreateTime, nil
}

func (usr *User) GetModifyTime() time.Time {
	usr.init()
	return usr.ModifyTime
}

func (usr *User) GetId() int {
	usr.init()
	return usr.Id
}

func (usr *User) GetType() int {
	usr.init()
	return usr.Type
}

func (usr *User) GetCon() *Connection {
	return usr.Con
}

func (usr *User) GetGroups() (Groups, error) {
	if err := usr.init(); err != nil {
		return nil, err
	}
	return usr.Groups, nil
}

func (usr *User) Delete() error {
	if err := deleteUser(usr.GetName(), usr.GetZone(), usr.Con); err != nil {
		return err
	}

	if err := usr.Con.RefreshUsers(); err != nil {
		return err
	}

	return nil
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

	typeMap := map[string]int{
		"rodsuser":   UserType,
		"rodsadmin":  AdminType,
		"groupadmin": GroupAdminType,
	}

	if infoMap, err := usr.FetchInfo(); err == nil {
		usr.Comment = infoMap["r_comment"]
		usr.CreateTime = timeStringToTime(infoMap["create_ts"])
		usr.ModifyTime = timeStringToTime(infoMap["modify_ts"])
		usr.Id, _ = strconv.Atoi(infoMap["user_id"])
		usr.Type = typeMap[infoMap["user_type_name"]]
		usr.Info = infoMap["user_info"]

		if zones, err := usr.Con.GetZones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], usr.Con); zne != nil {
				usr.Zone = zne
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
		usr.Groups = grps
	} else {
		return err
	}

	return nil
}

func (usrs Users) FindByName(name string, con *Connection) *User {
	for _, usr := range usrs {
		if usr.Name == name {
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
	return fmt.Sprintf("%v:%v#%v", getTypeString(usr.Type), usr.Name, usr.Zone)
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
				grp := grps.FindByName(gName, usr.Con)

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

	cUser := C.CString(usr.Name)
	defer C.free(unsafe.Pointer(cUser))

	ccon := usr.Con.GetCcon()

	if status := C.gorods_get_user(cUser, ccon, &result, &err); status != 0 {
		usr.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Users Failed: %v", C.GoString(err)))
	}

	usr.Con.ReturnCcon(ccon)

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
		return addToGroup(usr.Name, usr.Zone, grp.(string), usr.Con)
	case *Group:
		return addToGroup(usr.Name, usr.Zone, (grp.(*Group)).Name, usr.Con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods AddToGroup Failed: unknown type passed"))
}

func (usr *User) RemoveFromGroup(grp interface{}) error {
	switch grp.(type) {
	case string:
		return removeFromGroup(usr.Name, usr.Zone, grp.(string), usr.Con)
	case *Group:
		return removeFromGroup(usr.Name, usr.Zone, (grp.(*Group)).Name, usr.Con)
	default:
	}

	return newError(Fatal, fmt.Sprintf("iRods RemoveFromGroup Failed: unknown type passed"))
}

func (usr *User) ChangePassword(newPass string) error {
	var (
		err *C.char
	)

	cUserName := C.CString(usr.GetName())
	cNewPass := C.CString(newPass)
	cMyPass := C.CString(usr.GetCon().Options.Password)
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cNewPass))
	defer C.free(unsafe.Pointer(cMyPass))

	ccon := usr.GetCon().GetCcon()
	defer usr.GetCon().ReturnCcon(ccon)

	if status := C.gorods_change_user_password(cUserName, cNewPass, cMyPass, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods ChangePassword Failed: %v", C.GoString(err)))
	}

	return nil
}

func deleteUser(userName string, zone *Zone, con *Connection) error {
	var (
		err *C.char
	)

	cZoneName := C.CString(zone.GetName())
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
