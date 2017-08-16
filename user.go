/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

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

// User contains information relating to an iRODS user
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

	metaCol *MetaCollection
}

// Users is a slice of *User
type Users []*User

func initUser(name string, zone *Zone, con *Connection) (*User, error) {

	usr := new(User)

	usr.name = name
	usr.zone = zone
	usr.con = con

	return usr, nil
}

// Remove removed the user from it's parent slice.
func (usr *User) Remove() bool {
	for n, p := range *usr.parentSlice {
		if p.name == usr.name {
			usr.parentSlice.Remove(n)
			return true
		}
	}

	return false
}

// Name returns the users name.
func (usr *User) Name() string {
	return usr.name
}

// Path returns the users name. Used in gorods.MetaObj interface.
func (usr *User) Path() string {
	return usr.name
}

// Zone returns the *Zone to which the user belongs.
func (usr *User) Zone() *Zone {
	return usr.zone
}

// Comment loads data from iRODS if needed, and returns the user's comment attribute.
func (usr *User) Comment() (string, error) {
	if err := usr.init(); err != nil {
		return usr.comment, err
	}
	return usr.comment, nil
}

// Info loads data from iRODS if needed, and returns the user's info attribute.
func (usr *User) Info() (string, error) {
	if err := usr.init(); err != nil {
		return usr.info, err
	}
	return usr.info, nil
}

// CreateTime loads data from iRODS if needed, and returns the user's createTime attribute.
func (usr *User) CreateTime() (time.Time, error) {
	if err := usr.init(); err != nil {
		return usr.createTime, err
	}
	return usr.createTime, nil
}

// ModifyTime loads data from iRODS if needed, and returns the user's modifyTime attribute.
func (usr *User) ModifyTime() (time.Time, error) {
	if err := usr.init(); err != nil {
		return usr.modifyTime, err
	}
	return usr.modifyTime, nil
}

// Id loads data from iRODS if needed, and returns the user's id attribute.
func (usr *User) Id() (int, error) {
	if err := usr.init(); err != nil {
		return usr.id, err
	}
	return usr.id, nil
}

// Type loads data from iRODS if needed, and returns the user's typ attribute. Used in AccessObject and MetaObj interfaces.
func (usr *User) Type() int {
	if err := usr.init(); err != nil {
		return UnknownType
	}
	return usr.typ
}

// Con returns the connection used to initalize the user
func (usr *User) Con() *Connection {
	return usr.con
}

// Groups loads data from iRODS if needed, and returns the user's groups slice.
func (usr *User) Groups() (Groups, error) {
	if err := usr.init(); err != nil {
		return nil, err
	}
	return usr.groups, nil
}

// Delete deletes the user from the iCAT server
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

// RefreshInfo fetches user data from iCAT, and unloads it into the *User fields
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

		if zones, err := usr.con.Zones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], usr.con); zne != nil {
				usr.zone = zne
			} else {
				return newError(Fatal, -1, fmt.Sprintf("iRODS Refresh User Info Failed: Unable to locate zone in cache"))
			}
		}
	} else {
		return err
	}

	return nil
}

// RefreshGroups fetches group data from iCAT, and unloads it into the *User fields
func (usr *User) RefreshGroups() error {

	if grps, err := usr.FetchGroups(); err == nil {
		usr.groups = grps
	} else {
		return err
	}

	return nil
}

// FindByName searches the slice for a user by name.
// If no match is found, a new user with that name is created and returned.
// This was designed to resolve issues of casting resources for DataObjects and Collections, even though the cache was empty due to permissions.
func (usrs Users) FindByName(name string, con *Connection) *User {
	for _, usr := range usrs {
		if usr.name == name {
			return usr
		}
	}

	zne, err := con.LocalZone()
	if err != nil {
		return nil
	}

	usr, _ := initUser(name, zne, con)

	return usr
}

// Remove deletes an item from the slice based on the index.
func (usrs *Users) Remove(index int) {
	*usrs = append((*usrs)[:index], (*usrs)[index+1:]...)
}

