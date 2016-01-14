/* For copyright information please refer to files in the COPYRIGHT directory
 */
#include <stdlib.h>
#include <errno.h>
#include "arithmetics.h"
#include "datetime.h"
#include "hashtable.h"
#include "rules.h"
#include "index.h"
#include "cache.h"
#include "functions.h"
#include "configuration.h"
#include "locks.h"
#include <regex.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <sys/time.h>
#include "re.h"

#ifndef DEBUG
int rule(char *name);

int main(int argv, char **args) {
	printf("loading test rules\n");
	if(loadRuleFromCacheOrFile(RULE_ENGINE_INIT_CACHE, args[1], &coreRuleStrct)!=0) {
		printf("[ERROR]\n");
		return 0;
	}
	rule(args[2]);
    clearResources(RESC_CORE_RULE_SET | RESC_CORE_RULE_SET | RESC_EXT_RULE_SET
        		     | RESC_CORE_FUNC_DESC_INDEX | RESC_APP_FUNC_DESC_INDEX | RESC_EXT_FUNC_DESC_INDEX
        		     | RESC_REGION_APP | RESC_REGION_CORE | RESC_REGION_EXT | RESC_CACHE);
    return 0;
}

int rule(char *name) {
    Region *r = make_region(0, NULL);

    ruleExecInfo_t rei;
    memset(&rei,0, sizeof(ruleExecInfo_t));
    Res* res;
    int ret;
    res = parseAndComputeExpressionAdapter(name, NULL, 0, &rei, 0, r);
    char *str = convertResToString(res);
    if(res->nodeType == N_ERROR) {
        printf("ERROR: %s\n", str);
        ret = RES_ERR_CODE(res);
    } else {

        printf("RET: %s\n", str);
        free(str);
        ret = 0;
    }
    region_free(r);
    return ret;

}
#endif

