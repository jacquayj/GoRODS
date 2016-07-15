/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #cgo CFLAGS: -ggdb -I/usr/include/irods -I${SRCDIR}/lib/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/build -lgorods /usr/lib/libirods_client.a /usr/lib/libirods_client_api.a /usr/lib/irods/externals/libboost_thread.a /usr/lib/irods/externals/libboost_system.a /usr/lib/irods/externals/libboost_filesystem.a /usr/lib/irods/externals/libboost_program_options.a /usr/lib/irods/externals/libboost_chrono.a /usr/lib/irods/externals/libboost_regex.a /usr/lib/irods/externals/libjansson.a /usr/lib/libirods_client_core.a /usr/lib/libirods_client_plugins.a -lz -lssl -lcrypto -ldl -lpthread -lm -lrt -lstdc++ -rdynamic -Wno-write-strings -DBOOST_SYSTEM_NO_DEPRECATED
// #include "wrapper.h"
import "C"

import (
	"fmt"
	"unsafe"
	"sync"
)

// EnvironmentDefined and UserDefined constants are used when calling
// gorods.New(ConnectionOptions{ Type: ... })
// When EnvironmentDefined is specified, the options stored in ~/.irods/irods_environment.json will be used.
// When UserDefined is specified you must also pass Host, Port, Username, and Zone.
// Password should be set regardless.
const (
	EnvironmentDefined = iota
	UserDefined
)

// Used when calling Type() on different gorods objects
const (
	DataObjType = iota
	CollectionType
	ResourceType
	ResourceGroupType
	UserType
)

func getTypeString(t int) string {
	switch t {
		case DataObjType:
			return "d"
		case CollectionType:
			return "C"
		case ResourceType:
			return "R"
		case UserType:
			return "u"
		default:
			panic(newError(Fatal, "unrecognized meta type constant"))
	}
}

func isString(obj interface{}) bool {
	switch obj.(type) {
		case string:
			return true
		default:		
	}

	return false
}

// IRodsObj is a generic interface used to detect the object type and access common fields
type IRodsObj interface {
	GetType() int
	GetName() string
	GetPath() string
	GetCol() *Collection
	GetCon() *Connection

	// irm -rf
	Destroy() error

	// irm -f {-r}
	Delete(bool) error

	// irm {-r}
	Trash(bool) error

	// irm {-r} {-f}
	Rm(bool, bool) error

	Meta() (*MetaCollection, error)
	Attribute(string) (*Meta, error)
	AddMeta(Meta) (*Meta, error)
	DeleteMeta(string) (*MetaCollection, error)

	String() string
	Open() error
	Close() error
}

type IRodsObjs []IRodsObj

// Exists checks to see if a collection exists in the slice
// and returns true or false
func (objs IRodsObjs) Exists(path string) bool {
	if o := objs.Find(path); o != nil {
		return true
	}

	return false
}

// Find gets a collection from the slice and returns nil if one is not found.
// Both the collection name or full path can be used as input.
func (objs IRodsObjs) Find(path string) IRodsObj {
	
	// Strip trailing forward slash
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for i, obj := range objs {
		if obj.GetPath() == path || obj.GetName() == path {
			return objs[i]
		}
	}

	return nil
}

// Each
func (objs IRodsObjs) Each(iterator func(IRodsObj)) error {
	for _, value := range objs {
		iterator(value)
	}

	return nil
}

// FindRecursive acts just like Find, but also searches sub collections recursively.
// If the collection was not explicitly loaded recursively, only the first level of sub collections will be searched.
func (objs IRodsObjs) FindRecursive(path string) IRodsObj {
	
	var wg sync.WaitGroup
	wg.Add(2)

	var (
		f1 IRodsObj
		f2 IRodsObj
	)

	half := len(objs) / 2

	go func() {
		defer wg.Done()
		f1 = findRecursiveHelper(objs[:half], path)
	}()

	go func() {
		defer wg.Done()
		f2 = findRecursiveHelper(objs[half:], path)
	}()

	wg.Wait()

	if f1 != nil {
		return f1
	}
	
	return f2
}

func findRecursiveHelper(objs IRodsObjs, path string) IRodsObj {
	// Strip trailing forward slash
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for i, obj := range objs {
		if obj.GetPath() == path || obj.GetName() == path {
			return objs[i]
		}

		if obj.GetType() == CollectionType {
			col := obj.(*Collection)

			// Use .DataObjects and not All() so we don't init the non-recursive collections
			if subCol := col.DataObjects.FindRecursive(path); subCol != nil {
				return subCol
			}
		}
	}

	return nil
}


// ConnectionOptions are used when creating iRods iCAT server connections see gorods.New() docs for more info.
type ConnectionOptions struct {
	Type int

	Host string
	Port int
	Zone string

	Username string
	Password string
	Ticket   string
}

type Connection struct {
	ccon *C.rcComm_t

	Connected   bool
	Options     *ConnectionOptions
	OpenedObjs  IRodsObjs
}


