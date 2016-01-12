/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/* directAccessFileDriver.c - The Direct Access UNIX file driver.
 */

/* If the system is not compiled with RUN_SERVER_AS_ROOT, this driver
   will just call the same unixFileDriver.c routines. */

/* for osauthGetUid() */
#include "osauth.h"

#include "directAccessFileDriver.h"


/* Need to sequence the operations that are done with changed
   user context (i.e. after setuid()), since the user context
   is shared by all threads, and calls to setuid(user)/setuid(root)
   might happend out of order wrt a particular function. */
#include <pthread.h>
static pthread_mutex_t DirectAccessMutex;
static int DirectAccessMutexInitDone = 0;

int
directAccessAcquireLock()
{
  int rc;
  
  if (!DirectAccessMutexInitDone) {
    rc = pthread_mutex_init(&DirectAccessMutex, NULL);
    if (rc) {
      rodsLog(LOG_ERROR, "directAccessAcquireLock: error in pthread_mutex_init: %s",
              strerror(errno));
      return rc;
    }
    DirectAccessMutexInitDone = 1;
  }
  
  rc = pthread_mutex_lock(&DirectAccessMutex);
  if (rc) {
    rodsLog(LOG_ERROR, "directAccessAcquireLock: error in pthread_mutex_lock: %s",
            strerror(errno));
  }

  return rc;
}

int
directAccessReleaseLock()
{
  int rc = 0;

  if (DirectAccessMutexInitDone) {
    rc = pthread_mutex_unlock(&DirectAccessMutex);
    if (rc) {
      rodsLog(LOG_ERROR, "directAccessReleaseLock: error in pthread_mutex_unlock: %s",
              strerror(errno));
    }
  }
  
  return rc;
}


/* These global variables and 4 functions are used to track 
   per file state across operations (if needed). Right now 
   this is only needed to store the file mode for new
   files that are read-only, as they can't be opened for 
   write after the create call */

static int DirectAccessFileStateArraySize = 0;
static directAccessFileState_t *DirectAccessFileStateArray = NULL;

static int
directAccessFileArrayResize() 
{
    directAccessFileState_t *newArray;
    int newArraySize = DirectAccessFileStateArraySize + DIRECT_ACCESS_FILE_STATE_ARRAY_SIZE;
    int i;

    newArray = (directAccessFileState_t*)calloc(newArraySize, sizeof(directAccessFileState_t));
    if (newArray == NULL) {
        rodsLog(LOG_ERROR, "directAccessFileArrayResize: could not calloc new array: errno=%d",
                errno);
        return -1;
    }
    memset(newArray, 0, newArraySize*sizeof(directAccessFileState_t));

    for (i = 0; i < DirectAccessFileStateArraySize; i++) {
        newArray[i] = DirectAccessFileStateArray[i];
    }
    for (; i < newArraySize; i++) {
        /* indicate unused with -1 */
        newArray[i].fd = -1;
    }
    
    if (DirectAccessFileStateArray) {
        free(DirectAccessFileStateArray);
    }
    DirectAccessFileStateArray = newArray;
    DirectAccessFileStateArraySize = newArraySize;
    
    return 0;
}

static directAccessFileState_t *
directAccessFileNewState()
{
    int count;
    int index;
    
    /* only try to resize count times should never be an issue */
    count = 5;
    index = 0;

    while (count) {
        if (index == DirectAccessFileStateArraySize) {
            if (directAccessFileArrayResize()) {
                /* whoops ... malloc error */
                return NULL;
            }
            count--;
        }
        if (DirectAccessFileStateArray[index].fd == -1) {
            DirectAccessFileStateArray[index].fd = -2;
            return &DirectAccessFileStateArray[index];
        }
        index++;
    }

    return NULL;
}

void
directAccessFreeFileState(directAccessFileState_t *fileState)
{
    if (fileState) {
        memset(fileState, 0, sizeof(directAccessFileState_t));
        fileState->fd = -1;
    }
}

static directAccessFileState_t *
directAccessFileGetState(int fd)
{
    int i;

    for (i = 0; i < DirectAccessFileStateArraySize; i++) {
        if (DirectAccessFileStateArray[i].fd == fd) {
            return &DirectAccessFileStateArray[i];
        }
    }
    
    return NULL;
}


