/*** Copyright (c) 2016, The BioTeam, Inc.                    ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

import "testing"
import "strings"

//import "fmt"

func TestDataObjCreateDeleteWrite(t *testing.T) {
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

		wrErr := do.Write([]byte("test123content"))
		if wrErr != nil {
			t.Fatal(wrErr)
		}

		_, statErr := do.Stat()
		if statErr != nil {
			t.Fatal(statErr)
		}

		if chErr := do.Chmod("developers", Write, false); chErr != nil {
			t.Fatal(chErr)
		}

		acl, aclErr := do.ACL()
		if aclErr != nil {
			t.Fatal(aclErr)
		}

		acl[0].String()

		if acl[1].User().Name() != "rods" {
			t.Errorf("Expected string 'rods', got '%s'", acl[1].User().Name())
		}

		if acl[0].Group().Name() != "developers" {
			t.Errorf("Expected string 'developers', got '%s'", acl[0].Group().Name())
		}

		cont, readErr := do.Read()
		if readErr != nil {
			t.Fatal(readErr)
		}

		if string(cont) != "test123content" {
			t.Errorf("Expected string 'test123content', got '%s'", string(cont))
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

func TestDataObjReadBytes(t *testing.T) {
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
		if contents, readErr := myFile.ReadBytes(7, 6); readErr == nil {

			c := string(contents)

			if c != "World!" {
				t.Errorf("Expected string 'World!', got '%s'", c)
			}
		} else {
			t.Fatal(readErr)
		}

	}); openErr != nil {
		t.Fatal(openErr)
	}

}
