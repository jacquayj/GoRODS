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

type DataObjOptions struct {
	Name string
	Size int
	Mode int
	Force bool
	Resource string
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

func CreateDataObj(opts *DataObjOptions, coll *Collection) *DataObj {
	
	var errMsg *C.char
	var handle C.int
	
	var force int

	if opts.Force {
		force = 1
	} else {
		force = 0
	}
	
	path := coll.Path + "/" + opts.Name

	if status := C.gorods_create_dataobject(C.CString(path), C.rodsLong_t(opts.Size), C.int(opts.Mode), C.int(force), C.CString(opts.Resource), &handle, coll.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Create DataObject Failed: %v, Does the file already exist?", C.GoString(errMsg)))
	}

	coll.ReadCollection()

	return coll.DataObjs().Find(opts.Name)

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

func (obj *DataObj) Stat() map[string]interface{} {

	var err *C.char
	var statResult *C.rodsObjStat_t

	if status := C.gorods_stat_dataobject(C.CString(obj.Path), &statResult, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Close Stat Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	result := make(map[string]interface{})

	result["objSize"]      = int(statResult.objSize)
	result["dataMode"]     = int(statResult.dataMode)

	result["dataId"]       = C.GoString(&statResult.dataId[0])
	result["chksum"]       = C.GoString(&statResult.chksum[0])
	result["ownerName"]    = C.GoString(&statResult.ownerName[0])
	result["ownerZone"]    = C.GoString(&statResult.ownerZone[0])
	result["createTime"]   = C.GoString(&statResult.createTime[0])
	result["modifyTime"]   = C.GoString(&statResult.modifyTime[0])

	C.freeRodsObjStat(statResult)

	return result
}

