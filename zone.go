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
	name string

	typ         int
	conString   string
	comment     string
	createTime  time.Time
	modifyTime  time.Time
	id          int
	parentSlice *Zones

	hasInit bool

	con *Connection
}

type Zones []*Zone

func (znes Zones) FindByName(name string, con *Connection) *Zone {
	for _, zne := range znes {
		if zne.name == name {
			return zne
		}
	}

	zne, _ := initZone(name, con)

	return zne
}

func (znes *Zones) Remove(index int) {
	*znes = append((*znes)[:index], (*znes)[index+1:]...)
}

func initZone(name string, con *Connection) (*Zone, error) {
	zne := new(Zone)

	zne.con = con
	zne.name = name

	return zne, nil
}

func (zne *Zone) init() error {
	if !zne.hasInit {
		if err := zne.RefreshInfo(); err != nil {
			return err
		}
		zne.hasInit = true
	}

	return nil
}

func (zne *Zone) Remove() bool {
	for n, p := range *zne.parentSlice {
		if p.name == zne.name {
			zne.parentSlice.Remove(n)
			return true
		}
	}

	return false
}

func (zne *Zone) String() string {
	zne.init()
	return fmt.Sprintf("%v:%v", getTypeString(zne.typ), zne.name)
}

func (zne *Zone) Name() string {
	return zne.name
}

func (zne *Zone) Comment() (string, error) {
	if err := zne.init(); err != nil {
		return zne.comment, err
	}
	return zne.comment, nil
}

func (zne *Zone) CreateTime() (time.Time, error) {
	if err := zne.init(); err != nil {
		return zne.createTime, err
	}
	return zne.createTime, nil
}

func (zne *Zone) ModifyTime() (time.Time, error) {
	if err := zne.init(); err != nil {
		return zne.modifyTime, err
	}
	return zne.modifyTime, nil
}

func (zne *Zone) Id() (int, error) {
	if err := zne.init(); err != nil {
		return zne.id, err
	}
	return zne.id, nil
}

func (zne *Zone) Type() (int, error) {
	if err := zne.init(); err != nil {
		return zne.typ, err
	}
	return zne.typ, nil
}

func (zne *Zone) ConString() (string, error) {
	if err := zne.init(); err != nil {
		return zne.conString, err
	}
	return zne.conString, nil
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
		zne.comment = infoMap["r_comment"]
		zne.createTime = timeStringToTime(infoMap["create_ts"])
		zne.modifyTime = timeStringToTime(infoMap["modify_ts"])
		zne.id, _ = strconv.Atoi(infoMap["zone_id"])
		zne.typ = typeMap[infoMap["zone_type_name"]]
		zne.conString = infoMap["zone_conn_string"]
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

	cZone := C.CString(zne.name)
	defer C.free(unsafe.Pointer(cZone))

	ccon := zne.con.GetCcon()

	if status := C.gorods_get_zone(cZone, ccon, &result, &err); status != 0 {
		zne.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Zone Info Failed: %v", C.GoString(err)))
	}

	zne.con.ReturnCcon(ccon)

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
