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

// Resource holds information about a resource server registered with iCAT.
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

// Resources is a slice of *Resource.
type Resources []*Resource

// FindByName returns the matching *Resource based on the name.
func (rescs Resources) FindByName(name string) *Resource {
	for _, resc := range rescs {
		if resc.name == name {
			return resc
		}
	}
	return nil
}

// Remove removes the specific element from itself, based on index.
func (rescs *Resources) Remove(index int) {
	*rescs = append((*rescs)[:index], (*rescs)[index+1:]...)
}

func initResource(name string, con *Connection) (*Resource, error) {
	resc := new(Resource)

	zne, er := con.LocalZone()
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

// Remove removes the resource from it's parent slice
func (resc *Resource) Remove() bool {
	for n, p := range *resc.parentSlice {
		if p.name == resc.name {
			resc.parentSlice.Remove(n)
			return true
		}
	}

	return false
}

// String returns the resource name and zone name
func (resc *Resource) String() string {
	return fmt.Sprintf("%v#%v", resc.Name(), resc.Zone().Name())
}

// Name returns the resource name
func (resc *Resource) Name() string {
	return resc.name
}

// Comment loads data from iCAT if needed, and returns the resources comment attribute.
func (resc *Resource) Comment() (string, error) {
	if err := resc.init(); err != nil {
		return resc.comment, err
	}
	return resc.comment, nil
}

// CreateTime loads data from iCAT if needed, and returns the resources createTime attribute.
func (resc *Resource) CreateTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.createTime, err
	}
	return resc.createTime, nil
}

// ModifyTime loads data from iCAT if needed, and returns the resources modifyTime attribute.
func (resc *Resource) ModifyTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.modifyTime, err
	}
	return resc.modifyTime, nil
}

// Id loads data from iCAT if needed, and returns the resources id attribute.
func (resc *Resource) Id() (int, error) {
	if err := resc.init(); err != nil {
		return resc.id, err
	}
	return resc.id, nil
}

// Id loads data from iCAT if needed, and returns the resources id attribute.
func (resc *Resource) Type() (int, error) {
	if err := resc.init(); err != nil {
		return resc.typ, err
	}
	return resc.typ, nil
}

// Context loads data from iCAT if needed, and returns the resources context attribute.
func (resc *Resource) Context() (string, error) {
	if err := resc.init(); err != nil {
		return resc.context, err
	}
	return resc.context, nil
}

// Class loads data from iCAT if needed, and returns the resources class attribute.
func (resc *Resource) Class() (int, error) {
	if err := resc.init(); err != nil {
		return resc.class, err
	}
	return resc.class, nil
}

// Children loads data from iCAT if needed, and returns the resources children attribute.
func (resc *Resource) Children() (string, error) {
	if err := resc.init(); err != nil {
		return resc.children, err
	}
	return resc.children, nil
}

// FreeSpace loads data from iCAT if needed, and returns the resources freeSpace attribute.
func (resc *Resource) FreeSpace() (int, error) {
	if err := resc.init(); err != nil {
		return resc.freeSpace, err
	}
	return resc.freeSpace, nil
}

// Info loads data from iCAT if needed, and returns the resources info attribute.
func (resc *Resource) Info() (string, error) {
	if err := resc.init(); err != nil {
		return resc.info, err
	}
	return resc.info, nil
}

// Status loads data from iCAT if needed, and returns the resources status attribute.
func (resc *Resource) Status() (string, error) {
	if err := resc.init(); err != nil {
		return resc.status, err
	}
	return resc.status, nil
}

// ParentStr loads data from iCAT if needed, and returns the resources parentStr attribute.
func (resc *Resource) ParentStr() (string, error) {
	if err := resc.init(); err != nil {
		return resc.parentStr, err
	}
	return resc.parentStr, nil
}

// Net loads data from iCAT if needed, and returns the resources net attribute.
func (resc *Resource) Net() (string, error) {
	if err := resc.init(); err != nil {
		return resc.net, err
	}
	return resc.net, nil
}

// FreeSpaceTime loads data from iCAT if needed, and returns the resources freeSpaceTime attribute.
func (resc *Resource) FreeSpaceTime() (time.Time, error) {
	if err := resc.init(); err != nil {
		return resc.freeSpaceTime, err
	}
	return resc.freeSpaceTime, nil
}

// ObjCount loads data from iCAT if needed, and returns the resources status attribute.
func (resc *Resource) ObjCount() (int, error) {
	if err := resc.init(); err != nil {
		return resc.objCount, err
	}
	return resc.objCount, nil
}

// Status loads data from iCAT if needed, and returns the resources info attribute.
func (resc *Resource) StorageType() (string, error) {
	if err := resc.init(); err != nil {
		return resc.storageType, err
	}
	return resc.storageType, nil
}

// PhysPath loads data from iCAT if needed, and returns the resources physPath attribute.
func (resc *Resource) PhysPath() (string, error) {
	if err := resc.init(); err != nil {
		return resc.physPath, err
	}
	return resc.physPath, nil
}

// Zone returns the *Zone to which the resource belongs to.
func (resc *Resource) Zone() *Zone {
	return resc.zone
}

// RefreshInfo fetches fresh resource data from the iCAT server, and unloads it to the resource struct's fields.
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

		if zones, err := resc.con.Zones(); err != nil {
			return err
		} else {
			if zne := zones.FindByName(infoMap["zone_name"], resc.con); zne != nil {
				resc.zone = zne
			} else {
				return newError(Fatal, fmt.Sprintf("iRODS Refresh Resource Info Failed: Unable to locate zone in cache"))
			}
		}

	} else {
		return err
	}

	return nil
}

// FetchInfo fetches fresh resource info from the iCAT server and returns it as a map
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
		return nil, newError(Fatal, fmt.Sprintf("iRODS Get Resource Info Failed: %v", C.GoString(err)))
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
