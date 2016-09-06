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

type Resource struct {
	name string

	typ           int
	comment       string
	createTime    time.Time
	modifyTime    time.Time
	id            int
	context       string
	zone          *Zone
	class         int
	children      string
	freeSpace     int
	info          string
	status        string
	parentStr     string
	net           string
	freeSpaceTime time.Time
	objCount      int
	storageType   string
	physPath      string

	parentSlice *Resources
	hasInit     bool

	con *Connection
}

type Resources []*Resource

func (rescs Resources) FindByName(name string) *Resource {
	for _, resc := range rescs {
		if resc.name == name {
			return resc
		}
	}
	return nil
}

func (rescs *Resources) Remove(index int) {
	*rescs = append((*rescs)[:index], (*rescs)[index+1:]...)
}

func initResource(name string, con *Connection) (*Resource, error) {
	resc := new(Resource)

	zne, er := con.GetLocalZone()
	if er != nil {
		return nil, er
	}

	resc.con = con
	resc.name = name
	resc.zone = zne

	return resc, nil
}

func (resc *Resource) init() error {
	if !resc.hasInit {
		if err := resc.RefreshInfo(); err != nil {
			return err
		}
		resc.hasInit = true
	}

	return nil
}

func (resc *Resource) Remove() bool {
	for n, p := range *resc.parentSlice {
		if p.name == resc.name {
			resc.parentSlice.Remove(n)
			return true
		}
	}

	return false
}

func (resc *Resource) String() string {
	return fmt.Sprintf("%v#%v", resc.GetName(), resc.GetZone().GetName())
}

func (resc *Resource) GetName() string {
	return resc.name
}

func (resc *Resource) GetComment() (string, error) {
	if err := resc.init(); err != nil {
		return resc.comment, err
	}
	return resc.comment, nil
}

func (resc *Resource) GetCreateTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.createTime, err
	}
	return resc.createTime, nil
}

func (resc *Resource) GetModifyTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.modifyTime, err
	}
	return resc.modifyTime, nil
}

func (resc *Resource) GetId() (int, error) {
	if err := resc.init(); err != nil {
		return resc.id, err
	}
	return resc.id, nil
}

func (resc *Resource) GetType() (int, error) {
	if err := resc.init(); err != nil {
		return resc.typ, err
	}
	return resc.typ, nil
}

func (resc *Resource) GetContext() (string, error) {
	if err := resc.init(); err != nil {
		return resc.context, err
	}
	return resc.context, nil
}

func (resc *Resource) GetClass() (int, error) {
	if err := resc.init(); err != nil {
		return resc.class, err
	}
	return resc.class, nil
}

func (resc *Resource) GetChildren() (string, error) {
	if err := resc.init(); err != nil {
		return resc.children, err
	}
	return resc.children, nil
}

func (resc *Resource) GetFreeSpace() (int, error) {
	if err := resc.init(); err != nil {
		return resc.freeSpace, err
	}
	return resc.freeSpace, nil
}

func (resc *Resource) GetInfo() (string, error) {
	if err := resc.init(); err != nil {
		return resc.info, err
	}
	return resc.info, nil
}

func (resc *Resource) GetStatus() (string, error) {
	if err := resc.init(); err != nil {
		return resc.status, err
	}
	return resc.status, nil
}

func (resc *Resource) GetParentStr() (string, error) {
	if err := resc.init(); err != nil {
		return resc.parentStr, err
	}
	return resc.parentStr, nil
}

func (resc *Resource) GetNet() (string, error) {
	if err := resc.init(); err != nil {
		return resc.net, err
	}
	return resc.net, nil
}

func (resc *Resource) GetFreeSpaceTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.freeSpaceTime, err
	}
	return resc.freeSpaceTime, nil
}

func (resc *Resource) GetObjCount() (int, error) {
	if err := resc.init(); err != nil {
		return resc.objCount, err
	}
	return resc.objCount, nil
}

func (resc *Resource) GetStorageType() (string, error) {
	if err := resc.init(); err != nil {
		return resc.storageType, err
	}
	return resc.storageType, nil
}

func (resc *Resource) GetPhysPath() (string, error) {
	if err := resc.init(); err != nil {
		return resc.physPath, err
	}
	return resc.physPath, nil
}

func (resc *Resource) GetZone() *Zone {
	return resc.zone
}

func (resc *Resource) RefreshInfo() error {

	//map[
	// resc_context:
	// resc_id: 10022
	// zone_name: tempZone
	// resc_class_name: cache
	// modify_ts: 01471623935
	// resc_children:
	// free_space:
	// resc_info:
	// r_comment:
	// resc_status:
	// resc_parent:
	// resc_net: irods-resource
	// free_space_ts:
	// resc_objcount: 2
	// resc_name: irods-resourceResource
	// resc_type_name: unixfilesystem
	// resc_def_path: /var/lib/irods/iRODS/Vault
	// create_ts: 01471614567
	// ]

	typeMap := map[string]int{
		"cache":   Cache,
		"archive": Archive,
	}

	if infoMap, err := resc.FetchInfo(); err == nil {
		resc.comment = infoMap["r_comment"]
		resc.createTime = timeStringToTime(infoMap["create_ts"])
		resc.modifyTime = timeStringToTime(infoMap["modify_ts"])
		resc.id, _ = strconv.Atoi(infoMap["resc_id"])
		resc.typ = ResourceType

		resc.context = infoMap["resc_context"]
		resc.class = typeMap[infoMap["resc_class_name"]]
		resc.children = infoMap["resc_children"]
		resc.freeSpace, _ = strconv.Atoi(infoMap["free_space"])
		resc.info = infoMap["resc_info"]
		resc.status = infoMap["resc_status"]
		resc.parentStr = infoMap["resc_parent"]
		resc.net = infoMap["resc_net"]
		resc.freeSpaceTime = timeStringToTime(infoMap["free_space_ts"])
		resc.objCount, _ = strconv.Atoi(infoMap["resc_objcount"])
		resc.storageType = infoMap["resc_type_name"]
		resc.physPath = infoMap["resc_def_path"]

		if zones, err := resc.con.GetZones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], resc.con); zne != nil {
				resc.zone = zne
			} else {
				return newError(Fatal, fmt.Sprintf("iRods Refresh Resource Info Failed: Unable to locate zone in cache"))
			}
		}

	} else {
		return err
	}

	return nil
}

func (resc *Resource) FetchInfo() (map[string]string, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cResource := C.CString(resc.name)
	defer C.free(unsafe.Pointer(cResource))

	ccon := resc.con.GetCcon()

	if status := C.gorods_get_resource(cResource, ccon, &result, &err); status != 0 {
		resc.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Resource Info Failed: %v", C.GoString(err)))
	}

	resc.con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(map[string]string)

	for _, resourceInfo := range slice {

		resourceAttributes := strings.Split(strings.Trim(C.GoString(resourceInfo), " \n"), "\n")

		for _, attr := range resourceAttributes {

			split := strings.Split(attr, ": ")

			attrName := split[0]
			attrVal := split[1]

			response[attrName] = attrVal

		}
	}

	return response, nil
}
