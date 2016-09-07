/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// DataObj structs contain information about single data objects in an iRods zone.
type DataObj struct {
	path     string
	name     string
	checksum string
	size     int64
	offset   int64
	typ      int

	replNum    int
	rescHier   string
	replStatus int

	dataId   string
	resource *Resource
	phyPath  string

	openedAs C.int

	ownerName string
	owner     *User

	createTime time.Time
	modifyTime time.Time

	metaCol *MetaCollection

	// Con field is a pointer to the Connection used to fetch the data object
	con *Connection

	// Col field is a pointer to the Collection containing the data object
	col *Collection

	chandle C.int
}

// DataObjOptions is used for passing options to the CreateDataObj function
type DataObjOptions struct {
	Name     string
	Size     int64
	Mode     int
	Force    bool
	Resource interface{}
}

// String returns path of data object
func (obj *DataObj) String() string {
	return "DataObject: " + obj.path
}

// We don't init() here or return errors here because it takes forever. Lazy loading is better in this case.
func initDataObj(data *C.collEnt_t, col *Collection) *DataObj {

	dataObj := new(DataObj)

	dataObj.typ = DataObjType
	dataObj.col = col
	dataObj.con = dataObj.col.con
	dataObj.offset = 0
	dataObj.name = C.GoString(data.dataName)
	dataObj.path = C.GoString(data.collName) + "/" + dataObj.name
	dataObj.size = int64(data.dataSize)
	dataObj.chandle = C.int(-1)
	dataObj.checksum = C.GoString(data.chksum)
	dataObj.dataId = C.GoString(data.dataId)
	dataObj.phyPath = C.GoString(data.phyPath)
	dataObj.openedAs = C.int(-1)

	dataObj.replNum = int(data.replNum)
	dataObj.rescHier = C.GoString(data.resc_hier)
	dataObj.replStatus = int(data.replStatus)

	dataObj.ownerName = C.GoString(data.ownerName)
	dataObj.createTime = cTimeToTime(data.createTime)
	dataObj.modifyTime = cTimeToTime(data.modifyTime)

	if rsrcs, err := col.con.GetResources(); err != nil {
		return nil
	} else {
		if r := rsrcs.FindByName(C.GoString(data.resource)); r != nil {
			dataObj.resource = r
		}
	}

	if usrs, err := col.con.GetUsers(); err != nil {
		return nil
	} else {
		if u := usrs.FindByName(dataObj.ownerName, dataObj.con); u != nil {
			dataObj.owner = u
		}
	}

	return dataObj
}

// getDataObj initializes specified data object located at startPath using gorods.connection.
// Could be considered alias of Connection.DataObject()
func getDataObj(startPath string, con *Connection) (*DataObj, error) {

	collectionDir := filepath.Dir(startPath)
	dataObjName := filepath.Base(startPath)

	opts := CollectionOptions{
		Path:      collectionDir,
		Recursive: false,
	}

	if col, err := con.Collection(opts); err == nil {
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
		errMsg   *C.char
		handle   C.int
		force    int
		resource *C.char
	)

	if opts.Force {
		force = 1
	} else {
		force = 0
	}

	switch opts.Resource.(type) {
	case string:
		resource = C.CString(opts.Resource.(string))
	case *Resource:
		r := opts.Resource.(*Resource)
		resource = C.CString(r.Name())
	default:
		return nil, newError(Fatal, fmt.Sprintf("Wrong variable type passed in Resource field"))
	}

	path := C.CString(coll.path + "/" + opts.Name)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(resource))

	ccon := coll.con.GetCcon()

	if status := C.gorods_create_dataobject(path, C.rodsLong_t(opts.Size), C.int(opts.Mode), C.int(force), resource, &handle, ccon, &errMsg); status != 0 {
		coll.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Create DataObject Failed: %v, Does the file already exist?", C.GoString(errMsg)))
	}
	coll.con.ReturnCcon(ccon)

	if err := coll.Refresh(); err != nil {
		return nil, err
	}

	if do, err := getDataObj(C.GoString(path), coll.con); err != nil {
		return nil, err
	} else {
		return do, nil
	}

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
func (obj *DataObj) ACL() (ACLs, error) {

	var (
		result   C.goRodsACLResult_t
		err      *C.char
		zoneHint *C.char
	)

	zone, zErr := obj.con.GetLocalZone()
	if zErr != nil {
		return nil, zErr
	} else {
		zoneHint = C.CString(zone.Name())
	}

	cDataId := C.CString(obj.dataId)
	defer C.free(unsafe.Pointer(cDataId))
	defer C.free(unsafe.Pointer(zoneHint))

	ccon := obj.con.GetCcon()

	if status := C.gorods_get_dataobject_acl(ccon, cDataId, &result, zoneHint, &err); status != 0 {
		obj.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Data Object ACL Failed: %v", C.GoString(err)))
	}

	obj.con.ReturnCcon(ccon)

	return aclSliceToResponse(&result, obj.con)

}

