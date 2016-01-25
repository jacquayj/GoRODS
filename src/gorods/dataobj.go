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

	chandle C.int
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

func (obj *DataObj) Init() {
	if int(obj.chandle) == 0 {
		obj.Open()
	}
}

func (obj *DataObj) Open() {
	var errMsg *C.char

	if status := C.gorods_open_dataobject(C.CString(obj.Path), &obj.chandle, obj.collent.dataSize, obj.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Open DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}
}

func (obj *DataObj) Close() {
	var errMsg *C.char

	if status := C.gorods_close_dataobject(obj.chandle, obj.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Close DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}
}

func (obj *DataObj) Read() []byte {
	obj.Init()

	var buffer C.bytesBuf_t
	var err *C.char

	if status := C.gorods_read_dataobject(obj.chandle, obj.collent.dataSize, &buffer, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	data := C.GoBytes(unsafe.Pointer(buffer.buf), C.int(obj.collent.dataSize))

	obj.Close()

	return data
}
