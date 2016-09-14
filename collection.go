/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unsafe"
)

// Collection structs contain information about single collections in an iRods zone.
type Collection struct {
	options *CollectionOptions

	trimRepls   bool
	path        string
	name        string
	dataObjects IRodsObjs
	metaCol     *MetaCollection
	con         *Connection
	col         *Collection

	recursive bool
	hasInit   bool
	typ       int
	parent    *Collection

	ownerName  string
	owner      *User
	createTime time.Time
	modifyTime time.Time

	chandle C.int
}

// CollectionOptions stores options relating to collection initialization.
// Path is the full path of the collection you're requesting.
// Recursive if set to true will load sub collections into memory, until the end of the collection "tree" is found.
type CollectionOptions struct {
	Path      string
	Recursive bool
	GetRepls  bool
}

// String shows the contents of the collection.
//
// d = DataObj
//
// C = Collection
//
// Sample output:
//
// 	Collection: /tempZone/home/admin/gorods
// 		d: build.sh
// 		C: bin
// 		C: pkg
// 		C: src
func (obj *Collection) String() string {
	str := fmt.Sprintf("Collection: %v\n", obj.path)

	objs, _ := obj.All()

	for _, o := range objs {
		str += fmt.Sprintf("\t%v: %v\n", getTypeString(o.Type()), o.Name())
	}

	return str
}

