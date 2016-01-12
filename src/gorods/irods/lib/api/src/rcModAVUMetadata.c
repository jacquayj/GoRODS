/**
 * @file  rcModAVUMetadata.c
 *
 */
/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/* This was initially script-generated code.  */ 

/* See modAVUMetadata.h for a description of this API call.*/


/**
 * \fn rcModAVUMetadata (rcComm_t *conn, modAVUMetadataInp_t *modAVUMetadataInp)
 *
 * \brief Modify Attribute-Value-Unit information associated with an object; 
 * \n     the user-defined metadata.
 *
 * \user clients, in the 'C' code this is used by 'imeta'
 *
 * \category metadata operations
 *
 * \since 1.0
 *
 * \author  Wayne Schroeder
 * \date    2007
 *
 * \remark none
 *
 * \note none
 *
 * \usage
 * Modify AVUs:
 * \n See imeta.c as it sets fields and calls rcModAVUMetadata to perform
 * \n the various functions of imeta (see 'imeta -h').  rsModAVUMetadata
 * \n calls chlAddAVUMetadata, chlDeleteAVUMetadata, chlCopyAVUMetadata or
 * \n chlModAVUMetadata. 
 * \n Example client call, corresponding to 
 * \n 'imeta add -d f1 attrName attrValue':
 * \n modAVUMetadataInp.arg0 = "add";
 * \n modAVUMetadataInp.arg1 = "-d"
 * \n modAVUMetadataInp.arg2 = "/tempZone/home/rods/f1";
 * \n modAVUMetadataInp.arg3 = "attrName"
 * \n modAVUMetadataInp.arg4 = "attrValue";
 * \n modAVUMetadataInp.arg5 = "";
 * \n modAVUMetadataInp.arg6 = "";
 * \n modAVUMetadataInp.arg7 = "";
 * \n modAVUMetadataInp.arg8 = "";
 * \n modAVUMetadataInp.arg9 = "";
 * \n status = rcModAVUMetadata(conn, &modAVUMetadataInp);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] modAVUMetadataInp - The various elements of modAVUMetadata_t 
 * \n are used
 * \return integer
 * \retval 0 on success
 *
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

#include "modAVUMetadata.h"

int
rcModAVUMetadata (rcComm_t *conn, modAVUMetadataInp_t *modAVUMetadataInp)
{
    int status;
    status = procApiRequest (conn, MOD_AVU_METADATA_AN, modAVUMetadataInp,
			     NULL, (void **) NULL, NULL);

    return (status);
}