// Chmod changes the permissions/ACL of a data object
// accessLevel: Null | Read | Write | Own
func (obj *DataObj) Chmod(userOrGroup string, accessLevel int, recursive bool) error {
	return chmod(obj, userOrGroup, accessLevel, recursive)
}

// Type gets the type
func (obj *DataObj) Type() int {
	return obj.typ
}

// Connection returns the *Connection used to get data object
func (obj *DataObj) Con() *Connection {
	return obj.con
}

// GetName returns the Name of the data object
func (obj *DataObj) Name() string {
	return obj.name
}

// GetName returns the Path of the data object
func (obj *DataObj) Path() string {
	return obj.path
}

// GetName returns the *Collection of the data object
func (obj *DataObj) Col() *Collection {
	return obj.col
}

// GetOwnerName returns the owner name of the data object
func (obj *DataObj) OwnerName() string {
	return obj.ownerName
}

// GetOwnerName returns the owner name of the data object
func (obj *DataObj) Owner() *User {
	return obj.owner
}

func (obj *DataObj) Resource() *Resource {
	return obj.resource
}

func (obj *DataObj) DataId() string {
	return obj.dataId
}

func (obj *DataObj) PhyPath() string {
	return obj.phyPath
}

func (obj *DataObj) ReplNum() int {
	return obj.replNum
}

func (obj *DataObj) RescHier() string {
	return obj.rescHier
}

func (obj *DataObj) ReplStatus() int {
	return obj.replStatus
}

func (obj *DataObj) Checksum() string {
	return obj.checksum
}

func (obj *DataObj) Offset() int64 {
	return obj.offset
}

func (obj *DataObj) Size() int64 {
	return obj.size
}

// GetCreateTime returns the create time of the data object
func (obj *DataObj) CreateTime() time.Time {
	return obj.createTime
}

// GetModifyTime returns the modify time of the data object
func (obj *DataObj) ModifyTime() time.Time {
	return obj.modifyTime
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

	path := C.CString(obj.path)

	defer C.free(unsafe.Pointer(path))

	var (
		cForce     C.int
		cRecursive C.int
	)

	if force {
		cForce = C.int(1)
	}

	if recursive {
		cRecursive = C.int(1)
	}

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_rm(path, 0, cRecursive, cForce, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rm DataObject Failed: %v", C.GoString(errMsg)))
	}

	return nil
}

// Open opens a connection to iRods and sets the data object handle
func (obj *DataObj) Open() error {
	var errMsg *C.char

	path := C.CString(obj.path)
	resourceName := C.CString(obj.resource.Name())
	replNum := C.CString(strconv.Itoa(obj.replNum))
	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(resourceName))
	defer C.free(unsafe.Pointer(replNum))

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_open_dataobject(path, resourceName, replNum, C.O_RDONLY, &obj.chandle, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Open DataObject Failed: %v, %v", obj.path, C.GoString(errMsg)))
	}

	obj.openedAs = C.O_RDONLY

	return nil
}

// OpenRW opens a connection to iRods and sets the data object handle for read/write access
func (obj *DataObj) OpenRW() error {
	var errMsg *C.char

	path := C.CString(obj.path)
	resourceName := C.CString(obj.resource.Name())
	replNum := C.CString(strconv.Itoa(obj.replNum))
	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(resourceName))
	defer C.free(unsafe.Pointer(replNum))

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_open_dataobject(path, resourceName, replNum, C.O_RDWR, &obj.chandle, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods OpenRW DataObject Failed: %v, %v", obj.path, C.GoString(errMsg)))
	}

	obj.openedAs = C.O_RDWR

	return nil
}

