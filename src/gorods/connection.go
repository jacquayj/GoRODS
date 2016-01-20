package gorods

// #cgo LDFLAGS: -L./lib/build -lgorods
// #cgo CFLAGS: -I./lib/include -I./lib/irods/lib/core/include -I./lib/irods/lib/api/include -I./lib/irods/lib/md5/include -I./lib/irods/lib/sha1/include -I./lib/irods/server/core/include -I./lib/irods/server/icat/include -I./lib/irods/server/drivers/include -I./lib/irods/server/re/include
// #include "wrapper.h"
import "C"

import (
	"fmt"
)

type Connection struct {
	ccon *C.rcComm_t
	Connected bool

}

func (obj *Connection) String() string {
	return "Collection: "
}


func (con *Connection) GetCollection(startPath string) *Collection {
	return NewCollection(startPath, con)
}

func NewConnection() *Connection {
	con := &Connection { }

	var errMsg *C.char;

	if status := C.gorods_connect(&con.ccon, &errMsg); status == 0 {
		con.Connected = true
	} else {
		panic(fmt.Sprintf("iRods Connect Failed: %v", C.GoString(errMsg)))
	}

	return con
}





//


