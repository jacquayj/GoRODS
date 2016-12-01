# GoRODS

Golang binding for iRODS C API. Compatible with golang version >= 1.5

### Installation

**Step #1:** Install build dependencies (http://irods.org/download/): irods-dev-4.1.10

```
CentOS/RHEL (64 bit)
$ sudo yum install ftp://ftp.renci.org/pub/irods/releases/4.1.10/centos7/irods-dev-4.1.10-centos7-x86_64.rpm

Ubuntu (64 bit)
$ curl ftp://ftp.renci.org/pub/irods/releases/4.1.10/ubuntu14/irods-dev-4.1.10-ubuntu14-x86_64.deb > irods-dev-4.1.10-ubuntu14-x86_64.deb
$ sudo dpkg -i irods-dev-4.1.10-ubuntu14-x86_64.deb
```
**Step #2:** Install runtime dependencies (http://irods.org/download/): irods-icommands-4.1.10

```
CentOS/RHEL (64 bit)
$ sudo yum install ftp://ftp.renci.org/pub/irods/releases/4.1.10/centos7/irods-icommands-4.1.10-centos7-x86_64.rpm

Ubuntu (64 bit)
$ curl ftp://ftp.renci.org/pub/irods/releases/4.1.10/ubuntu14/irods-icommands-4.1.10-ubuntu14-x86_64.deb > irods-icommands-4.1.10-ubuntu14-x86_64.deb
$ sudo dpkg -i irods-icommands-4.1.10-ubuntu14-x86_64.deb
```

**Note:** The irods-icat-4.1.10 or irods-resource-4.1.10 packages also contain the required /var/lib/irods/plugins/network/libtcp.so shared object that is loaded at runtime. Be sure that at least one of those three packages is installed when deploying a GoRODS binary.


**Step #3:** Install GoRODS (assuming GOPATH is setup)

```
$ go get github.com/jjacquay712/GoRODS 
```

### Docs

https://godoc.org/github.com/jjacquay712/GoRODS

### Usage Guide and Examples

https://github.com/jjacquay712/GoRODS/blob/master/HOWTO.md

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
	"fmt"
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

## License & Copyright

Copyright (c) 2016, University of Florida Research Foundation, Inc. All Rights Reserved.

GoRODS is released under a 3-clause BSD License. For more information please refer to the LICENSE.md file
