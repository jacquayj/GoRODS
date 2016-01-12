/**
 * @file  rcExecCmd.c
 *
 */

/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* This is script-generated code.  */ 
/* See execCmd.h for a description of this API call.*/

#include "execCmd.h"

/**
 * \fn rcExecCmd (rcComm_t *conn, execCmd_t *execCmdInp, 
 * execCmdOut_t **execCmdOut)
 *
 * \brief Execute a command stored in the $(iRODS)/server/bin/cmd directory 
 *     of a iRODS server.
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
 * Execute the "hello" command on the server. The "hello" command is a simple
 * script store in the $(iRODS)/server/bin/cmd directory of a server.
 * \n int status;
 * \n execCmd_t execCmdInp; 
 * \n execCmdOut_t *execCmdOut = NULL;
 * \n bzero (&execCmdInp, sizeof (execCmdInp));
 * \n rstrcpy (execCmdInp.cmd, "hello", LONG_NAME_LEN);
 * \n rstrcpy (execCmdInp.cmdArgv, "John", LONG_NAME_LEN);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] execCmdInp - Elements of execCmd_t used :
 *    \li char \b cmd[LONG_NAME_LEN] - The cmd to execute.
 *    \li char \b cmdArgv[LONG_NAME_LEN] - The input argument for the cmd.
 *    \li char \b execAddr[LONG_NAME_LEN] - The address of the server to
 *           execute this cmd. This input can be empty
 *    \li char \b hintPath[MAX_NAME_LEN] - Execute where this file is located.
 *    \li int \b addPathToArgv - whether to add the resolved path to the argv.
 * \param[out] execCmdOut - A execCmdOut struct for the output.
 * \return integer
 * \retval 0 on success
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
rcExecCmd (rcComm_t *conn, execCmd_t *execCmdInp, execCmdOut_t **execCmdOut)
{
    int status;
    status = procApiRequest (conn, EXEC_CMD_AN, execCmdInp, NULL, 
        (void **) execCmdOut, NULL);

    return (status);
}

int
rcExecCmd241 (rcComm_t *conn, execCmd241_t *execCmdInp,
execCmdOut_t **execCmdOut)
{
    int status;
    status = procApiRequest (conn, EXEC_CMD241_AN, execCmdInp, NULL,
        (void **) execCmdOut, NULL);

    return (status);
}

