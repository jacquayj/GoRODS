/**
 * @file  rcGetHostForPut.c
 *
 */

/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* This is script-generated code.  */ 
/* See getHostForPut.h for a description of this API call.*/

#include "getHostForPut.h"

/**
 * \fn rcGetHostForPut (rcComm_t *conn, dataObjInp_t *dataObjInp,
 *   char **outHost)
 *
 * \brief Get the address of the best server to put (upload) a given
 * data object.
 *
 * \user client
 *
 * \category misc operations
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
 * Get the address of the best server to upload the data object
 * /myZone/home/john/myfile to myRescource:
 * \n dataObjInp_t dataObjInp;
 * \n char *outHost = NULL;
 * \n bzero (&dataObjInp, sizeof (dataObjInp));
 * \n rstrcpy (dataObjInp.objPath, "/myZone/home/john/myfile", MAX_NAME_LEN);
 * \n addKeyVal (&dataObjInp.condInput, DEST_RESC_NAME_KW, "myRescource");
 * \n status = rcGetHostForPut (conn, &dataObjInp, &outHost);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] dataObjInp - Elements of dataObjInp_t used :
 *    \li char \b objPath[MAX_NAME_LEN] - full path of the data object.
 *    \li keyValPair_t \b condInput - keyword/value pair input. Valid keywords:
 *    \n DEST_RESC_NAME_KW - The resource to store this data object
 *    \n RESC_NAME_KW - The resource of the data object to open.
 *    \n REPL_NUM_KW - the replica number of the copy to open.
 * \param[out] outHost - the best host address.
 *
 * \return integer
 * \retval 0 on success
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
rcGetHostForPut (rcComm_t *conn, dataObjInp_t *dataObjInp,
char **outHost)
{
    int status;
    status = procApiRequest (conn, GET_HOST_FOR_PUT_AN,  dataObjInp, NULL, 
        (void **) outHost, NULL);

    return (status);
}

