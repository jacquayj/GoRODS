package msi

type ParamType string

// String returns the string value of the ParamType
func (t ParamType) String() string {
	if t == UNDEFINED_T {
		return "UNDEFINED_PI"
	}

	return string(t)
}

// Types avaliable for use in msi.NewParam()
const (
	UNDEFINED_T                 ParamType = ""
	STR_MS_T                              = "STR_PI"
	INT_MS_T                              = "INT_PI"
	INT16_MS_T                            = "INT16_PI"
	CHAR_MS_T                             = "CHAR_PI"
	BUF_LEN_MS_T                          = "BUF_LEN_PI"
	STREAM_MS_T                           = "INT_PI"
	DOUBLE_MS_T                           = "DOUBLE_PI"
	FLOAT_MS_T                            = "FLOAT_PI"
	BOOL_MS_T                             = "BOOL_PI"
	DataObjInp_MS_T                       = "DataObjInp_PI"
	DataObjCloseInp_MS_T                  = "DataObjCloseInp_PI"
	DataObjCopyInp_MS_T                   = "DataObjCopyInp_PI"
	DataObjReadInp_MS_T                   = "dataObjReadInp_PI"
	DataObjWriteInp_MS_T                  = "dataObjWriteInp_PI"
	DataObjLseekInp_MS_T                  = "fileLseekInp_PI"
	DataObjLseekOut_MS_T                  = "fileLseekOut_PI"
	KeyValPair_MS_T                       = "KeyValPair_PI"
	TagStruct_MS_T                        = "TagStruct_PI"
	CollInp_MS_T                          = "CollInpNew_PI"
	ExecCmd_MS_T                          = "ExecCmd_PI"
	ExecCmdOut_MS_T                       = "ExecCmdOut_PI"
	RodsObjStat_MS_T                      = "RodsObjStat_PI"
	VaultPathPolicy_MS_T                  = "VaultPathPolicy_PI"
	StrArray_MS_T                         = "StrArray_PI"
	IntArray_MS_T                         = "IntArray_PI"
	GenQueryInp_MS_T                      = "GenQueryInp_PI"
	GenQueryOut_MS_T                      = "GenQueryOut_PI"
	XmsgTicketInfo_MS_T                   = "XmsgTicketInfo_PI"
	SendXmsgInfo_MS_T                     = "SendXmsgInfo_PI"
	GetXmsgTicketInp_MS_T                 = "GetXmsgTicketInp_PI"
	SendXmsgInp_MS_T                      = "SendXmsgInp_PI"
	RcvXmsgInp_MS_T                       = "RcvXmsgInp_PI"
	RcvXmsgOut_MS_T                       = "RcvXmsgOut_PI"
	StructFileExtAndRegInp_MS_T           = "StructFileExtAndRegInp_PI"
	RuleSet_MS_T                          = "RuleSet_PI"
	RuleStruct_MS_T                       = "RuleStruct_PI"
	DVMapStruct_MS_T                      = "DVMapStruct_PI"
	FNMapStruct_MS_T                      = "FNMapStruct_PI"
	MsrvcStruct_MS_T                      = "MsrvcStruct_PI"
	NcOpenInp_MS_T                        = "NcOpenInp_PI"
	NcInqIdInp_MS_T                       = "NcInqIdInp_PI"
	NcInqWithIdOut_MS_T                   = "NcInqWithIdOut_PI"
	NcInqInp_MS_T                         = "NcInqInp_PI"
	NcCloseInp_MS_T                       = "NcCloseInp_PI"
	NcGetVarInp_MS_T                      = "NcGetVarInp_PI"
	NcGetVarOut_MS_T                      = "NcGetVarOut_PI"
	NcInqOut_MS_T                         = "NcInqOut_PI"
	NcInqGrpsOut_MS_T                     = "NcInqGrpsOut_PI"
	Dictionary_MS_T                       = "Dictionary_PI"
	DictArray_MS_T                        = "DictArray_PI"
	GenArray_MS_T                         = "GenArray_PI"
	DataObjInfo_MS_T                      = "DataObjInfo_PI"
)

// iRODS constants for use in microservice return value
const (
	SUCCESS                 = 0
	SYS_INTERNAL_ERR        = -154000
	SYS_INVALID_INPUT_PARAM = -130000
)
