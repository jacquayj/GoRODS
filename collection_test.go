/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

import (
	"fmt"
	"testing"
)

func TestCollection(t *testing.T) {

	client, conErr := New(testCreds)

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		t.Fatal(conErr)
	}

	if openErr := client.OpenCollection(CollectionOptions{
		Path: fmt.Sprintf("/%v/home/%v", testCreds.Zone, testCreds.Username),
	}, func(col *Collection, con *Connection) {

	}); openErr != nil {
		t.Fatal(openErr)
	}

}
