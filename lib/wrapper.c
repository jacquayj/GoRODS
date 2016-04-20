/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

#include "wrapper.h"


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
		*err = "rcDataObjOpen failed";
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
		*err = "rcDataObjRead failed";
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
			elem->dataType = strcpy(gorods_malloc(strlen(elem->dataType) + 1), elem->dataType);
			elem->resource = strcpy(gorods_malloc(strlen(elem->resource) + 1), elem->resource);
			elem->rescGrp = strcpy(gorods_malloc(strlen(elem->rescGrp) + 1), elem->rescGrp);
			elem->phyPath = strcpy(gorods_malloc(strlen(elem->phyPath) + 1), elem->phyPath);
		}

		elem->ownerName = strcpy(gorods_malloc(strlen(elem->ownerName) + 1), elem->ownerName);
		elem->collName = strcpy(gorods_malloc(strlen(elem->collName) + 1), elem->collName);
		elem->createTime = strcpy(gorods_malloc(strlen(elem->createTime) + 1), elem->createTime);
		elem->modifyTime = strcpy(gorods_malloc(strlen(elem->modifyTime) + 1), elem->modifyTime);

		(*size)++;
		
		freeCollEnt(collEnt); 
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

int gorods_meta_collection(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err) {
	genQueryInp_t genQueryInp;
	genQueryOut_t *genQueryOut;
	int i1a[10];
	int i1b[10];
	int i2a[10];
	char *condVal[10];
	char v1[3000];
	char v2[3000];
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
			*err = "None";
			return -1;
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


int gorods_meta_dataobj(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err) {
	int testMode = 0; /* some some particular internal tests */
	int longMode = 0; /* more detailed listing */
	char zoneArgument[MAX_NAME_LEN + 2] = "";
	char *attrName = ""; // Get all attributes?
	// End global vars
	
	genQueryInp_t genQueryInp;
	genQueryOut_t *genQueryOut;
	int i1a[10];
	int i1b[10];
	int i2a[10];
	char *condVal[10];
	char v1[3000];
	char v2[3000];
	char v3[3000];
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

	if ( testMode ) {
		i1a[3] = COL_META_DATA_ATTR_ID;
		i1b[3] = 0;
	}
	if ( longMode ) {
		i1a[3] = COL_META_DATA_MODIFY_TIME;
		i1b[3] = 0;
		columnNames[3] = "time set";
	}

	genQueryInp.selectInp.inx = i1a;
	genQueryInp.selectInp.value = i1b;
	genQueryInp.selectInp.len = 3;

	if ( testMode ) {
		genQueryInp.selectInp.len = 4;
	}
	if ( longMode ) {
		genQueryInp.selectInp.len = 4;
	}

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
		status = splitPathByKey(fullName, myDirName, myFileName, '/');
		
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
			*err = "None";
			return -1;
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

