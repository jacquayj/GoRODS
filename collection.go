/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unsafe"
)

// Collection structs contain information about single collections in an iRods zone.
type Collection struct {
	Path        string
	Name        string
	DataObjects IRodsObjs
	MetaCol     *MetaCollection
	Con         *Connection
	Col         *Collection
	Recursive   bool
	Init 		bool
	Type        int

	chandle C.int
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
	str := fmt.Sprintf("Collection: %v\n", obj.Path)

	objs, _ := obj.All()

	for _, o := range objs {
		str += fmt.Sprintf("\t%v: %v\n", getTypeString(o.GetType()), o.GetName())
	}

	return str
}

// initCollection initializes collection from *C.collEnt_t. This is used internally in the gorods package.
func initCollection(data *C.collEnt_t, acol *Collection) (*Collection, error) {

	col := new(Collection)

	col.chandle = C.int(-1)
	col.Type = CollectionType
	col.Col = acol
	col.Con = col.Col.Con
	col.Path = C.GoString(data.collName)

	pathSlice := strings.Split(col.Path, "/")

	col.Name = pathSlice[len(pathSlice)-1]

	if acol.Recursive {
		col.Recursive = true

		if er := col.init(); er != nil {
			return nil, er
		}
	}

	return col, nil
}

// getCollection initializes specified collection located at startPath using gorods.Connection.
// Could be considered alias of Connection.Collection()
func getCollection(startPath string, recursive bool, con *Connection) (*Collection, error) {
	col := new(Collection)

	col.chandle = C.int(-1)
	col.Type = CollectionType
	col.Con = con
	col.Path = startPath
	col.Recursive = recursive

	if col.Path[len(col.Path)-1] == '/' {
		col.Path = col.Path[:len(col.Path)-1]
	}

	pathSlice := strings.Split(col.Path, "/")
	col.Name = pathSlice[len(pathSlice)-1]

	if col.Recursive {
		if er := col.init(); er != nil {
			return nil, er
		}
	} else {
		if er := col.Open(); er != nil {
			return nil, er
		}
	}

	return col, nil
}

// CreateCollection creates a collection in the specified collection using provided options. Returns the newly created collection object.
func CreateCollection(name string, coll *Collection) (*Collection, error) {

	var (
		errMsg *C.char
	)

	path := C.CString(coll.Path + "/" + name)

	defer C.free(unsafe.Pointer(path))

	if status := C.gorods_create_collection(path, coll.Con.ccon, &errMsg); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Create Collection Failed: %v, Does the collection already exist?", C.GoString(errMsg)))
	}

	coll.Refresh()

	newCol := coll.Cd(name)

	return newCol, nil

}

// init opens and reads collection information from iRods if it hasn't been init'd already
func (col *Collection) init() error {

	if !col.Init {
		if err := col.Open(); err != nil {
			return err
		}

		if err := col.ReadCollection(); err != nil {
			return err
		}
	}

	col.Init = true

	return nil
}

// GetCollections returns only the IRodsObjs that represent collections
func (col *Collection) GetCollections() (response IRodsObjs, err error) {
	if err = col.init(); err != nil {
		return
	}

	for i, obj := range col.DataObjects {
		if obj.GetType() == CollectionType {
			response = append(response, col.DataObjects[i])
		}
	}

	return
}

// GetDataObjs returns only the data objects contained within the collection
func (col *Collection) GetDataObjs() (response IRodsObjs, err error) {
	if err = col.init(); err != nil {
		return
	}

	for i, obj := range col.DataObjects {
		if obj.GetType() == DataObjType {
			response = append(response, col.DataObjects[i])
		}
	}

	return
}

// Returns generic interface slice containing both data objects and collections combined
func (col *Collection) All() (IRodsObjs, error) {
	if err := col.init(); err != nil {
		return col.DataObjects, err
	}

	return col.DataObjects, nil
}

// Type gets the type
func (col *Collection) GetType() int {
	return col.Type
}