/* helper function that gets the UNIX uid for the operation
   based on the clientUser in the rsComm structure. Will 
   return the uid to use, or -1 if there is a problem 
   (e.g. user doesn't exist, or remote zone user */
static int
directAccessGetOperationUid(rsComm_t *rsComm)
{
    if (rsComm->clientUser.authInfo.authFlag == LOCAL_PRIV_USER_AUTH) {
        /* a local iRODS admin ... do operation as root */
        return 0;
    }

    if (rsComm->clientUser.authInfo.authFlag != LOCAL_USER_AUTH) {
        /* we don't allow remote zone or unauthenticated users 
           to do anything in direct access vaults. */
        return -1;
    }

    /* Now we look up the user's uid based on the provided
       client user name. Depends on the fact that iRODS user
       names are the same as the UNIX user name. 
       osauthGetUid() will return -1 on error.
    */
    return osauthGetUid(rsComm->clientUser.userName);
}


/* These functions need to do some operations as root in order to
   manage file permissions correctly. */

int
directAccessFileCreate (rsComm_t *rsComm, char *fileName, int mode, 
                        rodsLong_t mySize, keyValPair_t *condInput)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileCreate";
    int fd;
    mode_t myMask;
    uid_t fileUid;
    gid_t fileGid;
    mode_t fileMode;
    int opUid;
    directAccessFileState_t *fileState;
    char *fileUidStr;
    char *fileGidStr;
    char *fileModeStr = NULL;
    int myerrno;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    /* initially create the file as root to avoid any
       directory permission issues. We'll chown it later
       if necessary. */
    directAccessAcquireLock();
    changeToRootUser();

    myMask = umask((mode_t) 0000);
    fd = open (fileName, O_RDWR|O_CREAT|O_EXCL, mode);
    /* reset the old mask */
    (void) umask((mode_t) myMask);
    
    if (fd == 0) {
        close (fd);
	rodsLog (LOG_NOTICE, "%s: 0 descriptor", fname);
        open ("/dev/null", O_RDWR, 0);
        myMask = umask((mode_t) 0000);
        fd = open (fileName, O_RDWR|O_CREAT|O_EXCL, mode);
        (void) umask((mode_t) myMask);
    }
    
    if (fd < 0) {
        myerrno = errno;
        changeToServiceUser();
        directAccessReleaseLock();
        fd = UNIX_FILE_CREATE_ERR - myerrno;
	if (errno == EEXIST) {
	    rodsLog (LOG_DEBUG, "%s: open error for %s, file exists",
                     fname, fileName);
        }
        else if (errno == ENOENT) {
	    rodsLog (LOG_DEBUG, "%s: open error for %s, path component doesn't exist",
                     fname, fileName, fd);
	} 
        else {
	    rodsLog (LOG_NOTICE, "%s: open error for %s, status = %d",
                     fname, fileName, fd);
	}
	return (fd);
    }

    /* if meta-data was passed, use chown/chmod to set the
       meta-data on the file. */
    if (condInput) {
        fileUidStr = getValByKey(condInput, FILE_UID_KW);
        fileGidStr = getValByKey(condInput, FILE_GID_KW);
        fileModeStr = getValByKey(condInput, FILE_MODE_KW);
        if (fileUidStr && fileGidStr && fileModeStr) {
            fileUid = atoi(fileUidStr);
            fileGid = atoi(fileGidStr);
            fileMode = atoi(fileModeStr);
            if (fchown(fd, fileUid, fileGid)) {
                rodsLog(LOG_ERROR, "%s: could not set owner/group on %s. errno=%d",
                        fname, fileName, errno);
            }
            if (fchmod(fd, fileMode)) {
                rodsLog(LOG_ERROR, "%s: could not set mode on %s. errno=%d",
                        fname, fileName, errno);
            }
        }
    }

    close(fd);

    /* now re-open the file as the user making the call */
    changeToUser(opUid);
    fd = open(fileName, O_RDWR);

    if (fd < 0) {
        if (errno == EACCES && fileModeStr) {
            /* possible that the file mode applied from the
               passed meta-data has a read-only mode for the
               operation user. If we're creating the file
               with meta-data, it's probably from iput or 
               irepl, so open as root so the file can be 
               properly created. */
            changeToRootUser();
            fd = open(fileName, O_RDWR);
            if (fd > 0) {
                /* set file state to tell close() to do close
                   as the root user */
                fileState = directAccessFileNewState();
                if (fileState) {
                    fileState->fd = fd;
                    fileState->fileMode = fileMode;
                }
            }
        }
    }

    if (fd < 0) {
        fd = UNIX_FILE_CREATE_ERR - errno;
        rodsLog (LOG_NOTICE, "%s: open error for %s, status = %d",
                 fname, fileName, fd);
    }

    changeToServiceUser();
    directAccessReleaseLock();
    
    return (fd);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileCreate(rsComm, fileName, mode, mySize, condInput);

