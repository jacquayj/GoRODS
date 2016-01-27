package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
	"reflect"
)

type DataObj struct {
	Path string
	Name string
	Size int64

	Con *Connection
	Col *Collection

	chandle C.int
}

type DataObjOptions struct {
	Name string
	Size int64
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

	dataObj.Col = col
	dataObj.Con = dataObj.Col.Con
	dataObj.Name = C.GoString(data.dataName)
	dataObj.Path = C.GoString(data.collName) + "/" + dataObj.Name
	dataObj.Size = int64(data.dataSize)

	return dataObj
}

func CreateDataObj(opts DataObjOptions, coll *Collection) *DataObj {
	
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

	dataObj := new(DataObj)

	dataObj.Col = coll
	dataObj.Con = dataObj.Col.Con
	dataObj.Name = opts.Name
	dataObj.Path = path
	dataObj.Size = opts.Size
	dataObj.chandle = handle

	coll.Add(dataObj)

	return dataObj

}

func (obj *DataObj) Init() {
	if int(obj.chandle) == 0 {
		obj.Open()
	}
}

func (obj *DataObj) Open() {
	var errMsg *C.char

	if status := C.gorods_open_dataobject(C.CString(obj.Path), &obj.chandle, C.rodsLong_t(obj.Size), obj.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Open DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}
}

func (obj *DataObj) Close() {
	var errMsg *C.char

	if status := C.gorods_close_dataobject(obj.chandle, obj.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Close DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}

	obj.chandle = C.int(0)

}

func (obj *DataObj) Read() []byte {
	obj.Init()

	var buffer C.bytesBuf_t
	var err *C.char

	if status := C.gorods_read_dataobject(obj.chandle,  C.rodsLong_t(obj.Size), &buffer, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	data := C.GoBytes(unsafe.Pointer(buffer.buf), C.int(obj.Size))

	obj.Close()

	return data
}

func (obj *DataObj) Write(data []byte) {
	obj.Init()

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0])

	var err *C.char
	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Size = size

	obj.Close()
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


// Supports Collection struct and string
func (obj *DataObj) CopyTo(iRodsCollection interface{}) *DataObj {
	
	var err *C.char
	var destination string
	var destinationCollectionString string

	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		destinationCollectionString = iRodsCollection.(string)

		if destinationCollectionString[len(destinationCollectionString) - 1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + obj.Name

	} else {
		destinationCollectionString = (iRodsCollection.(*Collection)).Path + "/"
		destination = destinationCollectionString + obj.Name
	}

	C.gorods_copy_dataobject(C.CString(obj.Path), C.CString(destination), obj.Con.ccon, &err)
	
	// reload destination collection
	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		// Find collection recursivly
		if expiredCollection := obj.Con.OpenedCollections.FindRecursive(destinationCollectionString[:len(destinationCollectionString) - 1]); expiredCollection != nil {
			expiredCollection.Init()
		}
	} else {
		(iRodsCollection.(*Collection)).Init()
	}

	return obj
}

// Supports Collection struct and string
func (obj *DataObj) MoveTo(iRodsCollection interface{}) *DataObj {
	
	var err *C.char
	var destination string
	var destinationCollectionString string
	var destinationCollection *Collection

	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		destinationCollectionString = iRodsCollection.(string)

		if destinationCollectionString[len(destinationCollectionString) - 1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + obj.Name

	} else {
		destinationCollectionString = (iRodsCollection.(*Collection)).Path + "/"
		destination = destinationCollectionString + obj.Name
	}

	C.gorods_move_dataobject(C.CString(obj.Path), C.CString(destination), obj.Con.ccon, &err)
	
	// Reload source collection, we are now detached
	obj.Col.Init()

	// Find & reload destination collection
	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		// Find collection recursivly
		if destinationCollection = obj.Con.OpenedCollections.FindRecursive(destinationCollectionString[:len(destinationCollectionString) - 1]); destinationCollection != nil {
			destinationCollection.Init()
		} else {
			// Can't find, load collection into memory
			destinationCollection = obj.Con.Collection(destinationCollectionString, false)
		}
	} else {
		destinationCollection = (iRodsCollection.(*Collection))
		destinationCollection.Init()
	}

	// Reassign .Col to destination collection
	obj.Col = destinationCollection

	return obj
}

// IMPLEMENT ME
func (obj *DataObj) Rename(name string) *DataObj {
	


	return obj
}


// IMPLEMENT ME
func (obj *DataObj) Delete(force bool) {
	



}

// IMPLEMENT ME
func (obj *DataObj) Chksum() *DataObj {
	


	return obj
}


// IMPLEMENT ME
func (obj *DataObj) MoveToResource(resource string) *DataObj {
	


	return obj
}

// IMPLEMENT ME
func (obj *DataObj) Replicate(resource string) *DataObj {
	


	return obj
}

// IMPLEMENT ME
func (obj *DataObj) ReplSettings(resource map[string]interface{}) *DataObj {
	//https://wiki.irods.org/doxygen_api/html/rc_data_obj_trim_8c_a7e4713d4b7617690e484fbada8560663.html


	return obj
}