// Close closes the data object, resets handler
func (obj *DataObj) Close() error {
	var errMsg *C.char

	if int(obj.chandle) > -1 {

		ccon := obj.con.GetCcon()
		defer obj.con.ReturnCcon(ccon)

		if status := C.gorods_close_dataobject(obj.chandle, ccon, &errMsg); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Close DataObject Failed: %v, %v", obj.path, C.GoString(errMsg)))
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
		buffer    C.bytesBuf_t
		err       *C.char
		bytesRead C.int
	)

	if er := obj.LSeek(0); er != nil {
		return nil, er
	}

	ccon := obj.con.GetCcon()

	if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(obj.size), &buffer, &bytesRead, ccon, &err); status != 0 {
		obj.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.path, C.GoString(err)))
	}

	obj.con.ReturnCcon(ccon)

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
		buffer    C.bytesBuf_t
		err       *C.char
		bytesRead C.int
	)

	if er := obj.LSeek(pos); er != nil {
		return nil, er
	}

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(length), &buffer, &bytesRead, ccon, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods ReadBytes DataObject Failed: %v, %v", obj.path, C.GoString(err)))
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

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_lseek_dataobject(obj.chandle, C.rodsLong_t(offset), ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods LSeek DataObject Failed: %v, %v", obj.path, C.GoString(err)))
	}

	obj.offset = offset

	return nil
}

// ReadChunk reads the entire data object in chunks (size of chunk specified by size parameter), passing the data into a callback function for each chunk. Use this to read/write large files.
func (obj *DataObj) ReadChunk(size int64, callback func([]byte)) error {
	if er := obj.init(); er != nil {
		return er
	}

	var (
		buffer    C.bytesBuf_t
		err       *C.char
		bytesRead C.int
	)

	if er := obj.LSeek(0); er != nil {
		return er
	}

	for obj.offset < obj.size {

		ccon := obj.con.GetCcon()

		if status := C.gorods_read_dataobject(obj.chandle, C.rodsLong_t(size), &buffer, &bytesRead, ccon, &err); status != 0 {
			obj.con.ReturnCcon(ccon)
			return newError(Fatal, fmt.Sprintf("iRods Read DataObject Failed: %v, %v", obj.path, C.GoString(err)))
		}

		obj.con.ReturnCcon(ccon)

		buf := unsafe.Pointer(buffer.buf)

		chunk := C.GoBytes(buf, bytesRead)

		C.free(buf)

		callback(chunk)

		if er := obj.LSeek(obj.offset + size); er != nil {
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
			return newError(Fatal, fmt.Sprintf("iRods Download DataObject Failed: %v, %v", obj.path, er))
		}
	}

	return nil
}

// Write writes the data to the data object, starting from the beginning. Returns error.
func (obj *DataObj) Write(data []byte) error {
	if er := obj.initRW(); er != nil {
		return er
	}

	if obj.openedAs != C.O_RDWR || obj.openedAs != C.O_WRONLY {
		obj.Close()
		obj.OpenRW()
	}

	if er := obj.LSeek(0); er != nil {
		return er
	}

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0]) // Do I need to free this? It might be done by go

	var err *C.char

	ccon := obj.con.GetCcon()

	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), ccon, &err); status != 0 {
		obj.con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.path, C.GoString(err)))
	}

	obj.con.ReturnCcon(ccon)

	obj.size = size

	return obj.Close()
}

