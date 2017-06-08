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

type Param struct {
	ptr      *C.msParam_t
	rodsType ParamType
}

func NewParam(paramType ParamType) *Param {
	p := new(Param)

	p.rodsType = paramType

	cTypeStr := C.CString(string(paramType))
	defer C.free(unsafe.Pointer(cTypeStr))

	p.ptr = C.NewParam(cTypeStr)

	runtime.SetFinalizer(p, paramDestructor)

	return p
}

func (param *Param) String() string {

	var cString *C.char

	switch param.rodsType {
	case STR_MS_T:
		cString = (*C.char)(param.ptr.inOutStruct)
	default:
		return (string(param.rodsType) + ".String() not supported")
	}

	return C.GoString(cString)
}

func (param *Param) Type() ParamType {
	return param.rodsType
}

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

func (param *Param) SetKVP(data map[string]string) *Param {
	if param.rodsType == KeyValPair_MS_T {
		kvp := (*C.keyValPair_t)(param.ptr.inOutStruct)

		for key, value := range data {
			C.addKeyVal(kvp, C.CString(key), C.CString(value))
		}

	}
	return param
}

func (param *Param) SetInt(val int) *Param {
	if param.rodsType == INT_MS_T {
		*((*C.int)(param.ptr.inOutStruct)) = C.int(val)
	}
	return param
}

func (param *Param) SetString(val string) *Param {
	if param.rodsType == STR_MS_T {
		param.ptr.inOutStruct = unsafe.Pointer(C.CString(val))
	}
	return param
}

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

func (param *Param) ConvertTo(t ParamType) *Param {
	cType := C.CString(string(t))

	C.ConvertParam(cType, &param.ptr)
	param.rodsType = t

	return param
}

func paramDestructor(param *Param) {
	C.FreeMsParam(param.ptr)
}

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
