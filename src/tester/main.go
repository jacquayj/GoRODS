package main

import (
   "fmt"
   "gorods"
)


func main() {


    irods := gorods.New(&gorods.Options {
    	Host: "localhost",
		Port: 1247,
		Zone: "tempZone",

		Username: "admin",
		Password: "password",
    })

    fmt.Printf("%v", irods)

    homeDir := irods.Collection("/tempZone/home/admin", true)

    for _, d := range homeDir.Collections() {
    	fmt.Printf("%v \n", d)
    }

   	fmt.Printf("%v \n", homeDir.Collections().Find("gorods").DataObjs().Find("build.sh")) 

   	fmt.Printf(string(homeDir.Collections().Find("gorods").DataObjs().Find("build.sh").Read()))

    // dataObj := irods.DataObj("/testZone/home/admin/irods-icat-4.1.7-centos7-x86_64.rpm")

    // collection.DataObjs()     -> type: DataObjs
    // collection.Collections()  -> type: Collections
    // collection.All()          -> type: []interface{}
    // collection.Both()         -> (type: DataObjs, type: Collections)

    // collections.Find(relPath) -> type: Collection

    // dataObjs.Find(relPath)    -> type: DataObj


}

func PrintCollectionTree(c *gorods.Collection) {
	for _, obj := range c.Collections() {
		
		fmt.Printf("%v \n", obj)
		
		PrintCollectionTree(obj)
	}
}