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

// Group holds info about iRODS groups
type Group struct {
	name        string
	createTime  time.Time
	modifyTime  time.Time
	id          int
	typ         int
	zone        *Zone
	info        string
	comment     string
	parentSlice *Groups

	hasInit bool

	users Users
	con   *Connection

	metaCol *MetaCollection
}

// Groups is a slice of *Group structs.
type Groups []*Group

// initGroup only accepts what's needed for lazing loading later on.
func initGroup(name string, con *Connection) (*Group, error) {

	grp := new(Group)

	grp.name = name
	grp.con = con
	grp.typ = GroupType

	if z, err := con.LocalZone(); err != nil {
		return nil, err
	} else {
		grp.zone = z
	}

	return grp, nil
}

// Name returns the name of the group.
func (grp *Group) Name() string {
	return grp.name
}

// Path returns the name of the group. Used in gorods.MetaObj interface.
func (grp *Group) Path() string {
	return grp.name
}

// Zone returns the *Zone struct that this group belongs to.
func (grp *Group) Zone() *Zone {
	return grp.zone
}

// Comment loads data from iRODS if needed, and returns the group's comment attribute.
func (grp *Group) Comment() (string, error) {
	if err := grp.init(); err != nil {
		return grp.comment, err
	}
	return grp.comment, nil
}

// CreateTime loads data from iRODS if needed, and returns the group's createTime attribute.
func (grp *Group) CreateTime() (time.Time, error) {
	if err := grp.init(); err != nil {
		return grp.createTime, err
	}
	return grp.createTime, nil
}

// ModifyTime loads data from iRODS if needed, and returns the group's modifyTime attribute.
func (grp *Group) ModifyTime() (time.Time, error) {
	if err := grp.init(); err != nil {
		return grp.modifyTime, err
	}
	return grp.modifyTime, nil
}

// Id loads data from iRODS if needed, and returns the group's Id attribute.
func (grp *Group) Id() (int, error) {
	if err := grp.init(); err != nil {
		return grp.id, err
	}
	return grp.id, nil
}

// Info loads data from iRODS if needed, and returns the group's Info attribute.
func (grp *Group) Info() (string, error) {
	if err := grp.init(); err != nil {
		return grp.info, err
	}
	return grp.info, nil
}

// Type returns GroupType, used in iRODSObject interfaces
func (grp *Group) Type() int {
	return grp.typ
}

// Con returns the *Connection used to fetch group info
func (grp *Group) Con() *Connection {
	return grp.con
}

// Users loads data from iRODS if needed, and returns the group's Users slice.
func (grp *Group) Users() (Users, error) {
	if err := grp.init(); err != nil {
		return nil, err
	}

	return grp.users, nil
}

// Delete deletes the group from iCAT server. Refreshes Connection group cache.
func (grp *Group) Delete() error {
	if err := deleteGroup(grp.Name(), grp.Zone(), grp.con); err != nil {
		return err
	}

	if err := grp.con.RefreshGroups(); err != nil {
		return err
	}

	return nil
}

func (grp *Group) init() error {
	if !grp.hasInit {
		if err := grp.RefreshInfo(); err != nil {
			return err
		}
		if err := grp.RefreshUsers(); err != nil {
			return err
		}
		grp.hasInit = true
	}

	return nil
}

// FindByName searches the slice (itself) and attempts to return a match based on name.
// If no match is found, a new group with that name is created and returned.
// This was designed to resolve issues of casting resources for DataObjects and Collections, even though the cache was empty due to permissions.
func (grps Groups) FindByName(name string, con *Connection) *Group {
	for _, grp := range grps {
		if grp.name == name {
			return grp
		}
	}

	grp, _ := initGroup(name, con)

	return grp
}

func (grps *Groups) Remove(index int) {
	*grps = append((*grps)[:index], (*grps)[index+1:]...)
}

// Remove deletes the group from the parent slice
func (grp *Group) Remove() bool {
	for n, p := range *grp.parentSlice {
		if p.name == grp.name {
			grp.parentSlice.Remove(n)
			return true
		}
	}

	return false
}

