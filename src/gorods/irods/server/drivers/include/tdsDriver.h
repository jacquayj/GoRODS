/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* tdsDriver.h - Header file for tdsDriver.c - the tds driver */

#ifndef TDS_DRIVER_H
#define TDS_DRIVER_H

#include <curl/curl.h>
#include <libxml/xmlmemory.h>
#include <libxml/parser.h>

#define THREDDS_DIR     "/thredds/"


#define NUM_URL_PATH    10

typedef struct {
    int inuse;
    int st_mode;        /* S_IFDIR or S_IFREG */
    char path[MAX_NAME_LEN];
} urlPath_t;

typedef struct {
    int len;
    char *httpResponse;
    xmlDocPtr doc;
    xmlNodePtr rootnode;
    xmlNodePtr curnode;
    char dirUrl[MAX_NAME_LEN];
    char curdir[MAX_NAME_LEN];
    CURL *easyhandle;
    urlPath_t urlPath[NUM_URL_PATH];
} tdsDirStruct_t;

#ifdef  __cplusplus
extern "C" {
#endif

int
tdsOpendir (rsComm_t *rsComm, char *dirUrl, void **outDirPtr);
int
tdsReaddir (rsComm_t *rsComm, void *dirPtr, struct dirent *direntPtr);
int
tdsClosedir (rsComm_t *rsComm, void *dirPtr);
int
tdsStat (rsComm_t *rsComm, char *urlPath, struct stat *statbuf);
int
getNextNode (tdsDirStruct_t *tdsDirStruct);
size_t
tdsDirRespHandler (void *buffer, size_t size, size_t nmemb, void *userp);
int
freeTdsDirStruct (tdsDirStruct_t **tdsDirStruct);
int
listPydapDir (rsComm_t *rsComm, char *dirUrl);
int
setTDSUrl (tdsDirStruct_t *tdsDirStruct, char *urlPath, int isDir);
int
allocTdsUrlPath (tdsDirStruct_t *tdsDirStruct);
int
freeTdsUrlPath (tdsDirStruct_t *tdsDirStruct, int inx);
int
tdsStageToCache (rsComm_t *rsComm, fileDriverType_t cacheFileType,
int mode, int flags, char *urlPath, char *cacheFilename, rodsLong_t dataSize,
keyValPair_t *condInput);
int
setTdsCurdir (tdsDirStruct_t *tdsDirStruct, char *name);
#ifdef  __cplusplus
}
#endif

#endif  /* TDS_DRIVER_H */

