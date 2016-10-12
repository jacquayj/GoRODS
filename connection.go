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
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
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
	ZoneType
	UserType
	AdminType
	GroupAdminType
	GroupType
	UnknownType
	Cache
	Archive
	Null
	Read
	Write
	Own
	Inherit
	NoInherit
	Local
	Remote
	PAMAuth
	PasswordAuth
)

// IRodsObj is a generic interface used to detect the object type and access common fields
type IRodsObj interface {
	Type() int
	Name() string
	Path() string
	Col() *Collection
	Con() *Connection
	ACL() (ACLs, error)

	OwnerName() string
	CreateTime() time.Time
	ModifyTime() time.Time

	CopyTo(interface{}) error
	MoveTo(interface{}) error
	DownloadTo(string) error

	// irm -rf
	Destroy() error

	// irm -f {-r}
	Delete(bool) error

	// irm {-r}
	Trash(bool) error

	// irm {-r} {-f}
	Rm(bool, bool) error

	Chmod(string, int, bool) error
	GrantAccess(AccessObject, int, bool) error

	Replicate(interface{}, DataObjOptions) error
	Backup(interface{}, DataObjOptions) error
	MoveToResource(interface{}) error
	TrimRepls(TrimOptions) error

	Meta() (*MetaCollection, error)
	Attribute(string) (Metas, error)
	AddMeta(Meta) (*Meta, error)
	DeleteMeta(string) (*MetaCollection, error)

	String() string
	Open() error
	Close() error
}

// IRodsObjs is a slice of IRodsObj (Either a *DataObj or *Collection).
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
		if obj.Path() == path || obj.Name() == path {
			return objs[i]
		}
	}

	return nil
}

// Each provides an iterator for IRodsObjs slice
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
		if obj.Path() == path || obj.Name() == path {
			return objs[i]
		}

		if obj.Type() == CollectionType {
			col := obj.(*Collection)

			// Use .DataObjects and not All() so we don't init the non-recursive collections
			if subCol := col.dataObjects.FindRecursive(path); subCol != nil {
				return subCol
			}
		}
	}

	return nil
}

func chmod(obj IRodsObj, user string, accessLevel int, recursive bool, includeZone bool) error {
	var (
		err        *C.char
		cRecursive C.int
		zone       string
	)

	if accessLevel != Null && accessLevel != Read && accessLevel != Write && accessLevel != Own && accessLevel != Inherit && accessLevel != NoInherit {
		return newError(Fatal, fmt.Sprintf("iRods Chmod DataObject Failed: accessLevel must be Null | Read | Write | Own"))
	}

	if includeZone {
		if z, err := obj.Con().LocalZone(); err == nil {
			zone = z.Name()
		} else {
			return err
		}
	}

	cUser := C.CString(user)
	cPath := C.CString(obj.Path())
	cZone := C.CString(zone)
	cAccessLevel := C.CString(getTypeString(accessLevel))
	defer C.free(unsafe.Pointer(cUser))
	defer C.free(unsafe.Pointer(cPath))
	defer C.free(unsafe.Pointer(cZone))
	defer C.free(unsafe.Pointer(cAccessLevel))

	if recursive {
		cRecursive = C.int(1)
	} else {
		cRecursive = C.int(0)
	}

	ccon := obj.Con().GetCcon()
	defer obj.Con().ReturnCcon(ccon)

	if status := C.gorods_chmod(ccon, cPath, cZone, cUser, cAccessLevel, cRecursive, &err); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Chmod DataObject Failed: %v", C.GoString(err)))
	}

	return nil
}

// ConnectionOptions are used when creating iRods iCAT server connections see gorods.New() docs for more info.
type ConnectionOptions struct {
	Type          int
	AuthType      int
	PAMPassFile   string
	PAMToken      string
	PAMPassExpire int
	Host          string
	Port          int
	Zone          string
	Username      string
	Password      string
	Ticket        string
	FastInit      bool
}

// Connection structs hold information about the iRODS iCAT server, and the user who's connecting. It also contains a cache of opened Collections and DataObjs
type Connection struct {
	ccon       *C.rcComm_t
	cconBuffer chan *C.rcComm_t
	users      Users
	groups     Groups
	zones      Zones
	resources  Resources

	Connected  bool
	Init       bool
	Options    *ConnectionOptions
	OpenedObjs IRodsObjs
}

