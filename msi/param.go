package msi

/*
#include <stdlib.h>
#include "call_microservice.h"
#include "rcMisc.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Param is the golang abstraction for *C.msParam_t types
type Param struct {
	ptr      *C.msParam_t
	rodsType ParamType
}

// NewParam creates a new *Param, with the provided type string
func NewParam(paramType ParamType) *Param {
	p := new(Param)

	p.rodsType = paramType

	cTypeStr := C.CString(string(paramType))
	defer C.free(unsafe.Pointer(cTypeStr))

	p.ptr = C.NewParam(cTypeStr)

	runtime.SetFinalizer(p, paramDestructor)

	return p
}

// Ptr returns an unsafe.Pointer of the underlying *C.msParam_t
func (param *Param) Ptr() unsafe.Pointer {
	return unsafe.Pointer(param.ptr)
}

// String converts STR_MS_T parameters to golang strings
func (param *Param) String() string {

	var cString *C.char

	switch param.rodsType {
	case STR_MS_T:
		cString = (*C.char)(param.ptr.inOutStruct)
	case KeyValPair_MS_T:
		cString = C.GetKVPStr(param.ptr)
	default:
		return (string(param.rodsType) + ".String() not supported")
	}

	return C.GoString(cString)
}

// Type returns the ParamType of the given *Param
func (param *Param) Type() ParamType {
	return param.rodsType
}

// Bytes returns the []byte of BUF_LEN_MS_T type parameters
func (param *Param) Bytes() []byte {
	var bytes []byte

	if param.rodsType == BUF_LEN_MS_T {

		internalBuff := param.ptr.inpOutBuf

		outBuff := unsafe.Pointer(internalBuff.buf)

		bufLen := int(internalBuff.len)

		bytes = (*[1 << 30]byte)(outBuff)[:bufLen:bufLen]
	}

	return bytes
}

// SetBytes sets the underlying BUF_LEN_MS_T struct to the provided byte slice
func (param *Param) SetBytes(bytes []byte) *Param {

	if param.rodsType == BUF_LEN_MS_T {
		length := C.int(len(bytes))

		cBuff := C.NewBytesBuff(length, unsafe.Pointer(&bytes[0]))

		C.fillBufLenInMsParam(param.ptr, length, cBuff)
	}

	return param
}

// SetKVP adds key-value pairs to the underlying KeyValPair_MS_T parameter
func (param *Param) SetKVP(data map[string]string) *Param {
	if param.rodsType == KeyValPair_MS_T {
		kvp := (*C.keyValPair_t)(param.ptr.inOutStruct)

		for key, value := range data {
			C.addKeyVal(kvp, C.CString(key), C.CString(value))
		}

	}
	return param
}

// Int returns the integer value of the underlying INT_MS_T parameter
func (param *Param) Int() int {
	if param.rodsType == INT_MS_T {
		return int(*((*C.int)(param.ptr.inOutStruct)))
	}
	return -1
}

// SetInt sets the integer value of the underlying INT_MS_T parameter
func (param *Param) SetInt(val int) *Param {
	if param.rodsType == INT_MS_T {
		*((*C.int)(param.ptr.inOutStruct)) = C.int(val)
	}
	return param
}

// SetString sets the string value of the underlying STR_MS_T parameter
func (param *Param) SetString(val string) *Param {
	if param.rodsType == STR_MS_T {
		param.ptr.inOutStruct = unsafe.Pointer(C.CString(val))
	}
	return param
}

// SetDataObjInp sets the underlying DataObjInp_MS_T struct fields from a map
// Valid keys and values are: {"objPath": string, "createMode": int, "openFlags": int}
func (param *Param) SetDataObjInp(input map[string]interface{}) *Param {
	if param.rodsType == DataObjInp_MS_T {
		var cInput *C.dataObjInp_t = (*C.dataObjInp_t)(param.ptr.inOutStruct)

		cPathByteStr := []byte(input["objPath"].(string))

		for i, c := range cPathByteStr {
			cInput.objPath[i] = C.char(c)
		}

		cInput.objPath[len(cPathByteStr)] = 0

		if _, ok := input["createMode"]; ok {
			cInput.createMode = C.int(input["createMode"].(int))
		}

		if _, ok := input["openFlags"]; ok {
			cInput.openFlags = C.int(input["openFlags"].(int))
		}

	}

	return param
}

// ConvertTo rebuilds the underlying data of msParam_t*, to the given ParamType.
// This is useful for setting the types of output parameters, since they are blank
// when passed to the microservice. If msParam_t* is nil, it is set to a newly
// allocated structure.
func (param *Param) ConvertTo(t ParamType) *Param {
	cType := C.CString(string(t))

	C.ConvertParam(cType, &param.ptr)
	param.rodsType = t

	return param
}

func paramDestructor(param *Param) {
	C.FreeMsParam(param.ptr)
}

// ToParam creates a new *msi.Param from an existing *C.msParam_t
func ToParam(gParam unsafe.Pointer) *Param {
	param := (*C.msParam_t)(gParam)

	typeStr := C.GoString(C.GetMSParamType(param))

	// Go won't let me access param->type directly
	typ := ParamType(typeStr)

	return &Param{
		param,
		typ,
	}
}