// initCollection initializes collection from *C.collEnt_t. This is used internally in the gorods package.
func initCollection(data *C.collEnt_t, acol *Collection) (*Collection, error) {

	col := new(Collection)

	col.chandle = C.int(-1)
	col.typ = CollectionType
	col.col = acol
	col.con = col.col.con
	col.path = C.GoString(data.collName)
	col.options = acol.options
	col.recursive = acol.recursive
	col.trimRepls = acol.trimRepls
	col.parent = acol

	col.ownerName = C.GoString(data.ownerName)
	col.createTime = cTimeToTime(data.createTime)
	col.modifyTime = cTimeToTime(data.modifyTime)

	col.name = filepath.Base(col.path)

	if usrs, err := col.con.GetUsers(); err != nil {
		return nil, err
	} else {
		if u := usrs.FindByName(col.ownerName, col.con); u != nil {
			col.owner = u
		} else {
			return nil, newError(Fatal, fmt.Sprintf("iRods initCollection Failed: Unable to locate user in cache"))
		}
	}

	if col.recursive {

		if er := col.init(); er != nil {
			return nil, er
		}
	}

	return col, nil
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
func (col *Collection) Stat() (map[string]interface{}, error) {

	var (
		err        *C.char
		statResult *C.rodsObjStat_t
	)

	path := C.CString(col.path)

	defer C.free(unsafe.Pointer(path))

	ccon := col.con.GetCcon()
	defer col.con.ReturnCcon(ccon)

	if status := C.gorods_stat_dataobject(path, &statResult, ccon, &err); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Stat Failed: %v, %v", col.path, C.GoString(err)))
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
	//result["rescHier"] = C.GoString(&statResult.rescHier[0])

	C.freeRodsObjStat(statResult)

	return result, nil
}

// getCollection initializes specified collection located at startPath using gorods.connection.
// Could be considered alias of Connection.collection()
func getCollection(opts CollectionOptions, con *Connection) (*Collection, error) {

	col := new(Collection)

	col.options = &opts

	col.chandle = C.int(-1)
	col.typ = CollectionType
	col.con = con
	col.path = strings.TrimRight(opts.Path, "/")
	col.name = filepath.Base(col.path)
	col.recursive = opts.Recursive
	col.trimRepls = !opts.GetRepls

	if col.recursive {
		if er := col.init(); er != nil {
			return nil, er
		}
	} else {
		if er := col.Open(); er != nil {
			return nil, er
		}
	}

	if info, err := col.Stat(); err == nil {
		col.ownerName = info["ownerName"].(string)

		cCreateTime := C.CString(info["createTime"].(string))
		defer C.free(unsafe.Pointer(cCreateTime))
		cModifyTime := C.CString(info["modifyTime"].(string))
		defer C.free(unsafe.Pointer(cModifyTime))

		col.createTime = cTimeToTime(cCreateTime)
		col.modifyTime = cTimeToTime(cModifyTime)

		if usrs, err := col.con.GetUsers(); err != nil {
			return nil, err
		} else {
			if u := usrs.FindByName(col.ownerName, col.con); u != nil {
				col.owner = u
			} else {
				return nil, newError(Fatal, fmt.Sprintf("iRods getCollection Failed: Unable to locate user in cache"))
			}
		}
	} else {
		return nil, err
	}

	return col, nil

}

// CreateCollection creates a collection in the specified collection using provided options. Returns the newly created collection object.
func CreateCollection(name string, coll *Collection) (*Collection, error) {

	var (
		errMsg *C.char
	)

	path := C.CString(coll.path + "/" + name)

	defer C.free(unsafe.Pointer(path))

	ccon := coll.con.GetCcon()

	if status := C.gorods_create_collection(path, ccon, &errMsg); status != 0 {
		coll.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Create Collection Failed: %v, Does the collection already exist?", C.GoString(errMsg)))
	}

	coll.con.ReturnCcon(ccon)

	coll.Refresh()

	newCol := coll.Cd(name)

	return newCol, nil

}

// init opens and reads collection information from iRods if it hasn't been init'd already
func (col *Collection) init() error {

	if !col.hasInit {
		if err := col.Open(); err != nil {
			return err
		}

		if err := col.ReadCollection(); err != nil {
			return err
		}
	}

	col.hasInit = true

	return nil
}

// GetCollections returns only the IRodsObjs that represent collections
func (col *Collection) Collections() (response IRodsObjs, err error) {
	if err = col.init(); err != nil {
		return
	}

	for i, obj := range col.dataObjects {
		if obj.Type() == CollectionType {
			response = append(response, col.dataObjects[i])
		}
	}

	return
}

// GetDataObjs returns only the data objects contained within the collection
func (col *Collection) DataObjs() (response IRodsObjs, err error) {
	if err = col.init(); err != nil {
		return
	}

	for i, obj := range col.dataObjects {
		if obj.Type() == DataObjType {
			response = append(response, col.dataObjects[i])
		}
	}

	return
}

// Returns generic interface slice containing both data objects and collections combined
func (col *Collection) All() (IRodsObjs, error) {
	if err := col.init(); err != nil {
		return col.dataObjects, err
	}

	return col.dataObjects, nil
}

// SetInheritance sets the inheritance option of the collection. If true, sub-collections and data objects inherit the permissions (ACL) of this collection.
func (col *Collection) SetInheritance(inherits bool, recursive bool) error {
	var ih int
	if inherits {
		ih = Inherit
	} else {
		ih = NoInherit
	}
	return chmod(col, "", ih, recursive, false)
}

// Inheritance returns true or false, depending on the collection's inheritance setting
func (col *Collection) Inheritance() (bool, error) {

	var (
		enabled C.int
		err     *C.char
	)

	collName := C.CString(col.path)
	defer C.free(unsafe.Pointer(collName))

	ccon := col.con.GetCcon()
	defer col.con.ReturnCcon(ccon)

	if status := C.gorods_get_collection_inheritance(ccon, collName, &enabled, &err); status != 0 {
		return false, newError(Fatal, fmt.Sprintf("iRods Get Collection Inheritance Failed: %v", C.GoString(err)))
	}

	if int(enabled) > 0 {
		return true, nil
	}

	return false, nil
}

// GrantAccess will add permissions (ACL) to the collection
func (col *Collection) GrantAccess(userOrGroup AccessObject, accessLevel int, recursive bool) error {
	return chmod(col, userOrGroup.Name(), accessLevel, recursive, true)
}

// Chmod changes the permissions/ACL of the collection
// accessLevel: Null | Read | Write | Own
func (col *Collection) Chmod(userOrGroup string, accessLevel int, recursive bool) error {
	return chmod(col, userOrGroup, accessLevel, recursive, true)
}

// ACL returns a slice of ACL structs. Example of slice in string format:
// [rods#tempZone:own
// developers#tempZone:modify object
// designers#tempZone:read object]
func (col *Collection) ACL() (ACLs, error) {

	var (
		result   C.goRodsACLResult_t
		err      *C.char
		zoneHint *C.char
		collName *C.char
	)

	zone, zErr := col.con.GetLocalZone()
	if zErr != nil {
		return nil, zErr
	} else {
		zoneHint = C.CString(zone.Name())
	}

	collName = C.CString(col.path)
	defer C.free(unsafe.Pointer(collName))
	defer C.free(unsafe.Pointer(zoneHint))

	ccon := col.con.GetCcon()

	if status := C.gorods_get_collection_acl(ccon, collName, &result, zoneHint, &err); status != 0 {
		col.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Collection ACL Failed: %v", C.GoString(err)))
	}

	col.con.ReturnCcon(ccon)

	return aclSliceToResponse(&result, col.con)
}

// Type gets the type
func (col *Collection) Type() int {
	return col.typ
}

// IsRecursive returns true or false
func (col *Collection) IsRecursive() bool {
	return col.recursive
}

// Connection returns the *Connection used to get collection
func (col *Collection) Con() *Connection {
	return col.con
}

// GetName returns the Name of the collection
func (col *Collection) Name() string {
	return col.name
}

// GetPath returns the Path of the collection
func (col *Collection) Path() string {
	return col.path
}

// GetOwnerName returns the owner name of the collection
func (col *Collection) OwnerName() string {
	return col.ownerName
}

// Owner returns a User struct, representing the user who owns the collection
func (col *Collection) Owner() *User {
	return col.owner
}

// GetCreateTime returns the create time of the collection
func (col *Collection) CreateTime() time.Time {
	return col.createTime
}

// GetModifyTime returns the modify time of the collection
func (col *Collection) ModifyTime() time.Time {
	return col.modifyTime
}

// GetCol returns the *Collection of the collection
func (col *Collection) Col() *Collection {
	return col.col
}

// Destroy is equivalent to irm -rf
func (col *Collection) Destroy() error {
	return col.Rm(true, true)
}

// Delete is equivalent to irm -f {-r}
func (col *Collection) Delete(recursive bool) error {
	return col.Rm(recursive, true)
}

// Trash is equivalent to irm {-r}
func (col *Collection) Trash(recursive bool) error {
	return col.Rm(recursive, false)
}

// Rm is equivalent to irm {-r} {-f}
func (col *Collection) Rm(recursive bool, force bool) error {
	var errMsg *C.char

	path := C.CString(col.path)

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

	ccon := col.con.GetCcon()
	defer col.con.ReturnCcon(ccon)

	if status := C.gorods_rm(path, 1, cRecursive, cForce, ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rm Collection Failed: %v", C.GoString(errMsg)))
	}

	return nil
}

// Attribute gets slice of Meta AVU triples, matching by Attribute name for Collection
func (col *Collection) Attribute(attr string) (Metas, error) {
	if mc, err := col.Meta(); err == nil {
		return mc.Get(attr)
	} else {
		return nil, err
	}

}

// Meta returns collection of all metadata AVU triples for Collection
func (col *Collection) Meta() (*MetaCollection, error) {
	if er := col.init(); er != nil {
		return nil, er
	}

	if col.metaCol == nil {
		if mc, err := newMetaCollection(col); err == nil {
			col.metaCol = mc
		} else {
			return nil, err
		}
	}

	return col.metaCol, nil
}

// AddMeta adds a single Meta triple struct
func (col *Collection) AddMeta(m Meta) (newMeta *Meta, err error) {
	var mc *MetaCollection

	if mc, err = col.Meta(); err != nil {
		return
	}

	newMeta, err = mc.Add(m)

	return
}

// DeleteMeta deletes a single Meta triple struct, identified by Attribute field
func (col *Collection) DeleteMeta(attr string) (*MetaCollection, error) {
	if mc, err := col.Meta(); err == nil {
		return mc, mc.Delete(attr)
	} else {
		return nil, err
	}
}

// DownloadTo recursively downloads all data objects and collections contained within the collection, into the path specified
func (col *Collection) DownloadTo(localPath string) error {

	if dir, err := os.Stat(localPath); err == nil && dir.IsDir() {
		if localPath[len(localPath)-1] != '/' {
			localPath += "/"
		}
		if objs, er := col.DataObjs(); er == nil {
			for _, obj := range objs {
				if e := obj.DownloadTo(localPath + obj.Name()); e != nil {
					return e
				}
			}
		} else {
			return er
		}

		if cols, er := col.Collections(); er == nil {
			for _, col := range cols {

				newDir := localPath + col.Name()

				if e := os.Mkdir(newDir, 0777); e != nil {
					return e
				}

				if e := col.DownloadTo(newDir); e != nil {
					return e
				}
			}
		} else {
			return er
		}

	} else {
		return newError(Fatal, fmt.Sprintf("iRods DownloadTo Failed: localPath doesn't exist or isn't a directory"))
	}

	return nil
}

// Open connects to iRods and sets the handle for Collection.
// Usually called by Collection.init()
func (col *Collection) Open() error {
	if int(col.chandle) < 0 {
		var (
			errMsg     *C.char
			cTrimRepls C.int
		)

		path := C.CString(col.path)

		if col.trimRepls {
			cTrimRepls = C.int(1)
		} else {
			cTrimRepls = C.int(0)
		}

		defer C.free(unsafe.Pointer(path))

		ccon := col.con.GetCcon()
		defer col.con.ReturnCcon(ccon)

		if status := C.gorods_open_collection(path, cTrimRepls, &col.chandle, ccon, &errMsg); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Open Collection Failed: %v, %v", col.path, C.GoString(errMsg)))
		}
	}

	return nil
}

