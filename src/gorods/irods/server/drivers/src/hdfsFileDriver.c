/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/* hdfsFileDriver.c - The HDFS file driver */


/*****************************************************************
Erik Scott
escott@renci.org
*****************************************************************/



#include "hdfsFileDriver.h"
/* escott */
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <dirent.h>

#include "hdfs.h"


void hdfsnotate(char *outline)
{
  FILE *fptr = fopen("/tmp/hdfsnotation","a");
  fprintf(fptr,"%s\n",outline);
  fclose(fptr);
}

typedef struct hdfsFDassociations {
  int fd;
  hdfsFile hf;
  char *fn;
} hdfsFDassociations_t;

static hdfsFDassociations_t hdfsFDpersistence[1024];
static int hdfsFDpersistenceIdx = 5;  /* trust me - 5 is a good starting point */
static hdfsFS persistentFS = NULL;

typedef struct hdfsFileInfoHolder {
  hdfsFileInfo *fi;
  int num;
  int current;
} hdfsFileInfoAssociations_t;

static hdfsFileInfoAssociations_t hdfsFileInfoPersistence[1024];
static int hdfsFileInfoPersistenceIdx = 0;



hdfsFS getFS()
{
  if (persistentFS == NULL) 
    {
      persistentFS = hdfsConnect("default",0);
    }
  return persistentFS;
}


int hdfsAssocPush(hdfsFile hf)
{
  hdfsFDpersistence[hdfsFDpersistenceIdx].fd = hdfsFDpersistenceIdx;
  hdfsFDpersistence[hdfsFDpersistenceIdx].hf = hf;
  hdfsFDpersistenceIdx++;
  return hdfsFDpersistenceIdx;
}

hdfsFile hdfsAssocGet(int idx)
{
  return hdfsFDpersistence[idx].hf;
}



/* escott */

int
hdfsFileCreate (rsComm_t *rsComm, char *fileName, int mode, rodsLong_t mySize, keyValPair_t *condInput)
{
  // This is an open with a create option wired to it.  Not a problem at all.
  // Interesting things are just that it does the open, then saves the hdfsFile
  // object that results from that in an array.  This lets us preserve the
  // unix-like int fd approach to keeping track of file descriptors while 
  // keeping the heavier-weight hdfsFile bits around.
  //
  // Oh, and a special trick - we add 1024 to the file descriptor (the int)
  // so just in case there are legitimate unix file descriptors open, we don't
  // collide with the numbering range.  Sort of a belt-and-suspenders deal.

  int fd; 
  rodsLog (LOG_DEBUG, "hdfsCreateFile starting.  fileName = %s, mode = %d mySize= %ld errno = %d", fileName, mode, (long)mySize, errno);
  hdfsFS fs = getFS();
  hdfsFDpersistence[hdfsFDpersistenceIdx].fd = hdfsFDpersistenceIdx;
  hdfsFDpersistence[hdfsFDpersistenceIdx].fn = (char *)malloc(strlen(fileName) + 1);
  strcpy(hdfsFDpersistence[hdfsFDpersistenceIdx].fn, fileName);
  hdfsFDpersistence[hdfsFDpersistenceIdx].hf = hdfsOpenFile(fs, fileName, O_WRONLY|O_CREAT, 0, 0, 0);
  if (! (hdfsFDpersistence[hdfsFDpersistenceIdx].hf ) )
    {
      rodsLog (LOG_NOTICE, "hdfsFileCreate failing. fileName = %s mode = %d mySize= %ld errno = %d", fileName, mode, (long)mySize, errno);
      fd =  HDFS_FILE_CREATE_ERR - errno;
    }
  else 
    {
      rodsLog (LOG_DEBUG, "hdfsCreateFile succeeding. Returned fd = %d", hdfsFDpersistenceIdx+1024);
      fd = hdfsFDpersistenceIdx + 1024;
      hdfsFDpersistenceIdx++;
    }
  return fd;

}

