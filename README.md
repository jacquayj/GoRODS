# GoRODS

Golang binding for iRODS C API. Compatible with golang version >= 1.5

[![GoDoc](https://godoc.org/github.com/jjacquay712/GoRODS?status.svg)](https://godoc.org/github.com/jjacquay712/GoRODS)

### Installation

Install GoRODS (assuming GOPATH is setup)

* [iRODS 4.1.10 Instructions](https://github.com/jjacquay712/GoRODS/blob/master/4-1-10_BUILD_README.md)
* [iRODS 4.2.0 Instructions](https://github.com/jjacquay712/GoRODS/blob/master/4-2-0_BUILD_README.md) (now master branch default)

```
$ go get github.com/jjacquay712/GoRODS 
```


### Docs

[iRODS client binding](https://godoc.org/github.com/jjacquay712/GoRODS)

[iRODS microservice binding](https://godoc.org/github.com/jjacquay712/GoRODS/msi)

### Usage Guide and Examples

[iRODS client binding](https://github.com/jjacquay712/GoRODS/blob/master/HOWTO.md)

[iRODS microservice binding](https://github.com/jjacquay712/irods-ugm-2017)

## Basic Usage

```go
package main

import (
	"fmt"
	"log"
	"github.com/jjacquay712/GoRODS"
)

func main() {
	
	client, conErr := gorods.New(gorods.ConnectionOptions{
		Type: gorods.UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "rods",
		Password: "password",
	})

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		log.Fatal(conErr)
	}


	// Open a collection reference for /tempZone/home/rods
	if openErr := client.OpenCollection(gorods.CollectionOptions{
		Path: "/tempZone/home/rods",
	}, func(col *gorods.Collection, con *gorods.Connection) {

		// Output collection's string representation
		fmt.Printf("String(): %v \n", col)

		// Loop over the data objects in the collection, print the file name
		col.EachDataObj(func(obj *gorods.DataObj) {
			fmt.Printf("%v \n", obj.Name())
		})

		// Loop over the subcollections in the collection, print the name
		col.EachCollection(func(subcol *gorods.Collection) {
			fmt.Printf("%v \n", subcol.Name())
		})

	}); openErr != nil {
		log.Fatal(openErr)
	}

}

```

**Output:**

![CLI GoRODS Output](https://raw.githubusercontent.com/jjacquay712/GoRODS/master/screenshots/cli.png)


## iRODS HTTP Mount

```go

package main

import (
	"github.com/jjacquay712/GoRODS"
	"log"
	"net/http"
)

func main() {

	client, conErr := gorods.New(gorods.ConnectionOptions{
		Type: gorods.UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "rods",
		Password: "password",
	})

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		log.Fatal(conErr)
	}

	mountPath := "/irods/"

	// Setup the GoRODS FileServer
	fs := gorods.FileServer(gorods.FSOptions{
		Path:   "/tempZone/home/rods",
		Client: client,
		Download: true,
		StripPrefix: mountPath,
	})

	// Create the URL router
	mux := http.NewServeMux()

	// Serve the iRODS collection at /irods/
	mux.Handle(mountPath, http.StripPrefix(mountPath, fs))

	// Start HTTP server on port 8080
	log.Fatal(http.ListenAndServe(":8080", mux))

}

```

**Output:**

![HTTP GoRODS Output](https://raw.githubusercontent.com/jjacquay712/GoRODS/master/screenshots/http.png)
![HTTP GoRODS Output](https://raw.githubusercontent.com/jjacquay712/GoRODS/master/screenshots/http2.png)

## Contributing

Send me a pull request!

## Todo

* See Github issues for todo list https://github.com/jjacquay712/GoRODS/issues


### Code Polish

* Complete unit tests

## Known Issues

* Bug list: https://godoc.org/github.com/jjacquay712/GoRODS#pkg-note-bug
* Missing functionality: https://github.com/jjacquay712/GoRODS/wiki

## License & Copyright

Copyright (c) 2016, University of Florida Research Foundation, Inc. and The BioTeam, Inc. All Rights Reserved.

GoRODS is released under a 3-clause BSD License. For more information please refer to the LICENSE.md file
