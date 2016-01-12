/* md5Checksum.c - checksumming routine on the client side
 */

#include "md5Checksum.h"
#include "rcMisc.h"
#include "base64.h"

#ifdef SHA256_FILE_HASH
#include "sha.h"
#endif

#define MD5_BUF_SZ      (4 * 1024)

#ifdef MD5_TESTING

int main (int argc, char *argv[])
{
    int i;
    char chksumStr[CHKSUM_LEN];

    if (argc != 2) {
	fprintf (stderr, "usage: md5checksum localFile\n");
	exit (1);
    }
 
    status = chksumLocFile (argv[i], chksumStr);

    if (status >= 0) {
        printf ("chksumStr = %s\n", chksumStr);
	exit (0);
    } else {
	fprintf (stderr, "chksumFile error for %s, status = %d.\n",
	  argv[i], status);
        exit (1);
    } 
}

#endif 	/* MD5_TESTING */

#ifdef SHA256_FILE_HASH
void sha256ToStr (unsigned char hash[SHA256_DIGEST_LENGTH],
		  char outputBuffer[CHKSUM_LEN]) {
    int len = strlen(SHA256_CHKSUM_PREFIX);
    rstrcpy(outputBuffer, SHA256_CHKSUM_PREFIX, CHKSUM_LEN);
    long unsigned int l = CHKSUM_LEN - len;
    base64_encode(hash, SHA256_DIGEST_LENGTH, (unsigned char *) outputBuffer + len, &l);

}
int extractHashFunction(keyValPair_t *condInput) {
	char *hash_function = getValByKey (condInput, HASH_KW);
	int use_sha256 = 
#if PREFER_SHA256_FILE_HASH > 0
        1
#else
        0
#endif
    ;
	if(hash_function != NULL) {
		if(strcmp(hash_function, "sha2") == 0) {
			use_sha256 = 1;
		} else { /* default to md5 */
			use_sha256 = 0;
		}
	}
	return use_sha256;
}
int extractHashFunction2(char *myChksum) {
	return strncmp(myChksum, SHA256_CHKSUM_PREFIX, strlen(SHA256_CHKSUM_PREFIX)) == 0?1:0;
}
int extractHashFunction3(rodsArguments_t *rodsArgs) {
    if(rodsArgs->hash == True) {
        return strcmp(rodsArgs->hashValue, "md5") != 0?1:0;
    }
	return 
#if PREFER_SHA256_FILE_HASH > 0
        1
#else
        0
#endif
    ;
}
int verifyHashUse(char *chksum) {
    return extractHashFunction2(chksum) == 1 ? 0 : UNSUPPORTED_HASH_TYPE_USED;
}
#else
int extractHashFunction(keyValPair_t *condInput) {
	return 0;
}
int extractHashFunction2(char *myChksum) {

	if(strncmp(myChksum, SHA256_CHKSUM_PREFIX, strlen(SHA256_CHKSUM_PREFIX)) == 0) {
	rodsLogError (LOG_ERROR, UNSUPPORTED_HASH_TYPE_USED,
        "File has a SHA256 file hash which is not enabled in this program");
	    return UNSUPPORTED_HASH_TYPE_USED;
	}
    

	if(strncmp(myChksum, SHA256_CHKSUM_PREFIX, strlen(SHA256_CHKSUM_PREFIX)) == 0) {
	rodsLogError (LOG_ERROR, UNSUPPORTED_HASH_TYPE_USED,
        "File has a SHA256 file hash which is not enabled in this program");
	    return UNSUPPORTED_HASH_TYPE_USED;
	}
    
	return 0;
}
int extractHashFunction3(rodsArguments_t *rodsArgs) {
	return 0;
}
int verifyHashUse(char *chksum) {
    return extractHashFunction2(chksum) == 0 ? 0 : UNSUPPORTED_HASH_TYPE_USED;
}
#endif


