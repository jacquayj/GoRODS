/**
 * @file  rcProcStat.c
 *
 */

/* This is script-generated code.  */ 
/* See procStat.h for a description of this API call.*/

#include "procStat.h"

/**
 * \fn rcProcStat (rcComm_t *conn, procStatInp_t *procStatInp,
 * genQueryOut_t **procStatOut)
 *
 * \brief Get all the information on client processes running in the federation.
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
 * Get the the client process information running on all servers.
 8 \n status, i;
 * \n procStatInp_t procStatInp;
 * \n genQueryOut_t *procStatOut = NULL;
 * \n sqlResult_t *clientName, *progName;
 * \n bzero (&procStatInp, sizeof (procStatInp));
 * \n addKeyVal (&dataObjInp.condInput, ALL_KW, "");
 * \n status = rcProcStat (conn, &dataObjInp, &outHost);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n }
 * \n # print the client programs and client user names. 
 * \n     if ((progName = getSqlResultByInx (procStatOut, PROG_NAME_INX)) == 
 *         NULL) {
 * \n         return (UNMATCHED_KEY_OR_INDEX);
 * \n     }
 * \n     if ((clientName = getSqlResultByInx (procStatOut, CLIENT_NAME_INX)) ==
 *          NULL) {
 * \n         return (UNMATCHED_KEY_OR_INDEX);
 * \n     }
 * \n     for (i = 0; i < rowCnt; i++) {
 * \n         char *clientNameVal, *progNameVal;
 * \n         progNameVal = progName->value + progName->len * i;
 * \n         clientNameVal = clientName->value + clientName->len * i;
 * \n         printf (client program=%s user=%s\n", progNameVal, clientNameVal);
 * \n     }
 * 
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] procStatInp - Elements of procStatInp_t used :
 *    \li char \b addr[LONG_NAME_LEN] - get the clients running only on this
 *        server. If empty, get the clients on the connected server. 
 *    \li char \b rodsZone[NAME_LEN] - get the clients running in this zone.
 *    \li keyValPair_t \b condInput - keyword/value pair input. Valid keywords:
 *    \n RESC_NAME_KW - get the clients on this resource server. 
 *    \n ALL_KW -  get the clients on all servers. This keyword has no value.
 * \param[out] procStatOut - arrays of client results given in a genQueryOut_t.
 * The index identifying the result arrays are:
 *	\n PID_INX - the pid of the agent
 *      \n STARTTIME_INX - the start time
 *	\n CLIENT_NAME_INX - the client user name
 *      \n CLIENT_ZONE_INX - the client user zone
 *      \n PROXY_NAME_INX - the proxy user name
 *	\n PROXY_ZONE_INX - the proxy user zone
 *	\n REMOTE_ADDR_INX - the client address
 *	\n PROG_NAME_INX - the client process name
 *	\n SERVER_ADDR_INX - the server address
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
rcProcStat (rcComm_t *conn, procStatInp_t *procStatInp,
genQueryOut_t **procStatOut)
{
    int status;
    status = procApiRequest (conn, PROC_STAT_AN, procStatInp, 
      NULL, (void **) procStatOut, NULL);

    return (status);
}
