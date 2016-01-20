package gorods

// #include "wrapper.h"
import "C"

type DataObj struct {
	Path string
	collent *C.collEnt_t
}

type DataObjs []*DataObj

func (obj *DataObj) String() string {
	return "DataObject: " + obj.Path
}

// func (obj *DataObj) Read() []byte {
	
// }

func NewDataObj(data *C.collEnt_t) *DataObj {

	dataObj := &DataObj {}

	dataObj.collent = data
	dataObj.Path = C.GoString(dataObj.collent.collName) + C.GoString(dataObj.collent.dataName)

	return dataObj
}