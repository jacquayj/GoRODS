#!/bin/bash

### Copyright (c) 2016, University of Florida Research Foundation, Inc. ###
### For more information please refer to the LICENSE.md file            ###

# Replace 'tester' with your application's own package name
YOUR_APP_PACKAGE=tester

# Get directory where build.sh is located (this file)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Get GoRODS lib directory path
if [ -f $DIR/src/github.com/jjacquay712/GoRODS/lib/wrapper.c ]; then
	GORODS_LIB_PATH=$DIR/src/github.com/jjacquay712/GoRODS/lib
else
	GORODS_LIB_PATH=./lib
fi

# Make sure it contains a required file
if [ ! -f $GORODS_LIB_PATH/wrapper.c ]; then
	echo "Couldn't find GoRods project directory"
	exit
fi

# Compile gorods.o, build libgorods.a with iRODS C API
(cd $GORODS_LIB_PATH; rm -f build/libgorods.a; rm -f build/gorods.o; gcc -ggdb -o build/gorods.o -c wrapper.c -I/usr/include/irods -Iinclude; ar rcs build/libgorods.a build/gorods.o)

C_BUILD_SUCCESS=$?

# Compile and install GoRODS, and your app
go install github.com/jjacquay712/GoRODS && go install $YOUR_APP_PACKAGE

GO_BUILD_SUCCESS=$?

if [ $GO_BUILD_SUCCESS == 0 ] && [ $C_BUILD_SUCCESS == 0 ]; then
	# Run binary
	if [ -f $DIR/../../../../bin/$YOUR_APP_PACKAGE ]; then
		echo "- Running from GoRods package dir -"
		echo
		$DIR/../../../../bin/$YOUR_APP_PACKAGE
	elif [ -f $DIR/bin/$YOUR_APP_PACKAGE ]; then
		echo "- Running from \$GOPATH dir -"
		echo
		$DIR/bin/$YOUR_APP_PACKAGE
	fi
else
	echo "Build failed"
fi