int
hdfsFileOpen (rsComm_t *rsComm, char *fileName, int flags, int mode, keyValPair_t *condInput)
{   

  char outbuf[1024];
  sprintf(outbuf,"hdfsFileOpen fileName = %s , flags = 0%o mode = 0%o",fileName, flags, mode);
  hdfsnotate(outbuf);

  /*
   int fd;

#if defined(osx_platform)
    // For osx, O_TRUNC = 0x0400, O_TRUNC = 0x200 for other system 
    if (flags & 0x200) {
        flags = flags ^ 0x200;
        flags = flags | O_TRUNC;
    }
#endif

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileOpen fileName = %s , flags = 0%o mode = 0%o",fileName, flags, mode);
    hdfsnotate(outbuf);

    fd = open (fileName, flags, mode);

    if (fd == 0) {
        close (fd);
        rodsLog (LOG_NOTICE, "hdfsFileOpen: 0 descriptor");
        open ("/dev/null", O_RDWR, 0);
        fd = open (fileName, flags, mode);
    }

    if (fd < 0) {
        fd = HDFS_FILE_OPEN_ERR - errno;
        rodsLog (LOG_NOTICE, "hdfsFileOpen: open error for %s, status = %d",
          fileName, fd);
    }

    return (fd);
*/

  int fd; 
  rodsLog (LOG_DEBUG, "hdfsOpenFile starting.  fileName = %s, flags = 0%o mode = 0%o ", fileName, flags, mode);
  hdfsFS fs = getFS();
  hdfsFDpersistence[hdfsFDpersistenceIdx].fd = hdfsFDpersistenceIdx;
  hdfsFDpersistence[hdfsFDpersistenceIdx].fn = (char *)malloc(strlen(fileName) + 1);
  strcpy(hdfsFDpersistence[hdfsFDpersistenceIdx].fn, fileName);
  hdfsFDpersistence[hdfsFDpersistenceIdx].hf = hdfsOpenFile(fs, fileName, flags , 0, 0, 0);
  if (! (hdfsFDpersistence[hdfsFDpersistenceIdx].hf ) )
    {
      rodsLog (LOG_NOTICE, "hdfsFileOpen failing. fileName = %s flags = 0%o mode = 0%o  errno = %d", fileName, flags, mode, errno);
      fd =  HDFS_FILE_CREATE_ERR - errno;
    }
  else 
    {
      rodsLog (LOG_DEBUG, "hdfsCreateFile succeeding. Returned fd = %d", hdfsFDpersistenceIdx+1024);
      fd = hdfsFDpersistenceIdx + 1024;
      hdfsFDpersistenceIdx++;
    }
  return fd;

}

int
hdfsFileRead (rsComm_t *rsComm, int fd, void *buf, int len)
{
  /*
    int status;

    status = read (fd, buf, len);

    if (status < 0) {
        status = HDFS_FILE_READ_ERR - errno;
        rodsLog (LOG_NOTICE, "hdfsFileRead: read error fd = %d, status = %d",
         fd, status);
    }

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileRead fd = %d len = %d status = %d", fd, len, status);
    hdfsnotate(outbuf);

    return (status);
  */

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileRead fd = %d len = %d", fd, len);
    hdfsnotate(outbuf);

    rodsLog (LOG_DEBUG,"hdfsFileRead() fd = %d len = %d", fd, len);
    hdfsFS fs = getFS();
  int idx = fd - 1024;
  hdfsFile readFile = hdfsFDpersistence[idx].hf;
  tSize numReadBytes = hdfsRead(fs, readFile, (void *)buf, len);
  rodsLog (LOG_DEBUG,"hdfsFileRead() read %d bytes.", numReadBytes);
  return numReadBytes;

}

int
hdfsnbFileRead (rsComm_t *rsComm, int fd, void *buf, int len)
{
    char outbuf[1024];
    sprintf(outbuf,"nbFileRead().  Dear God.");
    hdfsnotate(outbuf);
    rodsLog (LOG_DEBUG,"hdfsnbFileWrite() - the nonblocking version should never be called.  Not supported.");

    return -1; // indicate heinous error
}

