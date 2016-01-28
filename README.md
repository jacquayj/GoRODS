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

		// Or gorods.System to use the systems preconfigured environment
		Environment: gorods.UserDefined, 

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "admin",
		Password: "password",
	})

	// Open collection, preload sub collections into memory
	homeDir := irods.Collection("/tempZone/home/admin", true)

	// Read file from /tempZone/home/admin/gorods/build.sh
	file := homeDir.Cd("gorods").Get("build.sh").Read()

	// Print []Byte as string
	fmt.Printf(string(file))

	// Add local file to collection
	remoteFile := homeDir.Put("local_file.txt")

	// Copy file to gorods directory
	remoteFile.CopyTo("gorods")
	// or
	//
	// gorodsDir := homeDir.Collections().Find("gorods")
	// remoteFile.CopyTo(gorodsDir)

	// Move file
	remoteFile.MoveTo("gorods/local_file2.txt")

	// Rename file
	remoteFile.Rename("local_file3.txt")

	// Create file in home directory, overwrite if it exists
	test := gorods.CreateDataObj(gorods.DataObjOptions {
		Name: "test.txt",
		Mode: 0750,
		Force: true,
	}, homeDir)

	// Write string to test file
	test.Write([]byte("This is a test!"))

	// Stat the test.txt file
	fmt.Printf("%v \n", test.Stat())

	// Read the contents back, print to screen
	fmt.Printf("%v \n", string(test.Read()))

	// Delete the file
	test.Delete()

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
