/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/* directAccessFileDriver.h - header file for directAccessFileDriver.c
 */



#ifndef DIRECT_ACCESS_FILE_DRIVER_H
#define DIRECT_ACCESS_FILE_DRIVER_H

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
#include "unixFileDriver.h"

/* when we resize the state array, do in increments of this */
#define DIRECT_ACCESS_FILE_STATE_ARRAY_SIZE 256

typedef struct directAccessFileState {
    int fd;
    int fileMode;
} directAccessFileState_t;


int
directAccessFileCreate (rsComm_t *rsComm, char *fileName, int mode, 
                        rodsLong_t mySize, keyValPair_t *condInput);
int
directAccessFileOpen (rsComm_t *rsComm, char *fileName, int flags, int mode, 
                      keyValPair_t *condInput);
int
directAccessFileRead (rsComm_t *rsComm, int fd, void *buf, int len);
int
directAccessFileWrite (rsComm_t *rsComm, int fd, void *buf, int len);
int
directAccessFileClose (rsComm_t *rsComm, int fd);
int
directAccessFileUnlink (rsComm_t *rsComm, char *filename);
int
directAccessFileStat (rsComm_t *rsComm, char *filename, struct stat *statbuf);
int
directAccessFileFstat (rsComm_t *rsComm, int fd, struct stat *statbuf);
int
directAccessFileFsync (rsComm_t *rsComm, int fd);
int
directAccessFileMkdir (rsComm_t *rsComm, char *filename, int mode, 
                       keyValPair_t *condInput);
int
directAccessFileChmod (rsComm_t *rsComm, char *filename, int mode);
int
directAccessFileOpendir (rsComm_t *rsComm, char *dirname, void **outDirPtr);
int
directAccessFileReaddir (rsComm_t *rsComm, void *dirPtr, struct  dirent *direntPtr);
int
directAccessFileStage (rsComm_t *rsComm, char *path, int flag);
int
directAccessFileRename (rsComm_t *rsComm, char *oldFileName, char *newFileName);
rodsLong_t
directAccessFileGetFsFreeSpace (rsComm_t *rsComm, char *path, int flag);
rodsLong_t
directAccessFileLseek (rsComm_t *rsComm, int fd, rodsLong_t offset, int whence);
int
directAccessFileRmdir (rsComm_t *rsComm, char *filename);
int
directAccessFileClosedir (rsComm_t *rsComm, void *dirPtr);
int
directAccessFileTruncate (rsComm_t *rsComm, char *filename, rodsLong_t dataSize);

#endif	/* DIRECT_ACCESS_FILE_DRIVER_H */
