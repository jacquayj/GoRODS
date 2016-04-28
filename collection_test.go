/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

import "testing"


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

func TestHomeCollectionGetMeta(t *testing.T) {
	irods, err := New(ConnectionOptions {
		Type: EnvironmentDefined,
	})

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	homeDir, e := irods.Collection("/tempZone/home/admin/gorods", false)
	if e != nil {
		t.Fatalf("%v\n", e)
	}

	homeDir.Meta()

	homeDir.Close()
	irods.Disconnect()
}