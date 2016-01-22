package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
	"reflect"
)

// collection.DataObjs()     -> type: DataObjs
// collection.Collections()  -> type: Collections
// collection.All()          -> type: []interface{}
// collection.Both()         -> (type: DataObjs, type: Collections)
type Collection struct {
	Path string
	DataObjects []interface{}
	Con *Connection
	Col *Collection
	Recursive bool
	
	chandle C.int
	collent *C.collEnt_t
}

// collections.Find(relPath) -> type: Collection
type Collections []*Collection

func (colls Collections) Find(path string) *Collection {
	for i, col := range colls {
		if col.Path == path {
			return colls[i]
		}
	}

	return nil
}

func (obj *Collection) String() string {
	return "Collection: " + obj.Path
}

// Init from *C.collEnt_t
func NewCollection(data *C.collEnt_t, acol *Collection) *Collection {
	col := new(Collection)

	col.Col = acol
	col.Con = col.Col.Con
	col.Path = C.GoString(data.collName)
	
	if acol.Recursive {
		col.Recursive = true
		col.Init()
	}

	return col
}

// Called from connection
func GetCollection(startPath string, recursive bool, con *Connection) *Collection {
	col := new(Collection)

	col.Con = con
	col.Path = startPath
	col.Recursive = recursive

	if col.Recursive {
		col.Init()
	}

	return col
}

func (col *Collection) Init() {
	// If generic data object slice hasn't been built, build it!
	if len(col.DataObjects) == 0 {
		col.Open()
		col.ReadCollection()
	}
}

// Opens connection to collection, passes in flag options
func (col *Collection) Open() {
	var errMsg *C.char

	if status := C.gorods_open_collection(C.CString(col.Path), &col.chandle, col.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Open Collection Failed: %v, %v", col.Path, C.GoString(errMsg)))
	}
}

// Reads data into col.DataObjects
func (col *Collection) ReadCollection() {

	fmt.Printf("Debug: Reading %v \n", col.Path)

	// Init C varaibles
	var err *C.char
	var arr *C.collEnt_t
	var arrSize C.int

	// Read data objs from collection
	C.gorods_read_collection(col.Con.ccon, col.chandle, &arr, &arrSize, &err)
	
	// Get result length
	arrLen := int(arrSize)

	// Convert C array to slice, backed by arr *C.collEnt_t
	slice := (*[1 << 30]C.collEnt_t)(unsafe.Pointer(arr))[:arrLen:arrLen]

	for i, _ := range slice {
		obj := &slice[i]

		isCollection := (obj.objType != C.DATA_OBJ_T)

		if isCollection {
			col.DataObjects = append(col.DataObjects, NewCollection(obj, col))
		} else {
			col.DataObjects = append(col.DataObjects, NewDataObj(obj, col))
		}	
		
	}
}

func (col *Collection) DataObjs() DataObjs {
	col.Init()

	var response DataObjs

	for i, obj := range col.DataObjects {
		if reflect.TypeOf(obj).String() == "*gorods.DataObj" {
			response = append(response, col.DataObjects[i].(*DataObj))
		}
	}
	
	return response
}

func (col *Collection) Collections() Collections {
	col.Init()

	var response Collections

	for i, obj := range col.DataObjects {
		if reflect.TypeOf(obj).String() == "*gorods.Collection" {
			response = append(response, col.DataObjects[i].(*Collection))
		}
	}
	
	return response
}

func (col *Collection) All() []interface{} {
	col.Init()
	
	return col.DataObjects
}

func (col *Collection) Both() (DataObjs, Collections) {
	return col.DataObjs(), col.Collections()
}

