/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* pydapDriver.h - Header file for pydapDriver.c - the pydap driver */

#ifndef PYDAP_DRIVER_H
#define PYDAP_DRIVER_H

#include <curl/curl.h>

#define HLINK_PREFIX            "<a href="
#define PARENT_HLINK_DIR        "../"
typedef struct {
    int len;
    char *httpResponse;
    char *curPtr;
    CURL *easyhandle;
} httpDirStruct_t;

typedef struct {
    rodsLong_t len;
    int outFd;
    int mode;
    char outfile[MAX_NAME_LEN];
} httpDownloadStruct_t;

#ifdef  __cplusplus
extern "C" {
#endif

int
pydapOpendir (rsComm_t *rsComm, char *dirUrl, void **outDirPtr);
int
pydapReaddir (rsComm_t *rsComm, void *dirPtr, struct dirent *direntPtr);
int
pydapClosedir (rsComm_t *rsComm, void *dirPtr);
int
pydapStat (rsComm_t *rsComm, char *urlPath, struct stat *statbuf);
int
getNextHlink (httpDirStruct_t *httpDirStruct, char *hlink);
int
freeHttpDirStruct (httpDirStruct_t **httpDirStruct);
size_t
httpDirRespHandler (void *buffer, size_t size, size_t nmemb, void *userp);
int
listPydapDir (rsComm_t *rsComm, char *dirUrl);
int
pydapStageToCache (rsComm_t *rsComm, fileDriverType_t cacheFileType,
int mode, int flags, char *urlPath, char *cacheFilename, rodsLong_t dataSize,
keyValPair_t *condInput);
int
httpDownloadFunc (void *buffer, size_t size, size_t nmemb, void *userp);
#ifdef  __cplusplus
}
#endif

#endif  /* PYDAP_DRIVER_H */

