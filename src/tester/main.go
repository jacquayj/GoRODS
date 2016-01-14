package main

import (
   "fmt"
   "gorods"
)


func main() {

    irods := gorods.NewConnection()

    fmt.Printf("%v\n", irods.Connected)

    // rootColl := irods.Collection("/")

    // for _, file := range rootColl.GetDataObjs() {
    // 	fmt.Println(file.Name)
    // }

    //rootCol.GetColObjs()


}