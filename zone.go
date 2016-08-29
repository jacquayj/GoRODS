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

type Zone struct {
	Name string

	Type        int
	ConString   string
	Comment     string
	CreateTime  time.Time
	ModifyTime  time.Time
	Id          int
	ParentSlice *Zones

	Init bool

	Con *Connection
}

type Zones []*Zone

func (znes Zones) FindByName(name string) *Zone {
	for _, zne := range znes {
		if zne.Name == name {
			return zne
		}
	}
	return nil
}

func (znes *Zones) Remove(index int) {
	*znes = append((*znes)[:index], (*znes)[index+1:]...)
}

func initZone(name string, con *Connection) (*Zone, error) {
	zne := new(Zone)

	zne.Con = con
	zne.Name = name

	return zne, nil
}

func (zne *Zone) init() error {
	if !zne.Init {
		if err := zne.RefreshInfo(); err != nil {
			return err
		}
		zne.Init = true
	}

	return nil
}

func (zne *Zone) Remove() bool {
	for n, p := range *zne.ParentSlice {
		if p.Name == zne.Name {
			zne.ParentSlice.Remove(n)
			return true
		}
	}

	return false
}

func (zne *Zone) String() string {
	zne.init()
	return fmt.Sprintf("%v:%v", getTypeString(zne.Type), zne.Name)
}

func (zne *Zone) GetName() string {
	return zne.Name
}

func (zne *Zone) GetComment() string {
	zne.init()
	return zne.Comment
}

func (zne *Zone) GetCreateTime() time.Time {
	zne.init()
	return zne.CreateTime
}

func (zne *Zone) GetModifyTime() time.Time {
	zne.init()
	return zne.ModifyTime
}

func (zne *Zone) GetId() int {
	zne.init()
	return zne.Id
}

func (zne *Zone) GetType() int {
	zne.init()
	return zne.Type
}

func (zne *Zone) GetConString() string {
	zne.init()
	return zne.ConString
}

func (zne *Zone) RefreshInfo() error {

	// zone_name:tempZone
	// zone_type_name:local
	// zone_conn_string:
	// r_comment:
	// create_ts:1170000000
	// modify_ts:1170000000
	// zone_id:9000

	typeMap := map[string]int{
		"local":  Local,
		"remote": Remote,
	}

	if infoMap, err := zne.FetchInfo(); err == nil {
		zne.Comment = infoMap["r_comment"]
		zne.CreateTime = TimeStringToTime(infoMap["create_ts"])
		zne.ModifyTime = TimeStringToTime(infoMap["modify_ts"])
		zne.Id, _ = strconv.Atoi(infoMap["zone_id"])
		zne.Type = typeMap[infoMap["zone_type_name"]]
		zne.ConString = infoMap["zone_conn_string"]
	} else {
		return err
	}

	return nil
}

func (zne *Zone) FetchInfo() (map[string]string, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	cZone := C.CString(zne.Name)
	defer C.free(unsafe.Pointer(cZone))

	ccon := zne.Con.GetCcon()

	if status := C.gorods_get_zone(cZone, ccon, &result, &err); status != 0 {
		zne.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Zone Info Failed: %v", C.GoString(err)))
	}

	zne.Con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(map[string]string)

	for _, zoneInfo := range slice {

		zoneAttributes := strings.Split(strings.Trim(C.GoString(zoneInfo), " \n"), "\n")

		for _, attr := range zoneAttributes {

			split := strings.Split(attr, ": ")

			attrName := split[0]
			attrVal := split[1]

			response[attrName] = attrVal

		}
	}

	return response, nil
}
