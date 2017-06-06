/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. and The BioTeam, Inc.  ***
 *** For more information please refer to the LICENSE.md file                                   ***/

#include "rods.h"
#include "rodsErrorTable.h"
#include "rodsType.h"
#include "rodsClient.h"
#include "miscUtil.h"
#include "rodsPath.h"
#include "rcConnect.h"
#include "sslSockComm.h"
#include "dataObjOpen.h"
#include "dataObjRead.h"
#include "dataObjChksum.h"
#include "dataObjClose.h"
#include "lsUtil.h"
#include <malloc.h>

typedef struct {
	int size;
	int keySize;
	char** hashKeys;
	char** hashValues;
} goRodsHashResult_t;

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
	char** strArr;
} goRodsStringResult_t;

typedef struct {
	char* name;
	char* zone;
	char* dataAccess;
	char* acltype;
} goRodsACL_t;

typedef struct {
	int size;
	goRodsACL_t* aclArr;
} goRodsACLResult_t;

typedef struct {
	int size;
	char** pathArr;
} goRodsPathResult_t;


typedef struct GoRodsQueryOpts {
    int limit;
    int offset;
} goRodsQueryOpts_t;

int gorods_rclReadCollectionCols(rcComm_t *conn, collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts);
int gorods_rclReadCollectionObjs(rcComm_t *conn, collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts);
int gorods_readCollectionObjs( collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts);
int gorods_readCollectionCols( collHandle_t *collHandle, collEnt_t *collEnt, goRodsQueryOpts_t opts);
int gorods_genCollResInColl( queryHandle_t *queryHandle, collHandle_t *collHandle, goRodsQueryOpts_t opts);
int gorods_genDataResInColl( queryHandle_t *queryHandle, collHandle_t *collHandle, goRodsQueryOpts_t opts);
int gorods_queryCollInColl( queryHandle_t *queryHandle, char *collection,
                 int flags, genQueryInp_t *genQueryInp,
                 genQueryOut_t **genQueryOut, goRodsQueryOpts_t opts );
int gorods_queryDataObjInColl( queryHandle_t *queryHandle, char *collection,
                    int flags, genQueryInp_t *genQueryInp,
                    genQueryOut_t **genQueryOut, keyValPair_t *condInput, goRodsQueryOpts_t opts );
int gorods_getNextCollMetaInfo( collHandle_t *collHandle, collEnt_t *outCollEnt );
int gorods_getNextDataObjMetaInfo( collHandle_t *collHandle, collEnt_t *outCollEnt );


void display_mallinfo(void);
void* gorods_malloc(size_t size);
int gorods_connect(rcComm_t** conn, char** err);
int gorods_connect_env(rcComm_t** conn, char* host, int port, char* username, char* zone, char** err);
int gorods_clientLoginPam(rcComm_t* conn, char* password, int ttl, char** pamPass, char** err) ;

int gorods_get_groups(rcComm_t *conn, goRodsStringResult_t* result, char** err);
void gorods_build_group_result(genQueryOut_t *genQueryOut, goRodsStringResult_t* result);
void gorods_free_string_result(goRodsStringResult_t* result);
void gorods_build_group_user_result(genQueryOut_t *genQueryOut, goRodsStringResult_t* result);
int gorods_get_group(rcComm_t *conn, goRodsStringResult_t* result, char* groupName, char** err);

int gorods_build_iquest_result(genQueryOut_t * genQueryOut, goRodsHashResult_t* result, char** err);
int gorods_iquest_general(rcComm_t *conn, char *selectConditionString, int noDistinctFlag, int upperCaseFlag, char *zoneName, goRodsHashResult_t* result, char** err);
void gorods_free_map_result(goRodsHashResult_t* result);

int gorods_get_users(rcComm_t* conn, goRodsStringResult_t* result, char** err);
int gorods_get_user(char *user, rcComm_t* conn, goRodsStringResult_t* result, char** err);
int gorods_change_user_password(char* userName, char* newPassword, char* myPassword, rcComm_t *conn, char** err);

int gorods_get_resources(rcComm_t* conn, goRodsStringResult_t* result, char** err);
int gorods_get_resource(char* rescName, rcComm_t* conn, goRodsStringResult_t* result, char** err);

int gorods_get_resource_result(rcComm_t *Conn, genQueryOut_t *genQueryOut, goRodsStringResult_t* result);
int gorods_get_resources_new(rcComm_t* conn, goRodsStringResult_t* result, char** err);

int gorods_simple_query(simpleQueryInp_t simpleQueryInp, goRodsStringResult_t* result, rcComm_t* conn, char** err);

