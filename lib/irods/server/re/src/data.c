#include <libxslt/xslt.h>
#include <libxslt/xsltInternals.h>
#include <libxslt/transform.h>
#include <libxslt/xsltutils.h>

#include "data.h"

/*
 * S ::= struct Id? { F* }
 * F ::= Type Id Dim* Annotation* ;
 * Annotation ::= Id AnnotationParamList?
 * AnnotationParamList ::= ( Id , ... , Id )
 * Type ::= BaseType "*"*
 * BaseType ::= Id | struct Id
 * Dim ::= [ Id ]
 */

PARSER_FUNC_BEGIN(CDefSet)
	int rulegen = 1;
	int i=0;
	LOOP_BEGIN(d)
		TRY(t)
			TTEXT("typedef");
			NT(CStructDef);
			i++;
		OR(t)
			TTYPE(TK_EOS);
			printf("defset: eos\n");
			DONE(d);
		OR(t)
			NEXT_TOKEN_BASIC;
			printf("defset: token skipped = %s\n", token->text);
		END_TRY(t)
	LOOP_END(d)
	printf("defset : stackTop = %d, counter = %d, diff = %d\n", context->nodeStackTop, i, context->nodeStackTop - i);
	BUILD_NODE(C_DEF_SET, "{}", &start, i, i);
PARSER_FUNC_END(CDefSet)

PARSER_FUNC_BEGIN(CIfdefBlock)
	int rulegen = 1;
LOOP_BEGIN(loop)
	TRY(p)
		TTEXT2("#if", "#ifdef");
		printf("token = %s\n", token->text);
		/* assume that this is a #if 0 block */
		/* assume that the macro is not defined */
		LOOP_BEGIN(l1)
			TRY(e)
				TTEXT2("#else", "#endif");
				printf("token = %s\n", token->text);
				DONE(l1);
			OR(e)
				NEXT_TOKEN_BASIC;
				printf("token skipped = %s\n", token->text);
			END_TRY(e)
		LOOP_END(l1)
	OR(p)
		TTEXT("#ifndef");
		/* assume that the macro is not defined and it has not #else branch */
		NEXT_TOKEN_BASIC;
	OR(p)
		TTEXT("#endif");
	OR(p)
		DONE(loop);
	END_TRY(p)
	LOOP_END(loop)
PARSER_FUNC_END(CIfdefBlock)

PARSER_FUNC_BEGIN(CStructDef)
	int rulegen = 1;
	char structName[NAME_LEN];
	TTEXT("struct");
	TRY(n)
		TTYPE(TK_TEXT);
		rstrcpy(structName, token->text, NAME_LEN);
	OR(n)
		rstrcpy(structName, "", NAME_LEN);
	END_TRY(n)
	TTEXT("{");
	REPEAT_BEGIN(f)
		NT(CIfdefBlock);
		NT(CVariableDef);
	REPEAT_END(f)
	NT(CIfdefBlock);
	TTEXT("}");
	if(structName[0] == '\0') {
		TTYPE(TK_TEXT);
		rstrcpy(structName, token->text, NAME_LEN);
		structName[strlen(structName)-2] = '\0';
		if(isalpha(structName[0])) {
			structName[0] = toupper(structName[0]);
		}
	}
	printf("struct : stackTop = %d, counter = %d, diff = %d\n", context->nodeStackTop, COUNTER(f), context->nodeStackTop - COUNTER(f));
	BUILD_NODE(C_STRUCT_DEF, structName, &start, COUNTER(f), COUNTER(f));
PARSER_FUNC_END(CStructDef)

PARSER_FUNC_BEGIN(CVariableDef)
	int rulegen = 1;
	char memberName[NAME_LEN];
	NT(CType);
	TTYPE(TK_TEXT);
	rstrcpy(memberName, token->text, NAME_LEN);
	REPEAT_BEGIN(f)
		NT(CArrayDim);
		BUILD_NODE(C_ARRAY_TYPE, "[]", &start, 2, 2);
	REPEAT_END(f)
	REPEAT_BEGIN(a)
		NT(CGAnnotation);
	REPEAT_END(a)
	BUILD_NODE(CG_ANNOTATIONS, "annotations", &start, COUNTER(a), COUNTER(a));
	TTEXT(";");
	BUILD_NODE(C_STRUCT_MEMBER, memberName, &start, 2, 2);