// Close closes the Collection connection and resets the handle
func (col *Collection) Close() error {
	var errMsg *C.char

	for _, c := range col.dataObjects {
		if err := c.Close(); err != nil {
			return err
		}
	}

	if int(col.chandle) > -1 {

		ccon := col.con.GetCcon()
		defer col.con.ReturnCcon(ccon)

		if status := C.gorods_close_collection(col.chandle, ccon, &errMsg); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Close Collection Failed: %v, %v", col.path, C.GoString(errMsg)))
		}

		col.chandle = C.int(-1)
	}

	return nil
}

// CopyTo copies all collections and data objects contained withing the collection to the specified collection.
// Accepts string or *Collection types.
func (col *Collection) CopyTo(iRodsCollection interface{}) error {

	// Get reference to destination collection (just like MoveTo)
	var (
		destination                 string
		destinationCollectionString string
		destinationCollection       *Collection
	)

	switch iRodsCollection.(type) {
	case string:
		destinationCollectionString = iRodsCollection.(string)

		// Is this a relative path?
		if destinationCollectionString[0] != '/' {
			destinationCollectionString = path.Dir(col.path) + "/" + destinationCollectionString
		}

		if destinationCollectionString[len(destinationCollectionString)-1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + col.name
	case *Collection:
		destinationCollectionString = (iRodsCollection.(*Collection)).path + "/"
		destination = destinationCollectionString + col.name
	default:
		return newError(Fatal, fmt.Sprintf("iRods CopyTo Failed, unknown variable type passed as collection"))
	}

	var colEr error

	// load destination collection into memory
	if destinationCollection, colEr = col.con.Collection(CollectionOptions{
		Path:      destinationCollectionString,
		Recursive: false,
	}); colEr != nil {
		return colEr
	}

	// Create collection with same name in destination as sub-collection
	if newCol, err := destinationCollection.CreateSubCollection(col.name); err == nil {

		// loop through data objects, copy each to new sub-collection
		if objs, er := col.DataObjs(); er == nil {
			for _, obj := range objs {
				if e := obj.CopyTo(newCol); e != nil {
					return e
				}
			}
		} else {
			return er
		}

		// Loop through collections -> run recursive copyTo util?
		if cols, er := col.Collections(); er == nil {
			for _, aCol := range cols {
				if er := aCol.CopyTo(newCol); er != nil {
					return er
				}
			}
		} else {
			return er
		}

		newCol.Refresh() // <- is this required?

	} else {
		return err
	}

	return nil
}

// TrimRepls recursively trims data object replicas (removes from resource servers), using the rules defined in opts.
func (col *Collection) TrimRepls(opts TrimOptions) error {
	// loop through data objects
	if objs, er := col.DataObjs(); er == nil {
		for _, obj := range objs {
			if e := obj.TrimRepls(opts); e != nil {
				return e
			}
		}
	} else {
		return er
	}

	// Loop through collections
	if cols, er := col.Collections(); er == nil {
		for _, aCol := range cols {
			if er := aCol.TrimRepls(opts); er != nil {
				return er
			}

			c := aCol.(*Collection)

			c.Refresh()

		}
	} else {
		return er
	}

	col.Refresh()

	return nil
}

// MoveToResource recursively moves all data objects contained within the collection to the specified resource.
// Accepts string or *Resource type.
func (col *Collection) MoveToResource(targetResource interface{}) error {

	// loop through data objects
	if objs, er := col.DataObjs(); er == nil {
		for _, obj := range objs {
			if e := obj.MoveToResource(targetResource); e != nil {
				return e
			}
		}
	} else {
		return er
	}

	// Loop through collections
	if cols, er := col.Collections(); er == nil {
		for _, aCol := range cols {
			if er := aCol.MoveToResource(targetResource); er != nil {
				return er
			}

			c := aCol.(*Collection)

			c.Refresh()

		}
	} else {
		return er
	}

	col.Refresh()

	return nil
}

// Replicate recursively copies all data objects contained within the collection to the specified resource.
// Accepts string or *Resource type for targetResource parameter.
func (col *Collection) Replicate(targetResource interface{}, opts DataObjOptions) error {

	// loop through data objects
	if objs, er := col.DataObjs(); er == nil {
		for _, obj := range objs {
			if e := obj.Replicate(targetResource, opts); e != nil {
				return e
			}
		}
	} else {
		return er
	}

	// Loop through collections
	if cols, er := col.Collections(); er == nil {
		for _, aCol := range cols {
			if er := aCol.Replicate(targetResource, opts); er != nil {
				return er
			}

			c := aCol.(*Collection)

			if !c.trimRepls {
				c.Refresh()
			}
		}
	} else {
		return er
	}

	if !col.trimRepls {
		col.Refresh()
	}

	return nil
}

// Backup is similar to Replicate. In backup mode, if a good copy already exists in this resource group or resource, don't make another one.
func (col *Collection) Backup(targetResource interface{}, opts DataObjOptions) error {

	// loop through data objects
	if objs, er := col.DataObjs(); er == nil {
		for _, obj := range objs {
			if e := obj.Backup(targetResource, opts); e != nil {
				return e
			}
		}
	} else {
		return er
	}

	// Loop through collections
	if cols, er := col.Collections(); er == nil {
		for _, aCol := range cols {
			if er := aCol.Backup(targetResource, opts); er != nil {
				return er
			}

			c := aCol.(*Collection)

			if !c.trimRepls {
				c.Refresh()
			}
		}
	} else {
		return er
	}

	if !col.trimRepls {
		col.Refresh()
	}

	return nil
}

// MoveTo moves the collection to the specified collection. Supports Collection struct or string as input. Also refreshes the source and destination collections automatically to maintain correct state. Returns error.
func (col *Collection) MoveTo(iRodsCollection interface{}) error {

	var (
		err                         *C.char
		destination                 string
		destinationCollectionString string
		destinationCollection       *Collection
	)

	switch iRodsCollection.(type) {
	case string:
		destinationCollectionString = iRodsCollection.(string)

		// Is this a relative path?
		if destinationCollectionString[0] != '/' {
			destinationCollectionString = path.Dir(col.path) + "/" + destinationCollectionString
		}

		if destinationCollectionString[len(destinationCollectionString)-1] != '/' {
			destinationCollectionString += "/"
		}

		destination += destinationCollectionString + col.name
	case *Collection:
		destinationCollectionString = (iRodsCollection.(*Collection)).path + "/"
		destination = destinationCollectionString + col.name
	default:
		return newError(Fatal, fmt.Sprintf("iRods Move Collection Failed, unknown variable type passed as collection"))
	}

	path := C.CString(col.path)
	dest := C.CString(destination)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(dest))

	ccon := col.con.GetCcon()

	if status := C.gorods_move_dataobject(path, dest, C.RENAME_COLL, ccon, &err); status != 0 {
		col.con.ReturnCcon(ccon)
		return newError(Fatal, fmt.Sprintf("iRods Move Collection Failed: %v, D:%v, %v", col.path, destination, C.GoString(err)))
	}

	col.con.ReturnCcon(ccon)

	// Reload source collection, we are now detached
	col.parent.Refresh()

	// Find & reload destination collection
	switch iRodsCollection.(type) {
	case string:
		var colEr error

		// Can't find, load collection into memory
		destinationCollection, colEr = col.con.Collection(CollectionOptions{
			Path:      destinationCollectionString,
			Recursive: false,
		})
		if colEr != nil {
			return colEr
		}
	case *Collection:
		destinationCollection = (iRodsCollection.(*Collection))
	default:
		return newError(Fatal, fmt.Sprintf("iRods Move Collection Failed, unknown variable type passed as collection"))
	}

	destinationCollection.Refresh()

	// Reassign obj.col to destination collection
	col.parent = destinationCollection
	col.path = destinationCollection.path + "/" + col.name

	col.chandle = C.int(-1)

	return nil
}

