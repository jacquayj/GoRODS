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
	DataObjects []IRodsObj
	MetaCol     *MetaCollection
	Con         *Connection
	Col         *Collection
	Recursive   bool
	Type        int

	chandle C.int
}

// Collections is a slice of Collection structs
type Collections []*Collection

// Exists checks to see if a collection exists in the slice
// and returns true or false
func (colls Collections) Exists(path string) bool {
	if c := colls.Find(path); c != nil {
		return true
	}

	return false
}

// Find gets a collection from the slice and returns nil if one is not found.
// Both the collection name or full path can be used as input.
func (colls Collections) Find(path string) *Collection {
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for i, col := range colls {
		if col.Path == path || col.Name == path {
			return colls[i]
		}
	}

	return nil
}

// FindRecursive acts just like Find, but also searches sub collections recursively.
// If the collection was not explicitly loaded recursively, only the first level of sub collections will be searched.
func (colls Collections) FindRecursive(path string) *Collection {
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for i, col := range colls {
		if col.Path == path || col.Name == path {
			return colls[i]
		}

		if col.Recursive {
			// Use Collections() since we already loaded everything
			if subCol := col.Collections().FindRecursive(path); subCol != nil {
				return subCol
			}
		} else {
			// Use DataObjects so we don't load new collections
			var filtered Collections

			for n, obj := range col.DataObjects {
				if obj.GetType() == CollectionType {
					filtered = append(filtered, col.DataObjects[n].(*Collection))
				}
			}

			if subCol := filtered.FindRecursive(path); subCol != nil {
				return subCol
			}
		}
	}

	return nil
}

// String shows the contents of the collection.
//
// D = DataObj
//
// C = Collection
//
// Sample output:
//
// 	Collection: /tempZone/home/admin/gorods
// 		D: build.sh
// 		C: bin
// 		C: pkg
// 		C: src
func (obj *Collection) String() string {
	str := fmt.Sprintf("Collection: %v\n", obj.Path)

	for _, o := range obj.DataObjs() {
		str += fmt.Sprintf("\tD: %v\n", o.Name)
	}

	for _, o := range obj.Collections() {
		str += fmt.Sprintf("\tC: %v\n", o.Name)
	}

	return str
}

// initCollection initializes collection from *C.collEnt_t. This is used internally in the gorods package.
func initCollection(data *C.collEnt_t, acol *Collection) *Collection {

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
		col.init()
	}

	return col
}

// getCollection initializes specified collection located at startPath using gorods.Connection.
// Could be considered alias of Connection.Collection()
func getCollection(startPath string, recursive bool, con *Connection) *Collection {
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
		col.init()
	}

	return col
}

// init opens and reads collection information from iRods if the handle isn't set
func (col *Collection) init() *Collection {
	// If connection hasn't been opened, do it!
	if int(col.chandle) < 0 {
		col.Open()
		col.ReadCollection()
	}

	return col
}

// Type gets the type
func (col *Collection) GetType() int {
	return col.Type
}

