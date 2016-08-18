/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unsafe"
	"path/filepath"
	"strconv"
)

// DataObj structs contain information about single data objects in an iRods zone.
type DataObj struct {
	Path   string
	Name   string
	Checksum string
	Size   int64
	Offset int64
	Type   int
	DataId string
	Resource string
	PhyPath string

	OpenedAs C.int

	OwnerName string
	CreateTime int
	ModifyTime int

	MetaCol *MetaCollection

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


// String returns path of data object
func (obj *DataObj) String() string {
	return "DataObject: " + obj.Path
}

// We don't init() here or return errors here because it takes forever. Lazy loading is better in this case.
func initDataObj(data *C.collEnt_t, col *Collection) *DataObj {

	dataObj := new(DataObj)

	dataObj.Type = DataObjType
	dataObj.Col = col
	dataObj.Con = dataObj.Col.Con
	dataObj.Offset = 0
	dataObj.Name = C.GoString(data.dataName)
	dataObj.Path = C.GoString(data.collName) + "/" + dataObj.Name
	dataObj.Size = int64(data.dataSize)
	dataObj.chandle = C.int(-1)
	dataObj.Checksum = C.GoString(data.chksum)
	dataObj.DataId = C.GoString(data.dataId)
	dataObj.Resource = C.GoString(data.resource)
	dataObj.PhyPath = C.GoString(data.phyPath)
	dataObj.OpenedAs = C.int(-1)

	dataObj.OwnerName = C.GoString(data.ownerName)
	dataObj.CreateTime, _ = strconv.Atoi(C.GoString(data.createTime))
	dataObj.ModifyTime, _ = strconv.Atoi(C.GoString(data.modifyTime))

	return dataObj
}

// getDataObj initializes specified data object located at startPath using gorods.Connection.
// Could be considered alias of Connection.DataObject()
func getDataObj(startPath string, con *Connection) (*DataObj, error) {

	collectionDir := filepath.Dir(startPath)
	dataObjName := filepath.Base(startPath)

	if col, err := con.Collection(collectionDir, false); err == nil {
		if obj := col.FindObj(dataObjName); obj != nil {
			return obj, nil
		} else {
			return nil, newError(Fatal, fmt.Sprintf("Can't find DataObj within collection %v", collectionDir))
		}
	} else {
		return nil, err
	}

}


// CreateDataObj creates and adds a data object to the specified collection using provided options. Returns the newly created data object.
func CreateDataObj(opts DataObjOptions, coll *Collection) (*DataObj, error) {

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

	ccon := coll.Con.GetCcon()
	defer coll.Con.ReturnCcon(ccon)

	if status := C.gorods_create_dataobject(path, C.rodsLong_t(opts.Size), C.int(opts.Mode), C.int(force), resource, &handle, ccon, &errMsg); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Create DataObject Failed: %v, Does the file already exist?", C.GoString(errMsg)))
	}

	dataObj := new(DataObj)

	dataObj.Type = DataObjType
	dataObj.Col = coll
	dataObj.Con = dataObj.Col.Con
	dataObj.Offset = 0
	dataObj.Name = opts.Name
	dataObj.Path = C.GoString(path)
	dataObj.Size = opts.Size
	dataObj.chandle = handle
	dataObj.OpenedAs = C.int(-1)

	coll.add(dataObj)

	return dataObj, nil

}

func (obj *DataObj) init() error {
	if int(obj.chandle) < 0 {
		return obj.Open()
	}

	return nil
}

func (obj *DataObj) initRW() error {
	if int(obj.chandle) < 0 {
		return obj.OpenRW()
	}

	return nil
}