#endif /*RUN_SERVER_AS_ROOT */
}

int
directAccessFileOpen (rsComm_t *rsComm, char *fileName, int flags, int mode, 
                      keyValPair_t *condInput)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileOpen";
    int fd;
    int opUid;
    char *tmpStr;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    /* if we can retrieve the file's uid meta-data from the 
       condInput, then we'll assume all meta-data was set */
    if (condInput) {
        tmpStr = getValByKey(condInput, FILE_UID_KW);
    }
    else {
        tmpStr = NULL;
    }

    if ((flags & (O_RDWR|O_WRONLY)) && tmpStr) {
        /* If the files is being opened with a write flag, 
           and meta-data has been passed, try to create the file
           and set the correct meta-data. If the file exists then
           directAccessFileCreate() will have no effect. */
        fd = directAccessFileCreate(rsComm, fileName, mode, 0, condInput);
        if (fd != (UNIX_FILE_CREATE_ERR - EEXIST)) {
            rodsLog (LOG_NOTICE, "directAccessFileOpen: open error for %s, errno = %d",
                     fileName, errno);
            return fd;
        }
        close(fd);
    }

    /* perform the open as the user making the call */
    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    fd = unixFileOpen(rsComm, fileName, flags, mode, condInput);

    changeToServiceUser();
    directAccessReleaseLock();
    
    return (fd);

#else /* RUN_SERVER_AS_ROOT */
    
    return unixFileOpen(rsComm, fileName, flags, mode, condInput);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileMkdir (rsComm_t *rsComm, char *dirname, int mode,
                       keyValPair_t *condInput)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileMkdir";
    int opUid;
    int status;
    char *fileUidStr;
    char *fileGidStr;
    char *fileModeStr;
    int fileUid;
    int fileGid;
    int fileMode;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    /* initially create the file as root to avoid any
       directory permission issues. We'll chown it later. */
    directAccessAcquireLock();
    changeToRootUser();

    status = unixFileMkdir(rsComm, dirname, mode, condInput);

    /* if meta-data was passed, use chown/chmod to set the
       meta-data on the new directory. */
    if (status >= 0 && condInput) {
        fileUidStr = getValByKey(condInput, FILE_UID_KW);
        fileGidStr = getValByKey(condInput, FILE_GID_KW);
        fileModeStr = getValByKey(condInput, FILE_MODE_KW);
        if (fileUidStr && fileGidStr && fileModeStr) {
            fileUid = atoi(fileUidStr);
            fileGid = atoi(fileGidStr);
            fileMode = atoi(fileModeStr);
            if (chown(dirname, fileUid, fileGid)) {
                rodsLog(LOG_ERROR, "%s: could not set owner/group on %s. errno=%d",
                        fname, dirname, errno);
            }
            if (chmod(dirname, fileMode)) {
                rodsLog(LOG_ERROR, "%s: could not set mode on %s. errno=%d",
                        fname, dirname, errno);
            }
        }
    }
    
    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */
    
    return unixFileMkdir(rsComm, dirname, mode, condInput);

#endif /* RUN_SERVER_AS_ROOT */
}       


/* The following functions follow the same pattern. Get the uid
   for the operation (based on clientUser), and then call the
   unixFileDriver function as that user. */