// Rename is equivalent to the Linux mv command except that the collection must stay within it's current collection (directory), returns error.
func (col *Collection) Rename(newFileName string) error {

	if strings.Contains(newFileName, "/") {
		return newError(Fatal, fmt.Sprintf("Can't Rename DataObject, path detected in: %v", newFileName))
	}

	var err *C.char

	source := col.path
	destination := path.Dir(col.path) + "/" + newFileName

	s := C.CString(source)
	d := C.CString(destination)

	defer C.free(unsafe.Pointer(s))
	defer C.free(unsafe.Pointer(d))

	ccon := col.con.GetCcon()
	defer col.con.ReturnCcon(ccon)

	if status := C.gorods_move_dataobject(s, d, C.RENAME_COLL, ccon, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rename Collection Failed: %v, %v", col.path, C.GoString(err)))
	}

	col.name = newFileName
	col.path = destination

	col.chandle = C.int(-1)

	return nil
}

// Refresh is an alias of ReadCollection()
func (col *Collection) Refresh() error {
	return col.ReadCollection()
}

// ReadCollection reads data (overwrites) into col.dataObjects field.
func (col *Collection) ReadCollection() error {

	if er := col.Open(); er != nil {
		return er
	}

	// Init C varaibles
	var (
		err     *C.char
		arr     *C.collEnt_t
		arrSize C.int
	)

	ccon := col.con.GetCcon()

	// Read data objs from collection
	C.gorods_read_collection(ccon, col.chandle, &arr, &arrSize, &err)

	col.con.ReturnCcon(ccon)

	// Get result length
	arrLen := int(arrSize)

	unsafeArr := unsafe.Pointer(arr)
	defer C.free(unsafeArr)

	// Convert C array to slice, backed by arr *C.collEnt_t
	slice := (*[1 << 30]C.collEnt_t)(unsafeArr)[:arrLen:arrLen]

	col.dataObjects = make([]IRodsObj, 0)

	for i := range slice {
		obj := &slice[i]

		isCollection := (obj.objType != C.DATA_OBJ_T)

		if isCollection {
			if newCol, er := initCollection(obj, col); er == nil {
				col.add(newCol)
			} else {
				return er
			}
		} else {
			col.add(initDataObj(obj, col))

			// Strings only in DataObj types
			C.free(unsafe.Pointer(obj.dataName))
			C.free(unsafe.Pointer(obj.dataId))
			C.free(unsafe.Pointer(obj.chksum))
			//C.free(unsafe.Pointer(obj.dataType))
			C.free(unsafe.Pointer(obj.resource))
			//C.free(unsafe.Pointer(obj.rescGrp))
			C.free(unsafe.Pointer(obj.phyPath))
		}

		// String in both object types
		C.free(unsafe.Pointer(obj.ownerName))
		C.free(unsafe.Pointer(obj.collName))
		C.free(unsafe.Pointer(obj.createTime))
		C.free(unsafe.Pointer(obj.modifyTime))

	}

	return col.Close()
}

