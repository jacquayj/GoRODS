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
	Name string

	Type          int
	Comment       string
	CreateTime    time.Time
	ModifyTime    time.Time
	Id            int
	Context       string
	Zone          *Zone
	Class         int
	Children      string
	FreeSpace     int
	Info          string
	Status        string
	ParentStr     string
	Net           string
	FreeSpaceTime time.Time
	ObjCount      int
	StorageType   string
	PhysPath      string

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

	ParentSlice *Resources
	Init        bool

	Con *Connection
}

type Resources []*Resource

func (rescs Resources) FindByName(name string) *Resource {
	for _, resc := range rescs {
		if resc.Name == name {
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

	resc.Con = con
	resc.Name = name

	return resc, nil
}

func (resc *Resource) init() error {
	if !resc.Init {
		if err := resc.RefreshInfo(); err != nil {
			return err
		}
		resc.Init = true
	}

	return nil
}

func (resc *Resource) Remove() bool {
	for n, p := range *resc.ParentSlice {
		if p.Name == resc.Name {
			resc.ParentSlice.Remove(n)
			return true
		}
	}

	return false
}

func (resc *Resource) String() string {
	return fmt.Sprintf("%v#%v", resc.GetName(), resc.GetZone().GetName())
}

func (resc *Resource) GetName() string {
	return resc.Name
}

func (resc *Resource) GetComment() string {
	resc.init()
	return resc.Comment
}

func (resc *Resource) GetCreateTime() time.Time {
	resc.init()
	return resc.CreateTime
}

func (resc *Resource) GetModifyTime() time.Time {
	resc.init()
	return resc.ModifyTime
}

func (resc *Resource) GetId() int {
	resc.init()
	return resc.Id
}

func (resc *Resource) GetType() int {
	resc.init()
	return resc.Type
}

func (resc *Resource) GetContext() string {
	resc.init()
	return resc.Context
}

func (resc *Resource) GetClass() int {
	resc.init()
	return resc.Class
}

func (resc *Resource) GetChildren() string {
	resc.init()
	return resc.Children
}

func (resc *Resource) GetFreeSpace() int {
	resc.init()
	return resc.FreeSpace
}

func (resc *Resource) GetInfo() string {
	resc.init()
	return resc.Info
}

func (resc *Resource) GetStatus() string {
	resc.init()
	return resc.Status
}

func (resc *Resource) GetParentStr() string {
	resc.init()
	return resc.ParentStr
}

func (resc *Resource) GetNet() string {
	resc.init()
	return resc.Net
}

func (resc *Resource) GetFreeSpaceTime() time.Time {
	resc.init()
	return resc.FreeSpaceTime
}

func (resc *Resource) GetObjCount() int {
	resc.init()
	return resc.ObjCount
}

func (resc *Resource) GetStorageType() string {
	resc.init()
	return resc.StorageType
}

func (resc *Resource) GetPhysPath() string {
	resc.init()
	return resc.PhysPath
}

func (resc *Resource) GetZone() *Zone {
	resc.init()
	return resc.Zone
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
		resc.Comment = infoMap["r_comment"]
		resc.CreateTime = timeStringToTime(infoMap["create_ts"])
		resc.ModifyTime = timeStringToTime(infoMap["modify_ts"])
		resc.Id, _ = strconv.Atoi(infoMap["resc_id"])
		resc.Type = ResourceType

		resc.Context = infoMap["resc_context"]
		resc.Class = typeMap[infoMap["resc_class_name"]]
		resc.Children = infoMap["resc_children"]
		resc.FreeSpace, _ = strconv.Atoi(infoMap["free_space"])
		resc.Info = infoMap["resc_info"]
		resc.Status = infoMap["resc_status"]
		resc.ParentStr = infoMap["resc_parent"]
		resc.Net = infoMap["resc_net"]
		resc.FreeSpaceTime = timeStringToTime(infoMap["free_space_ts"])
		resc.ObjCount, _ = strconv.Atoi(infoMap["resc_objcount"])
		resc.StorageType = infoMap["resc_type_name"]
		resc.PhysPath = infoMap["resc_def_path"]

		if zones, err := resc.Con.GetZones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], resc.Con); zne != nil {
				resc.Zone = zne
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

	cResource := C.CString(resc.Name)
	defer C.free(unsafe.Pointer(cResource))

	ccon := resc.Con.GetCcon()

	if status := C.gorods_get_resource(cResource, ccon, &result, &err); status != 0 {
		resc.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Resource Info Failed: %v", C.GoString(err)))
	}

	resc.Con.ReturnCcon(ccon)

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