int
directAccessFileClose (rsComm_t *rsComm, int fd)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileClose";
    int status;
    int opUid;
    directAccessFileState_t *fileState;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    fileState = directAccessFileGetState(fd);

    directAccessAcquireLock();
    if (fileState != NULL || opUid == 0) {
        directAccessFreeFileState(fileState);
        changeToRootUser();
    }
    else {
        changeToUser(opUid);
    }
    
    status = unixFileClose(rsComm, fd);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileClose(rsComm, fd);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileUnlink (rsComm_t *rsComm, char *filename)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileUnlink";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileUnlink(rsComm, filename);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileUnlink(rsComm, filename);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileRename (rsComm_t *rsComm, char *oldFileName, char *newFileName)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileRename";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileRename(rsComm, oldFileName, newFileName);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileRename(rsComm, oldFileName, newFileName);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileTruncate (rsComm_t *rsComm, char *filename, rodsLong_t dataSize)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileTruncate";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileTruncate(rsComm, filename, dataSize);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileTruncate(rsComm, filename, dataSize);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileChmod (rsComm_t *rsComm, char *filename, int mode)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileChmod";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileChmod(rsComm, filename, mode);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileChmod(rsComm, filename, mode);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileStat (rsComm_t *rsComm, char *filename, struct stat *statbuf)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileStat";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileStat(rsComm, filename, statbuf);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileStat(rsComm, filename, statbuf);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileRmdir (rsComm_t *rsComm, char *dirname)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileRmdir";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileRmdir(rsComm, dirname);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileRmdir(rsComm, dirname);

#endif /* RUN_SERVER_AS_ROOT */
}

int
directAccessFileOpendir (rsComm_t *rsComm, char *dirname, void **outDirPtr)
{
#ifdef RUN_SERVER_AS_ROOT
    static char fname[] = "directAccessFileOpendir";
    int status;
    int opUid;

    opUid = directAccessGetOperationUid(rsComm);
    if (opUid < 0) {
        rodsLog(LOG_NOTICE, 
                "%s: remote zone users cannot modify direct access vaults. User %s#%s",
                fname, rsComm->clientUser.userName, rsComm->clientUser.rodsZone);
        return DIRECT_ACCESS_FILE_USER_INVALID_ERR;
    }

    directAccessAcquireLock();
    if (opUid) {
        changeToUser(opUid);
    }
    else {
        changeToRootUser();
    }
    
    status = unixFileOpendir(rsComm, dirname, outDirPtr);

    changeToServiceUser();
    directAccessReleaseLock();

    return (status);

#else /* RUN_SERVER_AS_ROOT */

    return unixFileOpendir(rsComm, dirname, outDirPtr);

#endif /* RUN_SERVER_AS_ROOT */
}


/* These functions pass directly through to the 
   corresponding unixFileDriver routine. */
int
directAccessFileRead (rsComm_t *rsComm, int fd, void *buf, int len)
{
    return unixFileRead(rsComm, fd, buf, len);
}

int
directAccessFileWrite (rsComm_t *rsComm, int fd, void *buf, int len)
{
    return unixFileWrite(rsComm, fd, buf, len);
}

int
directAccessFileFstat (rsComm_t *rsComm, int fd, struct stat *statbuf)
{
    return unixFileFstat(rsComm, fd, statbuf);
}

rodsLong_t
directAccessFileLseek (rsComm_t *rsComm, int fd, rodsLong_t offset, int whence)
{
    return unixFileLseek(rsComm, fd, offset, whence);
}

int
directAccessFileFsync (rsComm_t *rsComm, int fd)
{
    return unixFileFsync(rsComm, fd);
}

int
directAccessFileClosedir (rsComm_t *rsComm, void *dirPtr)
{
    return unixFileClosedir(rsComm, dirPtr);
}

int
directAccessFileReaddir (rsComm_t *rsComm, void *dirPtr, struct dirent *direntPtr)
{
    return unixFileReaddir(rsComm, dirPtr, direntPtr);
}

rodsLong_t
directAccessFileGetFsFreeSpace (rsComm_t *rsComm, char *path, int flag)
{
    return unixFileGetFsFreeSpace(rsComm, path, flag);
}

int
directAccessFileStage (rsComm_t *rsComm, char *path, int flag)
{
    return unixFileStage(rsComm, path, flag);
}