// NewConnection creates a connection to an iRods iCAT server. EnvironmentDefined and UserDefined
// constants are used in ConnectionOptions{ Type: ... }).
// When EnvironmentDefined is specified, the options stored in ~/.irods/irods_environment.json will be used.
// When UserDefined is specified you must also pass Host, Port, Username, and Zone. Password
// should be set unless using an anonymous user account with tickets.
func NewConnection(opts *ConnectionOptions) (*Connection, error) {
	con := new(Connection)

	con.Options = opts

	var (
		status    C.int
		errMsg    *C.char
		ipassword *C.char
		opassword *C.char
	)

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
		if status = C.gorods_connect_env(&con.ccon, host, port, username, zone, &errMsg); status != 0 {
			return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: %v", C.GoString(errMsg)))
		}
	} else {
		if status = C.gorods_connect(&con.ccon, &errMsg); status != 0 {
			return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: %v", C.GoString(errMsg)))
		}
	}

	ipassword = C.CString(con.Options.Password)
	defer C.free(unsafe.Pointer(ipassword))

	if con.Options.AuthType == 0 {
		con.Options.AuthType = PasswordAuth // Options: PasswordAuth PAMAuth
	}

	if con.Options.PAMPassExpire == 0 {
		con.Options.PAMPassExpire = 1 // Default expiration: 1 hour
	}

	var (
		pamPassFile *os.File
		pamFileErr  error
		size        int64
	)

	if con.Options.AuthType == PAMAuth {

		// Was a PAM token passed?
		if con.Options.PAMToken != "" {

			// Use it, pass directly to clientLoginWithPassword
			opassword = C.CString(con.Options.PAMToken)
			defer C.free(unsafe.Pointer(opassword))

		} else if con.Options.PAMPassFile == "" { // Continue with auth using .Password (ipassword) option, Check to see if PAMPassFile option is not set

			// It's not, fetch password and just keep in memory
			if status = C.gorods_clientLoginPam(con.ccon, ipassword, C.int(con.Options.PAMPassExpire), &opassword, &errMsg); status != 0 {
				return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: clientLoginPam error, invalid password?"))
			}

			defer C.free(unsafe.Pointer(opassword))

		} else { // There is a PAM file path set, save password to FS for subsequent use

			// Does the file/dir exist?
			if finfo, err := os.Stat(con.Options.PAMPassFile); err == nil {
				if !finfo.IsDir() {
					// Open file here
					pamPassFile, pamFileErr = os.OpenFile(con.Options.PAMPassFile, os.O_RDWR, 0666)
					if pamFileErr != nil {
						return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: Problem opening PAMPassFile at %v", con.Options.PAMPassFile))
					}

					size = finfo.Size()

				} else {
					return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: PAMPassFile is a directory durp"))
				}
			} else {
				// Create file here
				pamPassFile, pamFileErr = os.Create(con.Options.PAMPassFile)
				if pamFileErr != nil {
					return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: Problem creating PAMPassFile at %v", con.Options.PAMPassFile))
				}

			}

			// Is this an old password file?
			if size > 0 {

				fileBtz := make([]byte, size)
				if _, er := pamPassFile.Read(fileBtz); er != nil {
					return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: Problem reading PAMPassFile at %v", con.Options.PAMPassFile))
				}

				fileStr := string(fileBtz)
				fileSplit := strings.Split(fileStr, ":")
				unixTimeStamp, _ := strconv.Atoi(fileSplit[0])
				pamPassword := fileSplit[1]

				now := int(time.Now().Unix())

				// Check to see if the password has expired
				if (unixTimeStamp + (con.Options.PAMPassExpire * 60)) <= now {
					// we're expired, refresh
					opassword, pamFileErr = con.fetchAndWritePAMPass(pamPassFile, ipassword)
					if pamFileErr != nil {
						return nil, pamFileErr
					}
				} else {
					// It's still good, use it
					opassword = C.CString(pamPassword)
				}
				defer C.free(unsafe.Pointer(opassword))

			} else {

				// Nope, it's new. Write to the file
				opassword, pamFileErr = con.fetchAndWritePAMPass(pamPassFile, ipassword)
				if pamFileErr != nil {
					return nil, pamFileErr
				}

				defer C.free(unsafe.Pointer(opassword))
			}
		}
	} else if con.Options.AuthType == PasswordAuth {
		opassword = ipassword
	}

	if status = C.clientLoginWithPassword(con.ccon, opassword); status != 0 {

		// if status == C.CAT_PASSWORD_EXPIRED {
		// 	fmt.Printf("expired:%v\n", pamPassFile.Name())
		// }

		if con.Options.AuthType == PAMAuth {

			if pamPassFile != nil {

				// Failure, clear out file for another try.
				if er := pamPassFile.Truncate(int64(0)); er != nil {
					return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: Unable to truncate PAMPassFile: %v", er))
				}

				return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: clientLoginWithPassword error, expired password?"))
			}
		}

		return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: clientLoginWithPassword error, invalid password?"))
	}

	con.cconBuffer = make(chan *C.rcComm_t, 1)
	con.cconBuffer <- con.ccon

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

	if !con.Options.FastInit {
		if err := con.init(); err != nil {
			return nil, err
		}
	}

	return con, nil
}

