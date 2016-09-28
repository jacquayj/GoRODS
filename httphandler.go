/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func FileServer(opts FSOptions) http.Handler {

	handler := new(HttpHandler)

	handler.client = opts.Client
	handler.path = strings.TrimRight(opts.Path, "/")
	handler.opts = opts

	return handler

}

type FSOptions struct {
	Client   *Client
	Path     string
	Download bool
}

type HttpHandler struct {
	client *Client
	path   string
	opts   FSOptions
}

func (handler *HttpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {

	urlPath := strings.TrimRight(request.URL.Path, "/")
	openPath := strings.TrimRight(handler.path+"/"+urlPath, "/")

	if er := handler.client.OpenConnection(func(con *Connection) {
		if objType, err := con.PathType(openPath); err == nil {

			if objType == DataObjType {
				if obj, er := con.DataObject(openPath); er == nil {

					if handler.opts.Download {
						response.Header().Set("Content-Disposition", "attachment; filename="+obj.Name())
						response.Header().Set("Content-type", "application/octet-stream")
					} else {
						var mimeType string
						ext := filepath.Ext(openPath)

						if ext != "" {
							mimeType = mime.TypeByExtension(ext)

							if mimeType == "" {
								log.Printf("Can't find mime type for %s extension", ext)
								mimeType = "application/octet-stream"
							}
						} else {
							mimeType = "application/octet-stream"
						}

						response.Header().Set("Content-type", mimeType)
					}

					response.Header().Set("Content-Length", strconv.FormatInt(obj.Size(), 10))

					if readEr := obj.ReadChunk(1024000, func(chunk []byte) {
						response.Write(chunk)
					}); readEr != nil {
						log.Print(readEr)
					}

				} else {
					log.Print(er)
				}
			} else if objType == CollectionType {

				uP := request.URL.Path

				if uP != "/" && uP != "" && uP[len(uP)-1:] != "/" {
					http.Redirect(response, request, (uP + "/"), http.StatusFound)
					return
				}

				if col, er := con.Collection(CollectionOptions{
					Path:      openPath,
					Recursive: false,
					GetRepls:  false,
				}); er == nil {

					response.Header().Set("Content-Type", "text/html")

					response.Write([]byte("<h3>Collection: " + col.Path() + "</h3>"))
					response.Write([]byte("<br /><strong>Data Objects:</strong><br />"))

					col.EachDataObj(func(obj *DataObj) {
						response.Write([]byte("<a href=\"" + obj.Name() + "\">" + obj.Name() + "</a><br />"))
					})

					response.Write([]byte("<br /><strong>Sub Collections:</strong><br />"))

					col.EachCollection(func(subcol *Collection) {
						response.Write([]byte("<a href=\"" + subcol.Name() + "/\">" + subcol.Name() + "</a><br />"))
					})

				} else {
					log.Print(er)
				}
			}

		} else {

			response.Header().Set("Content-Type", "text/html")
			response.WriteHeader(http.StatusNotFound)

			response.Write([]byte("<h3>404 Not Found: " + openPath + "</h3>"))

			log.Print(err)
		}
	}); er != nil {
		log.Print(er)
		return
	}

}
