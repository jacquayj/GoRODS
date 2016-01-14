/**
 * @file  execCmd.h
 *
 */

/*** Copyright (c), The Unregents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* execCmd.h
 */

#ifndef EXEC_CMD_H
#define EXEC_CMD_H

/* This is Object File I/O type API call */

#include "rods.h"
#include "procApiRequest.h"
#include "apiNumber.h"
#include "initServer.h"
#include "dataObjInpOut.h"

#ifndef CMD_DIR
#define CMD_DIR		"cmd"
#endif
/* for backward compatibility */
typedef struct {
    char cmd[LONG_NAME_LEN];
    char cmdArgv[HUGE_NAME_LEN];
    char execAddr[LONG_NAME_LEN];	/* if non empty, exec at this addr */
    char hintPath[MAX_NAME_LEN];	/* exec where is file is located */
    int addPathToArgv;			/* whether to add the resolved phy
					 * path to the argv. */
    keyValPair_t condInput;
} execCmd241_t;

/**
 * \var execCmd_t
 * \brief Input struct for the rcExecCmd API for executing commands on the 
 *         server.
 * \since 1.0
 *
 * \remark none
 *
 * \note
 * Elements of execCmd_t:
 * \li char cmd[LONG_NAME_LEN] - The cmd to execute. 
 * \li char cmdArgv[LONG_NAME_LEN] - The input argument for the cmd
 * \li char execAddr[LONG_NAME_LEN] - The address of the server to execute 
 *        this cmd. 
 * \li char hintPath[MAX_NAME_LEN] - Execute where this file is located. 
 * \li int addPathToArgv - whether to add the resolved path to the argv
 * \sa none
 * \bug  no known bugs
 */

typedef struct ExecCmd {
    char cmd[LONG_NAME_LEN];
    char cmdArgv[HUGE_NAME_LEN];
    char execAddr[LONG_NAME_LEN];       /* if non empty, exec at this addr */
    char hintPath[MAX_NAME_LEN];        /* exec where is file is located */
    int addPathToArgv;                  /* whether to add the resolved phy
                                         * path to the argv. */
    int dummy;				/* 64 bit alignment */
    keyValPair_t condInput;
} execCmd_t;

/**
 * \var execCmdOut_t
 * \brief Output struct for the rcExecCmd API for executing commands on the 
 *         server.
 * \since 1.0
 *
 * \remark none
 *
 * \note
 * Elements of execCmdOut_t:
 * \li bytesBuf_t stdoutBuf - the buffer containing the stdout output from
 *        the command.
 * \li bytesBuf_t stderrBuf - the buffer containing the stderr output from
 *        the command.
 * \li int status - The status of the command
 * \sa none
 * \bug  no known bugs
 */

typedef struct ExecCmdOut {
    bytesBuf_t stdoutBuf;
    bytesBuf_t stderrBuf;
    int status;	    /* XXXXXX have to move status to back for 64 bit addr */
} execCmdOut_t;

#define ExecCmd241_PI "str cmd[LONG_NAME_LEN]; str cmdArgv[HUGE_NAME_LEN]; str execAddr[LONG_NAME_LEN]; str hintPath[MAX_NAME_LEN]; int addPathToArgv; struct KeyValPair_PI;"
#define ExecCmd_PI "str cmd[LONG_NAME_LEN]; str cmdArgv[HUGE_NAME_LEN]; str execAddr[LONG_NAME_LEN]; str hintPath[MAX_NAME_LEN]; int addPathToArgv; int dummy; struct KeyValPair_PI;"
#define ExecCmdOut_PI "struct BinBytesBuf_PI; struct BinBytesBuf_PI; int status;" 

#if defined(RODS_SERVER)
#define RS_EXEC_CMD241 rsExecCmd241
#define RS_EXEC_CMD rsExecCmd
/* prototype for the server handler */
int
rsExecCmd241 (rsComm_t *rsComm, execCmd241_t *execCmdInp, 
execCmdOut_t **execCmdOut);
int
rsExecCmd (rsComm_t *rsComm, execCmd_t *execCmdInp, 
execCmdOut_t **execCmdOut);
int
initExecCmdMutex ();
int
_rsExecCmd (rsComm_t *rsComm, execCmd_t *execCmdInp, 
execCmdOut_t **execCmdOut);
int
remoteExecCmd (rsComm_t *rsComm, execCmd_t *execCmdInp,
execCmdOut_t **execCmdOut, rodsServerHost_t *rodsServerHost);
int
execCmd (execCmd_t *execCmdInp, int stdOutFd, int stdErrFd);
int
initCmdArg (char *av[], char *cmdArgv, char *cmdPath);
#else
#define RS_EXEC_CMD241 NULL
#define RS_EXEC_CMD NULL
#endif

#ifdef  __cplusplus
extern "C" {
#endif

/* prototype for the client call */
int
rcExecCmd (rcComm_t *conn, execCmd_t *execCmdInp, execCmdOut_t **execCmdOut);

/* rcExecCmd - Execute a command on the server.
 * Input -
 *   rcComm_t *conn - The client connection handle.
 *   execCmd_t *execCmdInp - the execCmd input
 *
 * OutPut -
 *   bytesBuf_t *cmdOutBBuf - The stdout and stderr of the command is stored
 *    int this bytesBuf.
 *   int status - status of the operation.
 */
int
rcExecCmd241 (rcComm_t *conn, execCmd241_t *execCmdInp, 
execCmdOut_t **execCmdOut);

#ifdef  __cplusplus
}
#endif

#endif	/* EXEC_CMD_H */
