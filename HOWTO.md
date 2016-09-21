# GoRODS How Do I?

This document provides examples on a few common iRODS tasks that could be implemented with GoRODS.

#### 1. How do I read data from a file stored in iRODS?

This example opens a connection to the iRODS server, using "/tempZone/home/rods" as the starting collection. It then searches the collection for a data object (file) named "hello.txt", and prints the contents.

```go
package main

import (
	"fmt"
	"log"
	"github.com/jjacquay712/GoRods"
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

	if conErr == nil {
		client.OpenConnection(gorods.CollectionOptions{
			Path: "/tempZone/home/rods",
		}, func(col *gorods.Collection, con *gorods.Connection) {

			myFile := col.FindObj("hello.txt")

			if myFile != nil {
				if contents, readErr := myFile.Read(); readErr == nil {
					fmt.Printf("hello.txt file contents: '%v' \n", string(contents))
				} else {
					log.Fatal(readErr)
				}
			} else {
				log.Fatal(fmt.Errorf("Unable to locate file in collection %v", col.Path()))
			}

		})
	} else {
		log.Fatal(conErr)
	}

}

```


#### 2. Can I selectively read sections of a file stored in iRODS (seek certain byte range)? If so, how?
#### 3. How do I write a file into iRODS?
#### 4. How can I get a list of files in a directory in iRODS?
#### 5. How can I apply metadata to a file in iRODS?#### 6. How can I retrieve metadata from a file in iRODS?
#### 7. How can I search for a file by metadata and other attributes?
#### 8. How do I set access controls?