// String returns a user's type, name, and zone.
func (usr *User) String() string {
	usr.init()
	return fmt.Sprintf("%v:%v#%v", getTypeString(usr.typ), usr.name, usr.zone)
}

// FetchGroups fetches and returns fresh data about the user's groups from the iCAT server.
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
		return nil, newError(Fatal, status, fmt.Sprintf("iRODS Get Groups Failed: %v", C.GoString(err)))
	}

	usr.con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	if grps, err := usr.con.Groups(); err == nil {
		response := make(Groups, 0)

		for _, groupName := range slice {

			gName := C.GoString(groupName)

			if gName != usr.name {
				grp := grps.FindByName(gName, usr.con)

				if grp != nil {
					response = append(response, grp)
				} else {
					return nil, newError(Fatal, -1, fmt.Sprintf("iRODS FetchGroups Failed: Group in response not found in cache"))
				}
			}
		}

		return response, nil
	} else {
		return nil, err
	}

}

// FetchInfo fetches fresh user info from the iCAT server, and returns it as a map.
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
		return nil, newError(Fatal, status, fmt.Sprintf("iRODS Get Users Failed: %v", C.GoString(err)))
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

// AddToGroup adds the user to an existing iRODS group.
// Accepts string or *Group types.
func (usr *User) AddToGroup(grp interface{}) error {

	switch grp.(type) {
	case string:
		return addToGroup(usr.name, usr.zone, grp.(string), usr.con)
	case *Group:
		return addToGroup(usr.name, usr.zone, (grp.(*Group)).name, usr.con)
	default:
	}

	return newError(Fatal, -1, fmt.Sprintf("iRODS AddToGroup Failed: unknown type passed"))
}

// RemoveFromGroup removes the user from an iRODS group.
// Accepts string or *Group types.
func (usr *User) RemoveFromGroup(grp interface{}) error {
	switch grp.(type) {
	case string:
		return removeFromGroup(usr.name, usr.zone, grp.(string), usr.con)
	case *Group:
		return removeFromGroup(usr.name, usr.zone, (grp.(*Group)).name, usr.con)
	default:
	}

	return newError(Fatal, -1, fmt.Sprintf("iRODS RemoveFromGroup Failed: unknown type passed"))
}

// ChangePassword changes the user's password.
// You will need to be a rodsadmin for this to succeed (I think).
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
		return newError(Fatal, status, fmt.Sprintf("iRODS ChangePassword Failed: %v", C.GoString(err)))
	}

	return nil
}

// AddMeta adds a single Meta triple struct
func (usr *User) AddMeta(m Meta) (nm *Meta, err error) {
	var mc *MetaCollection

	if mc, err = usr.Meta(); err != nil {
		return
	}

	nm, err = mc.Add(m)

	return
}

// DeleteMeta deletes a single Meta triple struct, identified by Attribute field
func (usr *User) DeleteMeta(attr string) (*MetaCollection, error) {
	if mc, err := usr.Meta(); err == nil {
		return mc, mc.Delete(attr)
	} else {
		return nil, err
	}
}

// Meta returns collection of Meta AVU triple structs of the user object
func (usr *User) Meta() (*MetaCollection, error) {

	if usr.metaCol == nil {
		if mc, err := newMetaCollection(usr); err == nil {
			usr.metaCol = mc
		} else {
			return nil, err
		}
	}

	return usr.metaCol, nil
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
		return newError(Fatal, status, fmt.Sprintf("iRODS DeleteUser %v Failed: %v", userName, C.GoString(err)))
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
		return newError(Fatal, -1, fmt.Sprintf("iRODS CreateUser Failed: Unknown user type passed"))
	}

	cZoneName := C.CString(zoneName)
	cUserName := C.CString(userName)
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cType))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_create_user(cUserName, cZoneName, cType, ccon, &err); status != 0 {
		return newError(Fatal, status, fmt.Sprintf("iRODS CreateUser %v Failed: %v", userName, C.GoString(err)))
	}

	return nil
}
