/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

import "testing"

func TestEnvironmentDefinedConnection(t *testing.T) {
	irods, err := New(ConnectionOptions {
		Type: EnvironmentDefined,
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	irods.Disconnect()
}

func TestUserDefinedConnection(t *testing.T) {
	irods, err := New(ConnectionOptions {
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

func TestHomeCollection(t *testing.T) {
	irods, err := New(ConnectionOptions {
		Type: EnvironmentDefined,
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	homeDir, e := irods.Collection("/tempZone/home", false)
	if e != nil {
		t.Fatalf("%v\n", e)
	}

	homeDir.Close()
	irods.Disconnect()

}

func TestHomeCollectionRecursive(t *testing.T) {
	irods, err := New(ConnectionOptions {
		Type: EnvironmentDefined,
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	homeDir, e := irods.Collection("/tempZone/home", true)
	if e != nil {
		t.Fatalf("%v\n", e)
	}

	homeDir.Close()
	irods.Disconnect()

}

