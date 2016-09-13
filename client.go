/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	// "io/ioutil"
	// "path/filepath"
	// "strconv"
	// "strings"
	// "time"
	// "unsafe"
)

type Client struct {
	Options    *ConnectionOptions
	ConnectErr error
}

func (cli *Client) OpenConnection(opts CollectionOptions, handler func(*Collection, *Connection)) error {
	if cli.ConnectErr == nil {
		if con, err := NewConnection(cli.Options); err == nil {
			col, colEr := con.Collection(opts)

			if colEr != nil {
				return newError(Fatal, fmt.Sprintf("Can't open new connection: %v", colEr))
			}

			handler(col, con)

			if er := col.Close(); er != nil {
				return er
			}
			if er := con.Disconnect(); er != nil {
				return er
			}

			return nil
		} else {
			return newError(Fatal, fmt.Sprintf("Can't open new connection: %v", err))
		}
	}

	return newError(Fatal, fmt.Sprintf("Can't open new connection: %v", cli.ConnectErr))
}

func New(opts ConnectionOptions) (*Client, error) {
	cli := new(Client)

	cli.Options = &opts

	if _, err := NewConnection(cli.Options); err != nil {
		cli.ConnectErr = err
		return nil, err
	}

	return cli, nil
}