// IsRecursive returns true or false
func (col *Collection) IsRecursive() bool {
	return col.Recursive
}

// Connection returns the *Connection used to get collection
func (col *Collection) GetCon() *Connection {
	return col.Con
}

// GetName returns the Name of the collection
func (col *Collection) GetName() string {
	return col.Name
}

// GetPath returns the Path of the collection
func (col *Collection) GetPath() string {
	return col.Path
}

// GetCol returns the *Collection of the collection
func (col *Collection) GetCol() *Collection {
	return col.Col
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

	path := C.CString(col.Path)

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

	if status := C.gorods_rm(path, 1, cRecursive, cForce, col.Con.ccon, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Rm Collection Failed: %v", C.GoString(errMsg)))
	}

	return nil
}

// Attribute gets specific metadata AVU triple for Collection
func (col *Collection) Attribute(attr string) (*Meta, error) {
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

	if col.MetaCol == nil {
		if mc, err := newMetaCollection(col); err == nil {
			col.MetaCol = mc
		} else {
			return nil, err
		}
	}

	return col.MetaCol, nil
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

// Open connects to iRods and sets the handle for Collection.
// Usually called by Collection.init()
func (col *Collection) Open() error {
	if int(col.chandle) < 0 {
		var errMsg *C.char

		path := C.CString(col.Path)

		defer C.free(unsafe.Pointer(path))

		if status := C.gorods_open_collection(path, &col.chandle, col.Con.ccon, &errMsg); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Open Collection Failed: %v, %v", col.Path, C.GoString(errMsg)))
		}
	}

	return nil
}

// Close closes the Collection connection and resets the handle
func (col *Collection) Close() error {
	var errMsg *C.char

	for _, c := range col.DataObjects {
		if err := c.Close(); err != nil {
			return err
		}
	}

	if int(col.chandle) > -1 {
		if status := C.gorods_close_collection(col.chandle, col.Con.ccon, &errMsg); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Close Collection Failed: %v, %v", col.Path, C.GoString(errMsg)))
		}

		col.chandle = C.int(-1)
	}

	return nil
}

// Refresh is an alias of ReadCollection()
func (col *Collection) Refresh() error {
	return col.ReadCollection()
}

// ReadCollection reads data (overwrites) into col.DataObjects field.
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

	// Read data objs from collection
	C.gorods_read_collection(col.Con.ccon, col.chandle, &arr, &arrSize, &err)

	// Get result length
	arrLen := int(arrSize)

	unsafeArr := unsafe.Pointer(arr)
	defer C.free(unsafeArr)

	// Convert C array to slice, backed by arr *C.collEnt_t
	slice := (*[1 << 30]C.collEnt_t)(unsafeArr)[:arrLen:arrLen]

	col.DataObjects = make([]IRodsObj, 0)

	for i, _ := range slice {
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

// Put adds a local file to the remote iRods collection
func (col *Collection) Put(localFile string) (*DataObj, error) {
	if err := col.init(); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(localFile)
	if err != nil {
		return nil, newError(Fatal, fmt.Sprintf("Can't read file for Put(): %v", localFile))
	}

	fileName := filepath.Base(localFile)

	if newFile, er := col.CreateDataObj(DataObjOptions{
		Name:  fileName,
		Size:  int64(len(data)),
		Mode:  0750,
		Force: true,
	}); er == nil {
		newFile.Write(data)
		return newFile, nil
	} else {
		return nil, er
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
	col.DataObjects = append(col.DataObjects, dataObj)

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
	if cols, err := col.GetCollections(); err == nil {
		if c := cols.Find(path); c != nil {
			return c.(*Collection)
		}
	}

	return nil
}

// Get is a shortcut for calling collection.GetDataObjs().Find(path). It effectively returns the DataObj you specify collection-relatively or absolutely.
func (col *Collection) Get(path string) *DataObj {
	if cols, err := col.GetDataObjs(); err == nil {
		if d := cols.Find(path); d != nil {
			return d.(*DataObj)
		}
	}

	return nil
}
