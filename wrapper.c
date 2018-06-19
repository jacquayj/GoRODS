/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. and The BioTeam, Inc.  ***
 *** For more information please refer to the LICENSE.md file                                   ***/

#include "wrapper.h"

#define BIG_STR 3000

void* gorods_malloc(size_t size) {
	void* mem = malloc(size);
	if ( mem == NULL ) {
		printf("GoRods error: Unable to allocate %i bytes\n", (int)size);
		exit(1);
	}

	return mem;
}

int gorods_connect(rcComm_t** conn, char** host, int* port, char** username, char** zone, char** err) {
    rodsEnv myEnv;
    int status;

    status = getRodsEnv(&myEnv);
    if ( status != 0 ) {
        *err = "getRodsEnv failed";
        return status;
    }

    *host = &(myEnv.rodsHost[0]);
    *port = myEnv.rodsPort;
    *username = &(myEnv.rodsUserName[0]);
    *zone = &(myEnv.rodsZone[0]);

    rErrMsg_t errMsg;
    *conn = rcConnect(myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone, 1, &errMsg);

    if ( !*conn ) {
        *err = "rcConnect failed";
        return errMsg.status;
    }

    return 0;
}

int gorods_connect_env(rcComm_t** conn, char* host, int port, char* username, char* zone, char** err) {

    rErrMsg_t errMsg;
    *conn = rcConnect(host, port, username, zone, 1, &errMsg);

    if ( !*conn ) {
        *err = "rcConnect failed";
        return -1;
    }

    return 0;
}

void display_mallinfo(void) {
    struct mallinfo mi;

    mi = mallinfo();

    printf("Total (non-mmapped) allocated bytes (arena):       %d\n", mi.arena);
    // printf("# of free chunks (ordblks):            %d\n", mi.ordblks);
    // printf("# of free fastbin blocks (smblks):     %d\n", mi.smblks);
    // printf("# of mapped regions (hblks):           %d\n", mi.hblks);
    // printf("Bytes in mapped regions (hblkhd):      %d\n", mi.hblkhd);
    // printf("Max. total allocated space (usmblks):  %d\n", mi.usmblks);
    // printf("Free bytes held in fastbins (fsmblks): %d\n", mi.fsmblks);
    // printf("Total allocated space (uordblks):      %d\n", mi.uordblks);
    // printf("Total free space (fordblks):           %d\n", mi.fordblks);
    // printf("Topmost releasable block (keepcost):   %d\n", mi.keepcost);
}

int gorods_clientLoginPam(rcComm_t* conn, char* password, int ttl, char** pamPass, char** err) {

    pamAuthRequestInp_t pamAuthReqInp;
    pamAuthRequestOut_t *pamAuthReqOut = NULL;
    
    int status = 0;
    int doStty = 0;
    int len = 0;
    char myPassword[MAX_PASSWORD_LEN + 2];
    char userName[NAME_LEN * 2];

    strncpy(userName, conn->proxyUser.userName, NAME_LEN);
    
    if ( password[0] != '\0' ) {
        strncpy(myPassword, password, sizeof(myPassword));
    }

    len = strlen(myPassword);
    if ( myPassword[len - 1] == '\n' ) {
        myPassword[len - 1] = '\0'; /* remove trailing \n */
    }

    /* since PAM requires a plain text password to be sent
    to the server, ask the server to encrypt the current
    communication socket. */
    status = sslStart(conn);
    if ( status ) {
        *err = "sslStart error";
        return status;
    }

    memset(&pamAuthReqInp, 0, sizeof(pamAuthReqInp));

    pamAuthReqInp.pamPassword = myPassword;
    pamAuthReqInp.pamUser = userName;
    pamAuthReqInp.timeToLive = ttl;

    status = rcPamAuthRequest(conn, &pamAuthReqInp, &pamAuthReqOut);
    if ( status ) {
        *err = "rcPamAuthRequest error";
        sslEnd(conn);
        return status;
    }

    memset(myPassword, 0, sizeof(myPassword));

    /* can turn off SSL now. Have to request the server to do so.
    Will also ignore any error returns, as future socket ops
    are probably unaffected. */
    sslEnd(conn);


    *pamPass = strcpy(gorods_malloc(strlen(pamAuthReqOut->irodsPamPassword) + 1), pamAuthReqOut->irodsPamPassword);

    free(pamAuthReqOut->irodsPamPassword);
    free(pamAuthReqOut);
    
    return status;
}

int gorods_set_session_ticket(rcComm_t *myConn, char *ticket, char** err) {
    ticketAdminInp_t ticketAdminInp;
    int status;

    ticketAdminInp.arg1 = "session";
    ticketAdminInp.arg2 = ticket;
    ticketAdminInp.arg3 = "";
    ticketAdminInp.arg4 = "";
    ticketAdminInp.arg5 = "";
    ticketAdminInp.arg6 = "";

    status = rcTicketAdmin( myConn, &ticketAdminInp );

    if ( status != 0 ) {
        sprintf(*err, "set ticket error %d \n", status);
    }

    return status;
}

int gorods_iuserinfo(rcComm_t *myConn, char *name, userInfo_t* outInfo, char** err) {
    genQueryInp_t genQueryInp;
    genQueryOut_t *genQueryOut;
    int i1a[20];
    int i1b[20] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
    int i2a[20];
    char *condVal[10];
    char v1[BIG_STR];
    int i, j, status;

    char *columnNames[] = {"name", "id", "type", "zone", "info", "comment", "create time", "modify time"};

    memset(&genQueryInp, 0, sizeof(genQueryInp_t));
    memset(outInfo, 0, sizeof(userInfo_t));

    i = 0;
    i1a[i++] = COL_USER_NAME;
    i1a[i++] = COL_USER_ID;
    i1a[i++] = COL_USER_TYPE;
    i1a[i++] = COL_USER_ZONE;
    i1a[i++] = COL_USER_INFO;
    i1a[i++] = COL_USER_COMMENT;
    i1a[i++] = COL_USER_CREATE_TIME;
    i1a[i++] = COL_USER_MODIFY_TIME;
    genQueryInp.selectInp.inx = i1a;
    genQueryInp.selectInp.value = i1b;
    genQueryInp.selectInp.len = i;

    i2a[0] = COL_USER_NAME;
    snprintf(v1, BIG_STR, "='%s'", name);
    condVal[0] = v1;

    genQueryInp.sqlCondInp.inx = i2a;
    genQueryInp.sqlCondInp.value = condVal;
    genQueryInp.sqlCondInp.len = 1;

    genQueryInp.condInput.len = 0;

    genQueryInp.maxRows = 2;
    genQueryInp.continueInx = 0;

    status = rcGenQuery(myConn, &genQueryInp, &genQueryOut);
    if ( status == CAT_NO_ROWS_FOUND ) {
        freeGenQueryOut(&genQueryOut);

        i1a[0] = COL_USER_COMMENT;
        genQueryInp.selectInp.len = 1;

        status = rcGenQuery(myConn, &genQueryInp, &genQueryOut);
        if ( status == 0 ) {
            *err = "None";
            freeGenQueryOut(&genQueryOut);
            return -1;
        }
        if ( status == CAT_NO_ROWS_FOUND ) {
            *err = "User does not exist";
            freeGenQueryOut(&genQueryOut);
            return status;
        }
    }

    if ( status != 0 ) {
        *err = "Error in rcGenQuery";
        return status;
    }


    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            char *tResult;
            tResult = genQueryOut->sqlResult[j].value;
            tResult += i * genQueryOut->sqlResult[j].len;

            if ( strcmp( columnNames[j], "name" ) == 0 ) {
                strcpy(outInfo->userName, tResult);
            } else if ( strcmp( columnNames[j], "id" ) == 0 ) {
                // Convert to string
            } else if ( strcmp( columnNames[j], "type" ) == 0 ) {
                strcpy(outInfo->userType, tResult);
            } else if ( strcmp( columnNames[j], "zone" ) == 0 ) {
                strcpy(outInfo->rodsZone, tResult);
            } else if ( strcmp( columnNames[j], "info" ) == 0 ) {
                strcpy(outInfo->userOtherInfo.userInfo, tResult);
            } else if ( strcmp( columnNames[j], "comment" ) == 0 ) {
                strcpy(outInfo->userOtherInfo.userComments, tResult);
            } else if ( strcmp( columnNames[j], "create time" ) == 0 ) {
                strcpy(outInfo->userOtherInfo.userCreate, tResult);
            } else if ( strcmp( columnNames[j], "modify time" ) == 0 ) {
                strcpy(outInfo->userOtherInfo.userModify, tResult);
            }
        }
    }

    freeGenQueryOut(&genQueryOut);
    
    return status;
}


int gorods_phys_path_reg(rcComm_t* ccon, char* physPath, char* rodsPath, int force, int collection, int replica, char* resourceName, char* excludeFiles) {

    dataObjInp_t dataObjOprInp;
    memset(&dataObjOprInp, 0, sizeof(dataObjInp_t));

    addKeyVal(&dataObjOprInp.condInput, DATA_TYPE_KW, "generic");

    if ( force > 0 ) {
        addKeyVal(&dataObjOprInp.condInput, FORCE_FLAG_KW, "");
    }

    if ( collection > 0 ) {
        addKeyVal(&dataObjOprInp.condInput, COLLECTION_KW, "");
    }

    if ( replica > 0 ) {
        addKeyVal(&dataObjOprInp.condInput, REG_REPL_KW, "");
    }

    if ( excludeFiles != NULL && excludeFiles[0] != '\0' ) {
        addKeyVal(&dataObjOprInp.condInput, EXCLUDE_FILE_KW, excludeFiles);
    }

    if ( resourceName != NULL && resourceName[0] != '\0' ) {
        addKeyVal(&dataObjOprInp.condInput, DEST_RESC_NAME_KW, resourceName);
    }

    addKeyVal(&dataObjOprInp.condInput, FILE_PATH_KW, physPath);
    rstrcpy(dataObjOprInp.objPath, rodsPath, MAX_NAME_LEN);


    return rcPhyPathReg(ccon, &dataObjOprInp);
}


int gorods_open_collection(char* path, int trimRepls, collHandle_t* collHandle, rcComm_t* conn, char** err) {

    int flag;
    int status;

	bzero(collHandle, sizeof(collHandle_t)); 

	flag = VERY_LONG_METADATA_FG;

    if ( trimRepls == 0 ) {
        flag |= NO_TRIM_REPL_FG;;
    }

	status = rclOpenCollection(conn, path, flag, collHandle); 
	if ( status < 0 ) { 
		*err = "rcOpenCollection failed";
		return status;
	} 

	return 0;
}


int gorods_put_dataobject(char* inPath, char* outPath, rodsLong_t size, int mode, int force, char* resource, rcComm_t* conn, char** err) {
    
    int status;
    dataObjInp_t dataObjInp;
    char locFilePath[MAX_NAME_LEN];
    bzero(&dataObjInp, sizeof(dataObjInp)); 

    rstrcpy(dataObjInp.objPath, outPath, MAX_NAME_LEN); 
    rstrcpy(locFilePath, inPath, MAX_NAME_LEN);

    dataObjInp.createMode = mode;
    dataObjInp.dataSize = size;
    dataObjInp.numThreads = conn->transStat.numThreads;

    if ( resource != NULL && resource[0] != '\0' ) {
        addKeyVal(&dataObjInp.condInput, DEST_RESC_NAME_KW, resource); 
    }

    if ( force > 0 ) {
        addKeyVal(&dataObjInp.condInput, FORCE_FLAG_KW, ""); 
    }

    status = rcDataObjPut(conn, &dataObjInp, locFilePath); 
    if ( status < 0 ) { 
        *err = "rcDataObjPut failed";
    }

    return status;
}

int gorods_write_dataobject(int handle, void* data, int size, rcComm_t* conn, char** err) {
	
	openedDataObjInp_t dataObjWriteInp; 
	bytesBuf_t dataObjWriteOutBBuf; 

	bzero(&dataObjWriteInp, sizeof(dataObjWriteInp)); 
	bzero(&dataObjWriteOutBBuf, sizeof(dataObjWriteOutBBuf)); 

	dataObjWriteInp.l1descInx = handle;
	
	dataObjWriteOutBBuf.len = size; 
	dataObjWriteOutBBuf.buf = data; 
	
	int bytesWrite = rcDataObjWrite(conn, &dataObjWriteInp, &dataObjWriteOutBBuf); 
	if ( bytesWrite < 0 ) { 
		*err = "rcDataObjWrite failed";
		return bytesWrite;
	}

	return 0;
}

int gorods_create_dataobject(char* path, rodsLong_t size, int mode, int force, char* resource, int* handle, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 
	
	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 

    if ( mode == 0 ) {
        mode = 0750;
    }

	dataObjInp.createMode = mode; 
	dataObjInp.dataSize = size; 
    dataObjInp.numThreads = conn->transStat.numThreads;

	if ( resource != NULL && resource[0] != '\0' ) {
		addKeyVal(&dataObjInp.condInput, DEST_RESC_NAME_KW, resource); 
	}

	if ( force > 0 ) {
		addKeyVal(&dataObjInp.condInput, FORCE_FLAG_KW, ""); 
	}
	
	*handle = rcDataObjCreate(conn, &dataObjInp); 
	if ( *handle < 0 ) { 
		*err = "rcDataObjCreate failed";
		return *handle;
	}

	return 0;
}

int gorods_create_collection(char* path, rcComm_t* conn, char** err) {
	int status;

	collInp_t collCreateInp; 

	bzero(&collCreateInp, sizeof(collCreateInp));

	rstrcpy(collCreateInp.collName, path, MAX_NAME_LEN); 

	addKeyVal(&collCreateInp.condInput, RECURSIVE_OPR__KW, "");
	
	status = rcCollCreate(conn, &collCreateInp);
	if ( status < 0 ) { 
		*err = "rcCollCreate failed";
		return status;
	}

	return 0;
}

int gorods_open_dataobject(char* path, char* resourceName, char* replNum, int openFlag, int* handle, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 
	
	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 
	
	// O_RDONLY, O_WRONLY, O_RDWR, O_TRUNC
	dataObjInp.openFlags = openFlag; 
	dataObjInp.numThreads = conn->transStat.numThreads;

    addKeyVal(&dataObjInp.condInput, RESC_NAME_KW, resourceName); 
    addKeyVal(&dataObjInp.condInput, REPL_NUM_KW, replNum);

	int thehandle = rcDataObjOpen(conn, &dataObjInp); 
	if ( thehandle <= 0 ) { 
		*err = "rcDataObjOpen failed";
		return -1;
	}

    *handle = thehandle;

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
		return status;
	}

	return 0;
}

int gorods_close_collection(collHandle_t* collHandle, char** err) {
	int status = rclCloseCollection(collHandle);

	if ( status < 0 ) { 
		*err = "rcCloseCollection failed";
		return status;
	}

	return 0;
}

int gorods_get_local_zone(rcComm_t* conn, char** zoneName, char** err) {
    int status, i;
    simpleQueryInp_t simpleQueryInp;
    simpleQueryOut_t *simpleQueryOut;
    
    memset(&simpleQueryInp, 0, sizeof(simpleQueryInp_t));

    simpleQueryInp.form = 1;
    simpleQueryInp.sql = "select zone_name from R_ZONE_MAIN where zone_type_name=?";
    simpleQueryInp.arg1 = "local";
    simpleQueryInp.maxBufSize = 1024;

    status = rcSimpleQuery(conn, &simpleQueryInp, &simpleQueryOut);
    if ( status < 0 ) {
        *err = "Error getting local zone";
        return status;
    }

    *zoneName = strcpy(gorods_malloc(strlen(simpleQueryOut->outBuf) + 1), simpleQueryOut->outBuf);

    i = strlen(*zoneName);
    
    for ( ; i > 1; i-- ) {
        if ( *zoneName[i] == '\n' ) {
            *zoneName[i] = '\0';
            if ( *zoneName[i - 1] == ' ' ) {
                *zoneName[i - 1] = '\0';
            }
            break;
        }
    }

    return 0;
}

