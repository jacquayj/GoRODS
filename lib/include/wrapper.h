/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

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

typedef struct {
	char* name;
	char* value;
	char* units;
} goRodsMeta_t;

typedef struct {
	int size;
	goRodsMeta_t* metaArr;
} goRodsMetaResult_t;

typedef struct {
	int size;
	char** pathArr;
} goRodsPathResult_t;

void* gorods_malloc(size_t size);
int gorods_connect(rcComm_t** conn, char* password, char** err);
int gorods_connect_env(rcComm_t** conn, char* host, int port, char* username, char* zone, char* password, char** err);

int gorods_open_collection(char* path, int* collHandle, rcComm_t* conn, char** err);
int gorods_read_collection(rcComm_t* conn, int handleInx, collEnt_t** arr, int* size, char** err);
int gorods_close_collection(int handleInx, rcComm_t* conn, char** err);
int gorods_create_collection(char* path, rcComm_t* conn, char** err);

int gorods_open_dataobject(char* path, int* handle, rodsLong_t length, rcComm_t* conn, char** err);
int gorods_read_dataobject(int handleInx, rodsLong_t length, bytesBuf_t* buffer, rcComm_t* conn, char** err);
int gorods_lseek_dataobject(int handleInx, rodsLong_t offset, rcComm_t* conn, char** err);
int gorods_close_dataobject(int handleInx, rcComm_t* conn, char** err);
int gorods_stat_dataobject(char* path, rodsObjStat_t** rodsObjStatOut, rcComm_t* conn, char** err);
int gorods_create_dataobject(char* path, rodsLong_t size, int mode, int force, char* resource, int* handle, rcComm_t* conn, char** err);
int gorods_write_dataobject(int handle, void* data, int size, rcComm_t* conn, char** err);
int gorods_copy_dataobject(char* source, char* destination, rcComm_t* conn, char** err);
int gorods_move_dataobject(char* source, char* destination, rcComm_t* conn, char** err);
int gorods_unlink_dataobject(char* path, int force, rcComm_t* conn, char** err);
int gorods_checksum_dataobject(char* path, char** outChksum, rcComm_t* conn, char** err);
int gorods_rm(char* path, int isCollection, int recursive, int force, rcComm_t* conn, char** err);

void setGoRodsMeta(genQueryOut_t *genQueryOut, char *descriptions[], goRodsMetaResult_t* result);
void freeGoRodsMetaResult(goRodsMetaResult_t* result);
goRodsMeta_t* expandGoRodsMetaResult(goRodsMetaResult_t* result, int length);

int gorods_meta_dataobj(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err);
int gorods_meta_collection(char *name, char *cwd, goRodsMetaResult_t* result, rcComm_t* conn, char** err);
int gorods_mod_meta(char* type, char* path, char* oa, char* ov, char* ou, char* na, char* nv, char* nu, rcComm_t* conn, char** err);
int gorods_add_meta(char* type, char* path, char* na, char* nv, char* nu, rcComm_t* conn, char** err);
int gorods_rm_meta(char* type, char* path, char* oa, char* ov, char* ou, rcComm_t* conn, char** err);
int gorods_set_session_ticket(rcComm_t *myConn, char *ticket, char** err);

int gorods_query_collection(rcComm_t* conn, char* query, goRodsPathResult_t* result, char** err);
int gorods_query_dataobj(rcComm_t* conn, char* query, goRodsPathResult_t* result, char** err);

void getPathGenQueryResults(int status, genQueryOut_t *genQueryOut, char *descriptions[], goRodsPathResult_t* result);
void freeGoRodsPathResult(goRodsPathResult_t* result);
void build_cmd_token(char** cmdToken, int* tokenIndex, char* query);

int gorodsclearCollEnt( collEnt_t *collEnt );
int gorodsFreeCollEnt( collEnt_t *collEnt );
char* irods_env_str();
int irods_env(char** username, char** host, int* port, char** zone);
