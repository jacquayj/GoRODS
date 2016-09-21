# GoRODS How Do I?

This document provides examples on a few common iRODS tasks that could be implemented with GoRODS.

#### 1. How do I read data from a file stored in iRODS?

First we need to create our bare bones .go file, import the GoRODS package, and setup the client struct:

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

	// All example code in this document is written as if it was placed in this context / main func scope, with the client already setup.

}

```

**Remember that the following examples leave out the import statement, main() function declaration, and client struct setup, for the sake of brevity.**

This next example opens a connection to the iRODS server, using "/tempZone/home/rods" as the starting collection. It then searches the collection for a data object (file) named "hello.txt", and prints the contents.

**Example:**
```go

// Ensure the client initialized successfully and connected to the iCAT server
if conErr == nil {

	// Open a connection, with the starting collection of /tempZone/home/rods
	client.OpenConnection(gorods.CollectionOptions{
		Path: "/tempZone/home/rods",
	}, func(col *gorods.Collection, con *gorods.Connection) {

		// Search for the hello.txt data object within the collection
		myFile := col.FindObj("hello.txt")

		// Did we find it?
		if myFile != nil {

			// Yes, read the contents
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


```

**Output:**

```
hello.txt file contents: 'Hello, World!' 
```

#### 2. Can I selectively read sections of a file stored in iRODS (seek certain byte range)? If so, how?

This example is very similar to the one above, except it starts reading at an offset of 7 bytes, e.g. lseek(7) and reads the next 6 bytes. You'll also notice the "defer myFile.Close()" line, which is required since ReadBytes doesn't explicitly close the data object. This is to reduce the overhead of calling ReadBytes sequentially. See [ReadChunk](https://godoc.org/gopkg.in/jjacquay712/GoRods.v1#DataObj.ReadChunk) to read the entire file in chucks, without the need to call Close().

**Example:**

```go

// Ensure the client initialized successfully and connected to the iCAT server
if conErr == nil {

	// Open a connection, with the starting collection of /tempZone/home/rods
	client.OpenConnection(gorods.CollectionOptions{
		Path: "/tempZone/home/rods",
	}, func(col *gorods.Collection, con *gorods.Connection) {

		// Search for the hello.txt data object within the collection
		myFile := col.FindObj("hello.txt")

		// Did we find it?
		if myFile != nil {

			// The ReadBytes function doesn't explicitly close the data object, so we need to make sure it's closed when we're finished reading
			defer myFile.Close()

			// Yes, read 6 bytes starting with an offset of 7 bytes
			if contents, readErr := myFile.ReadBytes(7, 6); readErr == nil {
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
```

**Output:**

```
hello.txt file contents: 'World!' 
```


#### 3. How do I write a file into iRODS?
#### 4. How can I get a list of files in a directory in iRODS?
#### 5. How can I apply metadata to a file in iRODS?#### 6. How can I retrieve metadata from a file in iRODS?
#### 7. How can I search for a file by metadata and other attributes?
#### 8. How do I set access controls?