PARSER_FUNC_END(CVariableDef)

PARSER_FUNC_BEGIN(CGAnnotation)
	int rulegen = 1;
	int n;
	char annotationName[NAME_LEN];
	TTYPE(TK_TEXT);
	rstrcpy(annotationName, token->text, NAME_LEN);
	TRY(plt)
		LIST_BEGIN(pl)
			NT(CGAnnotation)
		LIST_DELIM(pl)
			TTEXT(",");
		LIST_END(pl)
		n = COUNTER(pl);
	OR(plt)
		n = 0;
	END_TRY(plt)
	BUILD_NODE(CG_ANNOTATION, annotationName, &start, n, n);
PARSER_FUNC_END(CGAnnotation)

PARSER_FUNC_BEGIN(CType)
	int rulegen = 1;
	NT(CBaseType);
	REPEAT_BEGIN(f)
		TTEXT("*");
		BUILD_NODE(C_POINTER_TYPE, "*", &start, 1, 1);
	REPEAT_END(f)
PARSER_FUNC_END(CType)

PARSER_FUNC_BEGIN(CBaseType)
	int rulegen = 1;
	char type[NAME_LEN];
	TRY(bt)
		TTEXT("struct");
		TTYPE(TK_TEXT);
		BUILD_NODE(C_STRUCT_TYPE, token->text, &start, 0, 0);
	OR(bt)
		TTEXT("unsigned");
		TTYPE(TK_TEXT);
		snprintf(type, NAME_LEN, "%s %s", "unsigned", token->text);
		BUILD_NODE(C_BASE_TYPE, type, &start, 0, 0);
	OR(bt)
		TTEXT("long");
		TTEXT("int");
		snprintf(type, NAME_LEN, "%s %s", "long", token->text);
		BUILD_NODE(C_BASE_TYPE, type, &start, 0, 0);
	OR(bt)
		TTYPE(TK_TEXT);
		BUILD_NODE(C_BASE_TYPE, token->text, &start, 0, 0);
	END_TRY(bt)
PARSER_FUNC_END(CBaseType)

PARSER_FUNC_BEGIN(CArrayDim)
	int rulegen = 1;
	char dimName[NAME_LEN];
	Label l;
	TTEXT("[");
	l = *FPOS;
	TTYPE(TK_TEXT);
	rstrcpy(dimName, token->text, NAME_LEN);
	TTEXT("]");
	BUILD_NODE(C_ARRAY_TYPE_DIM, dimName, &l, 0, 0);
PARSER_FUNC_END(CArrayDim)

codeBuffer_t *newCodeBuffer() {
	codeBuffer_t *buf = (codeBuffer_t *) malloc(sizeof(codeBuffer_t));
	buf->size = CODE_BUF_SIZE;
	buf->pos = 0;
	buf->buffer = (char *) malloc(buf->size);
	return buf;
}

void deleteCodeBuffer(codeBuffer_t *buf) {
	free(buf->buffer);
	free(buf);
}

int appendToCodeBuffer(codeBuffer_t *codeBuffer, char *code) {
	int n = strnlen(code, MAX_CODE_APPEND_LENGTH);
	if(n == MAX_CODE_APPEND_LENGTH) {
		// error
		return -1;
	} else {
		while(codeBuffer->size - codeBuffer->pos < n) {
			codeBuffer->buffer = (char *)realloc(codeBuffer->buffer, codeBuffer->size * 2);
			codeBuffer->size *= 2;
		}
		memcpy(codeBuffer->buffer+codeBuffer->pos, code, n);
		codeBuffer->pos += n;
		return 0;
	}
}

