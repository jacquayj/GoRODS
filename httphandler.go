/*** Copyright (c) 2016, University of Florida Research Foundation, Inc. ***
 *** For more information please refer to the LICENSE.md file            ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"fmt"
	"html/template"
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

var check func(error) = func(err error) {
	if err != nil {
		log.Print(err)
	}
}

var funcMap template.FuncMap = template.FuncMap{
	"prettySize": func(size int64) string {
		if size < 1024 {
			return fmt.Sprintf("%v bytes", size)
		} else if size < 1048576 { // 1 MiB
			return fmt.Sprintf("%.1f KiB", float64(size)/1024.0)
		} else if size < 1073741824 { // 1 GiB
			return fmt.Sprintf("%.1f MiB", float64(size)/1048576.0)
		} else if size < 1099511627776 { // 1 TiB
			return fmt.Sprintf("%.1f GiB", float64(size)/1073741824.0)
		} else {
			return fmt.Sprintf("%.1f TiB", float64(size)/1099511627776.0)
		}
	},
}

const tpl = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Collection: {{.Path}}</title>
	
	<!-- Latest compiled and minified CSS -->
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">

	<!-- Optional theme -->
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">

	<script src="https://code.jquery.com/jquery-3.1.1.min.js" integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8=" crossorigin="anonymous"></script>

	<!-- Latest compiled and minified JavaScript -->
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>

	<style type="text/css">
	.table td.fit, 
	.table th.fit {
	    white-space: nowrap;
	    width: 1%;
	}
	</style>

</head>
<body>
	<nav class="navbar navbar-default navbar-fixed-top">
		<div class="container">
			<div class="navbar-header">
				<a class="navbar-brand" href="#">GoRODS HTTP FileServer</a>
			</div>
			<div id="navbar" class="navbar-collapse collapse">
				<ul class="nav navbar-nav navbar-right">
					{{range .PathFragments}}
						<li><a href="#">{{ . }}</a></li>
					{{end}}
					<li class="active"><a href=".">{{.Name}}</a></li>
				</ul>
			</div><!--/.nav-collapse -->
		</div>
	</nav>

	<div class="container">
		<br /><br /><br />
		<h3>{{.Path}}</h3>

		<table class="table table-hover">
			<thead>
				<tr>
					<th>Name</th>
					<th>Size</th>
					<th>Type</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<th><a href="..">..</a></th>
					<td></td>
					<td>Collection</td>
				</tr>
				{{range .Collections}}
					<tr>
						<th><a href="{{.Name}}/">{{.Name}}</a></th>
						<td>{{prettySize .Size}}</td>
						<td>Collection</td>
					</tr>
				{{end}}
				{{range .DataObjs}}
					<tr>
						<th><a href="{{.Name}}">{{.Name}}</a></th>
						<td>{{prettySize .Size}}</td>
						<td>Data Object</td>
					</tr>
				{{end}}
			</tbody>
		</table>
	</div>

</body>
</html>
`

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

					t, err := template.New("collectionList").Funcs(funcMap).Parse(tpl)
					check(err)

					err = t.Execute(response, col)
					check(err)

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