// Put reads the entire file from localPath and adds it the collection, using the options specified.
func (col *Collection) Put(localPath string, opts DataObjOptions) (*DataObj, error) {

	var (
		errMsg   *C.char
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

	path := C.CString(col.path + "/" + opts.Name)
	cLocalPath := C.CString(localPath)

	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(resource))
	defer C.free(unsafe.Pointer(cLocalPath))

	ccon := col.con.GetCcon()

	if status := C.gorods_put_dataobject(cLocalPath, path, C.rodsLong_t(opts.Size), C.int(opts.Mode), C.int(force), resource, ccon, &errMsg); status != 0 {
		col.con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Put DataObject Failed: %v, Does the file already exist?", C.GoString(errMsg)))
	}
	col.con.ReturnCcon(ccon)

	if err := col.Refresh(); err != nil {
		return nil, err
	}

	if do, err := getDataObj(C.GoString(path), col.con); err != nil {
		return nil, err
	} else {
		return do, nil
	}

}

// CreateDataObj creates a data object within the collection using the options specified
func (col *Collection) CreateDataObj(opts DataObjOptions) (*DataObj, error) {
	return CreateDataObj(opts, col)
}

// CreateSubCollection creates a collection within the collection using the options specified
func (col *Collection) CreateSubCollection(name string) (*Collection, error) {
	return CreateCollection(name, col)
}

