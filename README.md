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
	irods := gorods.New(&gorods.ConnectionOptions {
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

	// Create file in home directory, overwrite if it exists
	test := gorods.CreateDataObj(&gorods.DataObjOptions {
		Name: "test.lol",
		Mode: 0750,
		Force: true,
	}, homeDir)

	// Stat the new test.lol file
	fmt.Printf("%v \n", test.Stat())

}

```

## Installation


```
$ git clone https://github.com/jjacquay712/GoRods.git
$ cd GoRods/
$ ./build.sh
```

## Contributing
