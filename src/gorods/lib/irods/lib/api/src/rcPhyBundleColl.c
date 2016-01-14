/**
 * @file  rcPhyBundleColl.c
 *
 */

/* This is script-generated code.  */ 
/* See phyBundleColl.h for a description of this API call.*/

#include "phyBundleColl.h"

/**
 * \fn rcPhyBundleColl (rcComm_t *conn, 
 * structFileExtAndRegInp_t *phyBundleCollInp)
 *
 * \brief Physically bundle files in a collection into
 * a number of tar files to make it more efficient to store these files on tape.
 * The tar files are placed into the /myZone/bundle/.... collection with file
 * names - collection.aRandomNumber. A new tar file will be created whenever
 * the number of subfiles exceeds 512 or the total size of the subfiles
 * exceeds 4 GBytes. A replica is registered for each bundled sub-files with
 * a fictitious resource - 'bundleResc' and a physical file path pointing to
 * the logical path of the tar bundle file. Whenever this copy of the subfile
 * is accessed, the tar file is untarred and staged automatically to the 
 * 'cache' resource. Each extracted file is registered as a replica of its
 * corresponding subfiles. This API is normally used by sysadmin to bundle
 * other users' files.
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
 * Phybun a collection /myZone/home/john/coll1 and put the bundle files in
 * myRescource:
 * \n int status;
 * \n structFileExtAndRegInp_t phyBundleCollInp;
 * \n bzero (&phyBundleCollInp, sizeof (phyBundleCollInp));
 * \n rstrcpy (collCreateInp.collName, "/myZone/home/john/coll1", MAX_NAME_LEN);
 * \n addKeyVal (&phyBundleCollInp.condInput, DEST_RESC_NAME_KW, "myRescource");
 * \n addKeyVal (&phyBundleCollInp.condInput, RESC_NAME_KW, "myRescource");
 * \n status = rcPhyBundleColl (conn, &collCreateInp);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] phyBundleCollInp - Elements of structFileExtAndRegInp_t used :
 *    \li char \b collection[MAX_NAME_LEN] - full path of the collection.
 *    \li keyValPair_t \b condInput - keyword/value pair input. Valid keywords:
 *    \n DEST_RESC_NAME_KW - The resource to store the new bundle file.
 *    \n RESC_NAME_KW - The copy stored in this resource to be used as source.
 *	  This must be the same as the resource specified with the
 *	  DEST_RESC_NAME_KW to make sure a copy in this resource is available 
 *        for bundling.
 * \return integer
 * \retval 0 on success
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
rcPhyBundleColl (rcComm_t *conn, 
structFileExtAndRegInp_t *phyBundleCollInp)
{
    int status;
    status = procApiRequest (conn, PHY_BUNDLE_COLL_AN, phyBundleCollInp, NULL, 
        (void **) NULL, NULL);

    return (status);
}