// Connection returns the *Connection used to get collection
func (col *Collection) Connection() *Connection {
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

// Attribute gets specific metadata AVU triple for Collection
func (col *Collection) Attribute(attr string) *Meta {
	col.init()

	return col.Meta().Get(attr)
}

// Meta returns collection of all metadata AVU triples for Collection
func (col *Collection) Meta() *MetaCollection {
	col.init()

	if col.MetaCol == nil {
		col.MetaCol = newMetaCollection(col)
	}

	return col.MetaCol
}

// AddMeta adds a single Meta triple struct
func (col *Collection) AddMeta(m Meta) (newMeta *Meta, err error) {
	newMeta, err = col.Meta().Add(m)

	return
}

// DeleteMeta deletes a single Meta triple struct, identified by Attribute field
func (col *Collection) DeleteMeta(attr string) (*MetaCollection, error) {
	return col.Meta().Delete(attr)
}

// Open connects to iRods and sets the handle for Collection.
// Usually called by Collection.init()
func (col *Collection) Open() *Collection {
	var errMsg *C.char

	path := C.CString(col.Path)

	defer C.free(unsafe.Pointer(path))

	if status := C.gorods_open_collection(path, &col.chandle, col.Con.ccon, &errMsg); status != 0 {
		panic(newError(Fatal, fmt.Sprintf("iRods Open Collection Failed: %v, %v", col.Path, C.GoString(errMsg))))
	}

	return col
}

// Close closes the Collection connection and resets the handle
func (col *Collection) Close() *Collection {
	var errMsg *C.char

	for _, c := range col.DataObjects {
		if c.GetType() == CollectionType {
			(c.(*Collection)).Close()
		} else if c.GetType() == DataObjType {
			(c.(*DataObj)).Close()
		}
	}

	if int(col.chandle) > -1 {
		if status := C.gorods_close_collection(col.chandle, col.Con.ccon, &errMsg); status != 0 {
			panic(newError(Fatal, fmt.Sprintf("iRods Close Collection Failed: %v, %v", col.Path, C.GoString(errMsg))))
		}

		col.chandle = C.int(-1)
	}

	return col
}

// Refresh is an alias of ReadCollection()
func (col *Collection) Refresh() {
	col.ReadCollection()
}

// ReadCollection reads data (overwrites) into col.DataObjects field.
func (col *Collection) ReadCollection() {

	if int(col.chandle) < 0 {
		col.Open()
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
			col.add(initCollection(obj, col))
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

	col.Close()
}

// DataObjs returns only the data objects contained within the collection
func (col *Collection) DataObjs() DataObjs {
	col.init()

	var response DataObjs

	for i, obj := range col.DataObjects {
		if obj.GetType() == DataObjType {
			response = append(response, col.DataObjects[i].(*DataObj))
		}
	}

	return response
}

// Collections returns only the collections contained within the collection
func (col *Collection) Collections() Collections {
	col.init()

	var response Collections

	for i, obj := range col.DataObjects {
		if obj.GetType() == CollectionType {
			response = append(response, col.DataObjects[i].(*Collection))
		}
	}

	return response
}

// Put adds a local file to the remote iRods collection
func (col *Collection) Put(localFile string) *DataObj {
	col.init()

	data, err := ioutil.ReadFile(localFile)
	if err != nil {
		panic(newError(Fatal, fmt.Sprintf("Can't read file for Put(): %v", localFile)))
	}

	fileName := filepath.Base(localFile)

	newFile := CreateDataObj(DataObjOptions{
		Name:  fileName,
		Size:  int64(len(data)),
		Mode:  0750,
		Force: true,
	}, col)

	newFile.Write(data)

	return newFile
}

// CreateDataObj creates a data object within the collection using the options specified
func (col *Collection) CreateDataObj(opts DataObjOptions) *DataObj {
	return CreateDataObj(opts, col)
}

func (col *Collection) add(dataObj IRodsObj) *Collection {
	col.init()

	col.DataObjects = append(col.DataObjects, dataObj)

	return col
}

// Returns generic interface slice containing both data objects and collections combined
func (col *Collection) All() []IRodsObj {
	col.init()

	return col.DataObjects
}

// Both returns two slices, the first for DataObjs and the second for Collections
func (col *Collection) Both() (DataObjs, Collections) {
	return col.DataObjs(), col.Collections()
}

// Exists returns true of false depending on whether the DataObj or Collection is found
func (col *Collection) Exists(path string) bool {
	return col.DataObjs().Exists(path) || col.Collections().Exists(path)
}

// Find returns either a DataObject or Collection using the collection-relative or absolute path specified.
func (col *Collection) Find(path string) IRodsObj {
	if d := col.DataObjs().Find(path); d != nil {
		return d
	}

	if c := col.Collections().Find(path); c != nil {
		return c
	}

	return nil
}

// Cd is a shortcut for calling collection.Collections().Find(path). It effectively returns (or changes to) the sub collection you specify collection-relatively or absolutely.
func (col *Collection) Cd(path string) *Collection {
	return col.Collections().Find(path)
}

// Get is a shortcut for calling collection.DataObjs().Find(path). It effectively returns the DataObj you specify collection-relatively or absolutely.
func (col *Collection) Get(path string) *DataObj {
	return col.DataObjs().Find(path)
}
