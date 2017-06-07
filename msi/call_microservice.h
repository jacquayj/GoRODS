
typedef struct {
	char* microserviceName;
	msParam_t** params;
	int paramsLen;
} msiCallInfo_t;

int call_microservice(*msiCallInfo_t);