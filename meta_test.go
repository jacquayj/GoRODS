/*** Copyright (c) 2016, The BioTeam, Inc.                     ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

//import "strings"

//import "fmt"

// func TestMetaRead(t *testing.T) {
// 	client, conErr := New(testCreds)

// 	// Ensure the client initialized successfully and connected to the iCAT server
// 	if conErr != nil {
// 		t.Fatal(conErr)
// 	}

// 	// Open a data object reference for /tempZone/home/rods/hello.txt
// 	if openErr := client.OpenDataObject(fmt.Sprintf("/%v/home/%v/hello.txt", testCreds.Zone, testCreds.Username), func(myFile *DataObj, con *Connection) {

// 		// read the contents
// 		if m, metaErr := myFile.Meta(); metaErr == nil {

// 			if ma, maEr := m.First("test"); maEr == nil {
// 				if ma.Value != "test" {
// 					t.Errorf("Expected string 'test', got '%s'", ma.Value)
// 				}
// 			} else {
// 				t.Fatal(metaErr)
// 			}

// 		} else {
// 			t.Fatal(metaErr)
// 		}

// 	}); openErr != nil {
// 		t.Fatal(openErr)
// 	}

// }

// func TestMetaWrite(t *testing.T) {
// 	client, conErr := New(testCreds)

// 	// Ensure the client initialized successfully and connected to the iCAT server
// 	if conErr != nil {
// 		t.Fatal(conErr)
// 	}

// 	// Open a data object reference for /tempZone/home/rods/hello.txt
// 	if openErr := client.OpenDataObject(fmt.Sprintf("/%v/home/%v/hello.txt", testCreds.Zone, testCreds.Username), func(myFile *DataObj, con *Connection) {

// 		// read the contents
// 		if m, metaErr := myFile.Meta(); metaErr == nil {

// 			if ma, maEr := m.First("test"); maEr == nil {
// 				if ma.Value != "test" {
// 					t.Errorf("Expected string 'test', got '%s'", ma.Value)
// 				}

// 				if _, maEr := ma.SetValue("test2"); maEr == nil {
// 					if ma.Value != "test2" {
// 						t.Errorf("Expected string 'test2', got '%s'", ma.Value)
// 					}
// 				} else {
// 					t.Fatal(metaErr)
// 				}

// 				if _, maEr := ma.SetValue("test"); maEr == nil {
// 					if ma.Value != "test" {
// 						t.Errorf("Expected string 'test', got '%s'", ma.Value)
// 					}
// 				} else {
// 					t.Fatal(metaErr)
// 				}

// 				if _, delErr := ma.Delete(); delErr == nil {

// 					if _, fErr := m.First("test"); fErr == nil {
// 						t.Fatal(fErr)
// 					}

// 					if _, addErr := m.Add(Meta{
// 						Attribute: "test",
// 						Value:     "test",
// 					}); addErr != nil {
// 						t.Fatal(addErr)
// 					}

// 				} else {
// 					t.Fatal(delErr)
// 				}

// 			} else {
// 				t.Fatal(metaErr)
// 			}

// 		} else {
// 			t.Fatal(metaErr)
// 		}

// 	}); openErr != nil {
// 		t.Fatal(openErr)
// 	}

// }
