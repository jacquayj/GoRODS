package gorods

type DataObj struct {
	Path string
}

type DataObjs []*DataObj

func NewDataObj() *DataObj {
	dataObj := &DataObj {"/"}

	return dataObj
}