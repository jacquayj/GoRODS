/*
 * data.h
 *
 *  Created on: May 31, 2013
 *      Author: ertri
 */

#ifndef DATA_H_
#define DATA_H_

#include "parser.h"
PARSER_FUNC_PROTO(CStructDef);
PARSER_FUNC_PROTO(CVariableDef);
PARSER_FUNC_PROTO(CType);
PARSER_FUNC_PROTO(CBaseType);
PARSER_FUNC_PROTO(CArrayDim);
PARSER_FUNC_PROTO(CGAnnotation);

#define MAX_CODE_APPEND_LENGTH 1024
#define CODE_BUF_SIZE 1024
typedef struct CodeBuffer {
	int size;
	int pos;
	char* buffer;
} codeBuffer_t;

codeBuffer_t *newCodeBuffer();

void deleteCodeBuffer(codeBuffer_t *buf);

int appendToCodeBuffer(codeBuffer_t *codeBuffer, char *code);


#endif /* DATA_H_ */
