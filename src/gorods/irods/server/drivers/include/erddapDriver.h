/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* erddapDriver.h - Header file for erddapDriver.c - the erddap driver */

#ifndef ERDDAP_DRIVER_H
#define ERDDAP_DRIVER_H

#include <curl/curl.h>

#ifdef  __cplusplus
extern "C" {
#endif

int
erddapOpendir (rsComm_t *rsComm, char *dirUrl, void **outDirPtr);
int
erddapReaddir (rsComm_t *rsComm, void *dirPtr, struct dirent *direntPtr);
int
erddapClosedir (rsComm_t *rsComm, void *dirPtr);
int
erddapStat (rsComm_t *rsComm, char *urlPath, struct stat *statbuf);
int
getNextHTTPlink (httpDirStruct_t *httpDirStruct, char *hlink);
int
erddapStageToCache (rsComm_t *rsComm, fileDriverType_t cacheFileType,
int mode, int flags, char *urlPath, char *cacheFilename, rodsLong_t dataSize,
keyValPair_t *condInput);
int
listErddapDir (rsComm_t *rsComm, char *dirUrl);
#ifdef  __cplusplus
}
#endif

#endif  /* ERDDAP_DRIVER_H */