// GetACL retuns a slice of ACL structs. Example of slice in string format:
// [rods#tempZone:own
// developers#tempZone:modify object
// designers#tempZone:read object]
func (obj *DataObj) GetACL() (ACLs, error) {

	var (
		result C.goRodsACLResult_t
		err    *C.char
	)

	zoneHint := C.CString("tempZone")
	cDataId := C.CString(obj.DataId)
	defer C.free(unsafe.Pointer(cDataId))
	defer C.free(unsafe.Pointer(zoneHint))

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_get_dataobject_acl(ccon, cDataId, &result, zoneHint, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Data Object ACL Failed: %v", C.GoString(err)))
	}

	unsafeArr := unsafe.Pointer(result.aclArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.goRodsACL_t
	slice := (*[1 << 30]C.goRodsACL_t)(unsafeArr)[:arrLen:arrLen]

	response := make([]*ACL, 0)

	for _, acl := range slice {

		response = append(response, &ACL {
			Name: C.GoString(acl.name),
			Zone: C.GoString(acl.zone),
			DataAccess: C.GoString(acl.dataAccess),
			ACLType: C.GoString(acl.acltype),
		})

	}

	C.gorods_free_acl_result(&result)

	return response, nil

}

// Type gets the type
func (obj *DataObj) GetType() int {
	return obj.Type
}

// Connection returns the *Connection used to get data object
func (obj *DataObj) GetCon() *Connection {
	return obj.Con
}

// GetName returns the Name of the data object
func (obj *DataObj) GetName() string {
	return obj.Name
}

// GetName returns the Path of the data object
func (obj *DataObj) GetPath() string {
	return obj.Path
}

// GetName returns the *Collection of the data object
func (obj *DataObj) GetCol() *Collection {
	return obj.Col
}

// GetOwnerName returns the owner name of the data object
func (obj *DataObj) GetOwnerName() string {
	return obj.OwnerName
}

// GetCreateTime returns the create time of the data object
func (obj *DataObj) GetCreateTime() int {
	return obj.CreateTime
}

// GetModifyTime returns the modify time of the data object
func (obj *DataObj) GetModifyTime() int {
	return obj.ModifyTime
}

// Destroy is equivalent to irm -rf
func (obj *DataObj) Destroy() error {
	return obj.Rm(true, true)
}

// Delete is equivalent to irm -f {-r}
func (obj *DataObj) Delete(recursive bool) error {
	return obj.Rm(recursive, true)
}

// Trash is equivalent to irm {-r}
func (obj *DataObj) Trash(recursive bool) error {
	return obj.Rm(recursive, false)
}

// Rm is equivalent to irm {-r} {-f}
func (obj *DataObj) Rm(recursive bool, force bool) error {
	var errMsg *C.char

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	var (
		cForce C.int
		cRecursive C.int
	)

	if force {
		cForce = C.int(1)
	}

	if recursive {
		cRecursive = C.int(1)
	}

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_rm(path, 0, cRecursive, cForce, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rm DataObject Failed: %v", C.GoString(errMsg)))
	}

	return nil
}

// Open opens a connection to iRods and sets the data object handle
func (obj *DataObj) Open() error {
	var errMsg *C.char

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_open_dataobject(path, C.O_RDONLY, &obj.chandle, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Open DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}

	obj.OpenedAs = C.O_RDONLY

	return nil
}

// OpenRW opens a connection to iRods and sets the data object handle for read/write access
func (obj *DataObj) OpenRW() error {
	var errMsg *C.char

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_open_dataobject(path, C.O_RDWR, &obj.chandle, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods OpenRW DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
	}

	obj.OpenedAs = C.O_RDWR

	return nil
}

// Close closes the data object, resets handler
func (obj *DataObj) Close() error {
	var errMsg *C.char

	if int(obj.chandle) > -1 {

		ccon := obj.Con.GetCcon()
		defer obj.Con.ReturnCcon(ccon)

		if status := C.gorods_close_dataobject(obj.chandle, ccon, &errMsg); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Close DataObject Failed: %v, %v", obj.Path, C.GoString(errMsg)))
		}

		obj.chandle = C.int(-1)
	}

	return nil
}

