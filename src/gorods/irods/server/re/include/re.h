/* For copyright information please refer to files in the COPYRIGHT directory
 */
#ifndef RE_H_
#define RE_H_



#ifdef DEBUG
#ifndef RULE_ENGINE_N
#define RULE_ENGINE_N
#endif
#include <string.h>
#include <time.h>
#include <rodsKeyWdDef.h>
#include <rodsErrorTable.h>
#include <rodsDef.h>
#include <rodsError.h>
#include <execCmd.h>
#include <objInfo.h>
#include <reGlobalsExtern.h>
#include <msParam.h>
#include <reFuncDefs.h>
#include <execMyRule.h>
#include <generalRowInsert.h>
#include <miscUtil.h>
//#define USE_BOOST 1
#define CACHE_ENABLE 1
#ifndef _POSIX_VERSION
#define _POSIX_VERSION 0
#endif
// #define DEBUG_INDEX
// #define solaris_platform

#undef RE_TEST_MACRO
#define RE_TEST_MACRO(x)

int print(msParam_t *s, ruleExecInfo_t *rei);
int writeLine(msParam_t* out, msParam_t *str, ruleExecInfo_t *rei);

// declare a function defined in SystemMS.c
int _delayExec(char *inActionCall, char *recoveryActionCall,
	       char *delayCondition,  ruleExecInfo_t *rei);

int chlSetAVUMetadata(rsComm_t *rsComm, char *type, char *ame, char *attr, char *value, char *unit);

#undef rodsLog
#undef rodsLogAndErrorMsg
#undef LOG_NOTICE
#undef LOG_DEBUG
#undef LOG_WARNING
#undef LOG_ERROR

#define LOG_NOTICE "NOTICE"
#define LOG_DEBUG "DEBUG"
#define LOG_WARNING "WARNING"
#define LOG_ERROR "ERROR"
#define rodsLog(a,...) printf("%s: ", a); printf(__VA_ARGS__)
#define rodsLogAndErrorMsg(a,b,c,...) printf(__VA_ARGS__)

#ifndef HAS_MICROSDEF_T
#define HAS_MICROSDEF_T
typedef struct {
  char action[MAX_ACTION_SIZE];
  int numberOfStringArgs;
  funcPtr callAction;
} microsdef_t;
#endif

extern int NumOfAction;
extern microsdef_t MicrosTable[];

#endif


#endif /* RE_H_ */
