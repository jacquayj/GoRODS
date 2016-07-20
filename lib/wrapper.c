/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

#include "wrapper.h"

#define BIG_STR 3000

void* gorods_malloc(size_t size) {
	void* mem = malloc(size);
	if ( mem == NULL ) {
		printf("GoRods error: Unable to allocate %i bytes\n", size);
		exit(1);
	}

	return mem;
}

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
    	status = clientLogin(*conn, 0, 0);
    }
    
    if ( status != 0 ) {
        *err = "clientLogin failed. Invalid password?";
        return -1;
    }

    return 0;
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
    	status = clientLogin(*conn, 0, 0);
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


int gorods_write_dataobject(int handle, void* data, int size, rcComm_t* conn, char** err) {
	
	openedDataObjInp_t dataObjWriteInp; 
	bytesBuf_t dataObjWriteOutBBuf; 

	bzero(&dataObjWriteInp, sizeof(dataObjWriteInp)); 
	bzero(&dataObjWriteOutBBuf, sizeof(dataObjWriteOutBBuf)); 

	dataObjWriteInp.l1descInx = handle;
	
	dataObjWriteOutBBuf.len = size; 
	dataObjWriteOutBBuf.buf = gorods_malloc(size); 
	
	// copy data to dataObjWriteOutBBuf.buf 
	memcpy(dataObjWriteOutBBuf.buf, data, size);
	
	int bytesWrite = rcDataObjWrite(conn, &dataObjWriteInp, &dataObjWriteOutBBuf); 
	if ( bytesWrite < 0 ) { 
		*err = "rcDataObjWrite failed";
		return -1;
	}

	free(dataObjWriteOutBBuf.buf);

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

int gorods_create_collection(char* path, rcComm_t* conn, char** err) {
	int status;

	collInp_t collCreateInp; 

	bzero(&collCreateInp, sizeof(collCreateInp));

	rstrcpy(collCreateInp.collName, path, MAX_NAME_LEN); 

	addKeyVal(&collCreateInp.condInput, RECURSIVE_OPR__KW, "");
	
	status = rcCollCreate(conn, &collCreateInp);
	if ( status < 0 ) { 
		*err = "rcCollCreate failed";
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

int gorods_close_collection(int handleInx, rcComm_t* conn, char** err) {
	int status = rcCloseCollection(conn, handleInx);

	if ( status < 0 ) { 
		*err = "rcCloseCollection failed";
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
		*err = "rcObjStat failed";
		return -1;
	}

	return 0;
}


int gorods_copy_dataobject(char* source, char* destination, rcComm_t* conn, char** err) {
	dataObjCopyInp_t dataObjCopyInp; 
	bzero(&dataObjCopyInp, sizeof(dataObjCopyInp)); 

	rstrcpy(dataObjCopyInp.destDataObjInp.objPath, destination, MAX_NAME_LEN); 
	rstrcpy(dataObjCopyInp.srcDataObjInp.objPath, source, MAX_NAME_LEN); 

	addKeyVal(&dataObjCopyInp.destDataObjInp.condInput, FORCE_FLAG_KW, ""); 
	addKeyVal(&dataObjCopyInp.destDataObjInp.condInput, REG_CHKSUM_KW, ""); 

	int status = rcDataObjCopy(conn, &dataObjCopyInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjCopy failed";
		return -1;
	}

	return 0;
}

int gorods_move_dataobject(char* source, char* destination, rcComm_t* conn, char** err) {
	dataObjCopyInp_t dataObjCopyInp; 
	bzero(&dataObjCopyInp, sizeof(dataObjCopyInp)); 

	rstrcpy(dataObjCopyInp.destDataObjInp.objPath, destination, MAX_NAME_LEN); 
	rstrcpy(dataObjCopyInp.srcDataObjInp.objPath, source, MAX_NAME_LEN); 

	int status = rcDataObjRename(conn, &dataObjCopyInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjRename failed";
		return -1;
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
	
	int status = rcDataObjUnlink(conn, &dataObjInp); 
	if ( status < 0 ) { 
		*err = "rcDataObjRename failed";
		return -1;
	}

	return 0;
}

int gorods_checksum_dataobject(char* path, char** outChksum, rcComm_t* conn, char** err) {

	dataObjInp_t dataObjInp; 

	*outChksum = NULL;

	bzero(&dataObjInp, sizeof(dataObjInp)); 
	rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN); 

	addKeyVal(&dataObjInp.condInput, FORCE_CHKSUM_KW, ""); 

	int status = rcDataObjChksum(conn, &dataObjInp, &(*outChksum)); 
	if ( status < 0 ) { 
		*err = "rcDataObjChksum failed";
		return -1;
	}

	return 0;
}

int gorods_read_collection(rcComm_t* conn, int handleInx, collEnt_t** arr, int* size, char** err) {

	int collectionResponseCapacity = 100;
	*size = 0;

	*arr = gorods_malloc(sizeof(collEnt_t) * collectionResponseCapacity);
	
	collEnt_t* collEnt = NULL;
	int status;

	while ( (status = rcReadCollection(conn, handleInx, &collEnt)) >= 0 ) { 

		// Expand array if needed
		if ( *size >= collectionResponseCapacity ) {
			collectionResponseCapacity *= 2;

			*arr = realloc(*arr, sizeof(collEnt_t) * collectionResponseCapacity);
		}
		collEnt_t* elem = &((*arr)[*size]);

		// Add element to array
		memcpy(elem, collEnt, sizeof(collEnt_t));
		if ( collEnt->objType == DATA_OBJ_T ) { 
			elem->dataName = strcpy(gorods_malloc(strlen(elem->dataName) + 1), elem->dataName);
			elem->dataId = strcpy(gorods_malloc(strlen(elem->dataId) + 1), elem->dataId);
			elem->chksum = strcpy(gorods_malloc(strlen(elem->chksum) + 1), elem->chksum);
			//elem->dataType = strcpy(gorods_malloc(strlen(elem->dataType) + 1), elem->dataType);
			elem->resource = strcpy(gorods_malloc(strlen(elem->resource) + 1), elem->resource);
			//elem->rescGrp = strcpy(gorods_malloc(strlen(elem->rescGrp) + 1), elem->rescGrp);
			elem->phyPath = strcpy(gorods_malloc(strlen(elem->phyPath) + 1), elem->phyPath);
		}

		elem->ownerName = strcpy(gorods_malloc(strlen(elem->ownerName) + 1), elem->ownerName);
		elem->collName = strcpy(gorods_malloc(strlen(elem->collName) + 1), elem->collName);
		elem->createTime = strcpy(gorods_malloc(strlen(elem->createTime) + 1), elem->createTime);
		elem->modifyTime = strcpy(gorods_malloc(strlen(elem->modifyTime) + 1), elem->modifyTime);

		(*size)++;
		
		gorodsFreeCollEnt(collEnt); 
	} 

	return 0;
}


int gorods_query_dataobj(rcComm_t* conn, char* query, goRodsPathResult_t* result, char** err) {
	
	char* cmdToken[40];

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

	if ( status != 0 && status != CAT_NO_ROWS_FOUND ) {
		*err = (char *)gorods_malloc(sizeof(char) * BIG_STR);
		sprintf(*err, "Error in rcGenQuery using '%s', check your syntax\n", query);
		return -1;
	}

	getPathGenQueryResults(status, genQueryOut, columnNames, result);

	while ( status == 0 && genQueryOut->continueInx > 0 ) {
		genQueryInp.continueInx = genQueryOut->continueInx;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
		
		getPathGenQueryResults(status, genQueryOut, columnNames, result);
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

	if ( status != 0 && status != CAT_NO_ROWS_FOUND ) {
		*err = (char *)gorods_malloc(sizeof(char) * BIG_STR);
		sprintf(*err, "Error in rcGenQuery using '%s', check your syntax\n", query);
		return -1;
	}

	getPathGenQueryResults(status, genQueryOut, columnNames, result);

	while ( status == 0 && genQueryOut->continueInx > 0 ) {
		genQueryInp.continueInx = genQueryOut->continueInx;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);

		getPathGenQueryResults(status, genQueryOut, columnNames, result);
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
			if ( descriptions[0] == "collection" ) {
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

				if ( descriptions[j] == "attribute" ) {
					lastItem->name = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
				}

				if ( descriptions[j] == "value" ) {
					lastItem->value = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
				}

				if ( descriptions[j] == "units" ) {
					lastItem->units = strcpy(gorods_malloc(strlen(tResult) + 1), tResult);
				}
			}
		}
	}

}

int gorods_rm(char* path, int isCollection, int recursive, int force, rcComm_t* conn, char** err) {

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

		dataObjInp.openFlags = O_RDONLY;

		rstrcpy(dataObjInp.objPath, path, MAX_NAME_LEN);

    	int status = rcDataObjUnlink(conn, &dataObjInp);

    	return status;
	}

}

int gorods_meta_collection(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err) {
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
	char *attrName = ""; // Get all attributes?
	char zoneArgument[MAX_NAME_LEN + 2] = "";

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

	status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
	if ( status == CAT_NO_ROWS_FOUND ) {
		i1a[0] = COL_COLL_COMMENTS;
		genQueryInp.selectInp.len = 1;
		genQueryInp.sqlCondInp.len = 1;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
		
		if ( status == 0 ) {
			*err = "No rows found";
			return CAT_NO_ROWS_FOUND;
		}

		if ( status == CAT_NO_ROWS_FOUND ) {
			*err = "Collection does not exist.\n";
			return -1;
		}
	}

	setGoRodsMeta(genQueryOut, columnNames, result); 

	while ( status == 0 && genQueryOut->continueInx > 0 ) {

		genQueryInp.continueInx = genQueryOut->continueInx;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);

		setGoRodsMeta(genQueryOut, columnNames, result); 
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
		return -1;
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
		return -1;
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
		return -1;
	}

	return 0;
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

	status = rcGenQuery(conn, &genQueryInp, &genQueryOut);
	if ( status == CAT_NO_ROWS_FOUND ) {
		i1a[0] = COL_D_DATA_PATH;
		genQueryInp.selectInp.len = 1;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);

		if ( status == 0 ) {
			*err = "No rows found";
			return CAT_NO_ROWS_FOUND;
		}

		if ( status == CAT_NO_ROWS_FOUND ) {
			*err = "Object does not exist.\n";
			return -1;
		}
	}

	setGoRodsMeta(genQueryOut, columnNames, result); 

	while ( status == 0 && genQueryOut->continueInx > 0 ) {

		genQueryInp.continueInx = genQueryOut->continueInx;
		status = rcGenQuery(conn, &genQueryInp, &genQueryOut);

		setGoRodsMeta(genQueryOut, columnNames, result); 
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
		return -1;
	}

	*username = myEnv.rodsUserName;
	*host = myEnv.rodsHost;
	*port = myEnv.rodsPort;
	*zone = myEnv.rodsZone;

	return 0;
 }

