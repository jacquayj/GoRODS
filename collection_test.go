/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

import "testing"

func TestCollection(t *testing.T) {
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

		col.Collections()

	}); openErr != nil {
		t.Fatal(openErr)
	}

	// Open a data object reference for /tempZone/home/rods/hello.txt
	if openErr := client.OpenCollection(CollectionOptions{
		Path:      "/tempZone/home/rods",
		Recursive: true,
	}, func(col *Collection, con *Connection) {

	}); openErr != nil {
		t.Fatal(openErr)
	}

}
