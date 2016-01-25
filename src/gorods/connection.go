package gorods

// #cgo LDFLAGS: -L./lib/build -lgorods
// #cgo CFLAGS: -I./lib/include -I./lib/irods/lib/core/include -I./lib/irods/lib/api/include -I./lib/irods/lib/md5/include -I./lib/irods/lib/sha1/include -I./lib/irods/server/core/include -I./lib/irods/server/icat/include -I./lib/irods/server/drivers/include -I./lib/irods/server/re/include
// #include "wrapper.h"
import "C"

import (
	"fmt"
)

type Options struct {
	Host string
	Port int
	Zone string

	Username string
	Password string
}

type Connection struct {
	ccon *C.rcComm_t

	Connected bool
	Options *Options
}

func New(opts *Options) *Connection {
	con := new(Connection)

	con.Options = opts

	var status C.int
	var errMsg *C.char
	var password *C.char

	if con.Options.Password != "" {
		password = C.CString(con.Options.Password)
	}

	// Are we passing env values?
	if con.Options.Host != "" {
		host := C.CString(con.Options.Host)
		port := C.int(con.Options.Port)
		username := C.CString(con.Options.Username)
		zone := C.CString(con.Options.Zone)

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

func (obj *Connection) String() string {
	envString := C.GoString(C.irods_env_str())

	return fmt.Sprintf("Connection Env:\n%v\tConnected: %v\n", envString, obj.Connected)
}

func (con *Connection) Collection(startPath string, recursive bool) *Collection {
	return GetCollection(startPath, recursive, con)
}
