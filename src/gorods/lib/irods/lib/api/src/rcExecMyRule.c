/**
 * @file  rcExecMyRule.c
 *
 */

/* This is script-generated code.  */ 
/* See execMyRule.h for a description of this API call.*/

#include "execMyRule.h"
#include "oprComplete.h"
#include "dataObjPut.h"
#include "dataObjGet.h"

/**
 * \fn rcExecMyRule (rcComm_t *conn, execMyRuleInp_t *execMyRuleInp,
 * msParamArray_t **outParamArray)
 *
 * \brief Submit a user defined rule to be executed by an iRODS server.
 *
 * \user client
 *
 * \category rule operations
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
 * Execute a rule that create a collection /myZone/home/john/coll1
 * \n int status;
 * \n execMyRuleInp_t execMyRuleInp;
 * \n msParamArray_t msParamArray;
 * \n msParamArray_t *outParamArray = NULL;
 * \n bzero (&execMyRuleInp, sizeof (execMyRuleInp));
 * \n rstrcpy (execMyRuleInp.myRule, "myTestRule { msiCollCreate(*Path,"0",*Status); writeLine("stdout","Create collection *Path"); }", META_STR_LEN);
 * \n memset (&msParamArray, 0, sizeof (msParamArray));
 * \n execMyRuleInp.inpParamArray = &msParamArray;
 * \n addMsParamToArray (&msParamArray, "Path", STR_MS_T, "/myZone/home/john/coll1", 1);
 * \n rstrcpy (execMyRuleInp.outParamDesc, "ruleExecOut", LONG_NAME_LEN);
 * \n status = rcExecMyRule (conn, &execMyRuleInp, &outParamArray);
 * \n if (status < 0) {
 * \n .... handle the error
 * \n } else {
 * \n  ... hnadle outParamArray output
 * \n }
 *
 * \param[in] conn - A rcComm_t connection handle to the server.
 * \param[in] execMyRuleInp - Elements of execMyRuleInp_t used :
 *    \li char \b myRule[META_STR_LEN] - The string representing the rule to 
 *       be executed..
 *    \li rodsHostAddr_t \b addr - The address to execute the rule. Can be
 *       left blank.
 *    \li char \b outParamDesc[LONG_NAME_LEN] - the list of output parameters
 *       separated by "%".
 *    \li msParamArray_t \b *inpParamArray - Input parameters for the rule in
 *       the form of array of msParam.
 *    \li keyValPair_t \b condInput - keyword/value pair input. Valid keywords:
 *    \n "looptest" - Just a test.
 * \param[out] outParamArray - A msParamArray_t containing an array of msParam_t.
 * \return integer
 * \retval 0 on success
 * \sideeffect none
 * \pre none
 * \post none
 * \sa none
 * \bug  no known bugs
**/

int
rcExecMyRule (rcComm_t *conn, execMyRuleInp_t *execMyRuleInp, 
msParamArray_t **outParamArray)
{
    int status;
    char myDir[MAX_NAME_LEN], myFile[MAX_NAME_LEN];

    status = procApiRequest (conn, EXEC_MY_RULE_AN, execMyRuleInp, NULL, 
        (void **)outParamArray, NULL);
 
    while (status == SYS_SVR_TO_CLI_MSI_REQUEST) {
	/* it is a server request */
	char *locFilePath;
        msParam_t *myMsParam;
        dataObjInp_t *dataObjInp = NULL;


	if ((myMsParam = getMsParamByLabel (*outParamArray, CL_PUT_ACTION))
	  != NULL) { 

	    dataObjInp = (dataObjInp_t *) myMsParam->inOutStruct;
	    if ((locFilePath = getValByKey (&dataObjInp->condInput, 
	      LOCAL_PATH_KW)) == NULL) {
                if ((status = splitPathByKey (dataObjInp->objPath,
                  myDir, myFile, '/')) < 0) {
                    rodsLogError (LOG_ERROR, status,
                      "rcExecMyRule: splitPathByKey for %s error",
                      dataObjInp->objPath);
                    rcOprComplete (conn, USER_FILE_DOES_NOT_EXIST);
                } else {
                    locFilePath = (char *) myFile;
                }
	    }
	    status = rcDataObjPut (conn, dataObjInp, locFilePath);
	    rcOprComplete (conn, status);
	} else if ((myMsParam = getMsParamByLabel (*outParamArray, 
	  CL_GET_ACTION)) != NULL) {
            dataObjInp = (dataObjInp_t *) myMsParam->inOutStruct;
            if ((locFilePath = getValByKey (&dataObjInp->condInput,
              LOCAL_PATH_KW)) == NULL) {
                if ((status = splitPathByKey (dataObjInp->objPath, 
                  myDir, myFile, '/')) < 0) {
        	    rodsLogError (LOG_ERROR, status,
                      "rcExecMyRule: splitPathByKey for %s error",
                      dataObjInp->objPath);
                    rcOprComplete (conn, USER_FILE_DOES_NOT_EXIST);
		} else {
		    locFilePath = (char *) myFile;
		}
            }
            status = rcDataObjGet (conn, dataObjInp, locFilePath);
            rcOprComplete (conn, status);
	} else {
	    rcOprComplete (conn, SYS_SVR_TO_CLI_MSI_NO_EXIST);
	}
	/* free outParamArray */
	if (dataObjInp != NULL) {
	    clearKeyVal (&dataObjInp->condInput); 
	}
	clearMsParamArray (*outParamArray, 1);
	free (*outParamArray);
	*outParamArray = NULL;

	/* read the reply from the eariler call */

        status = branchReadAndProcApiReply (conn, EXEC_MY_RULE_AN, 
        (void **)outParamArray, NULL);
        if (status < 0) {
            rodsLogError (LOG_DEBUG, status,
              "rcExecMyRule: readAndProcApiReply failed. status = %d", 
	      status);
        }
    }

	 
    return (status);
}

