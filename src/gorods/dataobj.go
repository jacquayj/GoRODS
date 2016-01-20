package gorods

type DataObj struct {
	Path string
}

type DataObjs []*DataObj

func (obj *DataObj) String() string {
	return "DataObject: " + obj.Path
}

func NewDataObj(path string) *DataObj {
	dataObj := &DataObj {Path: path,}

	return dataObj
}