int gorods_get_user_groups(rcComm_t *conn, char* name, goRodsStringResult_t* result, char** err);
int gorods_get_user_group_result(int status, goRodsStringResult_t* result, genQueryOut_t *genQueryOut, char *descriptions[]);

int gorods_remove_user_from_group(char* userName, char* zoneName, char* groupName, rcComm_t *conn, char** err);

int
gorods_queryDataObjAcl (rcComm_t *conn, char *dataId, char *zoneHint,
                 genQueryOut_t **genQueryOut);

int gorods_delete_group(char* groupName, char* zoneName, rcComm_t *conn, char** err);
int gorods_create_group(char* groupName, char* zoneName, rcComm_t *conn, char** err);

int gorods_get_zones(rcComm_t* conn, goRodsStringResult_t* result, char** err);
int gorods_get_zone(char* zoneName, rcComm_t* conn, goRodsStringResult_t* result, char** err);
int gorods_get_local_zone(rcComm_t* conn, char** zoneName, char** err);

int gorods_create_user(char* userName, char* zoneName, char* type, rcComm_t *conn, char** err);
int gorods_delete_user(char* userName, char* zoneName, rcComm_t *conn, char** err);

int gorods_general_admin(int userOption, char *arg0, char *arg1, char *arg2, char *arg3,
              char *arg4, char *arg5, char *arg6, char *arg7, char* arg8, char* arg9,
              rodsArguments_t* _rodsArgs, rcComm_t *conn, char** err);
int gorods_add_user_to_group(char* userName, char* zoneName, char* groupName, rcComm_t *conn, char** err);

int gorods_open_collection(char* path, int trimRepls, collHandle_t* collHandle, rcComm_t* conn, char** err);
int gorods_close_collection(collHandle_t* collHandle, char** err);
int gorods_create_collection(char* path, rcComm_t* conn, char** err);
int gorods_get_collection_acl(rcComm_t *conn, char *collName, goRodsACLResult_t* result, char* zoneHint, char** err);
int gorods_get_collection_inheritance(rcComm_t *conn, char *collName, int* enabled, char** err);

int gorods_get_dataobject(rcComm_t *conn, char *srcPath, collEnt_t* objData);
int gorods_get_dataobject_data(rcComm_t *conn, rodsArguments_t *rodsArgs, genQueryOut_t *genQueryOut, collEnt_t* objData);


int gorods_trimrepls_dataobject(rcComm_t *conn, char* objPath, char* ageStr, char* resource, char* keepCopiesStr, char** err);
int gorods_phymv_dataobject(rcComm_t *conn, char* objPath, char* sourceResource, char* destResource, char** err);
int gorods_repl_dataobject(rcComm_t *conn, char* objPath, char* resourceName, int backupMode, int createMode, rodsLong_t dataSize, char** err);
int gorods_put_dataobject(char* inPath, char* outPath, rodsLong_t size, int mode, int force, char* resource, rcComm_t* conn, char** err);
int gorods_open_dataobject(char* path, char* resourceName, char* replNum, int openFlag, int* handle, rcComm_t* conn, char** err);
int gorods_read_dataobject(int handleInx, rodsLong_t length, bytesBuf_t* buffer, int* bytesRead, rcComm_t* conn, char** err);
int gorods_lseek_dataobject(int handleInx, rodsLong_t offset, rcComm_t* conn, char** err);
int gorods_close_dataobject(int handleInx, rcComm_t* conn, char** err);
int gorods_stat_dataobject(char* path, rodsObjStat_t** rodsObjStatOut, rcComm_t* conn, char** err);
int gorods_create_dataobject(char* path, rodsLong_t size, int mode, int force, char* resource, int* handle, rcComm_t* conn, char** err);
int gorods_write_dataobject(int handle, void* data, int size, rcComm_t* conn, char** err);
int gorods_copy_dataobject(char* source, char* destination, int force, char* resource, rcComm_t* conn, char** err);
int gorods_move_dataobject(char* source, char* destination, int objType, rcComm_t* conn, char** err);
int gorods_unlink_dataobject(char* path, int force, rcComm_t* conn, char** err);
int gorods_checksum_dataobject(char* path, char** outChksum, rcComm_t* conn, char** err);
int gorods_rm(char* path, int isCollection, int recursive, int force, rcComm_t* conn, char** err);
int gorods_get_dataobject_acl(rcComm_t* conn, char* dataId, goRodsACLResult_t* result, char* zoneHint, char** err);
void gorods_free_acl_result(goRodsACLResult_t* result);

int gorods_chmod(rcComm_t *conn, char* path, char* zone, char* ugName, char* accessLevel, int recursive, char** err);


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
