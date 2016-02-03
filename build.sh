### Copyright (c) 2016, University of Florida Research Foundation, Inc. ###
### For more information please refer to the LICENSE.md file            ###

# Get directory where build.sh is located (this file)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Compile gorods.o, build libgorods.a with iRods C API
(cd $DIR/src/gorods/lib; gcc -c -fPIC wrapper.c -o build/gorods.o -I./include -I./irods/lib/core/include -I./irods/lib/api/include -I./irods/lib/md5/include -I./irods/lib/sha1/include -I./irods/server/core/include -I./irods/server/icat/include -I./irods/server/drivers/include -I./irods/server/re/include; cd build; ar rcs libgorods.a *.o)

# Compile & run Go code
(cd $DIR/src/gorods; go install tester && ../../bin/tester)