int
hdfsFileWrite (rsComm_t *rsComm, int fd, void *buf, int len)
{
  /*
    int status;

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileWrite fd = %d, len= %d", fd, len);
    hdfsnotate(outbuf);

    status = write (fd, buf, len);

    if (status < 0) {
        status = HDFS_FILE_WRITE_ERR - errno;
        rodsLog (LOG_NOTICE, "hdfsFileWrite: open write fd = %d, status = %d",
         fd, status);
    }
    return (status);
  */

  rodsLog (LOG_DEBUG,"hdfsWrite() fd = %d, len = %d", fd, len);
  hdfsFS fs = getFS();
  int idx = fd - 1024;
  hdfsFile writeFile = hdfsFDpersistence[idx].hf;
  tSize numWrittenBytes = hdfsWrite(fs, writeFile, (void *)buf, len);
  rodsLog (LOG_DEBUG,"hdfsWrite() numWrittenBytes = %ld", (long) numWrittenBytes);
  return numWrittenBytes;

}

int
hdfsnbFileWrite (rsComm_t *rsComm, int fd, void *buf, int len)
{
    char outbuf[1024];
    sprintf(outbuf,"nbFileWrite() - madness, part deux.");
    hdfsnotate(outbuf);
    rodsLog (LOG_DEBUG,"hdfsnbFileWrite() - the nonblocking version should never be called.  Not supported.");

    return -1;  // indicate some sort of greivous error.
}

int
hdfsFileClose (rsComm_t *rsComm, int fd)
{
  /*    int status;

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileClose fd=%d",fd);
    hdfsnotate(outbuf);

    status = close (fd);

    if (fd == 0) {
        rodsLog (LOG_NOTICE, "hdfsFileClose: 0 descriptor");
        open ("/dev/null", O_RDWR, 0);
    }
    if (status < 0) {
        status = HDFS_FILE_CLOSE_ERR - errno;
        rodsLog (LOG_NOTICE, "hdfsFileClose: open write fd = %d, status = %d",
         fd, status);
    }
    return (status);
  */

  rodsLog (LOG_DEBUG,"hdfsClose() fd = %d", fd);
  hdfsFS fs = getFS();
  int idx = fd - 1024;
  hdfsFile flushFile = hdfsFDpersistence[idx].hf;
  int flushStatus = hdfsFlush(fs, flushFile);
  rodsLog (LOG_DEBUG,"hdfsClose() flushed:  flushStatus = %d", flushStatus);
  hdfsCloseFile(fs, flushFile);


  return flushStatus;
}

int
hdfsFileUnlink (rsComm_t *rsComm, char *filename)
{
  /*    int status;

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileUnlink fileName = %s",filename);
    hdfsnotate(outbuf);

    status = unlink (filename);

    if (status < 0) {
        status = HDFS_FILE_UNLINK_ERR - errno;
        rodsLog (LOG_NOTICE, "hdfsFileUnlink: unlink of %s error, status = %d",
         filename, status);
    }

    return (status);

  */

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileStat fileName = %s",filename);
    hdfsnotate(outbuf);

    int status;
    rodsLog (LOG_DEBUG, "Starting hdfsFileUnlink filename = %s", filename);
    hdfsFS fs = getFS();

    int result = hdfsDelete(fs, filename);
    if (0 == result)
      {
	status = 0;
      }
    else 
      {
	status = HDFS_FILE_UNLINK_ERR - errno;
      }

    return(status);

}

