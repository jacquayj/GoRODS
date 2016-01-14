#include "rods.h"
#include "rodsErrorTable.h"
#include "rodsType.h"
#include "rodsClient.h"
#include "miscUtil.h"
#include "rodsPath.h"
#include "rcConnect.h"
#include "dataObjOpen.h"
#include "dataObjRead.h"
#include "dataObjChksum.h"
#include "dataObjClose.h"
#include "lsUtil.h"

int gorods_connect(rcComm_t* conn, char** err) {

    rodsEnv myEnv;
    int status;

    status = getRodsEnv(&myEnv);
    if ( status != 0 ) {
        *err = "getRodsEnv failed";
        return -1;
    }

    rErrMsg_t errMsg;
    conn = rcConnect(myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone, 1, &errMsg);

    if ( !conn ) {
        *err = "rcConnect failed";
        return -1;
    }
    
    status = clientLogin(conn);
    if ( status != 0 ) {
        *err = "clientLogin failed";
        return -1;
    }

    return 0;
}

int gorods_open_collection(char* path, int* handle, rcComm_t* conn, char** err) {

	collInp_t collOpenInp; 

	bzero(&collOpenInp, sizeof(collOpenInp)); 
	rstrcpy(collOpenInp.collName, path, MAX_NAME_LEN);

	collOpenInp.flags = RECUR_QUERY_FG | VERY_LONG_METADATA_FG; 
	
	*handle = rcOpenCollection(conn, &collOpenInp); 
	if ( *handle < 0 ) { 
		*err = "rcOpenCollection failed";
		return -1;
	} 

	return 0;
}

char* irods_env_str() {
    rodsEnv myEnv;
    int status = getRodsEnv( &myEnv );
    if ( status != 0 ) {
        return (char *)"error getting env";
    }

     char* str = malloc(sizeof(char) * 255);

     sprintf(str, "Host: %s\nPort: %i\nUsername: %s\nZone: %s\n", myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone);


     return str;
}

