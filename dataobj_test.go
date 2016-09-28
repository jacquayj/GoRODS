/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

import "testing"
import "strings"

func TestDataObjCreateDelete(t *testing.T) {
	client, conErr := New(ConnectionOptions{
		Type: UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "rods",
		Password: "password",
	})

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		t.Fatal(conErr)
	}

	// Open a data object reference for /tempZone/home/rods/hello.txt
	if openErr := client.OpenCollection(CollectionOptions{
		Path: "/tempZone/home/rods",
	}, func(col *Collection, con *Connection) {

		do, createErr := col.CreateDataObj(DataObjOptions{
			Name: "test123.txt",
		})

		if createErr != nil {
			t.Fatal(createErr)
		}

		delErr := do.Delete(false)

		if delErr != nil {
			t.Fatal(delErr)
		}

	}); openErr != nil {
		t.Fatal(openErr)
	}

}

func TestDataObjRead(t *testing.T) {
	client, conErr := New(ConnectionOptions{
		Type: UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "rods",
		Password: "password",
	})

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		t.Fatal(conErr)
	}

	// Open a data object reference for /tempZone/home/rods/hello.txt
	if openErr := client.OpenDataObject("/tempZone/home/rods/hello.txt", func(myFile *DataObj, con *Connection) {

		// read the contents
		if contents, readErr := myFile.Read(); readErr == nil {

			c := string(contents)

			if strings.Trim(c, "\n") != "Hello, World!" {
				t.Errorf("Expected string 'Hello, World!', got '%s'", c)
			}
		} else {
			t.Fatal(readErr)
		}

	}); openErr != nil {
		t.Fatal(openErr)
	}

}

func TestDataObjMeta(t *testing.T) {
	client, conErr := New(ConnectionOptions{
		Type: UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "rods",
		Password: "password",
	})

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		t.Fatal(conErr)
	}

	// Open a data object reference for /tempZone/home/rods/hello.txt
	if openErr := client.OpenDataObject("/tempZone/home/rods/hello.txt", func(myFile *DataObj, con *Connection) {

		// read the contents
		if _, metaErr := myFile.Meta(); metaErr == nil {

		} else {
			t.Fatal(metaErr)
		}

	}); openErr != nil {
		t.Fatal(openErr)
	}

}