// String returns the group name
func (grp *Group) String() string {
	return fmt.Sprintf("%v", grp.name)
}

// RefreshInfo refreshes the attributes of the group, pulling fresh data from the iCAT server.
func (grp *Group) RefreshInfo() error {
	// r_comment:
	// create_ts:01471444167
	// modify_ts:01471444167
	// user_id:10019
	// user_name:designers
	// user_type_name:rodsgroup
	// zone_name:tempZone
	// user_info:

	if infoMap, err := grp.FetchInfo(); err == nil {
		grp.comment = infoMap["r_comment"]
		grp.createTime = timeStringToTime(infoMap["create_ts"])
		grp.modifyTime = timeStringToTime(infoMap["modify_ts"])
		grp.id, _ = strconv.Atoi(infoMap["user_id"])
		grp.info = infoMap["user_info"]

		if zones, err := grp.con.Zones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], grp.con); zne != nil {
				grp.zone = zne
			} else {
				return newError(Fatal, -1, fmt.Sprintf("iRODS Refresh Group Info Failed: Unable to locate zone in cache"))
			}
		}
	} else {
		return err
	}

	return nil
}

// RefreshUsers pulls fresh data from the iCAT server and sets the internal field returned by *Group.Users
func (grp *Group) RefreshUsers() error {

	if usrs, err := grp.FetchUsers(); err == nil {
		grp.users = usrs
	} else {
		return err
	}

	return nil
}

// FetchInfo returns a map of fresh group info from the iCAT server
func (grp *Group) FetchInfo() (map[string]string, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cGroup := C.CString(grp.name)
	defer C.free(unsafe.Pointer(cGroup))

	ccon := grp.con.GetCcon()

	if status := C.gorods_get_user(cGroup, ccon, &result, &err); status != 0 {
		grp.con.ReturnCcon(ccon)
		return nil, newError(Fatal, status, fmt.Sprintf("iRODS Get Group Info Failed: %v", C.GoString(err)))
	}

	grp.con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(map[string]string)

	for _, groupInfo := range slice {

		groupAttributes := strings.Split(strings.Trim(C.GoString(groupInfo), " \n"), "\n")

		for _, attr := range groupAttributes {

			split := strings.Split(attr, ": ")

			attrName := split[0]
			attrVal := split[1]

			response[attrName] = attrVal

		}
	}

	return response, nil
}

// FetchUsers returns a slice of fresh *User from the iCAT server
func (grp *Group) FetchUsers() (Users, error) {

	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cGroupName := C.CString(grp.name)
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := grp.con.GetCcon()

	if status := C.gorods_get_group(ccon, &result, cGroupName, &err); status != 0 {
		grp.con.ReturnCcon(ccon)
		if status == C.CAT_NO_ROWS_FOUND {
			return make(Users, 0), nil
		} else {

			return nil, newError(Fatal, status, fmt.Sprintf("iRODS Get Group %v Failed: %v", grp.name, C.GoString(err)))
		}

	}

	grp.con.ReturnCcon(ccon)
	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	if usrs, err := grp.con.Users(); err == nil {
		response := make(Users, 0)

		for _, userNames := range slice {

			usrFrags := strings.Split(C.GoString(userNames), "#")

			if usr := usrs.FindByName(usrFrags[0], grp.con); usr != nil {
				response = append(response, usr)
			} else {
				return nil, newError(Fatal, -1, fmt.Sprintf("iRODS FetchUsers Failed: User in response not found in cache"))
			}

		}

		return response, nil
	} else {
		return nil, err
	}

}

// AddUser adds an iRODS uset to the group. Accepts a string or *User struct.
func (grp *Group) AddUser(usr interface{}) error {

	switch usr.(type) {
	case string:
		// Need to lookup user by string in cache for zone info

		if usrs, err := grp.con.Users(); err == nil {
			usrName := usr.(string)

			if existingUsr := usrs.FindByName(usrName, grp.con); existingUsr != nil {
				zoneName := existingUsr.zone
				return addToGroup(usrName, zoneName, grp.name, grp.con)
			} else {
				return newError(Fatal, -1, fmt.Sprintf("iRODS AddUser Failed: can't find iRODS user by string"))
			}
		} else {
			return err
		}

	case *User:
		aUsr := usr.(*User)
		return addToGroup(aUsr.name, aUsr.zone, grp.name, aUsr.con)
	default:
	}

	return newError(Fatal, -1, fmt.Sprintf("iRODS AddUser Failed: unknown type passed"))
}

