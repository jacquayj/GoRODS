/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. and The BioTeam, Inc.  ***
 *** For more information please refer to the LICENSE.md file                                   ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"
)

// Meta structs contain information about a single iRODS metadata attribute-value-units (AVU) triple
type Meta struct {
	Attribute string
	Value     string
	Units     string
	Parent    *MetaCollection
}

// Metas is a slice of *Meta
type Metas []*Meta

// MatchOne returns a single Meta struct from the slice, matching on Attribute, Value, and Units
func (ms Metas) MatchOne(m *Meta) *Meta {

	if len(ms) > 0 {
		for _, am := range ms {
			if am.Attribute == m.Attribute && am.Value == m.Value && am.Units == m.Units {
				return am
			}
		}
	}

	return nil
}

// String shows the contents of the meta slice struct.
//
// Sample output:
//
// 	Attr1: Val (unit: foo)
func (ms Metas) String() string {
	result := "["
	for _, m := range ms {
		result += m.String() + ",\n"
	}
	result = strings.TrimRight(result, ",\n")
	result += "]"
	return result
}

// MetaCollection is a collection of metadata AVU triples for a single data object
type MetaCollection struct {
	Metas Metas
	Obj   MetaObj
	Con   *Connection
}

func newMetaCollection(obj MetaObj) (*MetaCollection, error) {

	result := new(MetaCollection)
	result.Obj = obj
	result.Con = obj.Con()

	if err := result.init(); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *Meta) getTypeRodsString() string {
	return GetShortTypeString(m.Parent.Obj.Type())
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

// Delete deletes the current Meta struct from iRODS object
func (m *Meta) Delete() (*MetaCollection, error) {

	mT := C.CString(m.getTypeRodsString())
	path := C.CString(m.Parent.Obj.Path())
	oa := C.CString(m.Attribute)
	ov := C.CString(m.Value)
	ou := C.CString(m.Units)

	defer C.free(unsafe.Pointer(mT))
	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(oa))
	defer C.free(unsafe.Pointer(ov))
	defer C.free(unsafe.Pointer(ou))

	var err *C.char

	ccon := m.Parent.Con.GetCcon()

	if status := C.gorods_rm_meta(mT, path, oa, ov, ou, ccon, &err); status < 0 {
		m.Parent.Con.ReturnCcon(ccon)

		return m.Parent, newError(Fatal, status, fmt.Sprintf("iRODS rm Meta Failed: %v, %v", m.Parent.Obj.Path(), C.GoString(err)))
	}

	m.Parent.Con.ReturnCcon(ccon)

	m.Parent.Refresh()

	return m.Parent, nil
}