int gorods_get_zones(rcComm_t* conn, goRodsStringResult_t* result, char** err) {
    
    simpleQueryInp_t simpleQueryInp;
    memset(&simpleQueryInp, 0, sizeof(simpleQueryInp_t));

    simpleQueryInp.control = 0;

    simpleQueryInp.form = 1;
    simpleQueryInp.sql = "select zone_name from R_ZONE_MAIN";
    simpleQueryInp.maxBufSize = 1024;

    return gorods_simple_query(simpleQueryInp, result, conn, err);
}

int gorods_get_zone(char* zoneName, rcComm_t* conn, goRodsStringResult_t* result, char** err) {

    simpleQueryInp_t simpleQueryInp;
    memset(&simpleQueryInp, 0, sizeof(simpleQueryInp_t));

    simpleQueryInp.control = 0;
    simpleQueryInp.form = 2;
    simpleQueryInp.sql = "select * from R_ZONE_MAIN where zone_name=?";
    simpleQueryInp.arg1 = zoneName;
    simpleQueryInp.maxBufSize = 1024;

    return gorods_simple_query(simpleQueryInp, result, conn, err);
}

int gorods_get_resources(rcComm_t* conn, goRodsStringResult_t* result, char** err) {
    simpleQueryInp_t simpleQueryInp;

    memset(&simpleQueryInp, 0, sizeof(simpleQueryInp_t));
    simpleQueryInp.control = 0;

    simpleQueryInp.form = 1;
    simpleQueryInp.sql = "select resc_name from R_RESC_MAIN";
    simpleQueryInp.maxBufSize = 1024;
    
    return gorods_simple_query(simpleQueryInp, result, conn, err);
}

int gorods_get_resource(char* rescName, rcComm_t* conn, goRodsStringResult_t* result, char** err) {
    simpleQueryInp_t simpleQueryInp;

    memset( &simpleQueryInp, 0, sizeof( simpleQueryInp_t ) );
    simpleQueryInp.control = 0;

    simpleQueryInp.form = 2;
    simpleQueryInp.sql = "select * from R_RESC_MAIN where resc_name=?";
    simpleQueryInp.arg1 = rescName;
    simpleQueryInp.maxBufSize = 1024;
    
   return gorods_simple_query(simpleQueryInp, result, conn, err);
}


int gorods_get_user(char *user, rcComm_t* conn, goRodsStringResult_t* result, char** err) {
    simpleQueryInp_t simpleQueryInp;

    memset(&simpleQueryInp, 0, sizeof(simpleQueryInp_t));
    simpleQueryInp.control = 0;
    
    simpleQueryInp.form = 2;
    simpleQueryInp.sql = "select * from R_USER_MAIN where user_name=?";
    simpleQueryInp.arg1 = user;
    simpleQueryInp.maxBufSize = 1024;
    
    return gorods_simple_query(simpleQueryInp, result, conn, err);
}

int gorods_get_users(rcComm_t* conn, goRodsStringResult_t* result, char** err) {
    simpleQueryInp_t simpleQueryInp;

    memset(&simpleQueryInp, 0, sizeof(simpleQueryInp_t));
    simpleQueryInp.control = 0;
   
    simpleQueryInp.form = 1;
    simpleQueryInp.sql = "select user_name||'#'||zone_name from R_USER_MAIN where user_type_name != 'rodsgroup'";
    simpleQueryInp.maxBufSize = 1024;
    
    return gorods_simple_query(simpleQueryInp, result, conn, err);
}


int gorods_simple_query(simpleQueryInp_t simpleQueryInp, goRodsStringResult_t* result, rcComm_t* conn, char** err) {
   
    int status;
    simpleQueryOut_t *simpleQueryOut;
    status = rcSimpleQuery(conn, &simpleQueryInp, &simpleQueryOut);

    if ( status == CAT_NO_ROWS_FOUND ) {
        *err = "No rows found";
        //free(simpleQueryOut->outBuf);
        //free(simpleQueryOut);
        return status;
    }

    if ( status == SYS_NO_API_PRIV ) {
        *err = "rcSimpleQuery permission denied";
        //free(simpleQueryOut->outBuf);
        //free(simpleQueryOut);
        return status;
    }

    if ( status < 0 ) {
        *err = "rcSimpleQuery failed with error";
        //free(simpleQueryOut->outBuf);
        //free(simpleQueryOut);
        return status;
    }

    result->size++;
    result->strArr = gorods_malloc(result->size * sizeof(char*));

    result->strArr[0] = strcpy(gorods_malloc(strlen(simpleQueryOut->outBuf) + 1), simpleQueryOut->outBuf);
    free(simpleQueryOut->outBuf);


    if ( simpleQueryOut->control > 0 ) {

        simpleQueryInp.control = simpleQueryOut->control;

        for ( ; simpleQueryOut->control > 0 && status == 0; ) {
            status = rcSimpleQuery(conn, &simpleQueryInp, &simpleQueryOut);
            
            if ( status < 0 && status != CAT_NO_ROWS_FOUND ) {
                *err = "rcSimpleQuery failed with error";
                free(simpleQueryOut->outBuf);
                free(simpleQueryOut);
                return status;
            }

            if ( status == 0 ) {

            	int sz = result->size;

            	result->size++;
    			result->strArr = realloc(result->strArr, result->size * sizeof(char*));

    			result->strArr[sz] = strcpy(gorods_malloc(strlen(simpleQueryOut->outBuf) + 1), simpleQueryOut->outBuf);
                free(simpleQueryOut->outBuf);
            }
        }
    }

    free(simpleQueryOut);

    return status;
}


int gorods_get_user_groups(rcComm_t *conn, char* name, goRodsStringResult_t* result, char** err) {
    genQueryInp_t genQueryInp;
    genQueryOut_t *genQueryOut;

    int i1a[20];
    int i1b[20] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
    int i2a[20];
    char *condVal[10];
    char v1[BIG_STR];
    int i, status;

    char *columnNames[] = {"group"};

    memset(&genQueryInp, 0, sizeof(genQueryInp_t));

    i = 0;

    i1a[i++] = COL_USER_NAME;
    i1a[i++] = COL_USER_ID;
    i1a[i++] = COL_USER_TYPE;
    i1a[i++] = COL_USER_ZONE;
    i1a[i++] = COL_USER_INFO;
    i1a[i++] = COL_USER_COMMENT;
    i1a[i++] = COL_USER_CREATE_TIME;
    i1a[i++] = COL_USER_MODIFY_TIME;

    i1a[0] = COL_USER_GROUP_NAME;

    genQueryInp.selectInp.inx = i1a;
    genQueryInp.selectInp.value = i1b;
    genQueryInp.selectInp.len = 1;

    i2a[0] = COL_USER_NAME;

    snprintf(v1, BIG_STR, "='%s'", name);
    condVal[0] = v1;

    genQueryInp.sqlCondInp.inx = i2a;
    genQueryInp.sqlCondInp.value = condVal;
    genQueryInp.sqlCondInp.len = 1;
    genQueryInp.condInput.len = 0;
    genQueryInp.maxRows = 50;
    genQueryInp.continueInx = 0;
    
    int cont;

    status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

    if ( status == CAT_NO_ROWS_FOUND ) {
        *err = "Not a member of any group";
        freeGenQueryOut(&genQueryOut);
        return CAT_NO_ROWS_FOUND;
    }

    if ( status != 0 ) {
        *err = "rcGenQuery error";
        freeGenQueryOut(&genQueryOut);
        return status;
    }

    gorods_get_user_group_result(status, result, genQueryOut, columnNames);
    freeGenQueryOut(&genQueryOut);

    while ( status == 0 && cont > 0 ) {
        genQueryInp.continueInx = cont;
        
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;
       
        gorods_get_user_group_result(status, result, genQueryOut, columnNames);
        freeGenQueryOut(&genQueryOut);
    }

    return 0;
}

int gorods_get_user_group_result(int status, goRodsStringResult_t* result, genQueryOut_t *genQueryOut, char *descriptions[]) {
    
    int i, j;

    if ( result->size == 0 ) {
        result->size = genQueryOut->rowCnt;
        result->strArr = gorods_malloc(result->size * sizeof(char*));
    } else {
        result->size += genQueryOut->rowCnt;
        result->strArr = realloc(result->strArr, result->size * sizeof(char*));
    }

    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            
            char *tResult;
           
            tResult = genQueryOut->sqlResult[j].value;
            tResult += i * genQueryOut->sqlResult[j].len;
            
            result->strArr[i] = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);

        }
    }

    return 0;
}


int gorods_get_groups(rcComm_t *conn, goRodsStringResult_t* result, char** err) {
    genQueryInp_t  genQueryInp;
    genQueryOut_t *genQueryOut = 0;
    int selectIndexes[10];
    int selectValues[10];
    int conditionIndexes[10];
    char *conditionValues[10];
    char conditionString1[BIG_STR];
    char conditionString2[BIG_STR];
    int status;
    memset(&genQueryInp, 0, sizeof(genQueryInp_t));

    // if ( groupName != NULL && *groupName != '\0' ) {
    //     printf( "Members of group %s:\n", groupName );
    // }

    selectIndexes[0] = COL_USER_NAME;
    selectValues[0] = 0;
    selectIndexes[1] = COL_USER_ZONE;
    selectValues[1] = 0;
    genQueryInp.selectInp.inx = selectIndexes;
    genQueryInp.selectInp.value = selectValues;
   
    // if ( groupName != NULL && *groupName != '\0' ) {
    //     genQueryInp.selectInp.len = 2;
    // }
    // else {
        genQueryInp.selectInp.len = 1;
    // }

    conditionIndexes[0] = COL_USER_TYPE;
    sprintf(conditionString1, "='rodsgroup'");
    conditionValues[0] = conditionString1;

    genQueryInp.sqlCondInp.inx = conditionIndexes;
    genQueryInp.sqlCondInp.value = conditionValues;
    genQueryInp.sqlCondInp.len = 1;

    // if ( groupName != NULL && *groupName != '\0' ) {

    //     sprintf( conditionString1, "!='rodsgroup'" );

    //     conditionIndexes[1] = COL_USER_GROUP_NAME;
    //     sprintf( conditionString2, "='%s'", groupName );
    //     conditionValues[1] = conditionString2;
    //     genQueryInp.sqlCondInp.len = 2;
    // }

    genQueryInp.maxRows = 50;
    genQueryInp.continueInx = 0;
    genQueryInp.condInput.len = 0;

    int cont;

    status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    if ( status == CAT_NO_ROWS_FOUND ) {
        freeGenQueryOut(&genQueryOut);
        *err = "No rows found";
        return status;
    } 

    cont = genQueryOut->continueInx;

    gorods_build_group_result(genQueryOut, result);
    
    freeGenQueryOut(&genQueryOut);

    while ( status == 0 && cont > 0 ) {
        genQueryInp.continueInx = cont;
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;
        if ( status == 0 ) {
            gorods_build_group_result(genQueryOut, result);
        }
        freeGenQueryOut(&genQueryOut);
    }

    

    return 0;
}

int gorods_get_group(rcComm_t *conn, goRodsStringResult_t* result, char* groupName, char** err) {
    genQueryInp_t  genQueryInp;
    genQueryOut_t *genQueryOut = 0;
    int selectIndexes[10];
    int selectValues[10];
    int conditionIndexes[10];
    char *conditionValues[10];
    char conditionString1[BIG_STR];
    char conditionString2[BIG_STR];
    int status;
    memset(&genQueryInp, 0, sizeof(genQueryInp_t));

    selectIndexes[0] = COL_USER_NAME;
    selectValues[0] = 0;
    selectIndexes[1] = COL_USER_ZONE;
    selectValues[1] = 0;
    genQueryInp.selectInp.inx = selectIndexes;
    genQueryInp.selectInp.value = selectValues;
   
   	genQueryInp.selectInp.len = 2;
 
    conditionIndexes[0] = COL_USER_TYPE;
    sprintf(conditionString1, "='rodsgroup'");
    conditionValues[0] = conditionString1;

    genQueryInp.sqlCondInp.inx = conditionIndexes;
    genQueryInp.sqlCondInp.value = conditionValues;
    genQueryInp.sqlCondInp.len = 1;

    sprintf(conditionString1, "!='rodsgroup'");

    conditionIndexes[1] = COL_USER_GROUP_NAME;
    sprintf(conditionString2, "='%s'", groupName);
    conditionValues[1] = conditionString2;
    genQueryInp.sqlCondInp.len = 2;

    genQueryInp.maxRows = 50;
    genQueryInp.continueInx = 0;
    genQueryInp.condInput.len = 0;

    int cont;
    status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

    if ( status == CAT_NO_ROWS_FOUND ) {
        freeGenQueryOut(&genQueryOut);
        *err = "No rows found";
        return status;
    } 

    gorods_build_group_user_result(genQueryOut, result);

    freeGenQueryOut(&genQueryOut);

    while ( status == 0 && cont > 0 ) {
        genQueryInp.continueInx = cont;
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;
        if ( status == 0 ) {
            gorods_build_group_user_result(genQueryOut, result);
        }
        
        freeGenQueryOut(&genQueryOut);
    }

    return 0;
}

int gorods_add_user_to_group(char* userName, char* zoneName, char* groupName, rcComm_t *conn, char** err) {
	int status;

	status = gorods_general_admin(1, "modify", "group", groupName, "add", userName, zoneName, "", "", "", "", 0, conn, err);


	return status;
}

int gorods_remove_user_from_group(char* userName, char* zoneName, char* groupName, rcComm_t *conn, char** err) {
    int status;

    status = gorods_general_admin(1, "modify", "group", groupName, "remove", userName, zoneName, "", "", "", "", 0, conn, err);


    return status;
}

int gorods_delete_group(char* groupName, char* zoneName, rcComm_t *conn, char** err) {
    int status;

    status = gorods_general_admin(0, "rm", "user", groupName,
        zoneName, "", "", "", "", "", "", 0, conn, err);

    // generalAdmin( 0, "rm", "user", cmdToken[1],
    //                  myEnv.rodsZone, "", "", "", "", "", "" );

    return status;
}

int gorods_create_group(char* groupName, char* zoneName, rcComm_t *conn, char** err) {
    int status;

    status = gorods_general_admin(0, "add", "user", groupName, "rodsgroup",
        zoneName, "", "", "", "", "", 0, conn, err);

    // generalAdmin( 0, "add", "user", cmdToken[1], "rodsgroup",
    //                   myEnv.rodsZone, "", "", "", "", "" );

    return status;
}

int gorods_delete_user(char* userName, char* zoneName, rcComm_t *conn, char** err) {
    int status;

    status = gorods_general_admin(0, "rm", "user", userName,
        zoneName, "", "", "", "", "", "", 0, conn, err);

    // generalAdmin( 0, "rm", "user", cmdToken[1],
    //                  myEnv.rodsZone, "", "", "", "", "", "" );

    return status;
}

int gorods_create_user(char* userName, char* zoneName, char* type, rcComm_t *conn, char** err) {
    int status;

    status = gorods_general_admin(0, "add", "user", userName, type,
        zoneName, "", "", "", "", "", 0, conn, err);

    // generalAdmin( 0, "add", "user", cmdToken[1], "rodsgroup",
    //                   myEnv.rodsZone, "", "", "", "", "" );

    return status;
}

