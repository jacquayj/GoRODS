/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"path/filepath"
	"unsafe"
)

// Meta structs contain information about a single iRods metadata attribute-value-units (AVU) triple
type Meta struct {
	Attribute string
	Value     string
	Units     string
	Parent    *MetaCollection
}

type Metas []*Meta

// MetaCollection is a collection of metadata AVU triples for a single data object
type MetaCollection struct {
	Metas Metas
	Obj   IRodsObj
	Con   *Connection
}

func newMetaCollection(obj IRodsObj) (*MetaCollection, error) {

	result := new(MetaCollection)
	result.Obj = obj
	result.Con = obj.GetCon()

	if err := result.init(); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *Meta) getTypeRodsString() string {
	return getTypeString(m.Parent.Obj.GetType())
}

// SetValue will modify metadata AVU value only
func (m *Meta) SetValue(value string) (*Meta, error) {
	return m.Set(value, m.Units)
}

// SetUnits will modify metadata AVU units only
func (m *Meta) SetUnits(units string) (*Meta, error) {
	return m.Set(m.Value, units)
}

// Set will modify metadata AVU value & units
func (m *Meta) Set(value string, units string) (*Meta, error) {
	return m.SetAll(m.Attribute, value, units)
}

// Rename will modify metadata AVU attribute name only
func (m *Meta) Rename(attributeName string) (*Meta, error) {
	return m.SetAll(attributeName, m.Value, m.Units)
}

// Delete deletes the current Meta struct from iRods object
func (m *Meta) Delete() (*MetaCollection, error) {

	mT := C.CString(m.getTypeRodsString())
	path := C.CString(m.Parent.Obj.GetPath())
	oa := C.CString(m.Attribute)
	ov := C.CString(m.Value)
	ou := C.CString(m.Units)

	defer C.free(unsafe.Pointer(mT))
	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(oa))
	defer C.free(unsafe.Pointer(ov))
	defer C.free(unsafe.Pointer(ou))

	var err *C.char

	if status := C.gorods_rm_meta(mT, path, oa, ov, ou, m.Parent.Con.ccon, &err); status < 0 {
		return m.Parent, newError(Fatal, fmt.Sprintf("iRods rm Meta Failed: %v, %v", m.Parent.Obj.GetPath(), C.GoString(err)))
	}

	m.Parent.Refresh()

	return m.Parent, nil
}

// SetAll will modify metadata AVU with all three paramaters (Attribute, Value, Unit)
func (m *Meta) SetAll(attributeName string, value string, units string) (newMeta *Meta, e error) {

	if attributeName != m.Attribute || value != m.Value || units != m.Units {
		mT := C.CString(m.getTypeRodsString())
		path := C.CString(m.Parent.Obj.GetPath())
		oa := C.CString(m.Attribute)
		ov := C.CString(m.Value)
		ou := C.CString(m.Units)
		na := C.CString(attributeName)
		nv := C.CString(value)
		nu := C.CString(units)

		defer C.free(unsafe.Pointer(mT))
		defer C.free(unsafe.Pointer(path))
		defer C.free(unsafe.Pointer(oa))
		defer C.free(unsafe.Pointer(ov))
		defer C.free(unsafe.Pointer(ou))
		defer C.free(unsafe.Pointer(na))
		defer C.free(unsafe.Pointer(nv))
		defer C.free(unsafe.Pointer(nu))

		var err *C.char

		if status := C.gorods_mod_meta(mT, path, oa, ov, ou, na, nv, nu, m.Parent.Con.ccon, &err); status < 0 {
			e = newError(Fatal, fmt.Sprintf("iRods Set Meta Failed: %v, %v", m.Parent.Obj.GetPath(), C.GoString(err)))
			return
		}

		m.Parent.Refresh()
	}

	newMeta, e = m.Parent.Get(attributeName)
	return
}

func (mc *MetaCollection) init() error {
	// If MetaCollection hasn't been opened, do it!
	if len(mc.Metas) < 1 {
		if err := mc.ReadMeta(); err != nil {
			return err
		}
	}

	return nil
}

// Refresh clears existing metadata triples and grabs updated copy from iCAT server.
// It's an alias of ReadMeta()
func (mc *MetaCollection) Refresh() error {
	return mc.ReadMeta()
}