int
hdfsFileStat (rsComm_t *rsComm, char *filename, struct stat *statbuf)
{   

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileStat fileName = %s",filename);
    hdfsnotate(outbuf);

    int status;
    rodsLog (LOG_DEBUG, "Starting hdfsFileStat filename = %s", filename);
    hdfsFS fs = getFS();
    hdfsFileInfo *fileInfo = NULL;
    if ((fileInfo = hdfsGetPathInfo(fs,filename)) != NULL)
      {
	statbuf->st_dev = 0;    // hdfs doesn't really have a device
	statbuf->st_ino = 101;  // totally made up inode number
	statbuf->st_mode = fileInfo->mPermissions;  // incomplete but still useful.
	statbuf->st_nlink = 1;  // hdfs doesn't allow links.  Yet.
	statbuf->st_uid = 0;   // dummy this for being ownded by root - all hdfs files are owned by the iAgent userid
	statbuf->st_gid = 0;  // dummy, as above
	statbuf->st_rdev = 0;
	statbuf->st_size = fileInfo->mSize;
	statbuf->st_blksize = (blksize_t)fileInfo->mBlockSize;
	// the next line - if the block size divides the size evenly, you have perfect block packing.  Report the number of blocks.
	// else, you should report what perfect would have been, plus one for the overflowing bytes.
	// I hope someone actually uses this.
	statbuf->st_blocks =   (fileInfo->mSize) % (fileInfo->mBlockSize)==0 ?  (fileInfo->mSize) / (fileInfo->mBlockSize) : ((fileInfo->mSize) / (fileInfo->mBlockSize))+1  ;
	statbuf->st_atime = fileInfo->mLastMod;  // no hdfs support, so this is a dummy value.
	statbuf->st_mtime = fileInfo->mLastMod;
	statbuf->st_ctime = 0;                   // January 1st, 1970.  Bellbottoms.  Wide Collars.  Party Vans.
	status = 0;
	rodsLog (LOG_DEBUG, "hdfsFileStat() success - some sort of error.\n\t filename %s\n\t type %c\n\treplication %d\n\tblocksize %ld\n\tsize %ld\n\tlastmod %s\n\towner %s\n\tgroup %s\n\tpermissions 0%o", fileInfo->mName, fileInfo->mKind, fileInfo->mReplication, fileInfo->mBlockSize, fileInfo->mSize, ctime(&fileInfo->mLastMod), fileInfo->mOwner, fileInfo->mGroup, fileInfo->mPermissions );
      }
    else
      {
	status = HDFS_FILE_STAT_ERR ;
	rodsLog (LOG_DEBUG, "hdfsFileStat() BAILING OUT - some sort of error.\n\t filename %s\n\t type %c\n\treplication %d\n\tblocksize %ld\n\tsize %ld\n\tlastmod %s\n\towner %s\n\tgroup %s\n\tpermissions 0%o", fileInfo->mName, fileInfo->mKind, fileInfo->mReplication, fileInfo->mBlockSize, fileInfo->mSize, ctime(&fileInfo->mLastMod), fileInfo->mOwner, fileInfo->mGroup, fileInfo->mPermissions );

      }

    return(status);

}

int
hdfsFileFstat (rsComm_t *rsComm, int fd, struct stat *statbuf)
{

  // All I'm doing is converting the file descriptor to a filename and calling hdfsFileStat() based on that.
  // just doing this because there's no way to stat an hdfsFile struct.

  char *filename;
  int idx = fd - 1024;  // put that back down in the range of reality so I can look it up.
  filename = hdfsFDpersistence[idx].fn;  // set filename to point to the fn malloc'ed region for the right file descr

  char outbuf[1024];
  sprintf(outbuf,"hdfsFileFstat fd = %d filename = %s", fd, filename);
  hdfsnotate(outbuf);

  return (hdfsFileStat(rsComm, filename, statbuf));

}