// WriteBytes writes to the data object wherever the object's offset pointer is currently set to. It advances the pointer to the end of the written data for supporting subsequent writes. Be sure to call obj.LSeek(0) before hand if you wish to write from the beginning. Returns error.
func (obj *DataObj) WriteBytes(data []byte) error {
	if er := obj.initRW(); er != nil {
		return er
	}

	if obj.openedAs != C.O_RDWR || obj.openedAs != C.O_WRONLY {
		obj.Close()
		obj.OpenRW()
	}

	size := int64(len(data))

	dataPointer := unsafe.Pointer(&data[0]) // Do I need to free this? It might be done by go

	var err *C.char

	ccon := obj.con.GetCcon()

	if status := C.gorods_write_dataobject(obj.chandle, dataPointer, C.int(size), ccon, &err); status != 0 {
		obj.con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Write DataObject Failed: %v, %v", obj.path, C.GoString(err)))
	}

	obj.con.ReturnCcon(ccon)

	obj.size = size + obj.offset

	return obj.LSeek(obj.size)
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

	path := C.CString(obj.path)

	defer C.free(unsafe.Pointer(path))

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_stat_dataobject(path, &statResult, ccon, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Close Stat Failed: %v, %v", obj.path, C.GoString(err)))
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

	if obj.metaCol == nil {
		if mc, err := newMetaCollection(obj); err == nil {
			obj.metaCol = mc
		} else {
			return nil, err
		}
	}

	return obj.metaCol, nil
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
			destinationCollectionString = obj.col.path + "/" + destinationCollectionString
		}

		if destinationCollectionString[len(destinationCollectionString)-1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + obj.name

	} else {
		destinationCollectionString = (iRodsCollection.(*Collection)).path + "/"
		destination = destinationCollectionString + obj.name
	}

	path := C.CString(obj.path)
	dest := C.CString(destination)
	resource := C.CString("")

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(dest))
	defer C.free(unsafe.Pointer(resource))

	ccon := obj.con.GetCcon()

	if status := C.gorods_copy_dataobject(path, dest, C.int(0), resource, ccon, &err); status != 0 {
		obj.con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Copy DataObject Failed: %v, %v", destination, C.GoString(err)))
	}

	obj.con.ReturnCcon(ccon)

	// reload destination collection
	if isString(iRodsCollection) {
		// Find collection recursivly
		if dc := obj.con.OpenedObjs.FindRecursive(destinationCollectionString); dc != nil {
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
			destinationCollectionString = obj.col.path + "/" + destinationCollectionString
		}

		if destinationCollectionString[len(destinationCollectionString)-1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + obj.name

	} else {
		destinationCollectionString = (iRodsCollection.(*Collection)).path + "/"
		destination = destinationCollectionString + obj.name
	}

	path := C.CString(obj.path)
	dest := C.CString(destination)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(dest))

	ccon := obj.con.GetCcon()

	if status := C.gorods_move_dataobject(path, dest, ccon, &err); status != 0 {
		obj.con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Move DataObject Failed S:%v, D:%v, %v", obj.path, destination, C.GoString(err)))
	}

	obj.con.ReturnCcon(ccon)

	// Reload source collection, we are now detached
	obj.col.Refresh()

	// Find & reload destination collection
	if isString(iRodsCollection) {
		// Find collection recursivly
		if dc := obj.con.OpenedObjs.FindRecursive(destinationCollectionString); dc != nil {
			destinationCollection = dc.(*Collection)

			destinationCollection.Refresh()
		} else {
			opts := CollectionOptions{
				Path:      destinationCollectionString,
				Recursive: false,
			}
			// Can't find, load collection into memory
			destinationCollection, _ = obj.con.Collection(opts)
		}
	} else {
		destinationCollection = (iRodsCollection.(*Collection))
		destinationCollection.Refresh()
	}

	// Reassign obj.col to destination collection
	obj.col = destinationCollection
	obj.path = destinationCollection.path + "/" + obj.name

	obj.chandle = C.int(-1)

	return nil
}

// Rename is equivalent to the Linux mv command except that the data object must stay within the current collection (directory), returns error.
func (obj *DataObj) Rename(newFileName string) error {

	if strings.Contains(newFileName, "/") {
		return newError(Fatal, fmt.Sprintf("Can't Rename DataObject, path detected in: %v", newFileName))
	}

	var err *C.char

	source := obj.path
	destination := obj.col.path + "/" + newFileName

	s := C.CString(source)
	d := C.CString(destination)

	defer C.free(unsafe.Pointer(s))
	defer C.free(unsafe.Pointer(d))

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_move_dataobject(s, d, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rename DataObject Failed: %v, %v", obj.path, C.GoString(err)))
	}

	obj.name = newFileName
	obj.path = destination

	obj.chandle = C.int(-1)

	return nil
}

// // Delete deletes the data object from the iRods server with a force flag
// func (obj *DataObj) Delete() error {

// 	var err *C.char

// 	path := C.CString(obj.path)

// 	defer C.free(unsafe.Pointer(path))

// 	if status := C.gorods_unlink_dataobject(path, C.int(1), obj.con.ccon, &err); status != 0 {
// 		return newError(Fatal, fmt.Sprintf("iRods Delete DataObject Failed: %v, %v", obj.path, C.GoString(err)))
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

	path := C.CString(obj.path)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(chksumOut))

	ccon := obj.con.GetCcon()
	defer obj.con.ReturnCcon(ccon)

	if status := C.gorods_checksum_dataobject(path, &chksumOut, ccon, &err); status != 0 {
		return "", newError(Fatal, fmt.Sprintf("iRods Chksum DataObject Failed: %v, %v", obj.path, C.GoString(err)))
	}

	obj.checksum = C.GoString(chksumOut)

	return obj.checksum, nil
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