// Read reads the entire data object into memory and returns a []byte slice. Don't use this for large files.
func (obj *DataObj) Read() ([]byte, error) {
	if er := obj.init(); er != nil {
		return nil, er
	}

	var (
		buffer C.bytesBuf_t
		err    *C.char
		bytesRead C.int
	)

	if er := obj.LSeek(0); er != nil {
		return nil, er
	}

	ccon := obj.Con.GetCcon()

	if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(obj.Size), &buffer, &bytesRead, ccon, &err); status != 0 {
		obj.Con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Con.ReturnCcon(ccon)

	buf := unsafe.Pointer(buffer.buf)
	defer C.free(buf)

	data := C.GoBytes(buf, bytesRead)

	return data, obj.Close()
}

// ReadBytes reads bytes from a data object at the specified position and length, returns []byte slice and error.
func (obj *DataObj) ReadBytes(pos int64, length int) ([]byte, error) {
	if er := obj.init(); er != nil {
		return nil, er
	}

	var (
		buffer C.bytesBuf_t
		err    *C.char
		bytesRead C.int
	)

	if er := obj.LSeek(pos); er != nil {
		return nil, er
	}


	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(length), &buffer, &bytesRead, ccon, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods ReadBytes DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	buf := unsafe.Pointer(buffer.buf)
	defer C.free(buf)

	data := C.GoBytes(buf, bytesRead)

	return data, nil
}

// LSeek sets the read/write offset pointer of a data object, returns error
func (obj *DataObj) LSeek(offset int64) error {
	if er := obj.init(); er != nil {
		return er
	}

	var (
		err *C.char
	)

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_lseek_dataobject(obj.chandle, C.rodsLong_t(offset), ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods LSeek DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Offset = offset

	return nil
}

// ReadChunk reads the entire data object in chunks (size of chunk specified by size parameter), passing the data into a callback function for each chunk. Use this to read/write large files.
func (obj *DataObj) ReadChunk(size int64, callback func([]byte)) error {
	if er := obj.init(); er != nil {
		return er
	}

	var (
		buffer   C.bytesBuf_t
		err      *C.char
		bytesRead C.int
	)

	if er := obj.LSeek(0); er != nil {
		return er
	}

	for obj.Offset < obj.Size {


		ccon := obj.Con.GetCcon()

		if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(size), &buffer, &bytesRead, ccon, &err); status != 0 {
			obj.Con.ReturnCcon(ccon)
			return newError(Fatal, fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
		}

		obj.Con.ReturnCcon(ccon)

		buf := unsafe.Pointer(buffer.buf)

		chunk := C.GoBytes(buf, bytesRead)
		
		C.free(buf)

		callback(chunk)

		if er := obj.LSeek(obj.Offset + size); er != nil {
			return er
		}
	}

	if er := obj.LSeek(0); er != nil {
		return er
	}

	return obj.Close()
}

// DownloadTo downloads and writes the entire data object to the provided path. Don't use this with large files unless you have RAM to spare, use ReadChunk() instead. Returns error.
func (obj *DataObj) DownloadTo(localPath string) error {
	if er := obj.init(); er != nil {
		return er
	}

	if objContents, err := obj.Read(); err != nil {
		return err
	} else {
		if er := ioutil.WriteFile(localPath, objContents, 0644); er != nil {
			return newError(Fatal, fmt.Sprintf("iRods Download DataObject Failed: %v, %v", obj.Path, er))
		}
	}

	return nil
}

// Write writes the data to the data object, starting from the beginning. Returns error.
func (obj *DataObj) Write(data []byte) error {
	if er := obj.initRW(); er != nil {
		return er
	}

	if obj.OpenedAs != C.O_RDWR || obj.OpenedAs != C.O_WRONLY {
		obj.Close()
		obj.OpenRW()
	}

	if er := obj.LSeek(0); er != nil {
		return er
	}

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0]) // Do I need to free this? It might be done by go

	var err *C.char

	ccon := obj.Con.GetCcon()

	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), ccon, &err); status != 0 {
		obj.Con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Con.ReturnCcon(ccon)

	obj.Size = size

	return obj.Close()
}

