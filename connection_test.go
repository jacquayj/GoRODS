/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

import (
	"flag"
	"fmt"
	"testing"
)

var testCreds = ConnectionOptions{
	Type: UserDefined,
}

var shouldTestPAM = false

func init() {
	flag.StringVar(&testCreds.Host, "irods.host", "localhost", "Hostname of iRODS server, e.g. localhost")
	flag.IntVar(&testCreds.Port, "irods.port", 1247, "Port of iRODS server, e.g. 1247")
	flag.StringVar(&testCreds.Zone, "irods.zone", "tempZone", "Zone to use in connection to iRODS server, e.g. tempZone")
	flag.StringVar(&testCreds.Username, "irods.username", "rods", "Username to use in connection to iRODS server, e.g. rods")
	flag.StringVar(&testCreds.Password, "irods.password", "testpassword", "Password to use in connection to iRODS server, e.g. testpassword")

	flag.BoolVar(&shouldTestPAM, "irods.testpam", false, "Should try to connect with PAM")

	flag.Parse()

	if shouldTestPAM {
		testCreds.AuthType = PAMAuth
	}

	fmt.Printf("Setup testing with params: %v\n", testCreds.String())
}

func TestUserDefinedConnection(t *testing.T) {

	if irods, err := NewConnection(&testCreds); err != nil {
		t.Fatal(err)
	} else {
		if er := irods.Disconnect(); er != nil {
			t.Fatal(er)
		}
	}

}