// fetchAndWritePAMPass attempts to authenticate with the iCAT server using PAM.
// If PAM authentication is successful, it writes the returned PAM authentication token to a file for subsequent use in connections.
func (con *Connection) fetchAndWritePAMPass(pamPassFile *os.File, ipassword *C.char) (*C.char, error) {

	var (
		opassword *C.char
		errMsg    *C.char
	)

	if status := C.gorods_clientLoginPam(con.ccon, ipassword, C.int(con.Options.PAMPassExpire), &opassword, &errMsg); status != 0 {
		return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: clientLoginPam error, invalid password?"))
	}

	if er := pamPassFile.Truncate(0); er != nil {
		return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: Unable to write new password to PAMPassFile"))
	}

	pamPassFormat := strconv.Itoa(int(time.Now().Unix())) + ":" + C.GoString(opassword)

	if _, er := pamPassFile.WriteString(pamPassFormat); er != nil {
		return nil, newError(Fatal, fmt.Sprintf("iRods Connect Failed: Unable to write new password to PAMPassFile"))
	}

	return opassword, nil
}

// GetCcon checks out the connection handle for use in all iRODS operations. Basically a mutex for connections.
// Other goroutines calling this function will block until the handle is returned with con.ReturnCcon, by the goroutine using it.
// This prevents errors in the net code since concurrent API calls aren't supported over a single iRODS connection.
func (con *Connection) GetCcon() *C.rcComm_t {
	return <-con.cconBuffer
}

// ReturnCcon returns the connection handle for use in other threads. Unlocks the mutex.
func (con *Connection) ReturnCcon(ccon *C.rcComm_t) {
	con.cconBuffer <- ccon
}

// SetTicket is equivalent to using the -t flag with icommands
func (con *Connection) SetTicket(t string) error {
	var (
		status C.int
		errMsg *C.char
	)

	con.Options.Ticket = t

	ticket := C.CString(t)
	defer C.free(unsafe.Pointer(ticket))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status = C.gorods_set_session_ticket(ccon, ticket, &errMsg); status != 0 {
		return newError(Fatal, fmt.Sprintf("iRods Set Ticket Failed: %v", C.GoString(errMsg)))
	}

	return nil
}