// WriteBytes writes to the data object wherever the object's offset pointer is currently set to. It advances the pointer to the end of the written data for supporting subsequent writes. Be sure to call obj.LSeek(0) before hand if you wish to write from the beginning. Returns error.
func (obj *DataObj) WriteBytes(data []byte) error {
	if er := obj.initRW(); er != nil {
		return er
	}

	if obj.OpenedAs != C.O_RDWR || obj.OpenedAs != C.O_WRONLY {
		obj.Close()
		obj.OpenRW()
	}

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0]) // Do I need to free this? It might be done by go

	var err *C.char

	ccon := obj.Con.GetCcon()

	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), ccon, &err); status != 0 {
		obj.Con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Con.ReturnCcon(ccon)

	obj.Size = size + obj.Offset

	return obj.LSeek(obj.Size)
}

// Stat returns a map (key/value pairs) of the system meta information. The following keys can be used with the map:
//
// "objSize"
//
// "dataMode"
//
// "dataId"
//
// "chksum"
//
// "ownerName"
//
// "ownerZone"
//
// "createTime"
//
// "modifyTime"
func (obj *DataObj) Stat() (map[string]interface{}, error) {

	var (
		err        *C.char
		statResult *C.rodsObjStat_t
	)

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_stat_dataobject(path, &statResult, ccon, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Close Stat Failed: %v, %v", obj.Path, C.GoString(err)))
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

	return result, nil
}

// Attribute gets slice of Meta AVU triples, matching by Attribute name for DataObj
func (obj *DataObj) Attribute(attrName string) (Metas, error) {
	if meta, err := obj.Meta(); err == nil {
		return meta.Get(attrName)
	} else {
		return nil, err
	}
}

// AddMeta adds a single Meta triple struct
func (obj *DataObj) AddMeta(m Meta) (nm *Meta, err error) {
	var mc *MetaCollection

	if mc, err = obj.Meta(); err != nil {
		return
	}

	nm, err = mc.Add(m)

	return
}

// DeleteMeta deletes a single Meta triple struct, identified by Attribute field
func (obj *DataObj) DeleteMeta(attr string) (*MetaCollection, error) {
	if mc, err := obj.Meta(); err == nil {
		return mc, mc.Delete(attr)
	} else {
		return nil, err
	}
}

// Meta returns collection of Meta AVU triple structs of the data object
func (obj *DataObj) Meta() (*MetaCollection, error) {
	if er := obj.init(); er != nil {
		return nil, er
	}

	if obj.MetaCol == nil {
		if mc, err := newMetaCollection(obj); err == nil {
			obj.MetaCol = mc
		} else {
			return nil, err
		}
	}

	return obj.MetaCol, nil
}

// CopyTo copies the data object to the specified collection. Supports Collection struct or string as input. Also refreshes the destination collection automatically to maintain correct state. Returns error.
func (obj *DataObj) CopyTo(iRodsCollection interface{}) error {

	var (
		err                         *C.char
		destination                 string
		destinationCollectionString string
	)

	if isString(iRodsCollection) {
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

	ccon := obj.Con.GetCcon()

	if status := C.gorods_copy_dataobject(path, dest, ccon, &err); status != 0 {
		obj.Con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Copy DataObject Failed: %v, %v", destination, C.GoString(err)))
	}

	obj.Con.ReturnCcon(ccon)

	// reload destination collection
	if isString(iRodsCollection) {
		// Find collection recursivly
		if dc := obj.Con.OpenedObjs.FindRecursive(destinationCollectionString); dc != nil {
			(dc.(*Collection)).Refresh()
		}
	} else {
		(iRodsCollection.(*Collection)).Refresh()
	}

	return nil
}

// MoveTo moves the data object to the specified collection. Supports Collection struct or string as input. Also refreshes the source and destination collections automatically to maintain correct state. Returns error.
func (obj *DataObj) MoveTo(iRodsCollection interface{}) error {

	var (
		err                         *C.char
		destination                 string
		destinationCollectionString string
		destinationCollection       *Collection
	)

	if isString(iRodsCollection) {
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

	ccon := obj.Con.GetCcon()

	if status := C.gorods_move_dataobject(path, dest, ccon, &err); status != 0 {
		obj.Con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Move DataObject Failed S:%v, D:%v, %v", obj.Path, destination, C.GoString(err)))
	}

	obj.Con.ReturnCcon(ccon)

	// Reload source collection, we are now detached
	obj.Col.Refresh()

	// Find & reload destination collection
	if isString(iRodsCollection) {
		// Find collection recursivly
		if dc := obj.Con.OpenedObjs.FindRecursive(destinationCollectionString); dc != nil {
			destinationCollection = dc.(*Collection)

			destinationCollection.Refresh()
		} else {
			// Can't find, load collection into memory
			destinationCollection, _ = obj.Con.Collection(destinationCollectionString, false)
		}
	} else {
		destinationCollection = (iRodsCollection.(*Collection))
		destinationCollection.Refresh()
	}

	// Reassign obj.Col to destination collection
	obj.Col = destinationCollection
	obj.Path = destinationCollection.Path + "/" + obj.Name

	obj.chandle = C.int(-1)

	return nil
}

// Rename is equivalent to the Linux mv command except that the data object must stay within the current collection (directory), returns error.
func (obj *DataObj) Rename(newFileName string) error {

	if strings.Contains(newFileName, "/") {
		return newError(Fatal, fmt.Sprintf("Can't Rename DataObject, path detected in: %v", newFileName))
	}

	var err *C.char

	source := obj.Path
	destination := obj.Col.Path + "/" + newFileName

	s := C.CString(source)
	d := C.CString(destination)

	defer C.free(unsafe.Pointer(s))
	defer C.free(unsafe.Pointer(d))

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_move_dataobject(s, d, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rename DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Name = newFileName
	obj.Path = destination

	obj.chandle = C.int(-1)

	return nil
}

// // Delete deletes the data object from the iRods server with a force flag
// func (obj *DataObj) Delete() error {

// 	var err *C.char

// 	path := C.CString(obj.Path)

// 	defer C.free(unsafe.Pointer(path))

// 	if status := C.gorods_unlink_dataobject(path, C.int(1), obj.Con.ccon, &err); status != 0 {
// 		return newError(Fatal, fmt.Sprintf("iRods Delete DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
// 	}

// 	return nil
// }

// Unlink deletes the data object from the iRods server, no force flag is used
func (obj *DataObj) Unlink() error {
	return obj.Rm(true, false)
}

// Chksum returns md5 hash string of data object
func (obj *DataObj) Chksum() (string, error) {

	var (
		err       *C.char
		chksumOut *C.char
	)

	path := C.CString(obj.Path)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(chksumOut))

	ccon := obj.Con.GetCcon()
	defer obj.Con.ReturnCcon(ccon)

	if status := C.gorods_checksum_dataobject(path, &chksumOut, ccon, &err); status != 0 {
		return "", newError(Fatal, fmt.Sprintf("iRods Chksum DataObject Failed: %v, %v", obj.Path, C.GoString(err)))
	}

	obj.Checksum = C.GoString(chksumOut)

	return obj.Checksum, nil
}

// Verify returns true or false depending on whether the checksum md5 string matches
func (obj *DataObj) Verify(md5Checksum string) bool {
	if chksum, err := obj.Chksum(); err == nil {
		chksumSplit := strings.Split(chksum, ":")
		return (md5Checksum == chksumSplit[1])
	}

	return false
}

// NEED TO IMPLEMENT
func (obj *DataObj) MoveToResource(destinationResource string) *DataObj {

	return obj
}

// NEED TO IMPLEMENT
func (obj *DataObj) Replicate(targetResource string) *DataObj {

	return obj
}

// NEED TO IMPLEMENT
func (obj *DataObj) ReplSettings(resource map[string]interface{}) *DataObj {
	//https://wiki.irods.org/doxygen_api/html/rc_data_obj_trim_8c_a7e4713d4b7617690e484fbada8560663.html

	return obj
}