func (col *Collection) add(dataObj IRodsObj) *Collection {
	col.dataObjects = append(col.dataObjects, dataObj)

	return col
}

// Exists returns true of false depending on whether the DataObj or Collection is found
func (col *Collection) Exists(path string) bool {
	if objs, err := col.All(); err == nil {
		return objs.Exists(path)
	}

	return false
}

// Find returns either a DataObject or Collection using the collection-relative or absolute path specified.
func (col *Collection) Find(path string) IRodsObj {
	if objs, err := col.All(); err == nil {
		return objs.Find(path)
	}

	return nil
}

// Find returns either a DataObject or Collection using the collection-relative or absolute path specified.
func (col *Collection) FindRecursive(path string) IRodsObj {
	if objs, err := col.All(); err == nil {
		return objs.FindRecursive(path)
	}

	return nil
}

func (col *Collection) FindCol(path string) *Collection {
	if c := col.Find(path); c != nil {
		return c.(*Collection)
	}

	return nil
}

func (col *Collection) FindObj(path string) *DataObj {
	if c := col.Find(path); c != nil {
		return c.(*DataObj)
	}

	return nil
}

// Cd is a shortcut for calling collection.GetCollections().Find(path). It effectively returns (or changes to) the sub collection you specify collection-relatively or absolutely.
func (col *Collection) Cd(path string) *Collection {
	if cols, err := col.Collections(); err == nil {
		if c := cols.Find(path); c != nil {
			return c.(*Collection)
		}
	}

	return nil
}

// Get is a shortcut for calling collection.GetDataObjs().Find(path). It effectively returns the DataObj you specify collection-relatively or absolutely.
func (col *Collection) Get(path string) *DataObj {
	if cols, err := col.DataObjs(); err == nil {
		if d := cols.Find(path); d != nil {
			return d.(*DataObj)
		}
	}

	return nil
}
