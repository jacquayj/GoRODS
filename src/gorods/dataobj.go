package gorods

// #include "wrapper.h"
import "C"

type DataObj struct {
	Path string
	IsCollection bool
	Con *Connection
	Col *Collection
	
	collent *C.collEnt_t
}

type DataObjs []*DataObj

func (obj *DataObj) String() string {
	if obj.IsCollection {
		return "Collection: " + obj.Path
	}
	return "DataObject: " + obj.Path
}

// func (obj *DataObj) Read() []byte {
	
// }

func NewDataObj(data *C.collEnt_t, col *Collection) *DataObj {
	
	dataObj := new(DataObj)
	
	dataObj.collent = data
	dataObj.Col = col
	dataObj.Con = dataObj.Col.Con
	dataObj.Path = C.GoString(dataObj.collent.collName)
	dataObj.IsCollection = (dataObj.collent.objType != C.DATA_OBJ_T)

	if !dataObj.IsCollection {
		dataObj.Path += "/" + C.GoString(dataObj.collent.dataName)
	}

	return dataObj
}