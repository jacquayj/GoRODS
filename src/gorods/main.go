package main

import (
   "fmt"
)

// #cgo LDFLAGS: -L./irods/lib/core/obj -lRodsAPIs
// #cgo CFLAGS: -I./irods/lib/core/include -I./irods/lib/api/include -I./irods/lib/md5/include -I./irods/lib/sha1/include -I./irods/server/core/include -I./irods/server/icat/include -I./irods/server/drivers/include -I./irods/server/re/include
// #include "rods.h"
// #include "rodsErrorTable.h"
// #include "rodsType.h"
// #include "rodsClient.h"
// #include "miscUtil.h"
// #include "rodsPath.h"
// #include "rcConnect.h"
// #include "dataObjOpen.h"
// #include "dataObjRead.h"
// #include "dataObjChksum.h"
// #include "dataObjClose.h"
//
// rcComm_t* irods_connect(char** err) {
//
//     rodsEnv myEnv;
//     int status = getRodsEnv( &myEnv );
//     if ( status != 0 ) {
//
//                 //*err = (char *)"getRodsEnv failed.\n";
//         return NULL;
//     }
//     rErrMsg_t errMsg;
//
//     *err = myEnv.rodsHost;
//     rcComm_t* conn = rcConnect( myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone, 1, &errMsg );
//
//
//     if ( !conn ) {
//         //*err = (char *)"rcConnect failed\n";
//         return NULL;
//     }
//
//     return conn;
// }
//
//char* irods_env_str() {
//      rodsEnv myEnv;
//     int status = getRodsEnv( &myEnv );
//     if ( status != 0 ) {
//         return (char *)"error getting env";
//     }
//
//      char str[255];
//
//      sprintf(str, "Host: %s\nPort: %i\nUsername: %s\nZone: %s\n", myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone);
//
//
//      return str;
//
// }
import "C"

func main() {

        var irodsEnv *C.char = C.irods_env_str()

        filename := C.GoString(C.getRodsEnvFileName())



        fmt.Printf(C.GoString(irodsEnv) + filename)


}