// RemoveUser removes an iRODS user from the group. Accepts a string or *User struct.
func (grp *Group) RemoveUser(usr interface{}) error {
	switch usr.(type) {
	case string:
		// Need to lookup user by string in cache for zone info

		if usrs, err := grp.con.Users(); err == nil {
			usrName := usr.(string)

			if existingUsr := usrs.FindByName(usrName, grp.con); existingUsr != nil {
				zoneName := existingUsr.zone
				return removeFromGroup(usrName, zoneName, grp.name, grp.con)
			} else {
				return newError(Fatal, -1, fmt.Sprintf("iRODS RemoveUser Failed: can't find iRODS user by string"))
			}
		} else {
			return err
		}

	case *User:
		aUsr := usr.(*User)
		return removeFromGroup(aUsr.name, aUsr.zone, grp.name, aUsr.con)
	default:
	}

	return newError(Fatal, -1, fmt.Sprintf("iRODS RemoveUser Failed: unknown type passed"))
}

// AddMeta adds a single Meta triple struct
func (grp *Group) AddMeta(m Meta) (nm *Meta, err error) {
	var mc *MetaCollection

	if mc, err = grp.Meta(); err != nil {
		return
	}

	nm, err = mc.Add(m)

	return
}

// DeleteMeta deletes a single Meta triple struct, identified by Attribute field
func (grp *Group) DeleteMeta(attr string) (*MetaCollection, error) {
	if mc, err := grp.Meta(); err == nil {
		return mc, mc.Delete(attr)
	} else {
		return nil, err
	}
}

// Meta returns collection of Meta AVU triple structs of the user object
func (grp *Group) Meta() (*MetaCollection, error) {

	if grp.metaCol == nil {
		if mc, err := newMetaCollection(grp); err == nil {
			grp.metaCol = mc
		} else {
			return nil, err
		}
	}

	return grp.metaCol, nil
}

func addToGroup(userName string, zone *Zone, groupName string, con *Connection) error {

	var (
		err *C.char
	)

	cUserName := C.CString(userName)
	cZoneName := C.CString(zone.Name())
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_add_user_to_group(cUserName, cZoneName, cGroupName, ccon, &err); status != 0 {
		return newError(Fatal, status, fmt.Sprintf("iRODS AddToGroup %v Failed: %v", groupName, C.GoString(err)))
	}

	return nil
}

func removeFromGroup(userName string, zone *Zone, groupName string, con *Connection) error {
	var (
		err *C.char
	)

	cUserName := C.CString(userName)
	cZoneName := C.CString(zone.Name())
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cUserName))
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_remove_user_from_group(cUserName, cZoneName, cGroupName, ccon, &err); status != 0 {
		return newError(Fatal, status, fmt.Sprintf("iRODS AddToGroup %v Failed: %v", groupName, C.GoString(err)))
	}

	return nil
}

func deleteGroup(groupName string, zone *Zone, con *Connection) error {
	var (
		err *C.char
	)

	cZoneName := C.CString(zone.Name())
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_delete_group(cGroupName, cZoneName, ccon, &err); status != 0 {
		return newError(Fatal, status, fmt.Sprintf("iRODS DeleteGroup %v Failed: %v", groupName, C.GoString(err)))
	}

	return nil
}

func createGroup(groupName string, zone *Zone, con *Connection) error {
	var (
		err *C.char
	)

	cZoneName := C.CString(zone.Name())
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cGroupName))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := C.gorods_create_group(cGroupName, cZoneName, ccon, &err); status != 0 {
		return newError(Fatal, status, fmt.Sprintf("iRODS CreateGroup %v Failed: %v", groupName, C.GoString(err)))
	}

	return nil
}