rodsLong_t
hdfsFileLseek (rsComm_t *rsComm, int fd, rodsLong_t offset, int whence)
{
    rodsLong_t  status;

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileLseek fd= %d offset = %ld whence = %d", fd, (long) offset, whence);
    hdfsnotate(outbuf);
    /*
    status = lseek (fd, offset, whence);

    if (status < 0) {
        status = HDFS_FILE_LSEEK_ERR - errno;
        rodsLog (LOG_NOTICE, 
	  "hdfsFileLseek: lseek of fd %d error, status = %d", fd, status);
    }

    */

    rodsLog (LOG_DEBUG,"hdfsFileLseek() fd = %d offset = %ld whence = %d", fd, offset, whence);
    if (whence != SEEK_SET)
      {
	rodsLog (LOG_DEBUG,"hdfsFileLseek() bad whence = %d", whence);
	return  HDFS_FILE_LSEEK_ERR;
      }

    hdfsFS fs = getFS();
    int idx = fd - 1024;
    hdfsFile seekFile = hdfsFDpersistence[idx].hf;
    int results = hdfsSeek (fs, seekFile, (tOffset) offset);
    if (0 != results)
      {
	status = HDFS_FILE_LSEEK_ERR - errno;
      }
    else 
      {
	status = 0;
      }

    return (status);
}

int
hdfsFileFsync (rsComm_t *rsComm, int fd)
{
  char outbuf[1024];
  sprintf(outbuf,"hdfsFileFsync fd = %d", fd);
  hdfsnotate(outbuf);
  /*

    status = fsync (fd);

    if (status < 0) {
    status = HDFS_FILE_FSYNC_ERR - errno;
    rodsLog (LOG_NOTICE, 
    "hdfsFileFsync: fsync of fd %d error, status = %d", fd, status);
    }
  */

  hdfsFS fs = getFS();
  int idx = fd - 1024;
  hdfsFile flushFile = hdfsFDpersistence[idx].hf;
  int flushStatus = hdfsFlush(fs, flushFile);

  return (flushStatus);
}

int
hdfsFileMkdir (rsComm_t *rsComm, char *filename, int mode, keyValPair_t *condInput)
{
  int status;
  //  mode_t myMask;

  char outbuf[1024];
  sprintf(outbuf,"hdfsFileMkdir filename = %s mode = %d", filename, mode);
  hdfsnotate(outbuf);

  /*
    myMask = umask((mode_t) 0000);


    status = mkdir (filename, mode);
    // reset the old mask 
    (void) umask((mode_t) myMask);


    if (status < 0) {
    status = HDFS_FILE_MKDIR_ERR - errno;
    if (errno != EEXIST)
    rodsLog (LOG_NOTICE,
    "hdfsFileMkdir: mkdir of %s error, status = %d", 
    filename, status);
    }
  */

  hdfsFS fs = getFS();
  int result = hdfsCreateDirectory(fs, filename);
  if (0 != result)
    {
      status = HDFS_FILE_MKDIR_ERR - errno;
      return status;
    }
  result = hdfsChmod(fs, filename, mode);
  if (0 != result)
    {
      status = HDFS_FILE_MKDIR_ERR - 99;
    }
  else
    {
      status = 0;
    }

  return (status);
}       

int
hdfsFileChmod (rsComm_t *rsComm, char *filename, int mode)
{
  int status;

  char outbuf[1024];
  sprintf(outbuf,"hdfsFileChmod fileName = %s mode = %d", filename, mode);
  hdfsnotate(outbuf);
  /*
    status = chmod (filename, mode);

    if (status < 0) {
    status = HDFS_FILE_CHMOD_ERR - errno;
    rodsLog (LOG_NOTICE,
    "hdfsFileChmod: chmod of %s error, status = %d", 
    filename, status);
    }
  */
  hdfsFS fs = getFS();

  int result = hdfsChmod (fs, filename, mode);
  if (0 != result)
    {
      status  = HDFS_FILE_CHMOD_ERR - errno;
    }
  else
    {
      status = 0;
    }

  return (status);
}