int gorods_change_user_password(char* userName, char* newPassword, char* myPassword, rcComm_t *conn, char** err) {

    char buf0[MAX_PASSWORD_LEN + 10];
    char buf1[MAX_PASSWORD_LEN + 10];
    char buf2[MAX_PASSWORD_LEN + 100];

    int i, len, lcopy;
    char *key2;
    /* this is a random string used to pad, arbitrary, but must match
       the server side: */
    char rand[] = "1gCBizHWbwIYyWLoysGzTe6SyzqFKMniZX05faZHWAwQKXf6Fs";

    strncpy(buf0, newPassword, MAX_PASSWORD_LEN);
    len = strlen(newPassword);
    lcopy = MAX_PASSWORD_LEN - 10 - len;
    
    if ( lcopy > 15 ) { /* server will look for 15 characters of random string */
        strncat(buf0, rand, lcopy);
    }

    strncpy(buf1, myPassword, MAX_PASSWORD_LEN);

    key2 = getSessionSignatureClientside();
    obfEncodeByKeyV2(buf0, buf1, key2, buf2);
    newPassword = buf2;
        
    return gorods_general_admin(0, "modify", "user", userName, "password", newPassword, "", "", "", "", "", 0, conn, err);

}

int gorods_general_admin(int userOption, char *arg0, char *arg1, char *arg2, char *arg3,
              char *arg4, char *arg5, char *arg6, char *arg7, char* arg8, char* arg9,
              rodsArguments_t* _rodsArgs, rcComm_t *conn, char** err) {
    /* If userOption is 1, try userAdmin if generalAdmin gets a permission
     * failure */

	//_rodsArgs = 0;

    generalAdminInp_t generalAdminInp;
    userAdminInp_t userAdminInp;
    int status;
    char *funcName;

    if ( _rodsArgs && _rodsArgs->dryrun ) {
        arg3 = "--dryrun";
    }

    generalAdminInp.arg0 = arg0;
    generalAdminInp.arg1 = arg1;
    generalAdminInp.arg2 = arg2;
    generalAdminInp.arg3 = arg3;
    generalAdminInp.arg4 = arg4;
    generalAdminInp.arg5 = arg5;
    generalAdminInp.arg6 = arg6;
    generalAdminInp.arg7 = arg7;
    generalAdminInp.arg8 = arg8;
    generalAdminInp.arg9 = arg9;

    status = rcGeneralAdmin(conn, &generalAdminInp);

    funcName = "rcGeneralAdmin";

    if ( userOption == 1 && status == SYS_NO_API_PRIV ) {
        userAdminInp.arg0 = arg0;
        userAdminInp.arg1 = arg1;
        userAdminInp.arg2 = arg2;
        userAdminInp.arg3 = arg3;
        userAdminInp.arg4 = arg4;
        userAdminInp.arg5 = arg5;
        userAdminInp.arg6 = arg6;
        userAdminInp.arg7 = arg7;
        userAdminInp.arg8 = arg8;
        userAdminInp.arg9 = arg9;
        status = rcUserAdmin(conn, &userAdminInp);
        funcName = "rcGeneralAdmin and rcUserAdmin";
    }

    // =-=-=-=-=-=-=-
    // JMC :: for 'dryrun' option on rmresc we need to capture the
    //     :: return value and simply output either SUCCESS or FAILURE
    // rm resource dryrun BOOYA
    if ( _rodsArgs &&
            _rodsArgs->dryrun == 1 &&
            0 == strcmp( arg0, "rm" ) &&
            0 == strcmp( arg1, "resource" ) ) {
        if ( 0 == status ) {
            printf( "DRYRUN REMOVING RESOURCE [%s - %d] :: SUCCESS\n", arg2, status );
        }
        else {
            printf( "DRYRUN REMOVING RESOURCE [%s - %d] :: FAILURE\n", arg2, status );
        } // else
    }
    else if ( status == USER_INVALID_USERNAME_FORMAT ) {
        *err = "Invalid username format";
    }
    else if ( status < 0 && status != CAT_SUCCESS_BUT_WITH_NO_INFO ) {
       // char *mySubName = NULL;
        //const char *myName = rodsErrorName(status, &mySubName);
        //rodsLog( LOG_ERROR, "%s failed with error %d %s %s", funcName, status, myName, mySubName );
        
    	// Need to change error msg depending on the args, since this is a general purpose func
        *err = "General failure";

        if ( status == CAT_INVALID_USER_TYPE ) {

        	*err = "Invalid user type specified";
            //fprintf( stderr, "See 'lt user_type' for a list of valid user types.\n" );
        }
        
        //	free(mySubName);

        
    } // else if status < 0

    //printErrorStack( Conn->rError );
    //freeRErrorContent( Conn->rError );

    return status;
}


void gorods_build_group_user_result(genQueryOut_t *genQueryOut, goRodsStringResult_t* result) {
    
	if ( result->size == 0 ) {
    	result->size = genQueryOut->rowCnt;
		result->strArr = gorods_malloc(result->size * sizeof(char*));
    } else {
    	result->size += genQueryOut->rowCnt;
    	result->strArr = realloc(result->strArr, result->size * sizeof(char*));
    }

    int i, j;
    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        char *tResult;
        
        char resultStr[255] = "";

        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            tResult = genQueryOut->sqlResult[j].value;
            tResult += i * genQueryOut->sqlResult[j].len;
            
            if ( j > 0 ) {
            	strcat(&resultStr[0], "#");
            	strcat(&resultStr[0], tResult);
            } else {
               	strcat(&resultStr[0], tResult);
            }
        }

        result->strArr[i] = strcpy(gorods_malloc(strlen(resultStr) + 1), resultStr);
    }
}


void gorods_free_string_result(goRodsStringResult_t* result) {
	int i;
	for ( i = 0; i < result->size; i++ ) {
		free(result->strArr[i]);
	}
	free(result->strArr);
}

void gorods_build_group_result(genQueryOut_t *genQueryOut, goRodsStringResult_t* result) {
    
    if ( result->size == 0 ) {
    	result->size = genQueryOut->rowCnt;
		result->strArr = gorods_malloc(result->size * sizeof(char*));
    } else {
    	result->size += genQueryOut->rowCnt;
    	result->strArr = realloc(result->strArr, result->size * sizeof(char*));
    }

    int i, j;
    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        char *tResult;
        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            tResult = genQueryOut->sqlResult[j].value;
            tResult += i * genQueryOut->sqlResult[j].len;
        }
        result->strArr[i] = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
    }
}


int gorods_get_dataobject(rcComm_t *conn, char *srcPath, collEnt_t* objData) {
    int status;
    genQueryOut_t *genQueryOut = NULL;
    char myColl[MAX_NAME_LEN], myData[MAX_NAME_LEN];
    char condStr[MAX_NAME_LEN];
    int queryFlags;

    genQueryInp_t aGenQuery;
    initCondForLs(&aGenQuery);
    genQueryInp_t* genQueryInp = &aGenQuery;

    rodsArguments_t args;
    bzero(&args, sizeof(rodsArguments_t));

    rodsArguments_t* rodsArgs = &args;

    rodsArgs->veryLongOption = 1;
    rodsArgs->longOption = 0;
    rodsArgs->accessControl = 0;

    queryFlags = setQueryFlag( rodsArgs );
    setQueryInpForData( queryFlags, genQueryInp );
    genQueryInp->maxRows = MAX_SQL_ROWS;

    memset( myColl, 0, MAX_NAME_LEN );
    memset( myData, 0, MAX_NAME_LEN );

    if ( ( status = splitPathByKey(
                        srcPath, myColl, MAX_NAME_LEN, myData, MAX_NAME_LEN, '/' ) ) < 0 ) {
        rodsLogError( LOG_ERROR, status,
                      "setQueryInpForLong: splitPathByKey for %s error, status = %d",
                      srcPath, status );
        return status;
    }

    snprintf( condStr, MAX_NAME_LEN, "='%s'", myColl );
    addInxVal( &genQueryInp->sqlCondInp, COL_COLL_NAME, condStr );
    snprintf( condStr, MAX_NAME_LEN, "='%s'", myData );
    addInxVal( &genQueryInp->sqlCondInp, COL_DATA_NAME, condStr );

    status = rcGenQuery( conn, genQueryInp, &genQueryOut );

    if ( status < 0 ) {
        if ( status == CAT_NO_ROWS_FOUND ) {
            rodsLog( LOG_ERROR, "%s does not exist or user lacks access permission",
                     srcPath );
        }
        else {
            rodsLogError( LOG_ERROR, status,
                          "gorods_get_dataobject: rcGenQuery error for %s", srcPath );
        }
        return status;
    }

    return gorods_get_dataobject_data(conn, rodsArgs, genQueryOut, objData);
}

int gorods_get_dataobject_data(rcComm_t *conn, rodsArguments_t *rodsArgs, genQueryOut_t *genQueryOut, collEnt_t* objData) {
    int i = 0;
    sqlResult_t *dataName = 0, *colName = 0, *replNum = 0, *dataSize = 0, *rescName = 0,
                 *replStatus = 0, *dataModify = 0, *dataCreate = 0, *dataOwnerName = 0, *dataId = 0;
    sqlResult_t *chksumStr = 0, *dataPath = 0, *dataType = 0,*rescHier;

    char *tmpDataId = 0;
    int queryFlags = 0;

    if ( genQueryOut == NULL ) {
        return USER__NULL_INPUT_ERR;
    }

    queryFlags = setQueryFlag( rodsArgs );

    if ( rodsArgs->veryLongOption == True ) {
        if ( ( chksumStr = getSqlResultByInx( genQueryOut, COL_D_DATA_CHECKSUM ) )
                == NULL ) {
            rodsLog( LOG_ERROR,
                     "printLsLong: getSqlResultByInx for COL_D_DATA_CHECKSUM failed" );
            return UNMATCHED_KEY_OR_INDEX;
        }

        if ( ( dataPath = getSqlResultByInx( genQueryOut, COL_D_DATA_PATH ) )
                == NULL ) {
            rodsLog( LOG_ERROR,
                     "printLsLong: getSqlResultByInx for COL_D_DATA_PATH failed" );
            return UNMATCHED_KEY_OR_INDEX;
        }

        if ( ( dataType = getSqlResultByInx( genQueryOut, COL_DATA_TYPE_NAME ) ) == NULL ) {

            rodsLog( LOG_ERROR,
                     "printLsLong: getSqlResultByInx for COL_DATA_TYPE_NAME failed" );
            return UNMATCHED_KEY_OR_INDEX;
        }
    }

    if ( ( dataId = getSqlResultByInx( genQueryOut, COL_D_DATA_ID ) ) == NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_D_DATA_ID failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataName = getSqlResultByInx( genQueryOut, COL_DATA_NAME ) ) == NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_DATA_NAME failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( colName = getSqlResultByInx( genQueryOut, COL_COLL_NAME ) ) == NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_COLL_NAME failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( replNum = getSqlResultByInx( genQueryOut, COL_DATA_REPL_NUM ) ) ==
            NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_DATA_REPL_NUM failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataSize = getSqlResultByInx( genQueryOut, COL_DATA_SIZE ) ) == NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_DATA_SIZE failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( rescName = getSqlResultByInx( genQueryOut, COL_D_RESC_NAME ) ) == NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_D_RESC_NAME failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( rescHier = getSqlResultByInx( genQueryOut, COL_D_RESC_HIER ) ) == NULL ) {
        // If the index is not found then COL_D_RESC_HIER was most likely stripped
        // from the query input to talk to an older zone.
        // use resource name instead
        rescHier = rescName;
    }

    if ( ( replStatus = getSqlResultByInx( genQueryOut, COL_D_REPL_STATUS ) ) ==
            NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_D_REPL_STATUS failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataModify = getSqlResultByInx( genQueryOut, COL_D_MODIFY_TIME ) ) ==
            NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_D_MODIFY_TIME failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataCreate = getSqlResultByInx( genQueryOut, COL_D_CREATE_TIME ) ) ==
            NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_D_CREATE_TIME failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataOwnerName = getSqlResultByInx( genQueryOut, COL_D_OWNER_NAME ) ) ==
            NULL ) {
        rodsLog( LOG_ERROR,
                 "printLsLong: getSqlResultByInx for COL_D_OWNER_NAME failed" );
        return UNMATCHED_KEY_OR_INDEX;
    }

    queryHandle_t query_handle;
    rclInitQueryHandle( &query_handle, conn );
    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {

        bzero(objData, sizeof(collEnt_t));

        objData->dataName = &dataName->value[dataName->len * i];
        objData->collName = &colName->value[colName->len * i];
        objData->dataId = &dataId->value[dataId->len * i];
        objData->replNum = atoi( &replNum->value[replNum->len * i] );
        objData->dataSize = strtoll( &dataSize->value[dataSize->len * i], 0, 0 );
        objData->resource = &rescName->value[rescName->len * i];
        objData->ownerName = &dataOwnerName->value[dataOwnerName->len * i];
        objData->replStatus = atoi( &replStatus->value[replStatus->len * i] );
        objData->modifyTime = &dataModify->value[dataModify->len * i];
        objData->createTime = &dataCreate->value[dataCreate->len * i];
        objData->resc_hier = &rescHier->value[rescHier->len * i];

        if ( rodsArgs->veryLongOption == True ) {
            objData->chksum = &chksumStr->value[chksumStr->len * i];
            objData->phyPath = &dataPath->value[dataPath->len * i];
            objData->dataType = &dataType->value[dataType->len * i];
        }

        return 0;
    }

    return 0;
}


int gorods_get_resources_new(rcComm_t* conn, goRodsStringResult_t* result, char** err) {

    genQueryInp_t genQueryInp;
    genQueryOut_t *genQueryOut;
    int i1a[20];
    int i1b[20] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
    int i2a[20];
    char *condVal[10];
    char v1[BIG_STR];
    int i, status;

    int cont;

    memset(&genQueryInp, 0, sizeof(genQueryInp_t));

    i = 0;
    i1a[i++] = COL_R_RESC_NAME;

    genQueryInp.selectInp.inx = i1a;
    genQueryInp.selectInp.value = i1b;
    genQueryInp.selectInp.len = i;

    genQueryInp.sqlCondInp.inx = i2a;
    genQueryInp.sqlCondInp.value = condVal;
    
    // =-=-=-=-=-=-=-
    // JMC - backport 4629
    i2a[0] = COL_R_RESC_NAME;
    sprintf(v1, "!='%s'", BUNDLE_RESC); /* all but bundleResc */
    condVal[0] = v1;
    genQueryInp.sqlCondInp.len = 1;
    // =-=-=-=-=-=-=-

    genQueryInp.maxRows = 50;
    genQueryInp.continueInx = 0;
    status = rcGenQuery(conn, &genQueryInp, &genQueryOut);

    if ( status == CAT_NO_ROWS_FOUND ) {
       
        freeGenQueryOut(&genQueryOut);

        i1a[0] = COL_R_RESC_INFO;
        genQueryInp.selectInp.len = 1;
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);

        if ( status == 0 ) {
            freeGenQueryOut(&genQueryOut);
            *err = "No rows found";
            return CAT_NO_ROWS_FOUND;
        }

        if ( status == CAT_NO_ROWS_FOUND ) {
            freeGenQueryOut(&genQueryOut);
            *err = "Resource does not exist";
            return status;
        }
    }

    gorods_get_resource_result(conn, genQueryOut, result);
    
    cont = genQueryOut->continueInx;
    
    freeGenQueryOut(&genQueryOut);

    while ( status == 0 && cont > 0 ) {
        
        genQueryInp.continueInx = cont;

        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        
        cont = genQueryOut->continueInx;

        gorods_get_resource_result(conn, genQueryOut, result);

        freeGenQueryOut(&genQueryOut);
    }

    

    return 0;
}