int verifyChksumLocFile(char *fileName, char *myChksum, char *chksumStr) {
    int status;
    char chksumBuf[CHKSUM_LEN];
    int use_sha256;
    
    if(chksumStr == NULL) chksumStr = chksumBuf;
    
    status = extractHashFunction2(myChksum);
    if (status < 0) {
        return status;
    }
    
    use_sha256 = status; 
    
    /* verify the chksum */
	status = chksumLocFile (fileName, chksumStr, use_sha256);
	if (status < 0) {
	    return (status);
	}
	if (strcmp (myChksum, chksumStr) != 0) {
	    return (USER_CHKSUM_MISMATCH);
	}
    return 0;
}

int
chksumLocFile (char *fileName, char *chksumStr, int use_sha256)
{
    FILE *file;
    MD5_CTX context;
    int len;
    unsigned char buffer[MD5_BUF_SZ], digest[16];
    int status;
#ifdef SHA256_FILE_HASH
    unsigned char sha256_hash[SHA256_DIGEST_LENGTH+10];
    SHA256_CTX sha256;
#endif

    if ((file = fopen (fileName, "rb")) == NULL) {
	status = UNIX_FILE_OPEN_ERR - errno;
	rodsLogError (LOG_NOTICE, status,
        "chksumFile; fopen failed for %s. status = %d", fileName, status);
	return (status);
    }

#ifdef SHA256_FILE_HASH
    if (use_sha256) {
       SHA256_Init(&sha256); 
       while ((len = fread (buffer, 1, MD5_BUF_SZ, file)) > 0) {
	  SHA256_Update(&sha256, buffer, len);
       }
       SHA256_Final(sha256_hash, &sha256);

       fclose (file);

       sha256ToStr (sha256_hash, chksumStr);
    }
    else {
       MD5Init (&context);
       while ((len = fread (buffer, 1, MD5_BUF_SZ, file)) > 0) {
	  MD5Update (&context, buffer, len);
       }
       MD5Final (digest, &context);

       fclose (file);

       md5ToStr (digest, chksumStr);
    }
#else
    MD5Init (&context);
    while ((len = fread (buffer, 1, MD5_BUF_SZ, file)) > 0) {
        MD5Update (&context, buffer, len);
    }
    MD5Final (digest, &context);

    fclose (file);

    md5ToStr (digest, chksumStr);
#endif

/*
  rodsLog(LOG_NOTICE, "Testing: chksumLocFile called checksum:%s", chksumStr);
*/

    return (0);
}

int
md5ToStr (unsigned char *digest, char *chksumStr)
{
    int i;
    char *outPtr = chksumStr;

    for (i = 0; i < 16; i++) {
        sprintf (outPtr, "%02x", digest[i]);
	outPtr += 2;
    }

    return (0);
}

int
hashToStr (unsigned char *digest, char *digestStr)
{
    int i;
    char *outPtr = digestStr;

    for (i = 0; i < 16; i++) {
        sprintf (outPtr, "%02x", digest[i]);
	outPtr += 2;
    }

    return (0);
}

/* rcChksumLocFile - chksum a local file and put the result in the
 * condInput.
 * Input - 
 *	char *fileName - the local file name
 *	char *chksumFlag - the chksum flag. valid flags are 
 *         VERIFY_CHKSUM_KW and REG_CHKSUM_KW.
 *	keyValPair_t *condInput - the result is put into this struct
 */

int 
rcChksumLocFile (char *fileName, char *chksumFlag, keyValPair_t *condInput, int use_sha256)
{
    char chksumStr[CHKSUM_LEN];
    int status;

    if (condInput == NULL || chksumFlag == NULL || fileName == NULL) {
	rodsLog (LOG_NOTICE,
	  "rcChksumLocFile: NULL input");
	return (USER__NULL_INPUT_ERR);
    }

    if (strcmp (chksumFlag, VERIFY_CHKSUM_KW) != 0 &&
      strcmp (chksumFlag, REG_CHKSUM_KW) != 0 && 
      strcmp (chksumFlag, RSYNC_CHKSUM_KW) != 0) {
         rodsLog (LOG_NOTICE,
          "rcChksumLocFile: bad input chksumFlag %s", chksumFlag);
        return (USER_BAD_KEYWORD_ERR);
    }
 
    status = chksumLocFile (fileName, chksumStr, use_sha256);

    if (status < 0) {
	return (status);
    }

    addKeyVal (condInput, chksumFlag, chksumStr);

    return (0);
}

