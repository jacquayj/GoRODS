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

type Resource struct {
	Name string

	Type       int
	Comment    string
	CreateTime time.Time
	ModifyTime time.Time
	Id         int
	Parent     *Resources

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

	Init bool

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
	for n, p := range *resc.Parent {
		if p.Name == resc.Name {
			resc.Parent.Remove(n)
			return true
		}
	}

	return false
}

func (resc *Resource) String() string {
	//resc.init()
	return fmt.Sprintf("%v", resc.Name)
}

func (resc *Resource) GetName() string {
	resc.init()
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

	if infoMap, err := resc.FetchInfo(); err == nil {
		resc.Comment = infoMap["r_comment"]
		resc.CreateTime = TimeStringToTime(infoMap["create_ts"])
		resc.ModifyTime = TimeStringToTime(infoMap["modify_ts"])
		resc.Id, _ = strconv.Atoi(infoMap["resc_id"])
		resc.Type = ZoneType
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