// Might be a bug here, needs to realloc if called more than once
int gorods_get_resource_result(rcComm_t *Conn, genQueryOut_t *genQueryOut, goRodsStringResult_t* result) {

    int i, j;

    result->size = genQueryOut->rowCnt;
    result->strArr = gorods_malloc(result->size * sizeof(char*));
    
    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
       
        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            char *tResult;
            tResult = genQueryOut->sqlResult[j].value;
            tResult += i * genQueryOut->sqlResult[j].len;

            result->strArr[i] = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
        }
    }

    return 0;
}



int gorods_chmod(rcComm_t *conn, char* path, char* zone, char* ugName, char* accessLevel, int recursive, char** err) {

	int status;
	modAccessControlInp_t modAccessControlInp;

	modAccessControlInp.recursiveFlag = recursive;
	modAccessControlInp.accessLevel = accessLevel;
	modAccessControlInp.userName = ugName;
	modAccessControlInp.zone = zone;
	modAccessControlInp.path = path;

	status = rcModAccessControl(conn, &modAccessControlInp);	

	if ( status < 0 ) {
		*err = "rcModAccessControl failed";
		return status;
	}

	return 0;
}


// typedef struct {
//     int size;
//     int keySize;
//     char** hashKeys;
//     char** hashValues;
// } goRodsHashResult_t;


void gorods_free_map_result(goRodsHashResult_t* result) {

    int i;
    for ( i = 0; i < result->keySize; i++ ) {
        free(result->hashKeys[i]);
    }
    free(result->hashKeys);

    for ( i = 0; i < (result->size * result->keySize); i++ ) {
        free(result->hashValues[i]);
    }
    free(result->hashValues);

}



int gorods_iquest_general(rcComm_t *conn, char *selectConditionString, int noDistinctFlag, int upperCaseFlag, char *zoneName, goRodsHashResult_t* result, char** err) {
    /*
      NoDistinctFlag is 1 if the user is requesting 'distinct' to be skipped.
     */
    int i;
    int cont;
    genQueryInp_t genQueryInp;
    genQueryOut_t *genQueryOut = NULL;

    memset(&genQueryInp, 0, sizeof(genQueryInp_t));

    i = fillGenQueryInpFromStrCond(selectConditionString, &genQueryInp);
    if ( i < 0 ) {
        return i;
    }

    if ( noDistinctFlag ) {
        genQueryInp.options = NO_DISTINCT;
    }

    if ( upperCaseFlag ) {
        genQueryInp.options = UPPER_CASE_WHERE;
    }

    if ( zoneName != 0 && zoneName[0] != '\0' ) {
        addKeyVal(&genQueryInp.condInput, ZONE_KW, zoneName);
    }

    genQueryInp.maxRows = MAX_SQL_ROWS;
    genQueryInp.continueInx = 0;
    i = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

    if ( i < 0 ) {
        freeGenQueryOut(&genQueryOut);
        return i;
    }

    i = gorods_build_iquest_result(genQueryOut, result, err);
    if ( i < 0 ) {
        freeGenQueryOut(&genQueryOut);
        return i;
    }

    freeGenQueryOut(&genQueryOut);

    while ( i == 0 && cont > 0 ) {

        genQueryInp.continueInx = cont;

        i = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;
        
        if ( i < 0 ) {
            freeGenQueryOut(&genQueryOut);
            return i;
        }

        i = gorods_build_iquest_result(genQueryOut, result, err);
        if ( i < 0 ) {
            freeGenQueryOut(&genQueryOut);
            return i;
        }

        freeGenQueryOut(&genQueryOut);
    }

    return 0;

}



void printBasicGenQueryOut( genQueryOut_t *genQueryOut, goRodsGenQueryResult_t* result ) {
    int i, j;

    result->result = (char***)gorods_malloc(genQueryOut->rowCnt * sizeof(char**));
    result->rowSize = genQueryOut->rowCnt;

    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        if ( i > 0 ) {
            printf( "----\n" );
        }

        result->result[i] = (char**)gorods_malloc(genQueryOut->attriCnt * sizeof(char*));
        result->rowSize = genQueryOut->attriCnt;

        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            char *tResult;
            tResult = genQueryOut->sqlResult[j].value;
            tResult += i * genQueryOut->sqlResult[j].len;

            result->result[i][j] = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);

            printf( "%s\n", tResult );
        }
    }
    

}

int gorods_exec_specific_query(rcComm_t *conn, char *sql, char *args[], int argsLen, char* zoneArgument, goRodsGenQueryResult_t* result, char** err) {
    
    specificQueryInp_t specificQueryInp;
    int status, i, argsOffset;
    genQueryOut_t *genQueryOut = NULL;
    char *cp;
    int nQuestionMarks, nArgs;
    char *format = "";
    char myFormat[300] = "";

    memset( &specificQueryInp, 0, sizeof( specificQueryInp_t ) );
    specificQueryInp.maxRows = MAX_SQL_ROWS;
    specificQueryInp.continueInx = 0;
    specificQueryInp.sql = sql;

    if ( zoneArgument != 0 && zoneArgument[0] != '\0' ) {
        addKeyVal( &specificQueryInp.condInput, ZONE_KW, zoneArgument );
    }

    /* To differentiate format from args, count the ? in the SQL and the
       arguments */
    cp = specificQueryInp.sql;
    nQuestionMarks = 0;
    while ( *cp != '\0' ) {
        if ( *cp++ == '?' ) {
            nQuestionMarks++;
        }
    }
    i = argsLen;
    nArgs = 0;
    while ( args[i] != NULL && strlen( args[i] ) > 0 ) {
        nArgs++;
        i++;
    }
    /* If the SQL is an alias, counting the ?'s won't be accurate so now
       the following is only done if nQuestionMarks is > 0.  But this means
       iquest won't be able to notice a Format statement when using aliases,
       but will instead assume all are parameters to the SQL. */
    if ( nQuestionMarks > 0 && nArgs > nQuestionMarks ) {
        format = args[argsOffset];  /* this must be the format */
        argsOffset++;
        strncpy( myFormat, format, 300 - 10 );
        strcat( myFormat, "\n" ); /* since \n is difficult to pass in
				on the command line, add one by default */
    }

    i = 0;
    while ( args[argsOffset] != NULL && strlen( args[argsOffset] ) > 0 ) {
        specificQueryInp.args[i++] = args[argsOffset];
        argsOffset++;
    }
    status = rcSpecificQuery( conn, &specificQueryInp, &genQueryOut );
    if ( status < 0 ) {
        return status;
    }

   printBasicGenQueryOut( genQueryOut, result );

    while ( status == 0 && genQueryOut->continueInx > 0 ) {

        specificQueryInp.continueInx = genQueryOut->continueInx;
        status = rcSpecificQuery( conn, &specificQueryInp, &genQueryOut );
        if ( status < 0 ) {
            return status;
        }

        printBasicGenQueryOut( genQueryOut, result );
    }

    return 0;

}

int gorods_build_iquest_result(genQueryOut_t * genQueryOut, goRodsHashResult_t* result, char** err) {
    int i = 0, n = 0, j = 0;
    sqlResult_t *v[MAX_SQL_ATTR];
    char * cname[MAX_SQL_ATTR];


    printf("Build iquest request\n");

    n = genQueryOut->attriCnt;
 
    for ( i = 0; i < n; i++ ) {
        v[i] = &genQueryOut->sqlResult[i];
        cname[i] = getAttrNameFromAttrId(v[i]->attriInx);
        if ( cname[i] == NULL ) {
            *err = "Error in gorods_build_iquest_result, column not found";
            printf("CRAPPP\n");
            return NO_COLUMN_NAME_FOUND;
        }
    }

    if ( result->size == 0 ) {
         printf("Build iquest request size: 0\n");
        result->size = genQueryOut->rowCnt;
        result->keySize = genQueryOut->attriCnt;
        
        int keySize = result->keySize * sizeof(char*);
        int valuesSize = keySize * result->size;

        result->hashKeys = gorods_malloc(keySize);
        result->hashValues = gorods_malloc(valuesSize);

        // Fill key arr
        for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
            result->hashKeys[j] = strcpy(gorods_malloc(strlen(cname[j]) + 1), cname[j]);
        }
    } else {
        result->size += genQueryOut->rowCnt;
        int newSize = result->size * result->keySize;

        result->hashValues = realloc(result->hashValues, newSize);
    }

 
     printf("Build iquest request row count %i\n", genQueryOut->rowCnt);

    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
     
        for ( j = 0; j < n; j++ ) {
            
            char* value = &v[j]->value[v[j]->len * i];

            int rowStart = i * result->keySize;

            result->hashValues[rowStart + j] = strcpy(gorods_malloc(strlen(value) + 1), value);
        }
        
    }

    return 0;
}

int gorods_trimrepls_dataobject(rcComm_t *conn, char* objPath, char* ageStr, char* resource, char* keepCopiesStr, char** err) {

    int status;
    dataObjInp_t dataObjInp; 
    bzero(&dataObjInp, sizeof(dataObjInp));

    rstrcpy(dataObjInp.objPath, objPath, MAX_NAME_LEN); 

    if ( keepCopiesStr != NULL && keepCopiesStr[0] != '\0' ) {
        addKeyVal(&dataObjInp.condInput, COPIES_KW, keepCopiesStr);
    }

    if ( ageStr != NULL && ageStr[0] != '\0' ) {
        addKeyVal(&dataObjInp.condInput, AGE_KW, ageStr);
    }

    if ( resource != NULL && resource[0] != '\0' ) {
        addKeyVal(&dataObjInp.condInput, RESC_NAME_KW, resource); 
    }

    dataObjInp.numThreads = conn->transStat.numThreads;
    
    status = rcDataObjTrim(conn, &dataObjInp);

    if ( status < 0 ) { 
        *err = "rcDataObjTrim failed";
        return status;
    }

    return 0;

}


int gorods_phymv_dataobject(rcComm_t *conn, char* objPath, char* sourceResource, char* destResource, char** err) {

    int status;
    dataObjInp_t dataObjInp; 
    bzero(&dataObjInp, sizeof(dataObjInp)); 

    rstrcpy(dataObjInp.objPath, objPath, MAX_NAME_LEN);

    addKeyVal(&dataObjInp.condInput, RESC_NAME_KW, sourceResource); 
    addKeyVal(&dataObjInp.condInput, DEST_RESC_NAME_KW, destResource); 

    dataObjInp.numThreads = conn->transStat.numThreads;

    status = rcDataObjPhymv(conn, &dataObjInp); 
    if ( status < 0 ) { 
        *err = "rcDataObjPhymv failed";
        return status;
    }

    return 0;

}

int gorods_repl_dataobject(rcComm_t *conn, char* objPath, char* resourceName, int backupMode, int createMode, rodsLong_t dataSize, char** err) {
    
    int status;
    dataObjInp_t dataObjInp; 
    bzero(&dataObjInp, sizeof(dataObjInp));

    rstrcpy(dataObjInp.objPath, objPath, MAX_NAME_LEN); 
    dataObjInp.createMode = createMode;
    dataObjInp.dataSize = dataSize;
    dataObjInp.numThreads = conn->transStat.numThreads;

    if ( backupMode > 0 ) {
        addKeyVal(&dataObjInp.condInput, BACKUP_RESC_NAME_KW, resourceName);
    } else {
        addKeyVal(&dataObjInp.condInput, DEST_RESC_NAME_KW, resourceName);
    }

    status = rcDataObjRepl(conn, &dataObjInp); 
    if ( status < 0 ) { 
        *err = "rcDataObjRepl failed";
        return status;
    }

    return 0;
}


