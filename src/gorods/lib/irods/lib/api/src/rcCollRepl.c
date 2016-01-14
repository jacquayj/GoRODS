/**
 * @file  rcCollRepl.c
 *
 */

/* This is script-generated code.  */ 
/* See collRepl.h for a description of this API call.*/

#include "collRepl.h"

/**
 * \fn rcCollRepl (rcComm_t *conn, collInp_t *collReplInp, int vFlag)
 *
 * \brief Recursively replicate a collection.
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
 * Replicate a collection /myZone/home/john/coll1recursively to 
 * myRescource:
 * \n int status;
 * \n collInp_t collReplInp;
 * \n bzero (&collReplInp, sizeof (collReplInp));
 * \n rstrcpy (collReplInp.collName, "/myZone/home/john/coll1", MAX_NAME_LEN);
 * \n addKeyVal (&dataObjInp.condInput, DEST_RESC_NAME_KW, "myRescource");
 * \n status = rcCollRepl (conn, &collReplInp);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] collReplInp - Elements of collInp_t used :
 *    \li char \b collName[MAX_NAME_LEN] - full path of the collection.
 *    \li keyValPair_t \b condInput - keyword/value pair input. Valid keywords:
 *    \n DEST_RESC_NAME_KW - The resource to store the new replica.
 *    \n BACKUP_RESC_NAME_KW - The resource to store the new replica.
 *             In backup mode. If a good copy already exists in this resource
 *             group or resource, don't make another one.
 *    \n ALL_KW - replicate to all resources in the resource group if the
 *             input resource (via DEST_RESC_NAME_KW) is a resource group.
 *            This keyWd has no value.
 *    \n IRODS_ADMIN_KW - admin user backup/replicate other user's files.
 *            This keyWd has no value.
 *    \n RBUDP_TRANSFER_KW - use RBUDP for data transfer. This keyWd has no
 *             value.
 *    \n RBUDP_SEND_RATE_KW - the number of RBUDP packet to send per second
 *          The default is 600000.
 *    \n RBUDP_PACK_SIZE_KW - the size of RBUDP packet. The default is 8192.
 * \param[in] vFlag - Vervose flag. Print progress status.
 * \return integer
 * \retval 0 on success
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
_rcCollRepl (rcComm_t *conn, collInp_t *collReplInp, 
collOprStat_t **collOprStat)
{
    int status;

    collReplInp->oprType = REPLICATE_OPR;

    status = procApiRequest (conn, COLL_REPL_AN, collReplInp, NULL, 
        (void **) collOprStat, NULL);

    return status;
}

int
rcCollRepl (rcComm_t *conn, collInp_t *collReplInp, int vFlag)
{
    int status, retval;
    collOprStat_t *collOprStat = NULL;

    retval = _rcCollRepl (conn, collReplInp, &collOprStat);

    status = cliGetCollOprStat (conn, collOprStat, vFlag, retval);

    return (status);
}

