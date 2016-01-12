/*** Copyright (c), The Regents of the University of California            ***
 *** For more information please refer to files in the COPYRIGHT directory ***/
/* md5Checksum.h - Extern global declaration for client API */

#ifndef MD5_CHECKSUM_H
#define MD5_CHECKSUM_H

#include <stdio.h>
#include <time.h>
#include <string.h>
#include "rods.h"
#include "global.h"
#include "md5.h"
#include "sha1.h"
#include "parseCommandLine.h"
#define SHA256_CHKSUM_PREFIX "sha2:"
#ifdef  __cplusplus
extern "C" {
#endif
int verifyChksumLocFile(char *fileName, char *myChksum, char *chksumStr);
int
chksumLocFile (char *fileName, char *chksumStr, int use_sha256);
int
md5ToStr (unsigned char *digest, char *chksumStr);
int
hashToStr (unsigned char *digest, char *digestStr);
int
rcChksumLocFile (char *fileName, char *chksumFlag, keyValPair_t *condInput, int use_sha256);

int extractHashFunction(keyValPair_t *condInput);
int extractHashFunction2(char *myChksum);
int extractHashFunction3(rodsArguments_t *rodsArgs);
int verifyHashUse(char *chksumStr);
#ifdef SHA256_FILE_HASH
void sha256ToStr (unsigned char *hash, char chksumStr[CHKSUM_LEN]);
#endif


#ifdef  __cplusplus
}
#endif

#endif	/* MD5_CHECKSUM_H */