int gorods_get_collection_inheritance(rcComm_t *conn, char *collName, int* enabled, char** err) {
    genQueryOut_t *genQueryOut = NULL;
    int status;
    sqlResult_t *inheritResult;
    char *inheritStr;

    status = queryCollInheritance(conn, collName, &genQueryOut);

    if ( status < 0 ) {
        freeGenQueryOut(&genQueryOut);
        return status;
    }

    if ( (inheritResult = getSqlResultByInx(genQueryOut, COL_COLL_INHERITANCE)) == NULL ) {
        *err = "printCollInheritance: getSqlResultByInx for COL_COLL_INHERITANCE failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    inheritStr = &inheritResult->value[0];

    if ( *inheritStr == '1' ) {
        *enabled = 1;
    } else {
        *enabled = -1;
    }

    freeGenQueryOut(&genQueryOut);

    return status;
}

int gorods_get_collection_acl(rcComm_t *conn, char *collName, goRodsACLResult_t* result, char* zoneHint, char** err) {
    genQueryOut_t *genQueryOut = NULL;
    int status;
    int i;
    sqlResult_t *userName, *userZone, *dataAccess, *userType;
    char *userNameStr, *userZoneStr, *dataAccessStr, *userTypeStr;

    /* First try a specific-query.  If this is defined, it should be
        used as it returns the group names without expanding them to
        individual users and this is important to some sites (iPlant,
        in particular).  If this fails, go on the the general-query.
     */
    status = queryCollAclSpecific(conn, collName, zoneHint, &genQueryOut);
    if ( status >= 0 ) {
        int i, j;

        result->size = genQueryOut->rowCnt;
    	result->aclArr = gorods_malloc(sizeof(goRodsACL_t) * result->size);

        for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
            char *tResult[10];
            char empty = 0;
            tResult[3] = 0;

            for ( j = 0; j < 10; j++ ) {
                tResult[j] = &empty;
                if ( j < genQueryOut->attriCnt ) {
                    tResult[j] = genQueryOut->sqlResult[j].value;
                    tResult[j] += i * genQueryOut->sqlResult[j].len;
                }
            }

            goRodsACL_t* acl = &(result->aclArr[i]);

	        acl->name = strcpy(gorods_malloc(strlen(tResult[0]) + 1), tResult[0]);
	        acl->zone = strcpy(gorods_malloc(strlen(tResult[1]) + 1), tResult[1]);
	        acl->dataAccess = strcpy(gorods_malloc(strlen(tResult[2]) + 1), tResult[2]);
            acl->acltype =  strcpy(gorods_malloc(strlen(tResult[3]) + 1), tResult[3]);
        }

        freeGenQueryOut(&genQueryOut);
        return status;
    }

    status = queryCollAcl(conn, collName, zoneHint, &genQueryOut);

    if ( status < 0 ) {
        *err = "Error in queryCollAcl";
        freeGenQueryOut(&genQueryOut);
        return status;
    }

    if ( ( userName = getSqlResultByInx( genQueryOut, COL_COLL_USER_NAME ) ) == NULL ) {
        *err = "printCollAcl: getSqlResultByInx for COL_COLL_USER_NAME failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }
    if ( ( userZone = getSqlResultByInx( genQueryOut, COL_COLL_USER_ZONE ) ) == NULL ) {
        *err = "printCollAcl: getSqlResultByInx for COL_COLL_USER_ZONE failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataAccess = getSqlResultByInx( genQueryOut, COL_COLL_ACCESS_NAME ) ) == NULL ) {
        *err = "printCollAcl: getSqlResultByInx for COL_COLL_ACCESS_NAME failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( userType = getSqlResultByInx( genQueryOut, COL_COLL_ACCESS_TYPE ) ) == NULL ) {
        *err = "printCollAcl: getSqlResultByInx for COL_COLL_ACCESS_TYPE failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    result->size = genQueryOut->rowCnt;
    result->aclArr = gorods_malloc(sizeof(goRodsACL_t) * result->size);

    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        userNameStr = &userName->value[userName->len * i];
        userZoneStr = &userZone->value[userZone->len * i];
        dataAccessStr = &dataAccess->value[dataAccess->len * i];
        userTypeStr = &dataAccess->value[userType->len * i];

        goRodsACL_t* acl = &(result->aclArr[i]);

        acl->name = strcpy(gorods_malloc(strlen(userNameStr) + 1), userNameStr);
        acl->zone = strcpy(gorods_malloc(strlen(userZoneStr) + 1), userZoneStr);
        acl->dataAccess = strcpy(gorods_malloc(strlen(dataAccessStr) + 1), dataAccessStr);
        acl->acltype = strcpy(gorods_malloc(strlen(userTypeStr) + 1), userTypeStr);
    }

    freeGenQueryOut(&genQueryOut);

    return status;
}


int
gorods_queryDataObjAcl (rcComm_t *conn, char *dataId, char *zoneHint,
                 genQueryOut_t **genQueryOut)
{
    genQueryInp_t genQueryInp;
    int status;
    char tmpStr[MAX_NAME_LEN];

    if (dataId == NULL || genQueryOut == NULL) {
        return (USER__NULL_INPUT_ERR);
    }

    memset (&genQueryInp, 0, sizeof (genQueryInp_t));

    if (zoneHint != NULL) {
       addKeyVal (&genQueryInp.condInput, ZONE_KW, zoneHint);
    }

    addInxIval (&genQueryInp.selectInp, COL_USER_NAME, 1);
    addInxIval (&genQueryInp.selectInp, COL_USER_ZONE, 1);
    addInxIval (&genQueryInp.selectInp, COL_DATA_ACCESS_NAME, 1);
    addInxIval (&genQueryInp.selectInp, COL_USER_TYPE, 1);

    snprintf (tmpStr, MAX_NAME_LEN, " = '%s'", dataId);

    addInxVal (&genQueryInp.sqlCondInp, COL_DATA_ACCESS_DATA_ID, tmpStr);

    snprintf (tmpStr, MAX_NAME_LEN, "='%s'", "access_type");

    /* Currently necessary since other namespaces exist in the token table */
    addInxVal (&genQueryInp.sqlCondInp, COL_DATA_TOKEN_NAMESPACE, tmpStr);

    genQueryInp.maxRows = MAX_SQL_ROWS;

    status =  rcGenQuery(conn, &genQueryInp, genQueryOut);

    return (status);

}


int gorods_get_dataobject_acl(rcComm_t* conn, char* dataId, goRodsACLResult_t* result, char* zoneHint, char** err) {
    genQueryOut_t *genQueryOut = NULL;
    int status;
    int i;
    sqlResult_t *userName, *userZone, *dataAccess, *userType;
    char *userNameStr, *userZoneStr, *dataAccessStr, *userTypeStr;

    status = gorods_queryDataObjAcl(conn, dataId, zoneHint, &genQueryOut);

    if ( status < 0 ) {
    	*err = "Error in queryDataObjAcl";
        freeGenQueryOut(&genQueryOut);
        return status;
    }

    if ( ( userName = getSqlResultByInx( genQueryOut, COL_USER_NAME ) ) == NULL ) {
        *err = "printDataAcl: getSqlResultByInx for COL_USER_NAME failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( userZone = getSqlResultByInx( genQueryOut, COL_USER_ZONE ) ) == NULL ) {
        *err = "printDataAcl: getSqlResultByInx for COL_USER_ZONE failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( dataAccess = getSqlResultByInx( genQueryOut, COL_DATA_ACCESS_NAME ) ) == NULL ) {
        *err = "printDataAcl: getSqlResultByInx for COL_DATA_ACCESS_NAME failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    if ( ( userType = getSqlResultByInx( genQueryOut, COL_USER_TYPE ) ) == NULL ) {
        *err = "printDataAcl: getSqlResultByInx for COL_USER_TYPE failed";
        freeGenQueryOut(&genQueryOut);
        return UNMATCHED_KEY_OR_INDEX;
    }

    result->size = genQueryOut->rowCnt;
    result->aclArr = gorods_malloc(sizeof(goRodsACL_t) * result->size);

    for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
        userNameStr = &userName->value[userName->len * i];
        userZoneStr = &userZone->value[userZone->len * i];
        dataAccessStr = &dataAccess->value[dataAccess->len * i];
        userTypeStr = &userType->value[userType->len * i];

        goRodsACL_t* acl = &(result->aclArr[i]);

        acl->name = strcpy(gorods_malloc(strlen(userNameStr) + 1), userNameStr);
        acl->zone = strcpy(gorods_malloc(strlen(userZoneStr) + 1), userZoneStr);
        acl->dataAccess = strcpy(gorods_malloc(strlen(dataAccessStr) + 1), dataAccessStr);
        acl->acltype = strcpy(gorods_malloc(strlen(userTypeStr) + 1), userTypeStr);
    }

    freeGenQueryOut(&genQueryOut);

    return status;
}

void gorods_free_acl_result(goRodsACLResult_t* result) {

	int i;
	for ( i = 0; i < result->size; i++ ) {
        goRodsACL_t* acl = &(result->aclArr[i]);

        free(acl->name);
        free(acl->zone);
        free(acl->dataAccess);
        free(acl->acltype);
    }

    free(result->aclArr);

}

int gorods_read_dataobject(int handleInx, rodsLong_t length, bytesBuf_t* buffer, int* bytesRead, rcComm_t* conn, char** err) {
	
	openedDataObjInp_t dataObjReadInp; 
	
	bzero(&dataObjReadInp, sizeof(openedDataObjInp_t)); 
	bzero(buffer, sizeof(bytesBuf_t)); 

	dataObjReadInp.l1descInx = handleInx; 
	dataObjReadInp.len = (int)length;

	*bytesRead = rcDataObjRead(conn, &dataObjReadInp, buffer); 

	if ( *bytesRead < 0 ) { 
		*err = "rcDataObjRead failed";
		return *bytesRead;
	}

	return 0;
}

int gorods_lseek_dataobject(int handleInx, rodsLong_t offset, rcComm_t* conn, char** err) {
	int status; 

	openedDataObjInp_t dataObjLseekInp;
	fileLseekOut_t *dataObjLseekOut = NULL; 

	bzero(&dataObjLseekInp, sizeof(dataObjLseekInp)); 
	
	dataObjLseekInp.l1descInx = handleInx; 
	
	if ( dataObjLseekInp.l1descInx < 0 ) { 
		*err = "rcDataObjLSeek failed, invalid handle passed";
		return -1;
	} 
	
	dataObjLseekInp.offset = offset; 
	dataObjLseekInp.whence = SEEK_SET; 
	
	status = rcDataObjLseek(conn, &dataObjLseekInp, &dataObjLseekOut); 
	if ( status < 0 ) { 
		*err = "rcDataObjLSeek failed";
		return status;
	}

	free(dataObjLseekOut);
    dataObjLseekOut = NULL;

	return 0;
}

int gorods_stat_dataobject(char* path, rodsObjStat_t** rodsObjStatOut, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 

	*rodsObjStatOut = NULL;

	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 
	
    dataObjInp.numThreads = conn->transStat.numThreads;

	// pass memory address of rodsObjStatOut pointer
	int status = rcObjStat(conn, &dataObjInp, rodsObjStatOut); 
	if ( status < 0 ) { 
		*err = "rcObjStat failed";
		return status;
	}

	return 0;
}


int gorods_copy_dataobject(char* source, char* destination, int force, char* resource, rcComm_t* conn, char** err) {
	dataObjCopyInp_t dataObjCopyInp; 
	bzero(&dataObjCopyInp, sizeof(dataObjCopyInp)); 

	rstrcpy(dataObjCopyInp.destDataObjInp.objPath, destination, MAX_NAME_LEN); 
	rstrcpy(dataObjCopyInp.srcDataObjInp.objPath, source, MAX_NAME_LEN); 

	addKeyVal(&dataObjCopyInp.destDataObjInp.condInput, REG_CHKSUM_KW, ""); 

    if ( resource != NULL && resource[0] != '\0' ) {
        addKeyVal(&dataObjCopyInp.destDataObjInp.condInput, DEST_RESC_NAME_KW, resource); 
    }

    if ( force > 0 ) {
        addKeyVal(&dataObjCopyInp.destDataObjInp.condInput, FORCE_FLAG_KW, ""); 
    }

	int status = rcDataObjCopy(conn, &dataObjCopyInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjCopy failed";
		return status;
	}

	return 0;
}

int gorods_move_dataobject(char* source, char* destination, int objType, rcComm_t* conn, char** err) {
	dataObjCopyInp_t dataObjRenameInp; 
	bzero(&dataObjRenameInp, sizeof(dataObjRenameInp)); 

    if ( objType == DATA_OBJ_T ) {
        dataObjRenameInp.srcDataObjInp.oprType = dataObjRenameInp.destDataObjInp.oprType = RENAME_DATA_OBJ;
    } else if ( objType == COLL_OBJ_T ) {
        dataObjRenameInp.srcDataObjInp.oprType = dataObjRenameInp.destDataObjInp.oprType = RENAME_COLL;
    }

	rstrcpy(dataObjRenameInp.destDataObjInp.objPath, destination, MAX_NAME_LEN); 
	rstrcpy(dataObjRenameInp.srcDataObjInp.objPath, source, MAX_NAME_LEN); 

	int status = rcDataObjRename(conn, &dataObjRenameInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjRename failed";
		return status;
	}

	return 0;
}

int gorods_unlink_dataobject(char* path, int force, rcComm_t* conn, char** err) {
	dataObjInp_t dataObjInp; 
	bzero(&dataObjInp, sizeof(dataObjInp));

	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 
	
	if ( force != 0 ) {
		addKeyVal(&dataObjInp.condInput, FORCE_FLAG_KW, ""); 
	}

    dataObjInp.numThreads = conn->transStat.numThreads;
	
	int status = rcDataObjUnlink(conn, &dataObjInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjRename failed";
		return status;
	}

	return 0;
}

int gorods_checksum_dataobject(char* path, char** outChksum, rcComm_t* conn, char** err) {

	dataObjInp_t dataObjInp; 

	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 

	addKeyVal(&dataObjInp.condInput, FORCE_CHKSUM_KW, ""); 

    dataObjInp.numThreads = conn->transStat.numThreads;

	int status = rcDataObjChksum(conn, &dataObjInp, outChksum); 
	if ( status < 0 ) { 
		*err = "rcDataObjChksum failed";
		return status;
	}

	return 0;
}


const char NON_ROOT_COLL_CHECK_STR[] = "<>'/'";

int gorods_rclReadCollectionObjs(rcComm_t *conn, collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts) {
    int status;

    collHandle->queryHandle.conn = conn;        /* in case it changed */
    status = gorods_readCollectionObjs( collHandle, collEnt, opts );

    return status;
}

int gorods_rclReadCollectionCols(rcComm_t *conn, collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts) {
    int status;

    collHandle->queryHandle.conn = conn;        /* in case it changed */
    status = gorods_readCollectionCols( collHandle, collEnt, opts );

    return status;
}

int gorods_readCollectionCols( collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts) {
    int status = 0;
    int savedStatus = 0;

    queryHandle_t *queryHandle = &collHandle->queryHandle;

    if ( queryHandle == NULL || collHandle == NULL || collEnt == NULL ) {
        rodsLog( LOG_ERROR,
                 "rclReadCollection: NULL queryHandle or collHandle input" );
        return USER__NULL_INPUT_ERR;
    }

    memset( collEnt, 0, sizeof( collEnt_t ) );

    if ( collHandle->state == COLL_CLOSED ) {
        return CAT_NO_ROWS_FOUND;
    }


    if ( collHandle->state == COLL_OPENED ) {
        status = gorods_genCollResInColl( queryHandle, collHandle, opts );
        if ( status < 0 && status != CAT_NO_ROWS_FOUND ) {
            rodsLog( LOG_ERROR, "genCollResInColl in readCollection failed with status %d", status );
        }
    }

    if ( collHandle->state == COLL_COLL_OBJ_QUERIED ) {
        status = gorods_getNextCollMetaInfo( collHandle, collEnt );

        if ( status >= 0 ) {
            return status;
        }
        else {
            if ( status != CAT_NO_ROWS_FOUND ) {
                rodsLog( LOG_ERROR,
                         "rclReadCollection: getNextCollMetaInfo error for %s. status = %d",
                         collHandle->dataObjInp.objPath, status );
            }
            /* cleanup */
            if ( collHandle->dataObjInp.specColl == NULL ) {
                clearGenQueryInp( &collHandle->genQueryInp );
            }
            // Leave open for Data Obj Reads
            collHandle->state = COLL_OPENED;
        }
        
        return status;
    }


    return CAT_NO_ROWS_FOUND;
}

int gorods_readCollectionObjs( collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts) {
    int status = 0;
    int savedStatus = 0;

    queryHandle_t *queryHandle = &collHandle->queryHandle;

    if ( queryHandle == NULL || collHandle == NULL || collEnt == NULL ) {
        rodsLog( LOG_ERROR,
                 "rclReadCollection: NULL queryHandle or collHandle input" );
        return USER__NULL_INPUT_ERR;
    }

    memset( collEnt, 0, sizeof( collEnt_t ) );

    if ( collHandle->state == COLL_CLOSED ) {
        return CAT_NO_ROWS_FOUND;
    }


    if ( collHandle->state == COLL_OPENED ) {
        status = gorods_genDataResInColl( queryHandle, collHandle, opts );
        if ( status < 0 && status != CAT_NO_ROWS_FOUND ) {
            rodsLog( LOG_ERROR, "genDataResInColl in readCollection failed with status %d", status );
        }
    }

    if ( collHandle->state == COLL_DATA_OBJ_QUERIED ) {
        status = gorods_getNextDataObjMetaInfo( collHandle, collEnt );

        if ( status >= 0 ) {
            return status;
        }
        else {
            if ( status != CAT_NO_ROWS_FOUND ) {
                rodsLog( LOG_ERROR,
                         "rclReadCollection: getNextDataObjMetaInfo error for %s. status = %d",
                         collHandle->dataObjInp.objPath, status );
            }
            /* cleanup */
            if ( collHandle->dataObjInp.specColl == NULL ) {
                clearGenQueryInp( &collHandle->genQueryInp );
            }
            /* Nothing else to do. cleanup */
            collHandle->state = COLL_CLOSED;
        }
        return status;
    }
    
    return CAT_NO_ROWS_FOUND;
}

int gorods_genCollResInColl( queryHandle_t *queryHandle, collHandle_t *collHandle, goRodsQueryOpts_t opts) {
    genQueryOut_t *genQueryOut = NULL;
    int status = 0;

    /* query for sub-collections */
    if ( collHandle->dataObjInp.specColl != NULL ) {
        if ( collHandle->dataObjInp.specColl->collClass == LINKED_COLL ) {
            memset( &collHandle->genQueryInp, 0, sizeof( genQueryInp_t ) );
            status = gorods_queryCollInColl( queryHandle,
                                      collHandle->linkedObjPath, collHandle->flags & ( ~RECUR_QUERY_FG ),
                                      &collHandle->genQueryInp, &genQueryOut, opts );
        }
        else {

            if ( strlen( collHandle->linkedObjPath ) > 0 ) {
                rstrcpy( collHandle->dataObjInp.objPath,
                         collHandle->linkedObjPath, MAX_NAME_LEN );
            }
            addKeyVal( &collHandle->dataObjInp.condInput,
                       SEL_OBJ_TYPE_KW, "collection" );
            collHandle->dataObjInp.openFlags = 0;    /* start over */
            status = ( *queryHandle->querySpecColl )(
                         ( rcComm_t * ) queryHandle->conn, &collHandle->dataObjInp,
                         &genQueryOut );
        }
    }
    else {
        memset( &collHandle->genQueryInp, 0, sizeof( genQueryInp_t ) );
        status = gorods_queryCollInColl( queryHandle,
                                  collHandle->dataObjInp.objPath, collHandle->flags,
                                  &collHandle->genQueryInp, &genQueryOut, opts );
    }

    collHandle->rowInx = 0;
    collHandle->state = COLL_COLL_OBJ_QUERIED;
    if ( status >= 0 ) {
        status = genQueryOutToCollRes( &genQueryOut,
                                       &collHandle->collSqlResult );
    }
    else if ( status != CAT_NO_ROWS_FOUND ) {
        rodsLog( LOG_ERROR,
                 "genCollResInColl: query collection error for %s. status = %d",
                 collHandle->dataObjInp.objPath, status );
    } else {
        // Set the total result size even if there are no rows in this response
        collHandle->collSqlResult.totalRowCount = genQueryOut->totalRowCount;
        collHandle->collSqlResult.rowCnt = genQueryOut->rowCnt;
    }
    freeGenQueryOut( &genQueryOut );
    return status;
}

