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

int gorods_connect(rcComm_t** conn, char* password, char** err);
int gorods_connect_env(rcComm_t** conn, char* host, int port, char* username, char* zone, char* password, char** err);
char* irods_env_str();

int gorods_open_collection(char* path, int* collHandle, rcComm_t* conn, char** err);
int gorods_read_collection(rcComm_t* conn, int handleInx, collEnt_t** arr, int* size, char** err);
int gorods_close_collection(int handleInx, rcComm_t* conn, char** err);

int gorods_open_dataobject(char* path, int* handle, rodsLong_t length, rcComm_t* conn, char** err);
int gorods_read_dataobject(int handleInx, rodsLong_t length, bytesBuf_t* buffer, rcComm_t* conn, char** err);
int gorods_close_dataobject(int handleInx, rcComm_t* conn, char** err);
int gorods_stat_dataobject(char* path, rodsObjStat_t** rodsObjStatOut, rcComm_t* conn, char** err);
int gorods_create_dataobject(char* path, rodsLong_t size, int mode, int force, char* resource, int* handle, rcComm_t* conn, char** err);
int gorods_write_dataobject(int handle, void* data, int size, rcComm_t* conn, char** err);