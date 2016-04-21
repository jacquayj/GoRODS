# GoRods

GoLang wrapper of iRods C API. Requires go version >= 1.5 for cgo compile flags variable support.

**Notice:** This package is incomplete and still under heavy development. API is subject to change without warning until a stable version is released. Please understand the risks before deploying this package in production systems.

### Installation

```
$ go get github.com/jjacquay712/GoRods
```

### Docs

https://godoc.org/github.com/jjacquay712/GoRods

## Example Usage

```go
package main

import (
	"fmt"
	"github.com/jjacquay712/GoRods"
)

func main() {

	// Connect to server, error provided by second parameter
	irods, _ := gorods.New(gorods.ConnectionOptions {

		// Or gorods.System to use the systems preconfigured environment
		Environment: gorods.UserDefined, 

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "admin",
		Password: "password",
	})

	// Open collection, preload sub collections into memory
	homeDir, _ := irods.Collection("/tempZone/home/admin", true)

	buildFile := homeDir.Cd("gorods").Get("build.sh")

	// Returns MetaCollection containing all metadata for buildFile DataObject
	metas := buildFile.Meta()

	// Returns pointer to Meta struct
	metas.Get("MyAttribute")

	// Or use a shortcut
	buildFile.Attribute("MyAttribute")
	
	// Returns true/false if checksum matches
	buildFile.Verify("GdU5GXvmky9/rw7rduk4JaEtEdlhhhhGufiez+2aI4o=")
	
	// Download remote file
	buildFile.DownloadTo("build.sh")

	// Read file from /tempZone/home/admin/gorods/build.sh
	contents := buildFile.Read()

	// Read file in 5 byte chunks
	var wholeFile []byte

	buildFile.ReadChunk(5, func(chunk []byte) {
		wholeFile = append(wholeFile, chunk...)
	})

	fmt.Printf(string(wholeFile))

	// Print []Byte as string
	fmt.Printf(string(contents))

	// Add local file to collection
	remoteFile := homeDir.Put("local_file.txt")

	// Copy file to gorods directory
	remoteFile.CopyTo("gorods")
	// or
	//
	// gorodsDir := homeDir.Cd("gorods")
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

	// Write 5 copies of "test" to file
	// Will start writing at last offset (seek) position (typically 0)
	for n := 0; n < 5; n++ {
		test.WriteBytes([]byte("test\n"))
	}

	// We must close the file explicitly after calling WriteBytes()
	test.Close()

	// Stat the test.txt file
	fmt.Printf("%v \n", test.Stat())

	// Read the contents back, print to screen
	fmt.Printf("%v \n", string(test.Read()))

	// Delete the file
	test.Delete()

}

```

## Contributing

Send me a pull request!

## Todo

### iRods API Coverage

* Implement DataObj: MoveToResource(), Replicate(), ReplSettings()
* Implement Meta: set operations, imeta qu
* Implement Collection: CreateCollection(), MoveTo(), CopyTo(), DownloadTo()
* Implement Connection: ilocate, interface to query meta (imeta qu), direct DataObj getter like Connection.Collection()?
* Implement structs: User, Group, Resource, Zone?
* Implement access control (rcModAccessControl) and tickets

### Code Polish

* Add more robust error handling: Add idiomatic return errors for public functions, custom error interface
* Finish godoc compatible comments to all functions so documentation can be generated
* Hide unneeded public functions
* Add unit tests

## Known Issues

* The static library included (lib/build/libgorods.a) in this repo won't work on 32-bit systems and OSX. Will need to integrate iRods build from scratch.
* Build script requires pre compiled .o files (not included in this repo) to generate new libgorods.a
* There might be memory leaks from C variables not being free()'d
* Bug list: https://godoc.org/github.com/jjacquay712/GoRods#pkg-note-bug

## License & Copyright

Copyright (c) 2016, University of Florida Research Foundation, Inc. All Rights Reserved.

GoRods is released under a 3-clause BSD License. For more information please refer to the LICENSE.md file
