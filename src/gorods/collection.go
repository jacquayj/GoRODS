package gorods

// #include "wrapper.h"
import "C"

type Collection struct {
	ccoll *C.collInp_t
	Path string
}

type Collections []*Collection

func NewCollection(startPath string) *Collection {
	col := &Collection {Path: startPath}


	return col
}

// func (col *Collection) GetDataObjs() DataObjs {

// }

// func (col *Collection) GetCollObjs() Collections {

// }