int gorods_genDataResInColl( queryHandle_t *queryHandle, collHandle_t *collHandle, goRodsQueryOpts_t opts ) {
    genQueryOut_t *genQueryOut = NULL;
    int status = 0;

    if ( collHandle->dataObjInp.specColl != NULL ) {
        if ( collHandle->dataObjInp.specColl->collClass == LINKED_COLL ) {
            memset( &collHandle->genQueryInp, 0, sizeof( genQueryInp_t ) );
            status = gorods_queryDataObjInColl( queryHandle,
                                         collHandle->linkedObjPath, collHandle->flags & ( ~RECUR_QUERY_FG ),
                                         &collHandle->genQueryInp, &genQueryOut,
                                         &collHandle->dataObjInp.condInput, opts );
        }
        else {
            if ( strlen( collHandle->linkedObjPath ) > 0 ) {
                rstrcpy( collHandle->dataObjInp.objPath,
                         collHandle->linkedObjPath, MAX_NAME_LEN );
            }
            addKeyVal( &collHandle->dataObjInp.condInput,
                       SEL_OBJ_TYPE_KW, "dataObj" );
            collHandle->dataObjInp.openFlags = 0;    /* start over */
            status = ( *queryHandle->querySpecColl )
                     ( ( rcComm_t * ) queryHandle->conn,
                       &collHandle->dataObjInp, &genQueryOut );
        }
    }
    else {
        memset( &collHandle->genQueryInp, 0, sizeof( genQueryInp_t ) );
        status = gorods_queryDataObjInColl( queryHandle,
                                     collHandle->dataObjInp.objPath, collHandle->flags,
                                     &collHandle->genQueryInp, &genQueryOut,
                                     &collHandle->dataObjInp.condInput, opts );
    }

    collHandle->rowInx = 0;
    collHandle->state = COLL_DATA_OBJ_QUERIED;
    if ( status >= 0 ) {
        status = genQueryOutToDataObjRes( &genQueryOut,
                                          &collHandle->dataObjSqlResult );
    }
    else if ( status != CAT_NO_ROWS_FOUND ) {
        rodsLog( LOG_ERROR,
                 "genDataResInColl: query dataObj error for %s. status = %d",
                 collHandle->dataObjInp.objPath, status );
    } else {
        // Set the total result size even if there are no rows in this response
        collHandle->dataObjSqlResult.totalRowCount = genQueryOut->totalRowCount;
        collHandle->dataObjSqlResult.rowCnt = genQueryOut->rowCnt;
    }
    freeGenQueryOut( &genQueryOut );
    return status;
}

int gorods_queryCollInColl( queryHandle_t *queryHandle, char *collection,
                 int flags, genQueryInp_t *genQueryInp,
                 genQueryOut_t **genQueryOut, goRodsQueryOpts_t opts ) {
    char collQCond[MAX_NAME_LEN];
    int status;

    if ( collection == NULL || genQueryOut == NULL ) {
        return USER__NULL_INPUT_ERR;
    }

    memset( genQueryInp, 0, sizeof( genQueryInp_t ) );

    snprintf( collQCond, MAX_NAME_LEN, "%s", NON_ROOT_COLL_CHECK_STR);
    addInxVal( &genQueryInp->sqlCondInp, COL_COLL_NAME, collQCond );

    if ( ( flags & RECUR_QUERY_FG ) != 0 ) {
        genAllInCollQCond( collection, collQCond );
        addInxVal( &genQueryInp->sqlCondInp, COL_COLL_NAME, collQCond );
    }
    else {
        snprintf( collQCond, MAX_NAME_LEN, "='%s'", collection );
        addInxVal( &genQueryInp->sqlCondInp, COL_COLL_PARENT_NAME, collQCond );
    }
    addInxIval( &genQueryInp->selectInp, COL_COLL_NAME, 1 );
    addInxIval( &genQueryInp->selectInp, COL_COLL_OWNER_NAME, 1 );
    addInxIval( &genQueryInp->selectInp, COL_COLL_CREATE_TIME, 1 );
    addInxIval( &genQueryInp->selectInp, COL_COLL_MODIFY_TIME, 1 );
    addInxIval( &genQueryInp->selectInp, COL_COLL_TYPE, 1 );
    addInxIval( &genQueryInp->selectInp, COL_COLL_INFO1, 1 );
    addInxIval( &genQueryInp->selectInp, COL_COLL_INFO2, 1 );

    genQueryInp->maxRows = opts.limit;
    genQueryInp->rowOffset = opts.offset;
    genQueryInp->options = RETURN_TOTAL_ROW_COUNT;

    status = ( *queryHandle->genQuery )(
                 ( rcComm_t * ) queryHandle->conn, genQueryInp, genQueryOut );

    return status;
}

/* queryDataObjInColl - query the DataObj in a collection.
 */
int
gorods_queryDataObjInColl( queryHandle_t *queryHandle, char *collection,
                    int flags, genQueryInp_t *genQueryInp,
                    genQueryOut_t **genQueryOut, keyValPair_t *condInput, goRodsQueryOpts_t opts ) {
    char collQCond[MAX_NAME_LEN];
    int status;
    char *rescName = NULL;
    if ( collection == NULL || genQueryOut == NULL ) {
        return USER__NULL_INPUT_ERR;
    }

    memset( genQueryInp, 0, sizeof( genQueryInp_t ) );

    if ( ( flags & RECUR_QUERY_FG ) != 0 ) {
        genAllInCollQCond( collection, collQCond );
        addInxVal( &genQueryInp->sqlCondInp, COL_COLL_NAME, collQCond );
    }
    else {
        snprintf( collQCond, MAX_NAME_LEN, " = '%s'", collection );
        addInxVal( &genQueryInp->sqlCondInp, COL_COLL_NAME, collQCond );
    }
    if ( ( flags & INCLUDE_CONDINPUT_IN_QUERY ) != 0 &&
            condInput != NULL &&
            ( rescName = getValByKey( condInput, RESC_NAME_KW ) ) != NULL ) {
        snprintf( collQCond, MAX_NAME_LEN, " = '%s'", rescName );
        addInxVal( &genQueryInp->sqlCondInp, COL_D_RESC_NAME, collQCond );
    }

    setQueryInpForData( flags, genQueryInp );

    genQueryInp->maxRows = opts.limit;
    genQueryInp->rowOffset = opts.offset;
    genQueryInp->options = RETURN_TOTAL_ROW_COUNT;

    status = ( *queryHandle->genQuery )(
                 ( rcComm_t * ) queryHandle->conn, genQueryInp, genQueryOut );

    return status;

}


int gorods_getNextCollMetaInfo( collHandle_t *collHandle, collEnt_t *outCollEnt ) {
    char *value;
    int len;
    char *collType, *collInfo1, *collInfo2;
    int status = 0;
    queryHandle_t *queryHandle = &collHandle->queryHandle;
    dataObjInp_t *dataObjInp =  &collHandle->dataObjInp;
    genQueryInp_t *genQueryInp = &collHandle->genQueryInp;
    collSqlResult_t *collSqlResult = &collHandle->collSqlResult;

    if ( outCollEnt == NULL ) {
        return USER__NULL_INPUT_ERR;
    }

    memset( outCollEnt, 0, sizeof( collEnt_t ) );

    outCollEnt->objType = COLL_OBJ_T;
    if ( collHandle->rowInx >= collSqlResult->rowCnt ) {
        genQueryOut_t *genQueryOut = NULL;
        int continueInx = collSqlResult->continueInx;
        return CAT_NO_ROWS_FOUND;
    }
    value = collSqlResult->collName.value;
    len = collSqlResult->collName.len;
    outCollEnt->collName = &value[len * ( collHandle->rowInx )];

    value = collSqlResult->collOwner.value;
    len = collSqlResult->collOwner.len;
    outCollEnt->ownerName = &value[len * ( collHandle->rowInx )];

    value = collSqlResult->collCreateTime.value;
    len = collSqlResult->collCreateTime.len;
    outCollEnt->createTime = &value[len * ( collHandle->rowInx )];

    value = collSqlResult->collModifyTime.value;
    len = collSqlResult->collModifyTime.len;
    outCollEnt->modifyTime = &value[len * ( collHandle->rowInx )];

    value = collSqlResult->collType.value;
    len = collSqlResult->collType.len;
    collType = &value[len * ( collHandle->rowInx )];

    if ( *collType != '\0' ) {
        value = collSqlResult->collInfo1.value;
        len = collSqlResult->collInfo1.len;
        collInfo1 = &value[len * ( collHandle->rowInx )];

        value = collSqlResult->collInfo2.value;
        len = collSqlResult->collInfo2.len;
        collInfo2 = &value[len * ( collHandle->rowInx )];

        if ( strcmp( collType, INHERIT_PAR_SPEC_COLL_STR ) == 0 ) {
            if ( dataObjInp->specColl == NULL ) {
                rodsLog( LOG_ERROR,
                         "getNextCollMetaInfo: parent specColl is NULL for %s",
                         outCollEnt->collName );
                outCollEnt->specColl.collClass = NO_SPEC_COLL;
            }
            else {
                outCollEnt->specColl = *dataObjInp->specColl;
            }
            status = 0;
        }
        else {
            status = resolveSpecCollType( collType, outCollEnt->collName,
                                          collInfo1, collInfo2, &outCollEnt->specColl );
        }
    }
    else {
        outCollEnt->specColl.collClass = NO_SPEC_COLL;
        status = 0;
    }
    ( collHandle->rowInx ) ++;
    return status;
}

int gorods_getNextDataObjMetaInfo( collHandle_t *collHandle, collEnt_t *outCollEnt ) {
    int status;
    char *value;
    int len;
    char *replStatus, *dataId;
    int dataIdLen, replStatusLen;
    queryHandle_t *queryHandle = &collHandle->queryHandle;
    dataObjInp_t *dataObjInp =  &collHandle->dataObjInp;
    genQueryInp_t *genQueryInp = &collHandle->genQueryInp;
    dataObjSqlResult_t *dataObjSqlResult = &collHandle->dataObjSqlResult;
    rodsObjStat_t *rodsObjStat = collHandle->rodsObjStat;
    char *prevdataId;
    int selectedInx = -1;

    if ( outCollEnt == NULL ) {
        return USER__NULL_INPUT_ERR;
    }
    prevdataId = collHandle->prevdataId;
    memset( outCollEnt, 0, sizeof( collEnt_t ) );
    outCollEnt->objType = DATA_OBJ_T;
    if ( collHandle->rowInx >= dataObjSqlResult->rowCnt ) {
        genQueryOut_t *genQueryOut = NULL;
        int continueInx = dataObjSqlResult->continueInx;

        return CAT_NO_ROWS_FOUND;
    }

    dataId = dataObjSqlResult->dataId.value;
    dataIdLen = dataObjSqlResult->dataId.len;
    replStatus = dataObjSqlResult->replStatus.value;
    replStatusLen = dataObjSqlResult->replStatus.len;

    if ( strlen( dataId ) > 0 && ( collHandle->flags & NO_TRIM_REPL_FG ) == 0 ) {
        int i;
        int gotCopy = 0;

        /* rsync type query ask for dataId. Others don't. Need to
         * screen out dup copies */

        for ( i = collHandle->rowInx; i < dataObjSqlResult->rowCnt; i++ ) {
            if ( selectedInx < 0 ) {
                /* nothing selected yet. pick this if different */
                if ( strcmp( prevdataId, &dataId[dataIdLen * i] ) != 0 ) {
                    rstrcpy( prevdataId, &dataId[dataIdLen * i], NAME_LEN );
                    selectedInx = i;
                    if ( atoi( &dataId[dataIdLen * i] ) != 0 ) {
                        gotCopy = 1;
                    }
                }
            }
            else {
                /* skip i to the next object */
                if ( strcmp( prevdataId, &dataId[dataIdLen * i] ) != 0 ) {
                    break;
                }
                if ( gotCopy == 0 &&
                        atoi( &replStatus[replStatusLen * i] ) > 0 ) {
                    /* pick a good copy */
                    selectedInx = i;
                    gotCopy = 1;
                }
            }
        }
        if ( selectedInx < 0 ) {
            return CAT_NO_ROWS_FOUND;
        }

        collHandle->rowInx = i;
    }
    else {
        selectedInx = collHandle->rowInx;
        collHandle->rowInx++;
    }

    value = dataObjSqlResult->collName.value;
    len = dataObjSqlResult->collName.len;
    outCollEnt->collName = &value[len * selectedInx];

    value = dataObjSqlResult->dataName.value;
    len = dataObjSqlResult->dataName.len;
    outCollEnt->dataName = &value[len * selectedInx];

    value = dataObjSqlResult->dataMode.value;
    len = dataObjSqlResult->dataMode.len;
    outCollEnt->dataMode = atoi( &value[len * selectedInx] );

    value = dataObjSqlResult->dataSize.value;
    len = dataObjSqlResult->dataSize.len;
    outCollEnt->dataSize = strtoll( &value[len * selectedInx], 0, 0 );

    value = dataObjSqlResult->createTime.value;
    len = dataObjSqlResult->createTime.len;
    outCollEnt->createTime = &value[len * selectedInx];

    value = dataObjSqlResult->modifyTime.value;
    len = dataObjSqlResult->modifyTime.len;
    outCollEnt->modifyTime = &value[len * selectedInx];

    outCollEnt->dataId = &dataId[dataIdLen * selectedInx];

    outCollEnt->replStatus = atoi( &replStatus[replStatusLen *
                                   selectedInx] );

    value = dataObjSqlResult->replNum.value;
    len = dataObjSqlResult->replNum.len;
    outCollEnt->replNum = atoi( &value[len * selectedInx] );

    value = dataObjSqlResult->chksum.value;
    len = dataObjSqlResult->chksum.len;
    outCollEnt->chksum = &value[len * selectedInx];

    value = dataObjSqlResult->dataType.value;
    len = dataObjSqlResult->dataType.len;
    outCollEnt->dataType = &value[len * selectedInx];

    if ( rodsObjStat->specColl != NULL ) {
        outCollEnt->specColl = *rodsObjStat->specColl;
    }
    if ( rodsObjStat->specColl != NULL &&
            rodsObjStat->specColl->collClass != LINKED_COLL ) {
        outCollEnt->resource = rodsObjStat->specColl->resource;
        outCollEnt->ownerName = rodsObjStat->ownerName;
        outCollEnt->replStatus = NEWLY_CREATED_COPY;
        outCollEnt->resc_hier = rodsObjStat->specColl->rescHier;
    }
    else {
        value = dataObjSqlResult->resource.value;
        len = dataObjSqlResult->resource.len;
        outCollEnt->resource = &value[len * selectedInx];

        value = dataObjSqlResult->resc_hier.value;
        len = dataObjSqlResult->resc_hier.len;
        outCollEnt->resc_hier = &value[len * selectedInx];

        value = dataObjSqlResult->ownerName.value;
        len = dataObjSqlResult->ownerName.len;
        outCollEnt->ownerName = &value[len * selectedInx];
    }

    value = dataObjSqlResult->phyPath.value;
    len = dataObjSqlResult->phyPath.len;
    outCollEnt->phyPath = &value[len * selectedInx];

    return 0;
}

