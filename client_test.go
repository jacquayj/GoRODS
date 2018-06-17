/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

import (
	"testing"
)

func TestClientConnection(t *testing.T) {

	cli, conErr := New(testCreds)

	// Ensure the client initialized successfully and connected to the iCAT server
	if conErr != nil {
		t.Fatal(conErr)
	}

	oconErr := cli.OpenConnection(func(con *Connection) {

	})

	if oconErr != nil {
		t.Fatal(oconErr)
	}
}
