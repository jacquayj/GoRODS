#!/bin/bash

### Copyright (c) 2016, University of Florida Research Foundation, Inc. ###
### For more information please refer to the LICENSE.md file            ###

# Get directory where build.sh is located (this file)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Compile gorods.o, build libgorods.a with iRods C API
(cd $DIR/src/github.com/jjacquay712/GoRods/lib; rm -f build/libgorods.a; rm -f build/gorods.o; gcc -ggdb -o build/gorods.o -c wrapper.c -I/usr/include/irods -Iinclude; ar rcs build/libgorods.a build/gorods.o)

# Replace 'tester' with your application's own package name
go install github.com/jjacquay712/GoRods && go install tester && bin/tester
