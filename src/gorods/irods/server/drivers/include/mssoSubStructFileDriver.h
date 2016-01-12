/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to structFiles in the COPYRIGHT directory ***/

/* mssoSubStructFileDriver.h - header structFile for mssoSubStructFileDriver.c
 */



#ifndef MSSO_STRUCT_FILE_DRIVER_H_H
#define MSSO_STRUCT_FILE_DRIVER_H_H

#include "rods.h"
#include "structFileDriver.h"
#include "execMyRule.h"

#define NUM_MSSO_SUB_FILE_DESC 1024
#define MSSO_CACHE_DIR_STR "mssoCacheDir"
#define MSSO_RUN_DIR_STR "runDir"
#define MSSO_RUN_FILE_STR "run"
#define MSSO_MPF_FILE_STR "mpf"


typedef struct mssoSubFileDesc {
    int inuseFlag;
    int structFileInx;
    int fd;                         /* the fd of the opened cached subFile */
    char cacheFilePath[MAX_NAME_LEN];   /* the phy path name of the cached
                                         * subFile */
} mssoSubFileDesc_t;

typedef struct mssoSubFileStack{
    int structFileInx;
    int fd;                         /* the fd of the opened cached subFile */
    char cacheFilePath[MAX_NAME_LEN];
    char *stageIn[100];   /* stages from irods to execuion area */
    char *stageOut[100];  /* stages into irods from execution area */
    char *copyToIrods[100];  /* copy into irods from execution area */
    char *cleanOut[100];  /* clear from execution area */
    char *checkForChange[100];  /* check if these files are newer than the change directort */
    char stageArea[MAX_NAME_LEN]; /* is cmd/bin by default */
    int stinCnt;
    int stoutCnt;
    int cpoutCnt;
    int clnoutCnt;
    int cfcCnt;
    int noVersions;  
    char newRunDirName[MAX_NAME_LEN]; /* name of new place for old  run dir  */
    int oldRunDirTime; /* time of old run dir which was moved to make place to new one */

} mssoSubFileStack_t;




int
mssoSubStructFileCreate (rsComm_t *rsComm, subFile_t *subFile);
int 
mssoSubStructFileOpen (rsComm_t *rsComm, subFile_t *subFile); 
int 
mssoSubStructFileRead (rsComm_t *rsComm, int fd, void *buf, int len);
int
mssoSubStructFileWrite (rsComm_t *rsComm, int fd, void *buf, int len);
int 
mssoSubStructFileClose (rsComm_t *rsComm, int fd);
int 
mssoSubStructFileUnlink (rsComm_t *rsComm, subFile_t *subFile); 
int
mssoSubStructFileStat (rsComm_t *rsComm, subFile_t *subFile,
rodsStat_t **subStructFileStatOut); 
int
mssoSubStructFileFstat (rsComm_t *rsComm, int fd, 
rodsStat_t **subStructFileStatOut);
rodsLong_t
mssoSubStructFileLseek (rsComm_t *rsComm, int fd, rodsLong_t offset, int whence);
int 
mssoSubStructFileRename (rsComm_t *rsComm, subFile_t *subFile, char *newFileName);
int
mssoSubStructFileMkdir (rsComm_t *rsComm, subFile_t *subFile);
int
mssoSubStructFileRmdir (rsComm_t *rsComm, subFile_t *subFile);
int
mssoSubStructFileOpendir (rsComm_t *rsComm, subFile_t *subFile);
int
mssoSubStructFileReaddir (rsComm_t *rsComm, int fd, rodsDirent_t **rodsDirent);
int
mssoSubStructFileClosedir (rsComm_t *rsComm, int fd);
int
mssoSubStructFileTruncate (rsComm_t *rsComm, subFile_t *subFile);
int
rsMssoStructFileOpen (rsComm_t *rsComm, specColl_t *specColl, subFile_t *subFile, int stage);
int
stageMssoStructFile (int structFileInx, subFile_t *subFile);
int
mkMssoCacheDir (int structFileInx, subFile_t *subFile);
int
extractMssoFile (int structFileInx, subFile_t *subFile, char *runDir, char* showFiles);
int
mssoStructFileSync (rsComm_t *rsComm, structFileOprInp_t *structFileOprInp);
int
mssoStructFileExtract (rsComm_t *rsComm, structFileOprInp_t *structFileOprInp);
int
syncCacheDirToMssofile (int structFileInx, int oprType);
int
initMssoSubFileDesc ();
int
allocMssoSubFileDesc ();
int
freeMssoSubFileDesc (int mssoSubFileInx);
int
getMssoSubStructFilePhyPath (char *phyPath, specColl_t *specColl,
			     char *subFilePath);

int
prepareForExecution(char *inRuleFile, char *inParamFile, char *runDir, char *showFiles,
		    execMyRuleInp_t *execMyRuleInp, msParamArray_t *msParamArray, int structFileInx, subFile_t *subFile);

int
getMpfFileName(subFile_t *subFile, char *mpfFileName);

int
mkMssoMpfRunDir (int structFileInx, subFile_t *subFile, char *runDir);
int
mkMssoMpfRunFile (int structFileInx, subFile_t *subFile);

int
matchMssoStructFileDesc (specColl_t *specColl);

#endif	/* MSSO_STRUCT_FILE_DRIVER_H_H */