char *getTag(NodeType nt) {
	switch(nt) {
	case C_STRUCT_DEF:
		return "struct";
	case C_STRUCT_MEMBER:
		return "struct_member";
	case C_ARRAY_TYPE:
		return "array_type";
	case C_ARRAY_TYPE_DIM:
		return "array_dim";
	case C_POINTER_TYPE:
		return "pointer_type";
	case C_STRUCT_TYPE:
		return "struct_type";
	case C_BASE_TYPE:
		return "base_type";
	case C_DEF_SET:
		return "def_set";
	case CG_ANNOTATIONS:
		return "annotations";
	case CG_ANNOTATION:
		return "annotation";
	default:
		return NULL;
	}
}
int toXML(Node *ast, codeBuffer_t *codeBuffer) {
	int i;
	char *tag = getTag(getNodeType(ast));
	if(tag == NULL) {
		printf("unsupported node type %d\n", getNodeType(ast));
		return RE_UNSUPPORTED_AST_NODE_TYPE;
	}

	char stag[1024], ftag[1024];
	snprintf(stag, 1024, "<%s name='%s'>", tag, ast->text);
	snprintf(ftag, 1024, "</%s>", tag);

	appendToCodeBuffer(codeBuffer, stag);
	for(i = 0; i < ast->degree; i++) {
		toXML(ast->subtrees[i], codeBuffer);
	}
	appendToCodeBuffer(codeBuffer, ftag);
	return 0;
}

int testData() {
	Region *r = make_region(0, NULL);

	rError_t errmsgBuf;
	errmsgBuf.errMsg = NULL;
	errmsgBuf.len = 0;

	ParserContext *pc = newParserContext(&errmsgBuf, r);

	Pointer *e = newPointer2("struct abc { int x; long * y; short * * y1; struct abc z; struct abc w [X][Y];int * * w1[X];}");

	nextRuleGenCStructDef(e, pc);

	Node *n = pc->nodeStack[0];

	printTree(n, 0);

	codeBuffer_t *cb = newCodeBuffer();
	toXML(n, cb);
	printf("\n%s\n", cb->buffer);

	return 0;

}
#ifdef CODE_GENERATION
int main(int argc, char **args) {
	if(argc<4) {
		printf("usage:\nRuleEngine <xslt file> <output file> <header file> ...\n");
		return 0;
	}
	char *xsltFileName = args[1];
	char *outputFileName = args[2];
	char **headerFileName = args+3;
	int nHFs = argc - 3;

	Region *r = make_region(0, NULL);

	rError_t errmsgBuf;
	errmsgBuf.errMsg = NULL;
	errmsgBuf.len = 0;
	int i;
	codeBuffer_t *cb = newCodeBuffer();
	appendToCodeBuffer(cb, "<root>");
	for(i=0;i<nHFs;i++) {
		ParserContext *pc = newParserContext(&errmsgBuf, r);

		FILE *fp = fopen(headerFileName[i], "r");
		if(fp==NULL) {
			printf("cannot open header file\n");
			return 0;
		}
		Pointer *e = newPointer(fp, headerFileName[i]);

		nextRuleGenCDefSet(e, pc);

		if(pc->errmsg->len > 0) {
			char buf[1024];
			errMsgToString(pc->errmsg, buf, 1024);

			printf("error: %s\n", buf);
		}
		if(pc->nodeStackTop != 1) {
			printf("error: nodeStackTop = %d\n", pc->nodeStackTop);

		}

		Node *n = pc->nodeStack[0];

		printTree(n, 0);

		toXML(n, cb);

		deletePointer(e);
		deleteParserContext(pc);

	}
	appendToCodeBuffer(cb, "</root>");
	printf("\n%s\n", cb->buffer);

	char *xml = cb->buffer;

	xsltStylesheetPtr cur = NULL;
	xmlDocPtr doc, res;

	xmlSubstituteEntitiesDefault(1);

	cur = xsltParseStylesheetFile((const xmlChar *) xsltFileName);

	doc = xmlParseDoc((const xmlChar *) xml);
	res = xsltApplyStylesheet(cur, doc, NULL);
	xsltSaveResultToFilename(outputFileName, res, cur, 0);

	xsltFreeStylesheet(cur);
	xmlFreeDoc(res);
	xmlFreeDoc(doc);

    xsltCleanupGlobals();
    xmlCleanupParser();

/*	if(fp==NULL) {
		printf("cannot open header file\n");
		return 0;
	}
		*/

	return 0;

}

#endif

