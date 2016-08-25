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
	"time"
	"unsafe"
)

func cTimeToTime(cTime *C.char) time.Time {
	unixStamp, _ := strconv.ParseInt(C.GoString(cTime), 10, 64)
	return time.Unix(unixStamp, 0)
}

func TimeStringToTime(ts string) time.Time {
	unixStamp, _ := strconv.ParseInt(ts, 10, 64)
	return time.Unix(unixStamp, 0)
}

func getTypeString(t int) string {
	switch t {
	case DataObjType:
		return "d"
	case CollectionType:
		return "C"
	case ResourceType:
		return "R"
	case UserType:
		return "u"
	default:
		panic(newError(Fatal, "unrecognized meta type constant"))
	}
}

func aclSliceToResponse(result *C.goRodsACLResult_t, con *Connection) (ACLs, error) {
	unsafeArr := unsafe.Pointer(result.aclArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.goRodsACL_t
	slice := (*[1 << 30]C.goRodsACL_t)(unsafeArr)[:arrLen:arrLen]

	response := make(ACLs, 0)

	for _, acl := range slice {

		typeString := C.GoString(acl.acltype)
		var aclType int
		switch typeString {
		case "rodsgroup":
			aclType = GroupType
		case "rodsadmin":
			aclType = UserType
		case "rodsuser":
			aclType = UserType
		default:
			aclType = UnknownType
		}

		accessString := C.GoString(acl.dataAccess)
		var accessLevel int
		switch accessString {
		case "own":
			accessLevel = Own
		case "modify object":
			accessLevel = Write
		case "read object":
			accessLevel = Read
		default:
			accessLevel = Null
		}

		var accessObject AccessObject
		if aclType == UserType {
			// ensure users are loaded
			if len(con.Users) == 0 {
				con.RefreshUsers()
			}

			if existingUsr := con.Users.FindByName(C.GoString(acl.name)); existingUsr != nil {
				accessObject = existingUsr
			} else {
				return nil, newError(Fatal, fmt.Sprintf("iRods GetACL Failed: can't find iRODS user by string"))
			}
		} else if aclType == GroupType {
			// ensure users are loaded
			if len(con.Groups) == 0 {
				con.RefreshGroups()
			}

			if existingGrp := con.Groups.FindByName(C.GoString(acl.name)); existingGrp != nil {
				accessObject = existingGrp
			} else {
				return nil, newError(Fatal, fmt.Sprintf("iRods GetACL Failed: can't find iRODS group by string"))
			}

		} else if aclType == UnknownType {
			return nil, newError(Fatal, fmt.Sprintf("iRods GetACL Failed: Unknown Type"))
		}

		response = append(response, &ACL{
			AccessObject: accessObject,
			AccessLevel:  accessLevel,
			Type:         aclType,
		})

	}

	C.gorods_free_acl_result(result)

	return response, nil
}

func isString(obj interface{}) bool {
	switch obj.(type) {
	case string:
		return true
	default:
	}

	return false
}
