package msi

// #cgo CFLAGS: -I/usr/include/irods
// #cgo CPPFLAGS: -I/usr/include/irods -I/opt/irods-externals/boost1.60.0-0/include
// #cgo LDFLAGS: -lirods_server -lirods_common -lpthread
import "C"

import (
	"fmt"
	//"log"
	"unsafe"
)

type Param struct {
	ptr unsafe.Pointer
}

var rei unsafe.Pointer

func Configure(ruleExecInfo unsafe.Pointer) {
	rei = ruleExecInfo
}

func Call(msiName string, params ...interface{}) error {
	if rei == nil {
		return fmt.Errorf("Unable to call %v, ruleExecInfo is nil, please set using msi.Configure", msiName)
	}

	//numParams := len(params)

	return nil

}
