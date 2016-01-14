/**
 * @file  rcModAccessControl.c
 *
 */
/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/

/* This was initially script-generated code.  */ 

/* See modAccessControl.h for a description of this API call.*/

#include "modAccessControl.h"

/**
 * \fn rcModAccessControl (rcComm_t *conn, modAccessControlInp_t *modAccessControlInp)
 *
 * \brief Modify the access control information for an object. 
 *
 * \user client
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
 * Modify object access control:
 * \n See ichmod.c as it sets fields and calls rcModAccessControl to perform
 * \n the various functions of ichmod (see 'ichmod -h). 
 * \n For example:
 * \n 
 * \n modAccessControlInp.recursiveFlag = myRodsArgs.recursive;
 * \n modAccessControlInp.accessLevel = argv[myRodsArgs.optind];
 * \n modAccessControlInp.userName = userName;
 * \n modAccessControlInp.zone = zoneName;
 * \n modAccessControlInp.path = rodsPathInp.srcPath[0].outPath;
 * \n status = rcModAccessControl(conn, &modAccessControlInp);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] modAccessControlInp - All elements of modAccessControl_t are used
 * \return integer
 * \retval 0 on success
 *
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
rcModAccessControl (rcComm_t *conn, modAccessControlInp_t *modAccessControlInp)
{
    int status;
    status = procApiRequest (conn, MOD_ACCESS_CONTROL_AN,  modAccessControlInp, NULL, 
        (void **) NULL, NULL);

    return (status);
}
