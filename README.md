# GoRods

Golang binding for iRods C API. Requires go version >= 1.5 for cgo support.

**Notice:** This package is incomplete and still under heavy development. API is subject to change without warning until a stable version is released.

### Installation

Install dependencies (http://irods.org/download/): irods-dev-4.1.8

```
CentOS/RHEL (64 bit)
$ sudo yum install ftp://ftp.renci.org/pub/irods/releases/4.1.8/centos7/irods-dev-4.1.8-centos7-x86_64.rpm

Ubuntu (64 bit)
$ curl ftp://ftp.renci.org/pub/irods/releases/4.1.8/ubuntu14/irods-dev-4.1.8-ubuntu14-x86_64.deb > irods-dev-4.1.8-ubuntu14-x86_64.deb
$ sudo dpkg -i irods-dev-4.1.8-ubuntu14-x86_64.deb
```

Install GoRods

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

		// Or gorods.EnvironmentDefined to use the systems preconfigured environment
		Type: gorods.UserDefined, 

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "admin",
		Password: "password",
	})

	// Open collection, preload sub collections into memory
	homeDir, _ := irods.Collection("/tempZone/home/admin", true)

	buildFile := homeDir.Cd("gorods").Get("build.sh")

	// Search collections & objects by metadata
	metaResult, _ := irods.QueryMeta("myattr = myval")
	metaResult.Each(func (irodsObj gorods.IRodsObj) {
		fmt.Printf("Found: %v\n", irodsObj)
	})

	// Returns MetaCollection containing all metadata for buildFile DataObject
	metas, _ := buildFile.Meta()

	// Returns pointer to Meta struct
	metas.Get("MyAttribute")

	// Add a meta AVU triple
	metas.Add(gorods.Meta {
		Attribute: "add-test",
		Value: "test",
		Units: "string",
	})

	// Or use a shortcut
	myAttr, _ := buildFile.Attribute("MyAttribute")

	myAttr.SetValue("New Value")

	myAttr.SetUnits("myUnit")

	myAttr.Set("New Value", "myUnit")

	myAttr.Rename("AnotherAttribute")

	// Delete the metadata AVU triple
	myAttr.Delete()
	
	// Returns true/false if checksum matches
	buildFile.Verify("GdU5GXvmky9/rw7rduk4JaEtEdlhhhhGufiez+2aI4o=")
	
	// Download remote file
	buildFile.DownloadTo("build.sh")

	// Read file from /tempZone/home/admin/gorods/build.sh
	contents, _ := buildFile.Read()

	// Read file in 5 byte chunks
	var wholeFile []byte

	buildFile.ReadChunk(5, func(chunk []byte) {
		wholeFile = append(wholeFile, chunk...)
	})

	fmt.Printf(string(wholeFile))

	// Print []Byte as string
	fmt.Printf(string(contents))

	// Add local file to collection
	remoteFile, _ := homeDir.Put("local_file.txt")

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
	test, _ := gorods.CreateDataObj(gorods.DataObjOptions {
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

	// Close data object/collection handles if open
	homeDir.Close()
	buildFile.Close()
	remoteFile.Close()

	// Disconnect from the iCAT server, important!
	irods.Disconnect()

}

```

## Contributing

Send me a pull request!

## Todo

* See Github issues for todo list https://github.com/jjacquay712/GoRods/issues


### Code Polish

* Complete unit tests

## Known Issues

* The static library included (lib/build/libgorods.a) in this repo won't work on 32-bit systems and OSX. Install irods-dev system package, and run the build.sh script to compile binaries for your system.
* Bug list: https://godoc.org/github.com/jjacquay712/GoRods#pkg-note-bug

## License & Copyright

Copyright (c) 2016, University of Florida Research Foundation, Inc. All Rights Reserved.

GoRods is released under a 3-clause BSD License. For more information please refer to the LICENSE.md file
