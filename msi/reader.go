package msi

import (
	"io"
)

// ObjReader provides a golang io.Reader interface for iRODS 
// data object, identified by the INT_MS_T object reference
type ObjReader struct {
	objHandle *Param
}

// NewObjReader accepts an iRODS data object path string, opens a reference to the 
// object using msiDataObjOpen, and returns an *ObjReader which satisfies the io.Reader interface.
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

// Close calls msiDataObjClose on the opened data object handle.
func (rdr *ObjReader) Close() error {
	closeOut := NewParam(INT_MS_T)

	return Call("msiDataObjClose", rdr.objHandle, closeOut)
}

// Read reads []bytes from data objects in iRODS. 
// It satisfies the io.Reader interface.
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
