/**
 * @file  apiDoc.c
 *
 */

/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/** \mainpage iRODS C APIs Documentation

This documentation is generated from the iRODS code.

\section mainpage Main Project Page
 - http://www.irods.org

\section sortedAPI iRODS C APIs by Alphabet
 - <a href="globals.html">Full Alphabetical List</a>

\section clientAPI iRODS Client C APIs - Normally called by iRODS clients

 \subsection dataObjOpr Data Object Operations
  - #rcDataObjCreate
  - #rcDataObjCreateAndStat
  - #rcDataObjOpen
  - #rcDataObjOpenAndStat
  - #rcDataObjRead
  - #rcDataObjWrite
  - #rcDataObjLseek
  - #rcDataObjFsync
  - #rcDataObjClose
  - #rcDataObjPut
  - #rcDataObjGet
  - #rcDataObjCopy
  - #rcDataObjPhymv
  - #rcDataObjRename
  - #rcDataObjRepl
  - #rcDataObjRsync
  - #rcDataObjTrim
  - #rcDataObjTruncate
  - #rcDataObjUnlink
  - #rcDataObjChksum
  - #rcBulkDataObjPut - Bulk put (upload) a large number of data objects
  - #rcPhyPathReg
  - #rcObjStat
  
 \subsection collectionOpr Collection Operations
  - #rcCollCreate
  - #rcRmColl
  - #rcOpenCollection
  - #rcReadCollection
  - #rcCloseCollection
  - #rcCollRepl
  - #rcPhyBundleColl
  - #rcSyncMountedColl

\subsection clientIcatOpr Client iCat metadata Operations
  - #rcGenQuery
  - #rcSimpleQuery
  - #rcSpecificQuery
  - #rcModAccessControl
  - #rcModAVUMetadata

\subsection dataBaseObjOpr - Database Object Operations
  - #rcDatabaseObjControl
  - #rcDatabaseRescClose
  - #rcDatabaseRescOpen
  - #rcEndTransaction

\subsection authenticationOpr Authentication Operations
  - #rcAuthCheck
  - #rcAuthRequest
  - #rcAuthResponse

\subsection adminOpr Administration Operations
  - #rcGeneralAdmin
  - #rcUserAdmin
  - #rcGeneralRowInsert
  - #rcGeneralRowPurge
  - #rcGeneralUpdate

\subsection clientRuleOpr Client Rule Operations
  - #rcExecMyRule

\subsection batchExecOpr Batch (delayed exec) Request Operations
  - #rcRuleExecDel
  - #rcRuleExecMod
  - #rcRuleExecSubmit

\subsection xmsgOpr Xmsg Operations
  - #rcGetXmsgTicket
  - #rcRcvXmsg
  - #rcSendXmsg

\subsection netcdfOpr NETCDF Operations
  - #rcNcCreate
  - #rcNcOpen
  - #rcNcClose
  - #rcNcInq
  - #rcNcInqId
  - #rcNcInqWithId
  - #rcNcGetVarsByType
  - #rcNcRegGlobalAttr
  - #rcNcOpenGroup
  - #rcNcInqGrps
  - #rcNccfGetVara

\subsection miscClientOpr Misc Client Operations
  - #rcExecCmd
  - #rcStreamRead
  - #rcStreamClose
  - #rcGetHostForGet
  - #rcGetHostForPut
  - #rcProcStat
  - #rcGetMiscSvrInfo

\section serverAPI iRODS Server C APIs - Normally called by iRODS servers (server-server)

\subsection dataTransferOpr Low Level Data Transfer Operations
  - #rcDataCopy
  - #rcDataGet
  - #rcDataPut
  - #rcL3FileGetSingleBuf
  - #rcL3FilePutSingleBuf
  - #rcOprComplete

\subsection fileLevelOpr Low Level File Driver Operations
  - #rcFileChksum
  - #rcFileChmod
  - #rcFileClose
  - #rcFileClosedir
  - #rcFileCreate
  - #rcFileFstat
  - #rcFileFsync
  - #rcFileGet
  - #rcFileGetFsFreeSpace
  - #rcFileLseek
  - #rcFileMkdir
  - #rcFileOpen
  - #rcFileOpendir
  - #rcFilePut
  - #rcFileRead
  - #rcFileReaddir
  - #rcFileRename
  - #rcFileRmdir
  - #rcFileStage
  - #rcFileStageToCache
  - #rcFileStat
  - #rcFileSyncToArch
  - #rcFileTruncate
  - #rcFileUnlink
  - #rcFileWrite

\subsection structFileOpr Low Level Structured File (e.g. tar) Operations
  - #rcStructFileBundle
  - #rcStructFileExtAndReg
  - #rcStructFileExtract
  - #rcStructFileSync
  - #rcSubStructFileClose
  - #rcSubStructFileClosedir
  - #rcSubStructFileCreate
  - #rcSubStructFileFstat
  - #rcSubStructFileGet
  - #rcSubStructFileLseek
  - #rcSubStructFileMkdir
  - #rcSubStructFileOpen
  - #rcSubStructFileOpendir
  - #rcSubStructFilePut
  - #rcSubStructFileRead
  - #rcSubStructFileReaddir
  - #rcSubStructFileRename
  - #rcSubStructFileRmdir
  - #rcSubStructFileStat
  - #rcSubStructFileTruncate
  - #rcSubStructFileUnlink
  - #rcSubStructFileWrite

\subsection serverIcatOpr iCat Operations
  - #rcRegDataObj
  - #rcUnregDataObj
  - #rcBulkDataObjReg
  - #rcRegReplica
  - #rcModDataObjMeta
  - #rcUnbunAndRegPhyBunfile
  - #rcChkObjPermAndStat
  - #rcRegColl
  - #rcModColl

\subsection miscServerOpr Misc Server Operations
  - #rcGetRemoteZoneResc
  - #rcGetRescQuota
  - #rcQuerySpecColl

\section inputDataStruct Input/Output Data Structures
  - #dataObjInp_t
  - #openedDataObjInp_t
  - #bytesBuf_t
  - #fileLseekInp_t
  - #dataObjCopyInp_t
  - #collInp_t
  - #collEnt_t
  - #structFileExtAndRegInp_t
  - #execMyRuleInp_t 
  - #execCmd_t 
  - #execCmdOut_t
  - #fileReadInp_t
  - #fileCloseInp_t 
  - #procStatInp_t 
  - #miscSvrInfo_t

\subsection inputMetadataStruct Metadata Data Structures
  - #modAccessControlInp_t
  - #modAVUMetadataInp_t
  - #specificQueryInp_t

**/