int gorods_query_dataobj(rcComm_t* conn, char* query, goRodsPathResult_t* result, char** err) {
	
	char* cmdToken[40];
    int cont;
	int i;
	for ( i = 0; i < 40; i++ ) {
		cmdToken[i] = "";
	}

	cmdToken[0] = "qu";
	cmdToken[1] = "-d";

	int tokenIndex = 2;

	build_cmd_token(cmdToken, &tokenIndex, query);

	genQueryInp_t genQueryInp;
	genQueryOut_t *genQueryOut;

	int i1a[20];
	int i1b[20];
	int i2a[20];
	char *condVal[20];
	char v1[BIG_STR];
	char v2[BIG_STR];
	char v3[BIG_STR];
	int status;
	char *columnNames[] = {"dataObj", "collection"};
	int cmdIx;
	int condIx;
	char vstr[20][BIG_STR];

	memset(&genQueryInp, 0, sizeof(genQueryInp_t));

	i1a[0] = COL_COLL_NAME;
	i1b[0] = 0;  /* (unused) */
	i1a[1] = COL_DATA_NAME;
	i1b[1] = 0;
	genQueryInp.selectInp.inx = i1a;
	genQueryInp.selectInp.value = i1b;
	genQueryInp.selectInp.len = 2;

	i2a[0] = COL_META_DATA_ATTR_NAME;
	snprintf(v1, sizeof(v1), "='%s'", cmdToken[2]);
	condVal[0] = v1;

	i2a[1] = COL_META_DATA_ATTR_VALUE;
	snprintf(v2, sizeof(v2), "%s '%s'", cmdToken[3], cmdToken[4]);
	condVal[1] = v2;

	genQueryInp.sqlCondInp.inx = i2a;
	genQueryInp.sqlCondInp.value = condVal;
	genQueryInp.sqlCondInp.len = 2;

	if ( strcmp(cmdToken[5], "or") == 0 ) {
		snprintf(v3, sizeof(v3), "|| %s '%s'", cmdToken[6], cmdToken[7]);
		rstrcat(v2, v3, BIG_STR);
	}

	cmdIx = 5;
	condIx = 2;
	while ( strcmp(cmdToken[cmdIx], "and") == 0 ) {
		i2a[condIx] = COL_META_DATA_ATTR_NAME;
		cmdIx++;
		snprintf(vstr[condIx], BIG_STR, "='%s'", cmdToken[cmdIx]);
		condVal[condIx] = vstr[condIx];
		condIx++;

		i2a[condIx] = COL_META_DATA_ATTR_VALUE;
		snprintf(vstr[condIx], BIG_STR, "%s '%s'", cmdToken[cmdIx+1], cmdToken[cmdIx+2]);
		cmdIx += 3;
		condVal[condIx] = vstr[condIx];
		condIx++;
		genQueryInp.sqlCondInp.len += 2;
	}

	if ( *cmdToken[cmdIx] != '\0' ) {
		*err = "Unrecognized input\n";
		return -1;
	}

	genQueryInp.maxRows = 10;
	genQueryInp.continueInx = 0;
	genQueryInp.condInput.len = 0;

	status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

	if ( status != 0 && status != CAT_NO_ROWS_FOUND ) {
        freeGenQueryOut(&genQueryOut);
		*err = "Error in rcGenQuery";
		return status;
	}

	getPathGenQueryResults(status, genQueryOut, columnNames, result);
    freeGenQueryOut(&genQueryOut);
    
	while ( status == 0 && cont > 0 ) {
		genQueryInp.continueInx = cont;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
		cont = genQueryOut->continueInx;

		getPathGenQueryResults(status, genQueryOut, columnNames, result);
        freeGenQueryOut(&genQueryOut);
	}

	// Clean up cmdToken strings
	tokenIndex--;
	for ( ; tokenIndex >= 2; tokenIndex-- ) {
		free(cmdToken[tokenIndex]);
	}

	return 0;
}

int gorods_query_collection(rcComm_t* conn, char* query, goRodsPathResult_t* result, char** err) {

	char* cmdToken[40];

    int cont;
	int i;
	for ( i = 0; i < 40; i++ ) {
		cmdToken[i] = "";
	}

	cmdToken[0] = "qu";
	cmdToken[1] = "-C";

	int tokenIndex = 2;

	build_cmd_token(cmdToken, &tokenIndex, query);

	genQueryInp_t genQueryInp;
	genQueryOut_t *genQueryOut;
	int i1a[20];
	int i1b[20];
	int i2a[20];
	char *condVal[20];
	char v1[BIG_STR];
	char v2[BIG_STR];
	char v3[BIG_STR];
	int status;
	char *columnNames[] = {"collection"};
	int cmdIx;
	int condIx;
	char vstr[20][BIG_STR];

	memset(&genQueryInp, 0, sizeof(genQueryInp_t));

	i1a[0] = COL_COLL_NAME;
	i1b[0] = 0;  /* (unused) */
	genQueryInp.selectInp.inx = i1a;
	genQueryInp.selectInp.value = i1b;
	genQueryInp.selectInp.len = 1;

	i2a[0] = COL_META_COLL_ATTR_NAME;
	snprintf(v1, sizeof(v1), "='%s'", cmdToken[2]);
	condVal[0] = v1;

	i2a[1] = COL_META_COLL_ATTR_VALUE;
	snprintf(v2, sizeof(v2), "%s '%s'", cmdToken[3], cmdToken[4]);
	condVal[1] = v2;

	genQueryInp.sqlCondInp.inx = i2a;
	genQueryInp.sqlCondInp.value = condVal;
	genQueryInp.sqlCondInp.len = 2;

	if ( strcmp(cmdToken[5], "or") == 0 ) {
		snprintf(v3, sizeof(v3), "|| %s '%s'", cmdToken[6], cmdToken[7]);
		rstrcat(v2, v3, BIG_STR);
	}

	cmdIx = 5;
	condIx = 2;
	while ( strcmp(cmdToken[cmdIx], "and") == 0 ) {
		i2a[condIx] = COL_META_COLL_ATTR_NAME;
		cmdIx++;
		snprintf(vstr[condIx], BIG_STR, "='%s'", cmdToken[cmdIx]);
		condVal[condIx] = vstr[condIx];
		condIx++;

		i2a[condIx] = COL_META_COLL_ATTR_VALUE;
		snprintf(vstr[condIx], BIG_STR, "%s '%s'", cmdToken[cmdIx+1], cmdToken[cmdIx+2]);
		cmdIx += 3;
		condVal[condIx] = vstr[condIx];
		condIx++;
		genQueryInp.sqlCondInp.len += 2;
	}

	if ( *cmdToken[cmdIx] != '\0' ) {
		*err = "Unrecognized input\n";
		return -1;
	}

	genQueryInp.maxRows = 10;
	genQueryInp.continueInx = 0;
	genQueryInp.condInput.len = 0;

	status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

	if ( status != 0 && status != CAT_NO_ROWS_FOUND ) {
		*err = "Error in rcGenQuery";
        freeGenQueryOut(&genQueryOut);
		return status;
	}

	getPathGenQueryResults(status, genQueryOut, columnNames, result);
    freeGenQueryOut(&genQueryOut);

	while ( status == 0 && cont > 0 ) {
		genQueryInp.continueInx = cont;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;

		getPathGenQueryResults(status, genQueryOut, columnNames, result);
        freeGenQueryOut(&genQueryOut);
	}


	// Clean up cmdToken strings
	tokenIndex--;
	for ( ; tokenIndex >= 2; tokenIndex-- ) {
		free(cmdToken[tokenIndex]);
	}

	return 0;
}

void build_cmd_token(char** cmdToken, int* tokenIndex, char* query) {

	int queryStringLen = strlen(query);
	
	char token[255] = "";
	int  n;

	int inString = 0;
	int ignoreNext = 0;

	// Build cmdToken array from query (string) input
	for ( n = 0; n <= queryStringLen; n++ ) {
		char c = query[n];

		if ( *tokenIndex == 40 ) {
			break;
		}
		
		if ( !inString && (c == '\'' || c == '"') ) {
			inString = 1;
			continue;
		}

		if ( inString && c == '\\' ) {
			ignoreNext = 1;
			continue;
		}

		if ( ignoreNext ) {
			ignoreNext = 0;
		} else {
			if ( inString && (c == '\'' || c == '"') ) {
				inString = 0;
				continue;
			}
		}

		// Did we find a space?
		if ( !inString && (c == ' ' || c == '\0') ) { // Yes, set cmdToken element, reset token
			cmdToken[*tokenIndex] = gorods_malloc(strlen(token) + 1);
			
			memcpy(cmdToken[*tokenIndex], token, strlen(token) + 1);
			memset(&token[0], 0, sizeof(token));
	
			(*tokenIndex)++;

		} else { // No, keep building token
			if ( strlen(token) == 250 ) continue;

			strncat(token, &c, 1);
		}
	}

}

char** expandGoRodsPathResult(goRodsPathResult_t* result, int length) {
	int newSize = result->size + length;

	result->pathArr = realloc(result->pathArr, sizeof(char*) * newSize);
	result->size = newSize;

	char** newItem = &(result->pathArr[newSize - 1]);

	return newItem;
}

void getPathGenQueryResults(int status, genQueryOut_t *genQueryOut, char *descriptions[], goRodsPathResult_t* result) {
	int i, j;

	if ( status != CAT_NO_ROWS_FOUND ) {
		for ( i = 0; i < genQueryOut->rowCnt; i++ ) {

			// This would be dataobjs + collection for dataobj
			if ( strcmp(descriptions[0], "collection") == 0 ) {
				char *tResult;
				tResult = genQueryOut->sqlResult[0].value;
				tResult += i * genQueryOut->sqlResult[0].len;

				char** item = expandGoRodsPathResult(result, 1);

				*item = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
			} else { 
				char *dobj;
				dobj = genQueryOut->sqlResult[1].value;
				dobj += i * genQueryOut->sqlResult[1].len;

				char *col;
				col = genQueryOut->sqlResult[0].value;
				col += i * genQueryOut->sqlResult[0].len;

				char** item = expandGoRodsPathResult(result, 1);

				*item = gorods_malloc(strlen(col) + strlen(dobj) + 2);

				strcpy(*item, col);
				strcat(*item, "/");
				strcat(*item, dobj);
			}

		}
	}
}

void freeGoRodsPathResult(goRodsPathResult_t* result) {
	int n;

	for ( n = 0; n < result->size; n++ ) {
		free(result->pathArr[n]);
	}

	free(result->pathArr);
}

// int gorods_query_dataobj(char *cmdToken[]) {
// 	return 0;
// }

// int gorods_query_user( char *cmdToken[] ) {

// }

// int gorods_query_resc( char *cmdToken[] ) {

// }

int gorodsFreeCollEnt( collEnt_t *collEnt ) {
    if ( collEnt == NULL ) {
        return 0;
    }

    gorodsclearCollEnt( collEnt );

    free( collEnt );

    return 0;
}

int gorodsclearCollEnt( collEnt_t *collEnt ) {
    if ( collEnt == NULL ) {
        return 0;
    }

    if ( collEnt->collName != NULL ) {
        free( collEnt->collName );
    }
    if ( collEnt->dataName != NULL ) {
        free( collEnt->dataName );
    }
    if ( collEnt->dataId != NULL ) {
        free( collEnt->dataId );
    }
    if ( collEnt->createTime != NULL ) {
        free( collEnt->createTime );
    }
    if ( collEnt->modifyTime != NULL ) {
        free( collEnt->modifyTime );
    }
    if ( collEnt->chksum != NULL ) {
        free( collEnt->chksum );
    }
    if ( collEnt->resource != NULL ) {
        free( collEnt->resource );
    }
    if ( collEnt->phyPath != NULL ) {
        free( collEnt->phyPath );
    }
    if ( collEnt->ownerName != NULL ) {
        free( collEnt->ownerName );
    }
    if ( collEnt->dataType != NULL ) {
       // free( collEnt->dataType );    // JMC - backport 4636
    }
    return 0;
}


// Free memory from meta result set
void freeGoRodsMetaResult(goRodsMetaResult_t* result) {
	int n;

	for ( n = 0; n < result->size; n++ ) {
		free(result->metaArr[n].name);
		free(result->metaArr[n].value);
		free(result->metaArr[n].units);
	}

	free(result->metaArr);
}

goRodsMeta_t* expandGoRodsMetaResult(goRodsMetaResult_t* result, int length) {
	int newSize = result->size + length;

	result->metaArr = realloc(result->metaArr, sizeof(goRodsMeta_t) * newSize);
	result->size = newSize;

	goRodsMeta_t* newItem = &(result->metaArr[newSize - 1]);

	return newItem;
}


void setGoRodsMeta(genQueryOut_t *genQueryOut, char *descriptions[], goRodsMetaResult_t* result) {

	int i, j;

	for ( i = 0; i < genQueryOut->rowCnt; i++ ) {
		
		goRodsMeta_t* lastItem = expandGoRodsMetaResult(result, 1);
		
		for ( j = 0; j < genQueryOut->attriCnt; j++ ) {
			char *tResult;

			tResult = genQueryOut->sqlResult[j].value;
			tResult += i * genQueryOut->sqlResult[j].len;

			if ( *descriptions[j] != '\0' ) {

				if ( strcmp(descriptions[j], "attribute") == 0 ) {
					lastItem->name = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
				}

				if ( strcmp(descriptions[j], "value") == 0 ) {
					lastItem->value = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
				}

				if ( strcmp(descriptions[j], "units") == 0 ) {
					lastItem->units = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
				}
			}
		}
	}

}

int gorods_rm(char* path, int isCollection, int recursive, int force, int trash, rcComm_t* conn, char** err) {

	if ( isCollection > 0 ) {
		collInp_t collInp;
		memset(&collInp, 0, sizeof(collInp_t));

		if ( force > 0 ) {
			addKeyVal(&collInp.condInput, FORCE_FLAG_KW, "");
		}

		if ( recursive > 0 ) {
			addKeyVal(&collInp.condInput, RECURSIVE_OPR__KW, "");
		} else {
			*err = "Recursive flag must be used on collections\n";
			return USER__NULL_INPUT_ERR;
		}

        if ( trash > 0 ) {
            addKeyVal(&collInp.condInput, RMTRASH_KW, "");
        }

		rstrcpy(collInp.collName, path, MAX_NAME_LEN);
		
		int status = rcRmColl(conn, &collInp, 0);

		return status;

	} else {
		dataObjInp_t dataObjInp;
		memset(&dataObjInp, 0, sizeof(dataObjInp_t ));

		if ( force > 0 ) {
			addKeyVal(&dataObjInp.condInput, FORCE_FLAG_KW, "");
		}

		if ( recursive > 0 ) {
			addKeyVal(&dataObjInp.condInput, RECURSIVE_OPR__KW, "");
		}

        if ( trash > 0 ) {
            addKeyVal(&dataObjInp.condInput, RMTRASH_KW, "");
        }

		dataObjInp.openFlags = O_RDONLY;
        dataObjInp.numThreads = conn->transStat.numThreads;

		rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN);

    	int status = rcDataObjUnlink(conn, &dataObjInp);

    	return status;
	}

}

