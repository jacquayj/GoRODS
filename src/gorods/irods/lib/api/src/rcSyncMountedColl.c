/**
 * @file  rcSyncMountedColl.c
 *
 */

/* This is script-generated code.  */ 
/* See syncMountedColl.h for a description of this API call.*/

#include "syncMountedColl.h"

/**
 * \fn rcSyncMountedColl (rcComm_t *conn, dataObjInp_t *dataObjInp)
 *
 * \brief Sync the mounted structured file with the cache and optionally
 * purge the cache used with the mounted structured file.
 *
 * \user client
 *
 * \category collection operations
 *
 * \since 1.0
 *
 * \author  Mike Wan
 * \date    2007
 *
 * \remark none
 *
 * \note none
 *
 * \usage
 * Sync the mounted structured file of the mounted collection 
 * /myZone/home/john/dir1 and purge the cache files afterward:
 * \n dataObjInp_t dataObjInp;
 * \n bzero (&dataObjInp, sizeof (dataObjInp));
 * \n rstrcpy (dataObjInp.objPath, "/myZone/home/john/dir1", MAX_NAME_LEN);
 * \n addKeyVal (&dataObjInp.condInput, PURGE_STRUCT_FILE_CACHE, "");
 * \n status = rcSyncMountedColl (conn, &dataObjInp);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] syncMountedCollInp - Elements of dataObjInp_t used :
 *    \li char \b objPath[MAX_NAME_LEN] - full path of the mounted collection
 *         to sync with the cache.
 *    \li keyValPair_t \b condInput - keyword/value pair input. Valid keywords:
 *    \n PURGE_STRUCT_FILE_CACHE - After the sync is done, purge the cache 
 *          associated with the mounted structured file. 
 *          This keyword has no value.
 * \return integer
 * \retval 0 on success

 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
rcSyncMountedColl (rcComm_t *conn, dataObjInp_t *syncMountedCollInp)
{
    int status;
    status = procApiRequest (conn, SYNC_MOUNTED_COLL_AN, syncMountedCollInp, 
      NULL, (void **) NULL, NULL);

    return (status);
}
