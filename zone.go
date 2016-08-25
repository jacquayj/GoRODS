/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

// Package gorods is a Golang binding for the iRods C API (iRods client library).
// GoRods uses cgo to call iRods client functions.
package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	//"strconv"
	//"strings"
	//"time"
	//"unsafe"
)

type Zone struct {
	Name string

	Con *Connection
}

type Zones []*Zone

func (znes Zones) FindByName(name string) *Zone {
	for _, zne := range znes {
		if zne.Name == name {
			return zne
		}
	}
	return nil
}

func initZone(name string, con *Connection) (*Zone, error) {
	zne := new(Zone)

	zne.Con = con
	zne.Name = name

	return zne, nil
}

func (zne *Zone) String() string {
	return fmt.Sprintf("%v", zne.Name)
}

func (zne *Zone) GetName() string {
	return zne.Name
}
