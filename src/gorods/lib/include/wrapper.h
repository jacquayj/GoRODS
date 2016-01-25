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
int gorods_open_collection(char* path, int* collHandle, rcComm_t* conn, char** err);
int gorods_read_collection(rcComm_t* conn, int handleInx, collEnt_t** arr, int* size, char** err);
char* irods_env_str();
