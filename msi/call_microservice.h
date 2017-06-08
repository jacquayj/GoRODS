#include "msParam.h"

typedef struct msiCallInfo_t {
	char* microserviceName;
	msParam_t** params;
	int paramsLen;
	void* rei;
} msiCallInfo_t;

int call_microservice(msiCallInfo_t*, char**);
msParam_t** NewParamList(int);
msParam_t* NewParam(char* type);
void SetMsParamListItem(msParam_t**, int, msParam_t*);
void FreeMsParam(msParam_t* msParam);
char* GetMSParamType(msParam_t*);
void ConvertParam(char*, msParam_t**);
void SetupParam(char*, msParam_t*);