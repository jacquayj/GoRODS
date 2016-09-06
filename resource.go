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

	zne, er := con.GetLocalZone()
	if er != nil {
		return nil, er
	}

	resc.Con = con
	resc.Name = name
	resc.Zone = zne

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

func (resc *Resource) GetComment() (string, error) {
	if err := resc.init(); err != nil {
		return resc.Comment, err
	}
	return resc.Comment, nil
}

func (resc *Resource) GetCreateTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.CreateTime, err
	}
	return resc.CreateTime, nil
}

func (resc *Resource) GetModifyTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.ModifyTime, err
	}
	return resc.ModifyTime, nil
}

func (resc *Resource) GetId() (int, error) {
	if err := resc.init(); err != nil {
		return resc.Id, err
	}
	return resc.Id, nil
}

func (resc *Resource) GetType() (int, error) {
	if err := resc.init(); err != nil {
		return resc.Type, err
	}
	return resc.Type, nil
}

func (resc *Resource) GetContext() (string, error) {
	if err := resc.init(); err != nil {
		return resc.Context, err
	}
	return resc.Context, nil
}

func (resc *Resource) GetClass() (int, error) {
	if err := resc.init(); err != nil {
		return resc.Class, err
	}
	return resc.Class, nil
}

func (resc *Resource) GetChildren() (string, error) {
	if err := resc.init(); err != nil {
		return resc.Children, err
	}
	return resc.Children, nil
}

func (resc *Resource) GetFreeSpace() (int, error) {
	if err := resc.init(); err != nil {
		return resc.FreeSpace, err
	}
	return resc.FreeSpace, nil
}

func (resc *Resource) GetInfo() (string, error) {
	if err := resc.init(); err != nil {
		return resc.Info, err
	}
	return resc.Info, nil
}

func (resc *Resource) GetStatus() (string, error) {
	if err := resc.init(); err != nil {
		return resc.Status, err
	}
	return resc.Status, nil
}

func (resc *Resource) GetParentStr() (string, error) {
	if err := resc.init(); err != nil {
		return resc.ParentStr, err
	}
	return resc.ParentStr, nil
}

func (resc *Resource) GetNet() (string, error) {
	if err := resc.init(); err != nil {
		return resc.Net, err
	}
	return resc.Net, nil
}

func (resc *Resource) GetFreeSpaceTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.FreeSpaceTime, err
	}
	return resc.FreeSpaceTime, nil
}

func (resc *Resource) GetObjCount() (int, error) {
	if err := resc.init(); err != nil {
		return resc.ObjCount, err
	}
	return resc.ObjCount, nil
}

func (resc *Resource) GetStorageType() (string, error) {
	if err := resc.init(); err != nil {
		return resc.StorageType, err
	}
	return resc.StorageType, nil
}

func (resc *Resource) GetPhysPath() (string, error) {
	if err := resc.init(); err != nil {
		return resc.PhysPath, err
	}
	return resc.PhysPath, nil
}

func (resc *Resource) GetZone() *Zone {
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
