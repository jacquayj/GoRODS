package msi

import (
	//"fmt"
	"io"
)

type ObjReader struct {
	objHandle *Param
}

func NewObjReader(irodsPath string) (*ObjReader, error) {
	reader := new(ObjReader)

	reader.objHandle = NewParam(INT_MS_T)
	openPath := NewParam(DataObjInp_MS_T).SetDataObjInp(map[string]interface{}{
		"objPath": irodsPath,
	})

	if err := Call("msiDataObjOpen", openPath, reader.objHandle); err != nil {
		return nil, err
	}

	return reader, nil
}

func (rdr *ObjReader) Close() error {
	closeOut := NewParam(INT_MS_T)

	return Call("msiDataObjClose", rdr.objHandle, closeOut)
}

func (rdr *ObjReader) Read(data []byte) (n int, err error) {
	length := len(data)

	outBuff := NewParam(BUF_LEN_MS_T)
	readLen := NewParam(INT_MS_T).SetInt(length)

	if err = Call("msiDataObjRead", rdr.objHandle, readLen, outBuff); err != nil {
		return
	}

	readBytes := outBuff.Bytes()
	actualLen := len(readBytes)

	n = actualLen

	if n == 0 {
		err = io.EOF
		return
	}

	copy(data, readBytes)

	return

}
