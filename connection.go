/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a GoLang wrapper of the iRods C API (iRods client library). 
// gorods uses cgo to call iRods client functions.
package gorods

// #cgo CFLAGS: -I${SRCDIR}/lib/include -I${SRCDIR}/lib/irods/lib/core/include -I${SRCDIR}/lib/irods/lib/api/include -I${SRCDIR}/lib/irods/lib/md5/include -I${SRCDIR}/lib/irods/lib/sha1/include -I${SRCDIR}/lib/irods/server/core/include -I${SRCDIR}/lib/irods/server/icat/include -I${SRCDIR}/lib/irods/server/drivers/include -I${SRCDIR}/lib/irods/server/re/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/build -lgorods
// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
	"errors"
)

// System and UserDefined constants are used when calling
// gorods.New(ConnectionOptions{ Environment: ... })
// When System is specified, the options stored in ~/.irods/.irodsEnv will be used. 
// When UserDefined is specified you must also pass Host, Port, Username, and Zone. 
// Password should be set regardless. The five type constants are used internally for metadata
// operations.
const (
	System = iota
	UserDefined

	DataObjType
	CollectionType
	ResourceType
	ResourceGroupType
	UserType
)

type ConnectionOptions struct {
	Environment int

	Host string
	Port int
	Zone string

	Username string
	Password string
}

type Connection struct {
	ccon *C.rcComm_t

	Connected         bool
	Options           *ConnectionOptions
	OpenedCollections Collections
}


func New(opts ConnectionOptions) (*Connection, error) {
	con := new(Connection)

	con.Options = &opts

	var (
		status   C.int
		errMsg   *C.char
		password *C.char
	)

	if con.Options.Password != "" {
		password = C.CString(con.Options.Password)

		defer C.free(unsafe.Pointer(password))
	}

	// Are we passing env values?
	if con.Options.Environment == UserDefined {
		host := C.CString(con.Options.Host)
		port := C.int(con.Options.Port)
		username := C.CString(con.Options.Username)
		zone := C.CString(con.Options.Zone)

		defer C.free(unsafe.Pointer(host))
		defer C.free(unsafe.Pointer(username))
		defer C.free(unsafe.Pointer(zone))

		// BUG(jjacquay712): iRods C API code outputs errors messages, need to implement connect wrapper (gorods_connect_env) from a lower level to suppress this output
		// https://github.com/irods/irods/blob/master/iRODS/lib/core/src/rcConnect.cpp#L109
		status = C.gorods_connect_env(&con.ccon, host, port, username, zone, password, &errMsg)
	} else {
		// BUG(jjacquay712): C.gorods_connect implements getRodsEnv() which I believe reads the old ~/.irods/.irodsEnv file format
		status = C.gorods_connect(&con.ccon, password, &errMsg)
	}

	if status == 0 {
		con.Connected = true
	} else {
		return nil, errors.New(fmt.Sprintf("iRods Connect Failed: %v", C.GoString(errMsg)))
	}

	return con, nil
}

// Disconnect closes connection to iRods
func (con *Connection) Disconnect() {
	C.rcDisconnect(con.ccon)
	con.Connected = false
}

func (obj *Connection) String() string {
	if obj.Options.Environment == UserDefined {
		return fmt.Sprintf("Host: %v:%v/%v, Connected: %v\n", obj.Options.Host, obj.Options.Port, obj.Options.Zone, obj.Connected)
	}
	return fmt.Sprintf("Host: ?, Connected: %v\n", obj.Connected)
}

func (con *Connection) Collection(startPath string, recursive bool) (collection *Collection, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()

	collection = getCollection(startPath, recursive, con)
	con.OpenedCollections = append(con.OpenedCollections, collection)

	return
}

