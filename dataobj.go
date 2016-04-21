/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"unsafe"
)

// DataObj structs contain information about single data objects in an iRods zone.
type DataObj struct {
	Path   string
	Name   string
	Size   int64
	Offset int64

	MetaCol MetaCollection

	// Con field is a pointer to the Connection used to fetch the data object
	Con *Connection

	// Col field is a pointer to the Collection containing the data object
	Col *Collection

	chandle C.int
}

// DataObjOptions is used for passing options to the CreateDataObj function
type DataObjOptions struct {
	Name     string
	Size     int64
	Mode     int
	Force    bool
	Resource string
}


type DataObjs []*DataObj

func (dos DataObjs) Exists(path string) bool {
	if d := dos.Find(path); d != nil {
		return true
	}

	return false
}

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

func initDataObj(data *C.collEnt_t, col *Collection) *DataObj {

	dataObj := new(DataObj)

	dataObj.Col = col
	dataObj.Con = dataObj.Col.Con
	dataObj.Offset = 0
	dataObj.Name = C.GoString(data.dataName)
	dataObj.Path = C.GoString(data.collName) + "/" + dataObj.Name
	dataObj.Size = int64(data.dataSize)

	return dataObj
}

func CreateDataObj(opts DataObjOptions, coll *Collection) *DataObj {

	var (
		errMsg *C.char
		handle C.int
		force  int
	)

	if opts.Force {
		force = 1
	} else {
		force = 0
	}

	path := C.CString(coll.Path + "/" + opts.Name)
	resource := C.CString(opts.Resource)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(resource))

	if status := C.gorods_create_dataobject(path, C.rodsLong_t(opts.Size), C.int(opts.Mode), C.int(force), resource, &handle, coll.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Create DataObject Failed: %v, Does the file already exist?", C.GoString(errMsg)))
	}

	dataObj := new(DataObj)

	dataObj.Col = coll
	dataObj.Con = dataObj.Col.Con
	dataObj.Offset = 0
	dataObj.Name = opts.Name
	dataObj.Path = C.GoString(path)
	dataObj.Size = opts.Size
	dataObj.chandle = handle

	coll.add(dataObj)

	return dataObj

}

func (obj *DataObj) init() {
	if int(obj.chandle) == 0 {
		obj.Open()
	}
}

func (obj *DataObj) Open() {
	var errMsg *C.char

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	if status := C.gorods_open_dataobject(path, &obj.chandle, C.rodsLong_t(obj.Size), obj.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Open DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}
}

func (obj *DataObj) Close() *DataObj {
	var errMsg *C.char

	if status := C.gorods_close_dataobject(obj.chandle, obj.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Close DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}

	obj.chandle = C.int(0)

	return obj
}

func (obj *DataObj) Read() []byte {
	obj.init()

	var (
		buffer C.bytesBuf_t
		err    *C.char
	)

	obj.LSeek(0)

	if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(obj.Size), &buffer, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	buf := unsafe.Pointer(buffer.buf)
	defer C.free(buf)

	data := C.GoBytes(buf, C.int(obj.Size))

	obj.Close()

	return data
}

func (obj *DataObj) ReadBytes(pos int64, length int) []byte {
	obj.init()

	var (
		buffer C.bytesBuf_t
		err    *C.char
	)

	obj.LSeek(pos)

	if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(length), &buffer, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods ReadBytes DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	buf := unsafe.Pointer(buffer.buf)
	defer C.free(buf)

	data := C.GoBytes(buf, C.int(obj.Size))

	return data
}

func (obj *DataObj) LSeek(offset int64) *DataObj {
	obj.init()

	var (
		err *C.char
	)

	if status := C.gorods_lseek_dataobject(obj.chandle, C.rodsLong_t(offset), obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods LSeek DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Offset = offset

	return obj
}

func (obj *DataObj) ReadChunk(size int64, callback func([]byte)) *DataObj {
	obj.init()

	var (
		buffer C.bytesBuf_t
		err    *C.char
	)

	obj.LSeek(0)

	for obj.Offset < obj.Size {
		if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(size), &buffer, obj.Con.ccon, &err); status != 0 {
			panic(fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
		}

		buf := unsafe.Pointer(buffer.buf)

		chunk := C.GoBytes(buf, C.int(size))
		C.free(buf)

		callback(chunk)

		obj.LSeek(obj.Offset + size)
	}

	obj.LSeek(0)

	obj.Close()

	return obj
}

func (obj *DataObj) DownloadTo(localPath string) *DataObj {
	obj.init()


	if err := ioutil.WriteFile(localPath, obj.Read(), 0644); err != nil {
		panic(fmt.Sprintf("iRods Download DataObject Failed: %v, %v", obj.Path, err))
	}

	return obj
}

func (obj *DataObj) Write(data []byte) *DataObj {
	obj.init()

	obj.LSeek(0)

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0]) // Do I need to free this? It might be done by go

	var err *C.char
	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Size = size

	obj.Close()

	return obj
}

func (obj *DataObj) WriteBytes(data []byte) *DataObj {
	obj.init()

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0]) // Do I need to free this? It might be done by go

	var err *C.char
	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Size = size + obj.Offset

	obj.LSeek(obj.Size)

	return obj
}