// New creates a connection to an iRods iCAT server. EnvironmentDefined and UserDefined
// constants are used in ConnectionOptions{ Type: ... }).
// When EnvironmentDefined is specified, the options stored in ~/.irods/irods_environment.json will be used.
// When UserDefined is specified you must also pass Host, Port, Username, and Zone. Password 
// should be set unless using an anonymous user account with tickets.
func New(opts ConnectionOptions) (*Connection, error) {
	con := new(Connection)

	con.Options = &opts

	var (
		status   C.int
		errMsg   *C.char
		password *C.char
	)

	if con.Options.Password != "" {
		password = C.CString(con.Options.Password)

		defer C.free(unsafe.Pointer(password))
	}

	// Are we passing env values?
	if con.Options.Type == UserDefined {
		host := C.CString(con.Options.Host)
		port := C.int(con.Options.Port)
		username := C.CString(con.Options.Username)
		zone := C.CString(con.Options.Zone)

		defer C.free(unsafe.Pointer(host))
		defer C.free(unsafe.Pointer(username))
		defer C.free(unsafe.Pointer(zone))

		// BUG(jjacquay712): iRods C API code outputs errors messages, need to implement connect wrapper (gorods_connect_env) from a lower level to suppress this output
		// https://github.com/irods/irods/blob/master/iRODS/lib/core/src/rcConnect.cpp#L109
		status = C.gorods_connect_env(&con.ccon, host, port, username, zone, password, &errMsg)
	} else {
		status = C.gorods_connect(&con.ccon, password, &errMsg)
	}

	if status == 0 {
		con.Connected = true
	} else {
		return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: %v", C.GoString(errMsg)))
	}

	if con.Options.Ticket != "" {
		if err := con.SetTicket(con.Options.Ticket); err != nil {
			return nil, err
		}
	}

	return con, nil
}

// SetTicket is equivalent to using the -t flag with icommands
func (con *Connection) SetTicket(t string) error {
	var (
		status   C.int
		errMsg   *C.char
	)

	con.Options.Ticket = t

	ticket := C.CString(t)
	defer C.free(unsafe.Pointer(ticket))

	if status = C.gorods_set_session_ticket(con.ccon, ticket, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Set Ticket Failed: %v", C.GoString(errMsg)))
	}

	return nil
}

// Disconnect closes connection to iRods iCAT server, returns error on failure or nil on success
func (con *Connection) Disconnect() error {
	if status := int(C.rcDisconnect(con.ccon)); status < 0 {
		return newError(Fatal, fmt.Sprintf("iRods rcDisconnect Failed"))
	}

	con.Connected = false

	return nil
}

// String provides connection status and options provided during initialization (gorods.New)
func (obj *Connection) String() string {

	if obj.Options.Type == UserDefined {
		return fmt.Sprintf("Host: %v@%v:%v/%v, Connected: %v\n", obj.Options.Username, obj.Options.Host, obj.Options.Port, obj.Options.Zone, obj.Connected)
	}

	var (
		username *C.char
		host     *C.char
		port     C.int
		zone     *C.char
	)

	defer C.free(unsafe.Pointer(username))
	defer C.free(unsafe.Pointer(host))
	defer C.free(unsafe.Pointer(zone))

	if status := C.irods_env(&username, &host, &port, &zone); status != 0 {
		panic(newError(Fatal, fmt.Sprintf("iRods getEnv Failed")))
	}

	return fmt.Sprintf("Host: %v@%v:%v/%v, Connected: %v\n", C.GoString(username), C.GoString(host), int(port), C.GoString(zone), obj.Connected)
}

// Collection initializes and returns an existing iRods collection using the specified path
func (con *Connection) Collection(startPath string, recursive bool) (*Collection, error) {

	// Check the cache
	if collection := con.OpenedObjs.FindRecursive(startPath); collection == nil {

		// Load collection, no cache found
		if col, err := getCollection(startPath, recursive, con); err == nil {
			con.OpenedObjs = append(con.OpenedObjs, col)

			return col, nil
		} else {
			return nil, err
		}
	} else {
		col := collection.(*Collection)

		// Init the cached collection if recursive is set
		if recursive {
			col.Recursive = true
			
			if er := col.init(); er != nil {
				return nil, er
			}
		}
		
		return col, nil
	}

}

// DataObject directly returns a specific DataObj without the need to traverse collections. Must pass full path of data object.
func (con *Connection) DataObject(dataObjPath string) (dataobj *DataObj, err error) {
	// We use the caching mechanism from Collection()
	dataobj, err = getDataObj(dataObjPath, con)
	
	return
}

// SearchDataObjects searchs for and returns DataObjs slice based on a search string. Use '%' as a wildcard. Equivalent to ilocate command
func (con *Connection) SearchDataObjects(dataObjPath string) (dataobj *DataObj, err error) {
	return nil, nil
}

// QueryMeta 
func (con *Connection) QueryMeta(qString string) (response IRodsObjs, err error) {

	var errMsg *C.char
	var query *C.char = C.CString(qString)
	var colresult C.goRodsPathResult_t
	var dresult C.goRodsPathResult_t

	defer C.free(unsafe.Pointer(query))
	defer C.freeGoRodsPathResult(&colresult)
	defer C.freeGoRodsPathResult(&dresult)

	if status := C.gorods_query_collection(con.ccon, query, &colresult, &errMsg); status != 0 {
		err = newError(Fatal, fmt.Sprintf(C.GoString(errMsg)))
		return
	}

	size := int(colresult.size)

	if size > 0 {
		slice := (*[1 << 30]*C.char)(unsafe.Pointer(colresult.pathArr))[:size:size]

		for _, colString := range slice {
			if c, er := con.Collection(C.GoString(colString), false); er == nil {
				response = append(response, c)
			} else {
				err = er
				return
			}
		}
	}

	if status := C.gorods_query_dataobj(con.ccon, query, &dresult, &errMsg); status != 0 {
		err = newError(Fatal, fmt.Sprintf(C.GoString(errMsg)))
		return
	}

	size = int(dresult.size)
	
	if size > 0 {
		slice := (*[1 << 30]*C.char)(unsafe.Pointer(dresult.pathArr))[:size:size]

		for _, colString := range slice {

			if c, er := con.DataObject(C.GoString(colString)); er == nil {
				response = append(response, c)
			} else {
				err = er
				return
			}
		}
	}

	return
}


// func (con *Connection) QueryMeta(query string) (collection *Collection, err error) {

// }
