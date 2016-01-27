# GoRods
GoLang wrapper of iRods C API

## Example Usage

```go
package main

import (
	"fmt"
	"gorods"
)

func main() {

	// Connect to server
	irods := gorods.New(gorods.ConnectionOptions {
		Environment: gorods.UserDefined,
		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "admin",
		Password: "password",
	})

	// Open collection
	homeDir := irods.Collection("/tempZone/home/admin", true)

	// Read file from /tempZone/home/admin/gorods/build.sh
	file := homeDir.Collections().Find("gorods").DataObjs().Find("build.sh").Read()

	// Print []Byte as string
	fmt.Printf(string(file))

	// Add local file to collection
	remoteFile := homeDir.Put("local_file.txt")

	// Copy file to gorods directory
	remoteFile.CopyTo("/tempZone/home/admin/gorods")
	// or
	//
	// gorodsDir := homeDir.Collections().Find("gorods")
	// remoteFile.CopyTo(gorodsDir)

	// Move file/rename
	remoteFile.MoveTo("/tempZone/home/admin/gorods/local_file2.txt")

	// Create file in home directory, overwrite if it exists
	test := gorods.CreateDataObj(gorods.DataObjOptions {
		Name: "test.lol",
		Mode: 0750,
		Force: true,
	}, homeDir)

	// Write string to test file
	test.Write([]byte("This is a test!"))

	// Stat the test.lol file
	fmt.Printf("%v \n", test.Stat())

	// Read the contents back, print to screen
	fmt.Printf("%v \n", string(test.Read()))

}

```

## Installation


```
$ git clone https://github.com/jjacquay712/GoRods.git
$ cd GoRods/
$ ./build.sh
```

## Contributing

Send me a pull request!

## Known Issues

* Build script requires pre compiled .o files, the ones included in this repo won't work on 32-bit systems. Will need to integrate iRods build from scratch.
* There are probably memory leaks from C variables not being free()'d
* If destination target in MoveTo() exists, the source DataObj will be left intact with no warnings
* MoveTo() and CopyTo() require absolute paths if Collection struct isn't passed, need to add support for paths relative to source Collection
