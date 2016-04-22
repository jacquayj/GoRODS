/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
)

// Meta structs contain information about a single iRods metadata attribute-value-units (AVU) triple
type Meta struct {
	Attribute string
	Value	  string
	Units 	  string
}

// MetaCollection is a collection of metadata AVU triples for a single data object
type MetaCollection []*Meta

func initMetaCollection(metatype int, objName string, objPath string, ccon *C.rcComm_t) MetaCollection {
	var (
		err        *C.char
		metaResult C.goRodsMetaResult_t
	)

	result := make(MetaCollection, 0)

	name := C.CString(objName)
	cwd := C.CString(objPath)

	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(cwd))

	switch metatype {
		case DataObjType:
			if status := C.gorods_meta_dataobj(name, cwd, &metaResult, ccon, &err); status != 0 {
				panic(newError(Fatal, fmt.Sprintf("iRods Get Meta Failed: %v, %v", objPath, C.GoString(err))))
			}
		case CollectionType:
			if status := C.gorods_meta_collection(name, cwd, &metaResult, ccon, &err); status != 0 {
				panic(newError(Fatal, fmt.Sprintf("iRods Get Meta Failed: %v, %v", objPath, C.GoString(err))))
			}
		case ResourceType:
			
		case ResourceGroupType:
			
		case UserType:
			
		default:
			panic(newError(Fatal, "unrecognized meta type constant"))
	}

	size := int(metaResult.size)

	slice := (*[1 << 30]C.goRodsMeta_t)(unsafe.Pointer(metaResult.metaArr))[:size:size]

	for _, meta := range slice {

		m := new(Meta)

		m.Attribute = C.GoString(meta.name)
		m.Value = C.GoString(meta.value)
		m.Units = C.GoString(meta.units)

		result = append(result, m)
	}

	C.freeGoRodsMetaResult(&metaResult)

	return result
}

// String shows the contents of the meta struct.
//
// Sample output:
//
// 	Attr1: Val (unit: foo)
func (m *Meta) String() string {
	return m.Attribute + ": " + m.Value + " (unit: " + m.Units + ")"
}

// String shows the contents of the meta collection.
//
// Sample output:
//
// 	Attr1: Val (unit: )
// 	Attr2: Yes (unit: bool)
func (metas MetaCollection) String() string {
	var str string

	for _, m := range metas {
		str += m.String() + "\n"
	}

	return str
}

// Get finds a single Meta struct by it's Attribute field. Similar to Attribute() function of other types.
func (metas MetaCollection) Get(attr string) *Meta {
	for i, m := range metas {
		if m.Attribute == attr {
			return metas[i]
		}
	}

	return nil
}