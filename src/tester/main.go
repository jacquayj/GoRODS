package main

import (
   "fmt"
   "gorods"
)


func main() {

    irods := gorods.NewConnection()

    rootColl := irods.GetCollection("/testZone/home/admin")

	for _, obj := range rootColl.GetDataObjs() {
		fmt.Printf("%v \n", obj)
	}

}