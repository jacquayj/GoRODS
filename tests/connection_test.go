/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

import "testing"

func TestEnvironmentDefinedConnection(t *testing.T) {
	irods, err := New(ConnectionOptions{
		Type: EnvironmentDefined,
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	irods.Disconnect()
}

func TestUserDefinedConnection(t *testing.T) {
	irods, err := New(ConnectionOptions{
		Type: UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "admin",
		Password: "password",
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	irods.Disconnect()
}