// Disconnect closes connection to iRods iCAT server, returns error on failure or nil on success
func (con *Connection) Disconnect() error {

	for _, obj := range con.OpenedObjs {
		if er := obj.Close(); er != nil {
			return er
		}
	}

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)

	if status := int(C.rcDisconnect(ccon)); status < 0 {
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
func (con *Connection) Collection(opts CollectionOptions) (*Collection, error) {

	startPath := opts.Path
	recursive := opts.Recursive

	// Check the cache
	if collection := con.OpenedObjs.FindRecursive(startPath); collection == nil {
		//if collection := con.OpenedObjs.FindRecursive(startPath); true {

		// Load collection, no cache found
		if col, err := getCollection(opts, con); err == nil {
			con.OpenedObjs = append(con.OpenedObjs, col)

			return col, nil
		} else {
			return nil, err
		}
	} else {
		col := collection.(*Collection)

		// Init the cached collection if recursive is set
		if recursive {
			col.recursive = true

			if er := col.init(); er != nil {
				return nil, er
			}
		}

		return col, nil
	}

}

// PathType returns DataObjType, CollectionType, or -1 (error) for the iRODS path specified
func (con *Connection) PathType(p string) (int, error) {
	var (
		err        *C.char
		statResult *C.rodsObjStat_t
	)

	path := C.CString(p)
	defer C.free(unsafe.Pointer(path))

	ccon := con.GetCcon()
	defer con.ReturnCcon(ccon)
	defer C.freeRodsObjStat(statResult)

	if status := C.gorods_stat_dataobject(path, &statResult, ccon, &err); status != 0 {
		return -1, newError(Fatal, fmt.Sprintf("iRods Stat Failed: %v, %v", p, C.GoString(err)))
	}

	if statResult.objType == C.DATA_OBJ_T {
		return DataObjType, nil
	} else if statResult.objType == C.COLL_OBJ_T {
		return CollectionType, nil
	}

	return -1, newError(Fatal, "Unknown type")

}

// IQuest accepts a SQL query fragment, returns results in slice of maps
// If upperCase is true, all records will be matched using their uppercase representation.
func (con *Connection) IQuest(query string, upperCase bool) ([]map[string]string, error) {
	var (
		result C.goRodsHashResult_t
		err    *C.char
		upper  int
	)

	result.size = C.int(0)

	z, zErr := con.LocalZone()
	if zErr != nil {
		return nil, zErr
	}

	if upperCase {
		upper = 1
	}

	cQueryString := C.CString(query)
	cZoneName := C.CString(z.Name())
	defer C.free(unsafe.Pointer(cZoneName))
	defer C.free(unsafe.Pointer(cQueryString))

	ccon := con.GetCcon()

	if status := C.gorods_iquest_general(ccon, cQueryString, C.int(0), C.int(upper), cZoneName, &result, &err); status != 0 {
		con.ReturnCcon(ccon)
		if status == C.CAT_NO_ROWS_FOUND {
			return make([]map[string]string, 0), nil
		} else {
			return nil, newError(Fatal, fmt.Sprintf("iRods iquest Failed: %v", C.GoString(err)))
		}
	}

	con.ReturnCcon(ccon)
	defer C.gorods_free_map_result(&result)

	unsafeKeyArr := unsafe.Pointer(result.hashKeys)
	keyArrLen := int(result.keySize)

	unsafeValArr := unsafe.Pointer(result.hashValues)
	valArrLen := int(result.size) * keyArrLen

	response := make([]map[string]string, int(result.size))

	// Convert C array to slice
	keySlice := (*[1 << 30]*C.char)(unsafeKeyArr)[:keyArrLen:keyArrLen]
	valSlice := (*[1 << 30]*C.char)(unsafeValArr)[:valArrLen:valArrLen]

	for n, val := range valSlice {
		mapInx := n / keyArrLen

		var key string

		if n == 0 {
			key = C.GoString(keySlice[0])
		} else {
			key = C.GoString(keySlice[int(math.Mod(float64(n), float64(keyArrLen)))])
		}

		if response[mapInx] == nil {
			response[mapInx] = make(map[string]string)
		}

		response[mapInx][key] = C.GoString(val)
	}

	return response, nil
}

// DataObject directly returns a specific DataObj without the need to traverse collections. Must pass full path of data object.
func (con *Connection) DataObject(dataObjPath string) (dataobj *DataObj, err error) {
	// We use the caching mechanism from Collection()
	dataobj, err = getDataObj(dataObjPath, con)

	return
}

// QueryMeta queries both data objects and collections for matching metadata. Returns IRodsObjs.
func (con *Connection) QueryMeta(qString string) (response IRodsObjs, err error) {

	var errMsg *C.char
	var query *C.char = C.CString(qString)
	var colresult C.goRodsPathResult_t
	var dresult C.goRodsPathResult_t
	var ccon *C.rcComm_t

	defer C.free(unsafe.Pointer(query))
	defer C.freeGoRodsPathResult(&colresult)
	defer C.freeGoRodsPathResult(&dresult)

	ccon = con.GetCcon()
	if status := C.gorods_query_collection(ccon, query, &colresult, &errMsg); status != 0 {
		con.ReturnCcon(ccon)
		err = newError(Fatal, fmt.Sprintf(C.GoString(errMsg)))
		return
	}
	con.ReturnCcon(ccon)

	size := int(colresult.size)

	if size > 0 {
		slice := (*[1 << 30]*C.char)(unsafe.Pointer(colresult.pathArr))[:size:size]

		for _, colString := range slice {

			opts := CollectionOptions{
				Path:      C.GoString(colString),
				Recursive: false,
			}

			if c, er := con.Collection(opts); er == nil {
				response = append(response, c)
			} else {
				err = er
				return
			}
		}
	}

	ccon = con.GetCcon()
	if status := C.gorods_query_dataobj(ccon, query, &dresult, &errMsg); status != 0 {
		con.ReturnCcon(ccon)
		err = newError(Fatal, fmt.Sprintf(C.GoString(errMsg)))
		return
	}
	con.ReturnCcon(ccon)

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

func (con *Connection) init() error {
	if !con.Init {
		con.Init = true

		// user must be rodsadmin
		if err := con.RefreshZones(); err != nil {
			//return err
		}

		if err := con.RefreshResources(); err != nil {
			return err
		}

		// user must be rodsadmin
		if err := con.RefreshUsers(); err != nil {
			//return err
		}

		if err := con.RefreshGroups(); err != nil {
			return err
		}

	}

	return nil
}

// Groups returns a slice of all *Group in the iCAT.
// You must have the proper groupadmin or rodsadmin privileges to use this function.
func (con *Connection) Groups() (Groups, error) {
	if err := con.init(); err != nil {
		return nil, err
	}
	return con.groups, nil
}

// Users returns a slice of all *User in the iCAT.
// You must have the proper rodsadmin privileges to use this function.
func (con *Connection) Users() (Users, error) {
	if err := con.init(); err != nil {
		return nil, err
	}
	return con.users, nil
}

// Zones returns a slice of all *Zone in the iCAT.
// You must have the proper rodsadmin privileges to use this function.
func (con *Connection) Zones() (Zones, error) {
	if err := con.init(); err != nil {
		return nil, err
	}
	return con.zones, nil
}

// Resources returns a slice of all *Resource in the iCAT.
// You must have the proper rodsadmin privileges to use this function.
func (con *Connection) Resources() (Resources, error) {
	if err := con.init(); err != nil {
		return nil, err
	}
	return con.resources, nil
}

// CreateGroup creates a group within the local zone.
// You must have the proper groupadmin or rodsadmin privileges to use this function.
func (con *Connection) CreateGroup(name string) (*Group, error) {

	if z, err := con.LocalZone(); err != nil {
		return nil, err
	} else {
		if err := createGroup(name, z, con); err != nil {
			return nil, err
		}

		if err := con.RefreshGroups(); err != nil {
			return nil, err
		}

		if grps, err := con.Groups(); err != nil {
			return nil, err
		} else {
			if grp := grps.FindByName(name, con); grp != nil {
				return grp, nil
			} else {
				return nil, newError(Fatal, fmt.Sprintf("iRods CreateGroup %v Failed: %v", name, "Unable to locate newly created group in cache"))
			}
		}

	}

}

// CreateUser creates an iRODS user within the local zone.
// You must have the proper rodsadmin privileges to use this function.
func (con *Connection) CreateUser(name string, typ int) (*User, error) {

	if z, err := con.LocalZone(); err != nil {
		return nil, err
	} else {
		if err := createUser(name, z.Name(), typ, con); err != nil {
			return nil, err
		}

		if err := con.RefreshUsers(); err != nil {
			return nil, err
		}

		if usrs, err := con.Users(); err != nil {
			return nil, err
		} else {
			if usr := usrs.FindByName(name, con); usr != nil {
				return usr, nil
			} else {
				return nil, newError(Fatal, fmt.Sprintf("iRods CreateUser %v Failed: %v", name, "Unable to locate newly created user in cache"))
			}
		}
	}

}

// RefreshResources updates the slice returned by con.Resources() with fresh data from the iCAT server.
func (con *Connection) RefreshResources() error {
	if resources, err := con.FetchResources(); err != nil {
		return err
	} else {
		con.resources = resources
	}

	return nil
}

// RefreshUsers updates the slice returned by con.Users() with fresh data from the iCAT server.
func (con *Connection) RefreshUsers() error {
	if users, err := con.FetchUsers(); err != nil {
		return err
	} else {
		con.users = users
	}

	return nil
}

// RefreshZones updates the slice returned by con.Zones() with fresh data from the iCAT server.
func (con *Connection) RefreshZones() error {
	if zones, err := con.FetchZones(); err != nil {
		return err
	} else {
		con.zones = zones
	}

	return nil
}

// RefreshGroups updates the slice returned by con.Groups() with fresh data from the iCAT server.
func (con *Connection) RefreshGroups() error {
	if groups, err := con.FetchGroups(); err != nil {
		return err
	} else {
		con.groups = groups
	}

	return nil
}

// FetchGroups returns a slice of *Group, fresh from the iCAT server.
func (con *Connection) FetchGroups() (Groups, error) {

	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	ccon := con.GetCcon()

	if status := C.gorods_get_groups(ccon, &result, &err); status != 0 {
		con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Groups Failed: %v", C.GoString(err)))
	}

	con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(Groups, 0)

	for _, groupName := range slice {
		if grp, er := initGroup(C.GoString(groupName), con); er == nil {
			response = append(response, grp)
		} else {
			return nil, er
		}

	}

	return response, nil

}

// FetchUsers returns a slice of *User, fresh from the iCAT server.
func (con *Connection) FetchUsers() (Users, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	ccon := con.GetCcon()

	if status := C.gorods_get_users(ccon, &result, &err); status != 0 {
		con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Users Failed: %v", C.GoString(err)))
	}

	con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(Users, 0)

	for _, userNames := range slice {

		nameZone := strings.Split(strings.Trim(C.GoString(userNames), " \n"), "\n")

		for _, name := range nameZone {

			split := strings.Split(name, "#")

			user := split[0]
			zonename := split[1]
			var zone *Zone

			if zones, err := con.Zones(); err != nil {
				return nil, err
			} else {
				if zne := zones.FindByName(zonename, con); zne != nil {
					zone = zne
				} else {
					return nil, newError(Fatal, fmt.Sprintf("iRods Fetch Users Failed: Unable to locate zone in cache"))
				}
			}

			if usr, err := initUser(user, zone, con); err == nil {
				response = append(response, usr)
			} else {
				return nil, err
			}

		}

	}

	return response, nil
}

// FetchResources returns a slice of *Resource, fresh from the iCAT server.
func (con *Connection) FetchResources() (Resources, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	ccon := con.GetCcon()

	if status := C.gorods_get_resources_new(ccon, &result, &err); status != 0 {
		con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods  Get Resources Failed: %v", C.GoString(err)))
	}

	con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(Resources, 0)

	for _, cResourceName := range slice {

		name := C.GoString(cResourceName)

		if resc, err := initResource(name, con); err == nil {
			response = append(response, resc)
		} else {
			return nil, err
		}

	}

	return response, nil
}

// FetchZones returns a slice of *Zone, fresh from the iCAT server.
func (con *Connection) FetchZones() (Zones, error) {
	var (
		result C.goRodsStringResult_t
		err    *C.char
	)

	result.size = C.int(0)

	ccon := con.GetCcon()

	if status := C.gorods_get_zones(ccon, &result, &err); status != 0 {
		con.ReturnCcon(ccon)
		return nil, newError(Fatal, fmt.Sprintf("iRods Get Zones Failed: %v", C.GoString(err)))
	}

	con.ReturnCcon(ccon)

	defer C.gorods_free_string_result(&result)

	unsafeArr := unsafe.Pointer(result.strArr)
	arrLen := int(result.size)

	// Convert C array to slice, backed by arr *C.char
	slice := (*[1 << 30]*C.char)(unsafeArr)[:arrLen:arrLen]

	response := make(Zones, 0)

	for _, cZoneName := range slice {

		zoneNames := strings.Split(strings.Trim(C.GoString(cZoneName), " \n"), "\n")

		for _, name := range zoneNames {

			if zne, err := initZone(name, con); err == nil {
				response = append(response, zne)
			} else {
				return nil, err
			}

		}

	}

	return response, nil
}

// LocalZone returns the *Zone. First it checks the ConnectionOptions.Zone and uses that, otherwise it pulls it fresh from the iCAT server.
func (con *Connection) LocalZone() (*Zone, error) {

	var (
		cZoneName *C.char
		err       *C.char
	)

	if con.Options.Zone == "" {
		ccon := con.GetCcon()
		if status := C.gorods_get_local_zone(ccon, &cZoneName, &err); status != 0 {
			con.ReturnCcon(ccon)
			return nil, newError(Fatal, fmt.Sprintf("iRods Get Local Zone Failed: %v", C.GoString(err)))
		}
		con.ReturnCcon(ccon)
	} else {
		cZoneName = C.CString(con.Options.Zone)
	}

	defer C.free(unsafe.Pointer(cZoneName))

	zoneName := strings.Trim(C.GoString(cZoneName), " \n")

	if znes, err := con.Zones(); err != nil {
		return nil, err
	} else {
		if zne := znes.FindByName(zoneName, con); zne == nil {
			return nil, newError(Fatal, fmt.Sprintf("iRods Get Local Zone Failed: Local zone not found in cache"))
		} else {
			return zne, nil
		}
	}

}
