/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/* hdfsFileDriver.h - header file for hdfsFileDriver.c */



#ifndef HDFS_FILE_DRIVER_H
#define HDFS_FILE_DRIVER_H

#include <stdio.h>
#ifndef _WIN32
#include <sys/file.h>
#include <sys/param.h>
#endif
#include <errno.h>
#include <sys/stat.h>
#include <string.h>
#ifndef _WIN32
#include <unistd.h>
#endif
#include <sys/types.h>
#if defined(osx_platform)
#include <sys/malloc.h>
#else
#include <malloc.h>
#endif
#include <fcntl.h>
#ifndef _WIN32
#include <sys/file.h>
#include <unistd.h>  
#endif
#include <dirent.h>
   
#if defined(solaris_platform)
#include <sys/statvfs.h>
#endif
#if defined(linux_platform)
#include <sys/vfs.h>
#endif
#if defined(aix_platform) || defined(sgi_platform)
#include <sys/statfs.h>
#endif
#if defined(osx_platform)
#include <sys/param.h>
#include <sys/mount.h>
#endif
#include <sys/stat.h>



#include "rods.h"
#include "rcConnect.h"
#include "msParam.h"
#include "miscServerFunct.h"

#define NB_READ_TOUT_SEC	60	/* 60 sec timeout */
#define NB_WRITE_TOUT_SEC	60	/* 60 sec timeout */

int
hdfsFileCreate (rsComm_t *rsComm, char *fileName, int mode, rodsLong_t mySize, keyValPair_t *condInput);
int
hdfsFileOpen (rsComm_t *rsComm, char *fileName, int flags, int mode, keyValPair_t *condInput);
int
hdfsFileRead (rsComm_t *rsComm, int fd, void *buf, int len);
int
hdfsFileWrite (rsComm_t *rsComm, int fd, void *buf, int len);
int
hdfsFileClose (rsComm_t *rsComm, int fd);
int
hdfsFileUnlink (rsComm_t *rsComm, char *filename);
int
hdfsFileStat (rsComm_t *rsComm, char *filename, struct stat *statbuf);
int
hdfsFileFstat (rsComm_t *rsComm, int fd, struct stat *statbuf);
int
hdfsFileFsync (rsComm_t *rsComm, int fd);
int
hdfsFileMkdir (rsComm_t *rsComm, char *filename, int mode, keyValPair_t *condInput);
int
hdfsFileChmod (rsComm_t *rsComm, char *filename, int mode);
int
hdfsFileOpendir (rsComm_t *rsComm, char *dirname, void **outDirPtr);
int
hdfsFileReaddir (rsComm_t *rsComm, void *dirPtr, struct  dirent *direntPtr);
int
hdfsFileStage (rsComm_t *rsComm, char *path, int flag);
int
hdfsFileRename (rsComm_t *rsComm, char *oldFileName, char *newFileName);
rodsLong_t
hdfsFileGetFsFreeSpace (rsComm_t *rsComm, char *path, int flag);
rodsLong_t
hdfsFileLseek (rsComm_t *rsComm, int fd, rodsLong_t offset, int whence);
int
hdfsFileRmdir (rsComm_t *rsComm, char *filename);
int
hdfsFileClosedir (rsComm_t *rsComm, void *dirPtr);
int
hdfsFileTruncate (rsComm_t *rsComm, char *filename, rodsLong_t dataSize);
int
hdfsStageToCache (rsComm_t *rsComm, fileDriverType_t cacheFileType,
int mode, int flags, char *filename,
char *cacheFilename, rodsLong_t dataSize,
keyValPair_t *condInput);
int
hdfsSyncToArch (rsComm_t *rsComm, fileDriverType_t cacheFileType,
int mode, int flags, char *filename,
char *cacheFilename,  rodsLong_t dataSize,
keyValPair_t *condInput);
int
hdfsFileCopy (int mode, char *srcFileName, char *destFileName);
int
hdfsnbFileRead (rsComm_t *rsComm, int fd, void *buf, int len);
int
hdfsnbFileWrite (rsComm_t *rsComm, int fd, void *buf, int len);

#endif	/* HDFS_FILE_DRIVER_H */