int
hdfsFileRmdir (rsComm_t *rsComm, char *filename)
{
  char outbuf[1024];
  sprintf(outbuf,"hdfsFileRmdir fileName = %s",filename);
  hdfsnotate(outbuf);

  rodsLog (LOG_DEBUG,"hdfsFileRmdir() filename = %s", filename);

  /*
    status = rmdir (filename);

    if (status < 0) {
    status = HDFS_FILE_RMDIR_ERR - errno;
    rodsLog (LOG_DEBUG,
    "hdfsFileRmdir: rmdir of %s error, status = %d",
    filename, status);
    }
  */

  return (hdfsFileUnlink(rsComm, filename));
}

int
hdfsFileOpendir (rsComm_t *rsComm, char *dirname, void **outDirPtr)
{
    int status;

        status = HDFS_FILE_OPENDIR_ERR - errno;
	hdfsFS fs = getFS();
	int numEntries;
	hdfsFileInfoPersistence[hdfsFileInfoPersistenceIdx].fi = hdfsListDirectory(fs, dirname, &numEntries);
	hdfsFileInfoPersistence[hdfsFileInfoPersistenceIdx].num = numEntries;
	hdfsFileInfoPersistence[hdfsFileInfoPersistenceIdx].current = 0;
	// sure, it's really an int, but the API really expects a pointer, so this won't work
	// on DG/UX or probably even Linux for Z/Series.  Then again, good luck with those anyway.
	*outDirPtr = (void *) hdfsFileInfoPersistenceIdx;

	hdfsFileInfoPersistenceIdx++;

	return (0);

}

int
hdfsFileClosedir (rsComm_t *rsComm, void *dirPtr)
{
  long victim = (long)dirPtr;   // cast it back to a long, but we have to keep it a pointer for the API
  if (victim == (hdfsFileInfoPersistenceIdx - 1) )
    {
      // it's deleting the most recently created one, so decrement and reuse that entry
      // this will practically always be the case (in fact, always?) so this isn't as stupid
      // as it looks.  Replace with STL list in the E-iRODS refactoring...
      free(hdfsFileInfoPersistence[victim].fi);
      hdfsFileInfoPersistenceIdx--;
    }
  else
    {
      free(hdfsFileInfoPersistence[victim].fi);  // still prevent the leak, anyway.
    }

  return 0;

}

int
hdfsFileReaddir (rsComm_t *rsComm, void *dirPtr, struct dirent *direntPtr)
{
    struct dirent *tmpDirentPtr;

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileReaddir");
    hdfsnotate(outbuf);

    long idx = (long) dirPtr;
    hdfsFileInfo *fileinfo;

    fileinfo = hdfsFileInfoPersistence[idx].fi;
    int total = hdfsFileInfoPersistence[idx].num;
    int curr = hdfsFileInfoPersistence[idx].current;

    if (curr == total) 
      {
	return -1;  // done, no errors.
      }

    tmpDirentPtr = (struct dirent *) malloc (sizeof(struct dirent));
    tmpDirentPtr->d_ino = 106;  // good number - dummy number, though.
    tmpDirentPtr->d_type = DT_REG;  // regular file.
    strncpy(tmpDirentPtr->d_name, hdfsFileInfoPersistence[idx].fi[curr].mName,255);
    tmpDirentPtr->d_name[255]='\0';  // just in case the name is too long

    *direntPtr = *tmpDirentPtr;

    return (0);



    /*
    errno = 0;
    tmpDirentPtr = readdir ((DIR*)dirPtr);

    if (tmpDirentPtr == NULL) {
	if (errno == 0) {
	    // just the end 
	    status = -1;
	} else {
            status = HDFS_FILE_READDIR_ERR - errno;
             rodsLog (LOG_NOTICE,
               "hdfsFileReaddir: readdir error, status = %d", status);
	}
    } else {
	status = 0;
	*direntPtr = *tmpDirentPtr;
#if defined(solaris_platform)
	rstrcpy (direntPtr->d_name, tmpDirentPtr->d_name, MAX_NAME_LEN);
#endif
    }
    return (status);
*/

}

int
hdfsFileStage (rsComm_t *rsComm, char *path, int flag)
{
  // you can't combine hdfs with SAM FS, doesn't make any sense.  So this path is considered "unreachable".
  return -1;
}

