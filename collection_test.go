/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

// func TestCollection(t *testing.T) {

// 	t.Log("TESTT")

// 	client, conErr := New(testCreds)

// 	// Ensure the client initialized successfully and connected to the iCAT server
// 	if conErr != nil {
// 		t.Fatal(conErr)
// 	}

// 	// Open a data object reference for /tempZone/home/rods/hello.txt
// 	if openErr := client.OpenCollection(CollectionOptions{
// 		Path: fmt.Sprintf("/%v/home/%v", testCreds.Zone, testCreds.Username),
// 	}, func(col *Collection, con *Connection) {

// 		col.Collections()

// 	}); openErr != nil {
// 		t.Fatal(openErr)
// 	}

// 	// Open a data object reference for /tempZone/home/rods/hello.txt test
// 	if openErr := client.OpenCollection(CollectionOptions{
// 		Path:      fmt.Sprintf("/%v/home/%v", testCreds.Zone, testCreds.Username),
// 		Recursive: true,
// 	}, func(col *Collection, con *Connection) {

// 	}); openErr != nil {
// 		t.Fatal(openErr)
// 	}

// }
