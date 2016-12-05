/*** Copyright (c) 2016, The Bio Team, Inc.                    ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

import "testing"

func TestClientConnection(t *testing.T) {
	cli, conErr := New(ConnectionOptions{
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

	oconErr := cli.OpenConnection(func(con *Connection) {

	})
	if oconErr != nil {
		t.Fatal(oconErr)
	}
}
