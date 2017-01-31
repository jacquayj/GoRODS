/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

import "testing"

func TestUserDefinedConnection(t *testing.T) {
	if irods, err := NewConnection(&ConnectionOptions{
		Type: UserDefined,

		Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "rods",
		Password: "password",
	}); err != nil {
		t.Fatal(err)
	} else {
		if er := irods.Disconnect(); er != nil {
			t.Fatal(er)
		}
	}

}

func TestPAMConnection(t *testing.T) {
	if irods, err := NewConnection(&ConnectionOptions{
		Type:     UserDefined,
		AuthType: PAMAuth,
		Host:     "localhost",
		Port:     1247,
		Zone:     "tempZone",

		Username: "rods",
		Password: "pamtest",
	}); err != nil {
		t.Fatal(err)
	} else {
		if er := irods.Disconnect(); er != nil {
			t.Fatal(er)
		}
	}

}