int
hdfsFileRename (rsComm_t *rsComm, char *oldFileName, char *newFileName)
{
    int status;

    char outbuf[1024];
    sprintf(outbuf,"hdfsFileRename oldFileName = %s newFileName = %s", oldFileName, newFileName);
    hdfsnotate(outbuf);


    /*
    status = rename (oldFileName, newFileName);

    if (status < 0) {
        status = HDFS_FILE_RENAME_ERR - errno;
        rodsLog (LOG_NOTICE,
         "hdfsFileRename: rename error, status = %d\n",
         status);
    }
    */
  hdfsFS fs = getFS();

  int result = hdfsRename (fs, newFileName, oldFileName);
  if (0 != result)
    {
      status  = HDFS_FILE_RENAME_ERR - errno;
    }
  else
    {
      status = 0;
    }

    return (status);
}

int
hdfsFileTruncate (rsComm_t *rsComm, char *filename, rodsLong_t dataSize)
{
  // this function should be stubbed out in the struct of function pointers.
  // HDFS files are immutable after creation - you cannot change (or even truncate)
  // one after creation.
  //
  // This should only be callable by the iRODS FUSE driver, and again - if you need
  // FUSE against an HDFS file store, you should be using the HDFS FUSE driver.
  // Skip some layers, get better results.

    return (-1);
}

rodsLong_t
hdfsFileGetFsFreeSpace (rsComm_t *rsComm, char *path, int flag)
{
  char outbuf[1024];
  sprintf(outbuf,"hdfsFileGetFsFreeSpace path = %s flag = %d", path, flag);
  hdfsnotate(outbuf);

  hdfsFS fs = getFS();
  rodsLong_t totalSpace = hdfsGetCapacity(fs);
  rodsLong_t usedSpace = hdfsGetUsed(fs);

  if ( (totalSpace < 0) || (usedSpace < 0) )
    {
      return USER_NO_SUPPORT_ERR;
    }
  else
    {
    return totalSpace - usedSpace;
    }

}



/* hdfsStageToCache - This routine is for testing the TEST_STAGE_FILE_TYPE.
 * Just copy the file from filename to cacheFilename. optionalInfo info
 * is not used.
 * 
 */
  
int
hdfsStageToCache (rsComm_t *rsComm, fileDriverType_t cacheFileType, 
int mode, int flags, char *filename, 
char *cacheFilename, rodsLong_t dataSize,
keyValPair_t *condInput)
{
  return -1;
}

/* hdfsSyncToArch - This routine is for testing the TEST_STAGE_FILE_TYPE.
 * Just copy the file from cacheFilename to filename. optionalInfo info
 * is not used.
 *
 */

int
hdfsSyncToArch (rsComm_t *rsComm, fileDriverType_t cacheFileType, 
int mode, int flags, char *filename,
char *cacheFilename,  rodsLong_t dataSize,
keyValPair_t *condInput)
{
  return -1;
}


// the good news is that this can only be called (by iRODS) such that
// both the course and the target are in the same filesystem.
// If they were in different filesystem types, they'd be accessed
// through different drivers.  And we only support having one HDFS
// instance per iRODS Zone.

int
hdfsFileCopy (int mode, char *srcFileName, char *destFileName)
{
  // this is just an excerpt of the old code...
  /*
    if (bytesCopied != statbuf.st_size) {
        rodsLog (LOG_ERROR,
         "hdfsFileCopy: Copied size %lld does not match source size %lld of %s",
         bytesCopied, statbuf.st_size, srcFileName);
        return SYS_COPY_LEN_ERR;
    } else {
	return 0;
    }
  */

  hdfsFS fs = getFS();
  int status = hdfsCopy(fs, srcFileName, fs, destFileName);

  if (status < 0)
    {
      return SYS_COPY_LEN_ERR;   // vague, but it's all we have here.
    }
  else
    {
      return 0;
    }


}

