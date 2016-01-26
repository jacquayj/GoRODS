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

int gorods_connect(rcComm_t** conn, char* password, char** err) {

    rodsEnv myEnv;
    int status;

    status = getRodsEnv(&myEnv);
    if ( status != 0 ) {
        *err = "getRodsEnv failed";
        return -1;
    }

    rErrMsg_t errMsg;
    *conn = rcConnect(myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone, 1, &errMsg);

    if ( !*conn ) {
        *err = "rcConnect failed";
        return -1;
    }
    
    if ( password != NULL ) {
    	status = clientLoginWithPassword(*conn, password);
    } else {
    	status = clientLogin(*conn);
    }
    
    if ( status != 0 ) {
        *err = "clientLogin failed. Invalid password?";
        return -1;
    }

    return 0;
}

int gorods_connect_env(rcComm_t** conn, char* host, int port, char* username, char* zone, char* password, char** err) {
    int status;

    rErrMsg_t errMsg;
    *conn = rcConnect(host, port, username, zone, 1, &errMsg);

    if ( !*conn ) {
        *err = "rcConnect failed";
        return -1;
    }
    
    if ( password != NULL ) {
    	status = clientLoginWithPassword(*conn, password);
    } else {
    	status = clientLogin(*conn);
    }
    
    if ( status != 0 ) {
        *err = "clientLogin failed. Invalid password?";
        return -1;
    }

    return 0;
}

int gorods_open_collection(char* path, int* handle, rcComm_t* conn, char** err) {
	collInp_t collOpenInp; 

	bzero(&collOpenInp, sizeof(collOpenInp)); 
	rstrcpy(collOpenInp.collName, path, MAX_NAME_LEN);

	collOpenInp.flags = VERY_LONG_METADATA_FG; 

	*handle = rcOpenCollection(conn, &collOpenInp); 
	if ( *handle < 0 ) { 
		*err = "rcOpenCollection failed";
		return -1;
	} 

	return 0;
}

int gorods_create_dataobject(char* path, rodsLong_t size, int mode, int force, char* resource, int* handle, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 
	
	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 

	dataObjInp.createMode = mode; 
	dataObjInp.dataSize = size; 

	if ( resource != NULL && resource[0] != '\0' ) {
		addKeyVal(&dataObjInp.condInput, DEST_RESC_NAME_KW, resource); 
	}

	if ( force > 0 ) {
		addKeyVal(&dataObjInp.condInput, FORCE_FLAG_KW, ""); 
	}
	
	*handle = rcDataObjCreate(conn, &dataObjInp); 
	if ( *handle < 0 ) { 
		*err = "rcDataObjCreate failed";
		return -1;
	}

	return 0;
}

int gorods_open_dataobject(char* path, int* handle, rodsLong_t length, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 
	
	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 
	
	dataObjInp.openFlags = O_RDWR; 
	
	*handle = rcDataObjOpen(conn, &dataObjInp); 
	if ( *handle < 0 ) { 
		*err = "rcDataObjOpen failed";
		return -1;
	}

	return 0;
}

int gorods_close_dataobject(int handleInx, rcComm_t* conn, char** err) {
	openedDataObjInp_t openedDataObjInp; 
	
	bzero(&openedDataObjInp, sizeof(openedDataObjInp)); 
	
	openedDataObjInp.l1descInx = handleInx; 
	
	int status;
	status = rcDataObjClose(conn, &openedDataObjInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjClose failed";
		return -1;
	}

	return 0;
}

int gorods_read_dataobject(int handleInx, rodsLong_t length, bytesBuf_t* buffer, rcComm_t* conn, char** err) {
	
	int bytesRead; 
	openedDataObjInp_t dataObjReadInp; 
	
	bzero(&dataObjReadInp, sizeof(dataObjReadInp)); 
	bzero(buffer, sizeof(*buffer)); 

	dataObjReadInp.l1descInx = handleInx; 
	dataObjReadInp.len = (int)length;

	bytesRead = rcDataObjRead(conn, &dataObjReadInp, buffer); 
	
	if ( bytesRead < 0 ) { 
		*err = "rcDataObjRead failed";
		return -1;
	}

	return 0;
}

int gorods_stat_dataobject(char* path, rodsObjStat_t** rodsObjStatOut, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 

	*rodsObjStatOut = NULL;

	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 
	
	// pass memory address of rodsObjStatOut pointer
	int status = rcObjStat(conn, &dataObjInp, &(*rodsObjStatOut)); 
	if ( status < 0 ) { 
		*err = "rcDataObjRead failed";
		return -1;
	}

	return 0;
}

int gorods_read_collection(rcComm_t* conn, int handleInx, collEnt_t** arr, int* size, char** err) {

	int collectionResponseCapacity = 100;
	*size = 0;

	*arr = malloc(sizeof(collEnt_t) * collectionResponseCapacity);
	
	collEnt_t* collEnt = NULL;
	int status;
	
	while ( (status = rcReadCollection(conn, handleInx, &collEnt)) >= 0 ) { 
		
			// Expand array if needed
			if ( *size >= collectionResponseCapacity ) {
				collectionResponseCapacity += 1;
				*arr = realloc(*arr, sizeof(collEnt_t) * collectionResponseCapacity);
			}

			collEnt_t* elem = &((*arr)[*size]);

			// Add element to array
    		memcpy(elem, collEnt, sizeof(collEnt_t));
			
			if ( collEnt->objType == DATA_OBJ_T ) { 
	    		elem->dataName = strcpy(malloc(strlen(elem->dataName) + 1), elem->dataName);
	    		elem->dataId = strcpy(malloc(strlen(elem->dataId) + 1), elem->dataId);
	    		elem->chksum = strcpy(malloc(strlen(elem->chksum) + 1), elem->chksum);
	    		elem->dataType = strcpy(malloc(strlen(elem->dataType) + 1), elem->dataType);
	    		elem->resource = strcpy(malloc(strlen(elem->resource) + 1), elem->resource);
   				elem->rescGrp = strcpy(malloc(strlen(elem->rescGrp) + 1), elem->rescGrp);
   				elem->phyPath = strcpy(malloc(strlen(elem->phyPath) + 1), elem->phyPath);
			}

			elem->ownerName = strcpy(malloc(strlen(elem->ownerName) + 1), elem->ownerName);
			elem->collName = strcpy(malloc(strlen(elem->collName) + 1), elem->collName);
			elem->createTime = strcpy(malloc(strlen(elem->createTime) + 1), elem->createTime);
	  		elem->modifyTime = strcpy(malloc(strlen(elem->modifyTime) + 1), elem->modifyTime);

    		(*size)++;
		
		freeCollEnt(collEnt); 
	} 

	rcCloseCollection(conn, handleInx); 

	return 0;
}

char* irods_env_str() {
    rodsEnv myEnv;
    int status = getRodsEnv( &myEnv );
    if ( status != 0 ) {
        return (char *)"error getting env";
    }

     char* str = malloc(sizeof(char) * 255);

     sprintf(str, "\tHost: %s\n\tPort: %i\n\tUsername: %s\n\tZone: %s\n", myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone);


     return str;
}

