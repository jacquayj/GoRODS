#include "rsModAVUMetadata.hpp"
#include "irods_re_structs.hpp"
#include "irods_ms_plugin.hpp"

int actionTableLookUp(irods::ms_table_entry&, char *action);

#include <boost/any.hpp>

extern "C" {

#include "call_microservice.h"

int call_microservice(msiCallInfo_t* callInfo, char** errStr) {
	
	irods::ms_table_entry ms_entry;
	
	int status = 0;
	int actInx = actionTableLookUp(ms_entry, callInfo->microserviceName);

	ruleExecInfo_t* rei = (ruleExecInfo_t*)(callInfo->rei);

	if ( actInx >= 0 ) {

		std::vector<msParam_t*> arguments;

		for ( int n = 0; n < callInfo->paramsLen; n++ ) {
			arguments.push_back(callInfo->params[n]);
		}

		// Call microservice
		status = ms_entry.call(rei, arguments);

		if ( status < 0 ) {
			*errStr = (char*)"Error calling microservice";
		}
	} else {
		*errStr = (char*)"Unable to locate microservice";
		status = -1;
	}
	
	return status;
}

msParam_t** NewParamList(int len) {
	msParam_t** paramArr;

	paramArr = (msParam_t**)malloc(sizeof(msParam_t*) * len);

	return paramArr;
}

msParam_t* NewParam(char* type) {
	msParam_t* param = (msParam_t*)malloc(sizeof(msParam_t));

	if ( strcmp(type, KeyValPair_MS_T) == 0 ) {
		keyValPair_t* data = (keyValPair_t*)malloc(sizeof(*data));
		memset(data, 0, sizeof(*data));
		
		fillMsParam(param, NULL, type, data, NULL);
	} else if ( strcmp(type, INT_MS_T) == 0 ) {
		int* data = (int*)malloc(sizeof(int));
		memset(data, 0, sizeof(int));
		
		fillMsParam(param, NULL, type, data, NULL);
	} else if ( strcmp(type, DataObjInp_MS_T) == 0 ) {
		dataObjInp_t* data = (dataObjInp_t*)malloc(sizeof(dataObjInp_t));
		memset(data, 0, sizeof(dataObjInp_t));
		
		fillMsParam(param, NULL, type, data, NULL);
	} else {
		fillMsParam(param, NULL, type, NULL, NULL);
	}

	return param;
}

void FreeMsParam(msParam_t* msParam) {

	if ( strcmp(msParam->type, KeyValPair_MS_T) == 0 ) {
		keyValPair_t* kvp = (keyValPair_t*)msParam->inOutStruct;

		for ( int n = 0; n < kvp->len; n++ ) {
			free(kvp->keyWord[n]);
			free(kvp->value[n]);
		}

		free(kvp->keyWord);
		free(kvp->value);

		free(kvp);

	} else if ( strcmp(msParam->type, STR_MS_T) == 0 ) {
		if ( msParam->inOutStruct != NULL ) {
			free(msParam->inOutStruct);
		}
	}

	free(msParam);
}

void SetMsParamListItem(msParam_t** list, int inx, msParam_t* ptr) {
	list[inx] = ptr;
}

char* GetMSParamType(msParam_t* param) {
	return param->type;
}

}