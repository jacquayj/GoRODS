package gorods

// #cgo LDFLAGS: -L./lib/build -lgorods
// #cgo CFLAGS: -I./lib/include -I./lib/irods/lib/core/include -I./lib/irods/lib/api/include -I./lib/irods/lib/md5/include -I./lib/irods/lib/sha1/include -I./lib/irods/server/core/include -I./lib/irods/server/icat/include -I./lib/irods/server/drivers/include -I./lib/irods/server/re/include
// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	System = iota
	UserDefined
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

func New(opts ConnectionOptions) *Connection {
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

		status = C.gorods_connect_env(&con.ccon, host, port, username, zone, password, &errMsg)
	} else {
		status = C.gorods_connect(&con.ccon, password, &errMsg)
	}

	if status == 0 {
		con.Connected = true
	} else {
		panic(fmt.Sprintf("iRods Connect Failed: %v", C.GoString(errMsg)))
	}

	return con
}

func (con *Connection) Disconnect() {
	C.rcDisconnect(con.ccon)
	con.Connected = false
}

func (obj *Connection) String() string {
	cEnvString := C.irods_env_str()

	defer C.free(unsafe.Pointer(cEnvString))

	envString := C.GoString(cEnvString)

	return fmt.Sprintf("Connection Env:\n%v\tConnected: %v\n", envString, obj.Connected)
}

func (con *Connection) Collection(startPath string, recursive bool) *Collection {
	collection := GetCollection(startPath, recursive, con)

	con.OpenedCollections = append(con.OpenedCollections, collection)

	return collection
}