func (obj *DataObj) Stat() map[string]interface{} {
	obj.init()

	var (
		err        *C.char
		statResult *C.rodsObjStat_t
	)

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	if status := C.gorods_stat_dataobject(path, &statResult, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Close Stat Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	result := make(map[string]interface{})

	result["objSize"] = int(statResult.objSize)
	result["dataMode"] = int(statResult.dataMode)

	result["dataId"] = C.GoString(&statResult.dataId[0])
	result["chksum"] = C.GoString(&statResult.chksum[0])
	result["ownerName"] = C.GoString(&statResult.ownerName[0])
	result["ownerZone"] = C.GoString(&statResult.ownerZone[0])
	result["createTime"] = C.GoString(&statResult.createTime[0])
	result["modifyTime"] = C.GoString(&statResult.modifyTime[0])

	C.freeRodsObjStat(statResult)

	return result
}

func (obj *DataObj) Attribute(attr string) *Meta {
	obj.init()

	return obj.Meta().Get(attr)
}

func (obj *DataObj) Meta() MetaCollection {
	obj.init()

	if obj.MetaCol == nil {
		obj.MetaCol = NewMetaCollection(DataObjType, obj.Name, obj.Col.Path, obj.Con.ccon)
	}
	
	return obj.MetaCol
}

// Supports Collection struct and string
func (obj *DataObj) CopyTo(iRodsCollection interface{}) *DataObj {

	var (
		err                         *C.char
		destination                 string
		destinationCollectionString string
		destinationCollection       *Collection
	)

	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		destinationCollectionString = iRodsCollection.(string)

		// Is this a relative path?
		if destinationCollectionString[0] != '/' {
			destinationCollectionString = obj.Col.Path + "/" + destinationCollectionString
		}

		if destinationCollectionString[len(destinationCollectionString)-1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + obj.Name

	} else {
		destinationCollectionString = (iRodsCollection.(*Collection)).Path + "/"
		destination = destinationCollectionString + obj.Name
	}

	path := C.CString(obj.Path)
	dest := C.CString(destination)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(dest))

	if status := C.gorods_copy_dataobject(path, dest, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Copy DataObject Failed: %v, %v", destination, C.GoString(err)))
	}

	// reload destination collection
	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		// Find collection recursivly
		if destinationCollection = obj.Con.OpenedCollections.FindRecursive(destinationCollectionString); destinationCollection != nil {
			destinationCollection.init()
		} else {
			// Can't find, load collection into memory
			destinationCollection, _ = obj.Con.Collection(destinationCollectionString, false)
		}
	} else {
		destinationCollection = (iRodsCollection.(*Collection))
		destinationCollection.init()
	}

	return destinationCollection.DataObjs().Find(obj.Name)
}

// Supports Collection struct and string
func (obj *DataObj) MoveTo(iRodsCollection interface{}) *DataObj {

	var (
		err                         *C.char
		destination                 string
		destinationCollectionString string
		destinationCollection       *Collection
	)

	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		destinationCollectionString = iRodsCollection.(string)

		// Is this a relative path?
		if destinationCollectionString[0] != '/' {
			destinationCollectionString = obj.Col.Path + "/" + destinationCollectionString
		}

		if destinationCollectionString[len(destinationCollectionString)-1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + obj.Name

	} else {
		destinationCollectionString = (iRodsCollection.(*Collection)).Path + "/"
		destination = destinationCollectionString + obj.Name
	}

	path := C.CString(obj.Path)
	dest := C.CString(destination)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(dest))

	if status := C.gorods_move_dataobject(path, dest, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Move DataObject Failed S:%v, D:%v, %v", obj.Path, destination, C.GoString(err)))
	}

	// Reload source collection, we are now detached
	obj.Col.init()

	// Find & reload destination collection
	if reflect.TypeOf(iRodsCollection).Kind() == reflect.String {
		// Find collection recursivly
		if destinationCollection = obj.Con.OpenedCollections.FindRecursive(destinationCollectionString); destinationCollection != nil {
			destinationCollection.init()
		} else {
			// Can't find, load collection into memory
			destinationCollection, _ = obj.Con.Collection(destinationCollectionString, false)
		}
	} else {
		destinationCollection = (iRodsCollection.(*Collection))
		destinationCollection.init()
	}

	// Reassign obj.Col to destination collection
	obj.Col = destinationCollection
	obj.Path = destinationCollection.Path + "/" + obj.Name

	obj.chandle = C.int(0)

	return obj
}

func (obj *DataObj) Rename(name string) *DataObj {

	if strings.Contains(name, "/") {
		panic(fmt.Sprintf("Can't Rename DataObject, path detected in: %v", name))
	}

	var err *C.char

	source := obj.Path
	destination := obj.Col.Path + "/" + name

	s := C.CString(source)
	d := C.CString(destination)

	defer C.free(unsafe.Pointer(s))
	defer C.free(unsafe.Pointer(d))

	if status := C.gorods_move_dataobject(s, d, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Rename DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Name = name
	obj.Path = destination

	obj.chandle = C.int(0)

	return obj
}

func (obj *DataObj) Delete() {

	var err *C.char

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	if status := C.gorods_unlink_dataobject(path, C.int(1), obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Delete DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

}

func (obj *DataObj) Unlink() {

	var err *C.char

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	if status := C.gorods_unlink_dataobject(path, C.int(0), obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Delete DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

}

func (obj *DataObj) Chksum() string {

	var (
		err       *C.char
		chksumOut *C.char
	)

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(chksumOut))

	if status := C.gorods_checksum_dataobject(path, &chksumOut, obj.Con.ccon, &err); status != 0 {
		panic(fmt.Sprintf("iRods Chksum DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	return C.GoString(chksumOut)
}

func (obj *DataObj) Verify(checksum string) bool {
	chksum := strings.Split(obj.Chksum(), ":")

	return (checksum == chksum[1])
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