// SetAll will modify metadata AVU with all three paramaters (Attribute, Value, Unit)
func (m *Meta) SetAll(attributeName string, value string, units string) (newMeta *Meta, e error) {

	if attributeName != m.Attribute || value != m.Value || units != m.Units {
		mT := C.CString(m.getTypeRodsString())
		path := C.CString(m.Parent.Obj.Path())
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

		ccon := m.Parent.Con.GetCcon()

		if status := C.gorods_mod_meta(mT, path, oa, ov, ou, na, nv, nu, ccon, &err); status < 0 {
			m.Parent.Con.ReturnCcon(ccon)
			e = newError(Fatal, status, fmt.Sprintf("iRODS Set Meta Failed: %v, %v", m.Parent.Obj.Path(), C.GoString(err)))
			return
		}

		m.Parent.Con.ReturnCcon(ccon)

		m.Attribute = attributeName
		m.Value = value
		m.Units = units

	}

	newMeta = m

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

	mc.Metas = make(Metas, 0)

	name := C.CString(mc.Obj.Name())

	defer C.free(unsafe.Pointer(name))

	switch mc.Obj.Type() {
	case DataObjType:

		ccon := mc.Con.GetCcon()
		defer mc.Con.ReturnCcon(ccon)

		cwdGo := (mc.Obj.(*DataObj)).Col().Path()
		cwd := C.CString(cwdGo)

		defer C.free(unsafe.Pointer(cwd))

		if status := C.gorods_meta_dataobj(name, cwd, &metaResult, ccon, &err); status != 0 {
			if status == C.CAT_NO_ROWS_FOUND {
				return nil
			} else {
				return newError(Fatal, status, fmt.Sprintf("iRODS Get Meta Failed: %v, %v, %v", cwdGo, C.GoString(err), status))
			}
		}
	case CollectionType:

		ccon := mc.Con.GetCcon()
		defer mc.Con.ReturnCcon(ccon)

		cwdGo := filepath.Dir(mc.Obj.Path())
		cwd := C.CString(cwdGo)

		defer C.free(unsafe.Pointer(cwd))

		if status := C.gorods_meta_collection(name, cwd, &metaResult, ccon, &err); status != 0 {
			if status == C.CAT_NO_ROWS_FOUND {
				return nil
			} else {
				return newError(Fatal, status, fmt.Sprintf("iRODS Get Meta Failed: %v, %v, %v", cwdGo, C.GoString(err), status))
			}
		}
	case ResourceType:

	case ResourceGroupType:

	case UserType, GroupType, AdminType, GroupAdminType:

		gZone, zErr := mc.Con.LocalZone()
		if zErr != nil {
			return zErr
		}

		ccon := mc.Con.GetCcon()
		defer mc.Con.ReturnCcon(ccon)

		zone := C.CString(gZone.Name())
		name := C.CString(mc.Obj.Name())
		defer C.free(unsafe.Pointer(name))
		defer C.free(unsafe.Pointer(zone))

		if status := C.gorods_meta_user(name, zone, &metaResult, ccon, &err); status != 0 {
			if status == C.CAT_NO_ROWS_FOUND {
				return nil
			} else {
				return newError(Fatal, status, fmt.Sprintf("iRODS Get Meta Failed: %v, %v, %v", C.GoString(name), C.GoString(err), status))
			}
		}

	default:
		return newError(Fatal, -1, "unrecognized meta type constant")
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

	str = "Metadata: " + mc.Obj.Path() + "\n"
	for _, m := range mc.Metas {
		str += "\t" + m.String() + "\n"
	}

	return str
}

// First finds the first matching Meta struct by it's Attribute field
func (mc *MetaCollection) First(attr string) (*Meta, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	for i, m := range mc.Metas {
		if m.Attribute == attr {
			return mc.Metas[i], nil
		}
	}

	return nil, newError(Fatal, -1, fmt.Sprintf("iRODS Get Meta Failed, no match"))
}

// Get returns a Meta struct slice (since attributes can share the same name in iRODS), matching by their Attribute field. Similar to Attribute() function of other types
func (mc *MetaCollection) Get(attr string) (Metas, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	result := make(Metas, 0)

	for _, m := range mc.Metas {
		if m.Attribute == attr {
			result = append(result, m)
		}
	}

	if len(result) == 0 {
		return result, newError(Fatal, -1, fmt.Sprintf("iRODS Get Meta Failed, no match"))
	}

	return result, nil
}

// All returns a slice of all Meta structs in the MetaCollection
func (mc *MetaCollection) All() (Metas, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	return mc.Metas, nil
}

// Each accepts an iterator function for looping over all Meta structs in the MetaCollection
func (mc *MetaCollection) Each(iterator func(*Meta)) error {
	if err := mc.init(); err != nil {
		return err
	}

	for _, value := range mc.Metas {
		iterator(value)
	}

	return nil
}

// Delete deletes the meta AVU triple from the data object, identified by it's Attribute field
func (mc *MetaCollection) Delete(attr string) (err error) {
	if err = mc.init(); err != nil {
		return
	}

	var meta Metas

	if meta, err = mc.Get(attr); err != nil {
		return
	}

	for _, m := range meta {
		if _, err = m.Delete(); err != nil {
			return
		}
	}

	return
}

// Add creates a new meta AVU triple, returns pointer to the created Meta struct
func (mc *MetaCollection) Add(m Meta) (*Meta, error) {
	if er := mc.init(); er != nil {
		return nil, er
	}

	if existingMeta, er := mc.Get(m.Attribute); er == nil {
		if len(existingMeta) > 0 {
			for _, am := range existingMeta {
				if m.Value == am.Value {
					return nil, newError(Fatal, -1, fmt.Sprintf("iRODS Add Meta Failed: Attribute + Value already exists"))
				}
			}
		}
	}

	if m.Attribute != "" && m.Value != "" {
		m.Parent = mc

		mT := C.CString(m.getTypeRodsString())
		path := C.CString(m.Parent.Obj.Path())
		na := C.CString(m.Attribute)
		nv := C.CString(m.Value)
		nu := C.CString(m.Units)

		defer C.free(unsafe.Pointer(mT))
		defer C.free(unsafe.Pointer(path))
		defer C.free(unsafe.Pointer(na))
		defer C.free(unsafe.Pointer(nv))
		defer C.free(unsafe.Pointer(nu))

		var err *C.char

		ccon := m.Parent.Con.GetCcon()

		if status := C.gorods_add_meta(mT, path, na, nv, nu, ccon, &err); status < 0 {
			m.Parent.Con.ReturnCcon(ccon)
			return nil, newError(Fatal, status, fmt.Sprintf("iRODS Add Meta Failed: %v, %v", m.Parent.Obj.Path(), C.GoString(err)))
		}

		m.Parent.Con.ReturnCcon(ccon)

		m.Parent.Refresh()

	} else {
		return nil, newError(Fatal, -1, fmt.Sprintf("iRODS Add Meta Failed: Please specify Attribute and Value fields"))
	}

	if attrs, er := m.Parent.Get(m.Attribute); er == nil {
		if am := attrs.MatchOne(&m); am != nil {
			return am, nil
		}
		return nil, newError(Fatal, -1, fmt.Sprintf("iRODS Add Meta Error: Unable to locate added meta triple"))
	} else {
		return nil, er
	}

}