// ReadMeta clears existing metadata triples and grabs updated copy from iCAT server.
func (mc *MetaCollection) ReadMeta() error {
	var (
		err        *C.char
		metaResult C.goRodsMetaResult_t
	)

	mc.Metas = make([]*Meta, 0)

	name := C.CString(mc.Obj.GetName())

	defer C.free(unsafe.Pointer(name))

	switch mc.Obj.GetType() {
	case DataObjType:
		cwd := C.CString(mc.Obj.GetCol().Path)
		defer C.free(unsafe.Pointer(cwd))

		if status := C.gorods_meta_dataobj(name, cwd, &metaResult, mc.Con.ccon, &err); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Get Meta Failed: %v, %v", cwd, C.GoString(err)))
		}
	case CollectionType:
		cwd := C.CString(filepath.Dir(mc.Obj.GetPath()))
		defer C.free(unsafe.Pointer(cwd))

		if status := C.gorods_meta_collection(name, cwd, &metaResult, mc.Con.ccon, &err); status != 0 {
			return newError(Fatal, fmt.Sprintf("iRods Get Meta Failed: %v, %v", cwd, C.GoString(err)))
		}
	case ResourceType:

	case ResourceGroupType:

	case UserType:

	default:
		return newError(Fatal, "unrecognized meta type constant")
	}

	size := int(metaResult.size)

	slice := (*[1 << 30]C.goRodsMeta_t)(unsafe.Pointer(metaResult.metaArr))[:size:size]

	for _, meta := range slice {

		m := new(Meta)

		m.Attribute = C.GoString(meta.name)
		m.Value = C.GoString(meta.value)
		m.Units = C.GoString(meta.units)
		m.Parent = mc

		mc.Metas = append(mc.Metas, m)
	}

	C.freeGoRodsMetaResult(&metaResult)

	return nil
}

// String shows the contents of the meta struct.
//
// Sample output:
//
// 	Attr1: Val (unit: foo)
func (m *Meta) String() string {
	return m.Attribute + ": " + m.Value + " (unit: " + m.Units + ")"
}

// String shows the contents of the meta collection.
//
// Sample output:
//
// 	Attr1: Val (unit: )
// 	Attr2: Yes (unit: bool)
func (mc *MetaCollection) String() string {
	mc.init()

	var str string

	str = "Metadata: " + mc.Obj.GetPath() + "\n"
	for _, m := range mc.Metas {
		str += "\t" + m.String() + "\n"
	}

	return str
}

// Get finds a single Meta struct by it's Attribute field. Similar to Attribute() function of other types.
func (mc *MetaCollection) Get(attr string) (*Meta, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	for i, m := range mc.Metas {
		if m.Attribute == attr {
			return mc.Metas[i], nil
		}
	}

	return nil, newError(Fatal, fmt.Sprintf("iRods Get Meta Failed, no match"))
}

// All
func (mc *MetaCollection) All() (Metas, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	return mc.Metas, nil
}

// Delete deletes the meta AVU triple from the data object, identified by it's Attribute field
func (mc *MetaCollection) Delete(attr string) (err error) {
	if err = mc.init(); err != nil {
		return
	}

	var meta *Meta

	if meta, err = mc.Get(attr); err != nil {
		return
	}

	if _, err = meta.Delete(); err != nil {
		return
	}

	return
}

// Add creates a new meta AVU triple, returns pointer to the created Meta struct
func (mc *MetaCollection) Add(m Meta) (*Meta, error) {
	if er := mc.init(); er != nil {
		return nil, er
	}

	_, er := mc.Get(m.Attribute)

	if m.Attribute != "" && m.Value != "" && er == nil {
		m.Parent = mc

		mT := C.CString(m.getTypeRodsString())
		path := C.CString(m.Parent.Obj.GetPath())
		na := C.CString(m.Attribute)
		nv := C.CString(m.Value)
		nu := C.CString(m.Units)

		defer C.free(unsafe.Pointer(mT))
		defer C.free(unsafe.Pointer(path))
		defer C.free(unsafe.Pointer(na))
		defer C.free(unsafe.Pointer(nv))
		defer C.free(unsafe.Pointer(nu))

		var err *C.char

		if status := C.gorods_add_meta(mT, path, na, nv, nu, m.Parent.Con.ccon, &err); status < 0 {
			return nil, newError(Fatal, fmt.Sprintf("iRods Add Meta Failed: %v, %v", m.Parent.Obj.GetPath(), C.GoString(err)))
		}

		m.Parent.Refresh()

	} else {
		return nil, newError(Fatal, fmt.Sprintf("iRods Add Meta Failed: Please specify Attribute and Value fields or the attribute already exists"))
	}

	if attr, er := m.Parent.Get(m.Attribute); er == nil {
		return attr, nil
	} else {
		return nil, er
	}

}
