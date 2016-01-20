package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
)

type Collection struct {
	//ccoll *C.collInp_t
	Path string
	Con *Connection
	chandle C.int
}

func (obj *Collection) String() string {
	return "Collection: " + obj.Path
}

func NewCollection(startPath string, con *Connection) *Collection {
	col := &Collection {Path: startPath}

	col.Con = con
	col.Path = startPath

	var errMsg *C.char

	if status := C.gorods_open_collection(C.CString(col.Path), &col.chandle, col.Con.ccon, &errMsg); status != 0 {
		panic(fmt.Sprintf("iRods Open Collection Failed: %v", C.GoString(errMsg)))
	}

	return col
}


func (col *Collection) GetDataObjs() DataObjs {

	var err *C.char
	var arr *C.collEnt_t
	var arrSize C.int

	// Read data objs from collection
	C.gorods_read_collection_dataobjs(col.Con.ccon, col.chandle, &arr, &arrSize, &err)
	
	// Get result length
	arrLen := int(arrSize)

	// Convert C array to slice
	slice := (*[1 << 30]C.collEnt_t)(unsafe.Pointer(arr))[:arrLen:arrLen]

	var objs DataObjs
	for _, item := range slice {
		objs = append(objs, NewDataObj(C.GoString(item.collName)))
	}
	
	return objs
}




// func (col *Collection) GetCollObjs() Collections {

// }