int gorods_meta_dataobj(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err) {
    char zoneArgument[MAX_NAME_LEN + 2] = "";
    char *attrName = ""; // Get all attributes?
    // End global vars
    
    genQueryInp_t genQueryInp;
    genQueryOut_t *genQueryOut;
    int i1a[10];
    int i1b[10];
    int i2a[10];
    char *condVal[10];
    char v1[BIG_STR];
    char v2[BIG_STR];
    char v3[BIG_STR];
    char fullName[MAX_NAME_LEN];
    char myDirName[MAX_NAME_LEN];
    char myFileName[MAX_NAME_LEN];
    int status;

    /* "id" only used in testMode, in longMode id is reset to be 'time set' :*/
    char *columnNames[] = {"attribute", "value", "units", "id"};

    memset(result, 0, sizeof(goRodsMetaResult_t));
    memset(&genQueryInp, 0, sizeof(genQueryInp_t));

    i1a[0] = COL_META_DATA_ATTR_NAME;
    i1b[0] = 0;
    i1a[1] = COL_META_DATA_ATTR_VALUE;
    i1b[1] = 0;
    i1a[2] = COL_META_DATA_ATTR_UNITS;
    i1b[2] = 0;


    genQueryInp.selectInp.inx = i1a;
    genQueryInp.selectInp.value = i1b;
    genQueryInp.selectInp.len = 3;


    i2a[0] = COL_COLL_NAME;
    sprintf(v1, "='%s'", cwd);
    condVal[0] = v1;

    i2a[1] = COL_DATA_NAME;
    sprintf(v2, "='%s'", name);
    condVal[1] = v2;

    strncpy(fullName, cwd, MAX_NAME_LEN);
    rstrcat(fullName, "/", MAX_NAME_LEN);
    rstrcat(fullName, name, MAX_NAME_LEN);


    if ( strstr(name, "/") != NULL ) {
        /* reset v1 and v2 for when full path or relative path entered */
        if ( *name == '/' ) {
            strncpy(fullName, name, MAX_NAME_LEN);
        }
        status = splitPathByKey(fullName, myDirName, 255, myFileName, 255, '/');
        
        sprintf(v1, "='%s'", myDirName);
        sprintf(v2, "='%s'", myFileName);
    }

    genQueryInp.sqlCondInp.inx = i2a;
    genQueryInp.sqlCondInp.value = condVal;
    genQueryInp.sqlCondInp.len = 2;

    if ( attrName != NULL && *attrName != '\0' ) {
        i2a[2] = COL_META_DATA_ATTR_NAME;

        sprintf(v3, "= '%s'", attrName);

        condVal[2] = v3;
        genQueryInp.sqlCondInp.len++;
    }

    genQueryInp.maxRows = 10;
    genQueryInp.continueInx = 0;
    genQueryInp.condInput.len = 0;

    if ( zoneArgument[0] != '\0' ) {
        addKeyVal(&genQueryInp.condInput, ZONE_KW, zoneArgument);
    }

    int cont;

    status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

    if ( status == CAT_NO_ROWS_FOUND ) {

        freeGenQueryOut(&genQueryOut);

        i1a[0] = COL_D_DATA_PATH;
        genQueryInp.selectInp.len = 1;

        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;

        if ( status == 0 ) {
            *err = "No rows found";
            freeGenQueryOut(&genQueryOut);
            return CAT_NO_ROWS_FOUND;
        }

        if ( status == CAT_NO_ROWS_FOUND ) {
            *err = "Object does not exist.\n";
            freeGenQueryOut(&genQueryOut);
            return status;
        }
    }

    setGoRodsMeta(genQueryOut, columnNames, result); 
    freeGenQueryOut(&genQueryOut);

    while ( status == 0 && cont > 0 ) {

        genQueryInp.continueInx = cont;

        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;

        setGoRodsMeta(genQueryOut, columnNames, result);
        freeGenQueryOut(&genQueryOut);
    }

    

    return 0;
}

int gorods_meta_collection(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err) {
	char *attrName = ""; // Get all attributes?
    char zoneArgument[MAX_NAME_LEN + 2] = "";

    genQueryInp_t genQueryInp;
	genQueryOut_t *genQueryOut;
	int i1a[10];
	int i1b[10];
	int i2a[10];
	char *condVal[10];
	char v1[BIG_STR];
	char v2[BIG_STR];
	char fullName[MAX_NAME_LEN];
	int  status;

	char *columnNames[] = {"attribute", "value", "units"};
	
    memset(result, 0, sizeof(goRodsMetaResult_t));
	memset(&genQueryInp, 0, sizeof(genQueryInp_t));

	i1a[0] = COL_META_COLL_ATTR_NAME;
	i1b[0] = 0; /* currently unused */
	i1a[1] = COL_META_COLL_ATTR_VALUE;
	i1b[1] = 0;
	i1a[2] = COL_META_COLL_ATTR_UNITS;
	i1b[2] = 0;
	genQueryInp.selectInp.inx = i1a;
	genQueryInp.selectInp.value = i1b;
	genQueryInp.selectInp.len = 3;

	strncpy(fullName, cwd, MAX_NAME_LEN);
	if ( strlen(name) > 0 ) {
		if ( *name == '/' ) {
			strncpy(fullName, name, MAX_NAME_LEN);
		}
		else {
			rstrcat(fullName, "/", MAX_NAME_LEN);
			rstrcat(fullName, name, MAX_NAME_LEN);
		}
	}
	i2a[0] = COL_COLL_NAME;
	snprintf(v1, sizeof(v1), "='%s'", fullName);
	condVal[0] = v1;

	genQueryInp.sqlCondInp.inx = i2a;
	genQueryInp.sqlCondInp.value = condVal;
	genQueryInp.sqlCondInp.len = 1;

	if ( attrName != NULL && *attrName != '\0' ) {
		i2a[1] = COL_META_COLL_ATTR_NAME;
		
		snprintf(v2, sizeof(v2), "= '%s'", attrName);
		
		condVal[1] = v2;
		genQueryInp.sqlCondInp.len++;
	}

	genQueryInp.maxRows = 10;
	genQueryInp.continueInx = 0;
	genQueryInp.condInput.len = 0;

	if ( zoneArgument[0] != '\0' ) {
		addKeyVal(&genQueryInp.condInput, ZONE_KW, zoneArgument);
	}

    int cont;

	status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

	if ( status == CAT_NO_ROWS_FOUND ) {
		freeGenQueryOut(&genQueryOut);

        i1a[0] = COL_COLL_COMMENTS;
		genQueryInp.selectInp.len = 1;
		genQueryInp.sqlCondInp.len = 1;
		
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;
		
		if ( status == 0 ) {
			*err = "No rows found";
            freeGenQueryOut(&genQueryOut);
			return CAT_NO_ROWS_FOUND;
		}

		if ( status == CAT_NO_ROWS_FOUND ) {
			*err = "Collection does not exist.\n";
            freeGenQueryOut(&genQueryOut);
			return status;
		}
	}

	setGoRodsMeta(genQueryOut, columnNames, result); 
    freeGenQueryOut(&genQueryOut);

	while ( status == 0 && cont > 0 ) {

		genQueryInp.continueInx = cont;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;

		setGoRodsMeta(genQueryOut, columnNames, result);
        freeGenQueryOut(&genQueryOut);
	}

    

	return 0;
}

int gorods_meta_user(char *name, char *zone, goRodsMetaResult_t* result, rcComm_t* conn, char** err) {
    char zoneArgument[MAX_NAME_LEN + 2] = "";

    genQueryOut_t *genQueryOut;
    int i1a[10];
    int i1b[10];
    int i2a[10];
    char *condVal[10];
    int status;
    char *columnNames[] = {"attribute", "value", "units"};

    char userName[NAME_LEN];
    char userZone[NAME_LEN];

    status = parseUserName( name, userName, userZone );
    if ( status ) {
        *err = "Invalid username format";
        return -1;
    }

    if ( userZone[0] == '\0' ) {
        snprintf( userZone, sizeof( userZone ), "%s", zone );
    }

    genQueryInp_t genQueryInp;
    memset(&genQueryInp, 0, sizeof(genQueryInp));
    memset(result, 0, sizeof(goRodsMetaResult_t));
    
    i1a[0] = COL_META_USER_ATTR_NAME;
    i1b[0] = 0; /* currently unused */
    i1a[1] = COL_META_USER_ATTR_VALUE;
    i1b[1] = 0;
    i1a[2] = COL_META_USER_ATTR_UNITS;
    i1b[2] = 0;
    genQueryInp.selectInp.inx = i1a;
    genQueryInp.selectInp.value = i1b;
    genQueryInp.selectInp.len = 3;

    char v1[MAX_NAME_LEN];
    char v2[MAX_NAME_LEN];

    i2a[0] = COL_USER_NAME;

    strcpy(v1, "='");
    strcat(v1, userName);
    strcat(v1, "'");

    i2a[1] = COL_USER_ZONE;

    strcpy(v2, "='");
    strcat(v2, userZone);
    strcat(v2, "'");

    genQueryInp.sqlCondInp.inx = i2a;
    genQueryInp.sqlCondInp.value = condVal;
    genQueryInp.sqlCondInp.len = 2;

    condVal[0] = v1;
    condVal[1] = v2;

    genQueryInp.maxRows = 10;
    genQueryInp.continueInx = 0;
    genQueryInp.condInput.len = 0;

    if ( zoneArgument[0] != '\0' ) {
        addKeyVal( &genQueryInp.condInput, ZONE_KW, zoneArgument );
    }

    int cont;

    status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
    cont = genQueryOut->continueInx;

    if ( status == CAT_NO_ROWS_FOUND ) {
        freeGenQueryOut(&genQueryOut);

        i1a[0] = COL_R_RESC_INFO;
        genQueryInp.selectInp.len = 1;
        genQueryInp.sqlCondInp.len = 1;
        
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;
        
        if ( status == 0 ) {
            *err = "No rows found";
            freeGenQueryOut(&genQueryOut);
            return CAT_NO_ROWS_FOUND;
        }

        if ( status == CAT_NO_ROWS_FOUND ) {
            *err = "User does not exist.\n";
            freeGenQueryOut(&genQueryOut);
            return status;
        }
    }

    setGoRodsMeta(genQueryOut, columnNames, result); 
    freeGenQueryOut(&genQueryOut);

    while ( status == 0 && cont > 0 ) {

        genQueryInp.continueInx = cont;
        status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
        cont = genQueryOut->continueInx;

        setGoRodsMeta(genQueryOut, columnNames, result);
        freeGenQueryOut(&genQueryOut);
    }


    return 0;
}

int gorods_mod_meta(char* type, char* path, char* oa, char* ov, char* ou, char* na, char* nv, char* nu, rcComm_t* conn, char** err) {

	if ( strlen(na) >= 252 || strlen(nv) >= 252 || strlen(nu) >= 252 ) {
		*err = "Attribute, Value, or Unit string length too long";
		return -1;
	}

	modAVUMetadataInp_t modAVUMetadataInp;

	char typeArg[255] = "-";
	modAVUMetadataInp.arg1 = strcat(typeArg, type);
    modAVUMetadataInp.arg0 = "mod";
    modAVUMetadataInp.arg2 = path;
    modAVUMetadataInp.arg3 = oa;
    modAVUMetadataInp.arg4 = ov;


    if ( ou[0] == '\0' || ou == NULL ) {
	    char nameArg[255] = "n:";
	    char valueArg[255] = "v:";
	    char unitArg[255] = "u:";

	    modAVUMetadataInp.arg5 = strcat(nameArg, na);
	    modAVUMetadataInp.arg6 = strcat(valueArg, nv);
	    modAVUMetadataInp.arg7 = strcat(unitArg, nu);
	    modAVUMetadataInp.arg8 = "";
	    modAVUMetadataInp.arg9 = "";
    } else {
    	modAVUMetadataInp.arg5 = ou;

	    char nameArg[255] = "n:";
	    char valueArg[255] = "v:";
	    char unitArg[255] = "u:";

	    modAVUMetadataInp.arg6 = strcat(nameArg, na);
	    modAVUMetadataInp.arg7 = strcat(valueArg, nv);
	    modAVUMetadataInp.arg8 = strcat(unitArg, nu);
	    modAVUMetadataInp.arg9 = "";
    }

    int status = rcModAVUMetadata(conn, &modAVUMetadataInp);
    if ( status != 0 ) {
		*err = "Unable to mod metadata";
		return status;
	}

	return 0;
}

int gorods_add_meta(char* type, char* path, char* na, char* nv, char* nu, rcComm_t* conn, char** err) {

	if ( strlen(na) >= 252 || strlen(nv) >= 252 || strlen(nu) >= 252 ) {
		*err = "Attribute, Value, or Unit string length too long";
		return -1;
	}

	modAVUMetadataInp_t modAVUMetadataInp;

	char typeArg[255] = "-";
	modAVUMetadataInp.arg1 = strcat(typeArg, type);
    modAVUMetadataInp.arg0 = "add";
    modAVUMetadataInp.arg2 = path;
    modAVUMetadataInp.arg3 = na;
    modAVUMetadataInp.arg4 = nv;
	modAVUMetadataInp.arg5 = nu;
	modAVUMetadataInp.arg6 = "";
	modAVUMetadataInp.arg7 = "";
	modAVUMetadataInp.arg8 = "";
	modAVUMetadataInp.arg9 = "";


    int status = rcModAVUMetadata(conn, &modAVUMetadataInp);
    if ( status != 0 ) {
		*err = "Unable to add metadata";
		return status;
	}

	return 0;
}

int gorods_rm_meta(char* type, char* path, char* oa, char* ov, char* ou, rcComm_t* conn, char** err) {

	if ( strlen(oa) >= 252 || strlen(ov) >= 252 || strlen(ou) >= 252 ) {
		*err = "Attribute, Value, or Unit string length too long";
		return -1;
	}

	modAVUMetadataInp_t modAVUMetadataInp;

	char typeArg[255] = "-";
	modAVUMetadataInp.arg1 = strcat(typeArg, type);
    modAVUMetadataInp.arg0 = "rm";
    modAVUMetadataInp.arg2 = path;
    modAVUMetadataInp.arg3 = oa;
    modAVUMetadataInp.arg4 = ov;
    if ( ou[0] == '\0' || ou == NULL ) {
		modAVUMetadataInp.arg5 = "";
	} else {
		modAVUMetadataInp.arg5 = ou;
	}
	modAVUMetadataInp.arg6 = "";
	modAVUMetadataInp.arg7 = "";
	modAVUMetadataInp.arg8 = "";
	modAVUMetadataInp.arg9 = "";


    int status = rcModAVUMetadata(conn, &modAVUMetadataInp);
    if ( status != 0 ) {
		*err = "Unable to rm metadata";
		return status;
	}

	return 0;
}



char* irods_env_str() {
	rodsEnv myEnv;
	int status = getRodsEnv(&myEnv);
	if ( status != 0 ) {
		return (char *)"error getting env";
	}

	char* str = gorods_malloc(sizeof(char) * 255);

	sprintf(str, "\tHost: %s\n\tPort: %i\n\tUsername: %s\n\tZone: %s\n", myEnv.rodsHost, myEnv.rodsPort, myEnv.rodsUserName, myEnv.rodsZone);


	return str;
 }

 int irods_env(char** username, char** host, int* port, char** zone) {
 	rodsEnv myEnv;
	
	int status = getRodsEnv(&myEnv);
	if ( status != 0 ) {
		return status;
	}

	*username = myEnv.rodsUserName;
	*host = myEnv.rodsHost;
	*port = myEnv.rodsPort;
	*zone = myEnv.rodsZone;

	return 0;
 }

