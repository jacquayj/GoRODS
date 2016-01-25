package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
)

type DataObj struct {
	Path string
	Name string

	Con *Connection
	Col *Collection
	collent *C.collEnt_t
}

type DataObjs []*DataObj

func (dos DataObjs) Find(path string) *DataObj {
	for i, do := range dos {
		if do.Path == path || do.Name == path {
			return dos[i]
		}
	}

	return nil
}

func (obj *DataObj) String() string {
	return "DataObject: " + obj.Path
}

func NewDataObj(data *C.collEnt_t, col *Collection) *DataObj {
	
	dataObj := new(DataObj)
	
	dataObj.collent = data

	dataObj.Col = col
	dataObj.Con = dataObj.Col.Con
	dataObj.Name = C.GoString(dataObj.collent.dataName)
	dataObj.Path = C.GoString(dataObj.collent.collName) + "/" + dataObj.Name

	return dataObj
}


func (obj *DataObj) Read() []byte {

	var buffer C.bytesBuf_t
	var err *C.char

	if status := C.gorods_read_dataobject(C.CString(obj.Path), obj.collent.dataSize, &buffer, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	data := C.GoBytes(unsafe.Pointer(buffer.buf), C.int(obj.collent.dataSize))

	return data
}