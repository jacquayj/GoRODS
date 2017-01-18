/*** Copyright (c) 2016, The Bio Team, Inc.                    ***
 *** For more information please refer to the LICENSE.md file  ***/

package gorods

// #include "wrapper.h"
import "C"

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

func FileServer(opts FSOptions) http.Handler {
	h := new(HandlerFactory)
	h.opts = opts
	return h
}

type FSOptions struct {
	Client         *Client
	Connection     *Connection
	Path           string
	Download       bool
	StripPrefix    string
	CollectionView string
}

type HandlerFactory struct {
	opts FSOptions
}

func (hf *HandlerFactory) ServeHTTP(response http.ResponseWriter, request *http.Request) {

	handler := new(HttpHandler)

	handler.client = hf.opts.Client
	handler.connection = hf.opts.Connection
	handler.path = strings.TrimRight(hf.opts.Path, "/")
	handler.opts = hf.opts

	if handler.opts.CollectionView != "" {
		tpl = string(handler.opts.CollectionView)
	}

	handler.ServeHTTP(response, request)

}

type HttpHandler struct {
	client     *Client
	connection *Connection
	path       string
	opts       FSOptions

	response    http.ResponseWriter
	request     *http.Request
	handlerPath string
	openPath    string
	query       url.Values
}

var check func(error) = func(err error) {
	if err != nil {
		log.Print(err)
	}
}

var tpl = `
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
	a {
		cursor: pointer;
	}
	.modal-dialog {
		margin-top: 100px;
	}
	.meta-del {
		color: red;
		cursor: pointer;
		font-size: 18px;
	}
	.li-sep {
		line-height: 50px;
		color:#777;
	}
	.li-sep:first-child {
		display: none;
	}

	.prog-bar {
		width: 0%;
	    height: 3px;
	    background-color: #337ab7;
	    position: absolute;
	    top: 0px;
	    z-index: 9999;
	}
	</style>

	<script type="text/javascript">

	var me = {{ .Con.Options.Username }};
	var users = {{ usersJSON }};
	var groups = {{ groupsJSON }};

	function escapeHtml(text) {
	    'use strict';
	    return text.replace(/[\"&<>]/g, function (a) {
	        return { '"': '&quot;', '&': '&amp;', '<': '&lt;', '>': '&gt;' }[a];
	    });
	}

	$(function() {

		var chmod = function(objName, formData, inctx) {

			var ajaxPath = document.location.pathname + objName + "?createacl=1";

			$.post(ajaxPath, formData, function(response, status) {
				if ( response.Success == true ) {
					var ctx = inctx.parent();

					$(".acl-name", inctx).val("");
					$(".acl-access", inctx).val("");

					refreshMetas(objName, ctx);
				} else {
					alert("An error has occured: " + response.Message);
				}
			});
		};


		var createCollectionHandler = function() {

			var colInput = $('.collection-name', $(this).parent().parent());

			var colName = colInput.val();

			var ajaxPath = document.location.pathname + "?createcol=1";

			$.post(ajaxPath, {colname: colName}, function(response, status) {
				if ( response.Success == true ) {
					document.location.reload(true);
				} else {
					alert("An error has occured: " + response.Message);
				}
			});
			
		};

		$('[data-toggle="popover"]').popover({
			content: function(){
				var cont = $($(this).data('contentwrapper'));

		        return cont.html();
		    },
			html: true
		}).on('shown.bs.popover', function() {
			var pID = $(this).attr("aria-describedby");

			$('.create-collection', $("#" + pID)).on('click', createCollectionHandler);
		});
		

		var deleteMetaHandler = function(objname) {
			return function() {	
				var self = $(this);

				var tr = self.parent().parent();

				var tds = $("td", tr);

				var a = tds.eq(0).text();
				var v = tds.eq(1).text();
				var u = tds.eq(2).text();

				var ajaxPath = document.location.pathname + objname + "?deletemeta=1";

				var formData = {
					"attribute": a,
					"value": v,
					"units": u
				};

				$.post(ajaxPath, formData, function(response, status) {
					if ( response.Success == true ) {
						var ctx = tr.parent().parent().parent();

						refreshMetas(objname, ctx);
					} else {
						alert("An error has occured: " + response.Message);
					}
				});

			};
		};

		var refreshMetas = function(objname, ctx, done) {
			var ajaxPath = document.location.pathname + objname + "?meta=1";

			$.ajax({
				"url": ajaxPath,
				"complete": function(response, status) {
					var metaTbl = $(".meta-tbl tbody", ctx).html("");
					var aclTbl = $(".acl-tbl tbody", ctx).html("");


					if ( Object.keys(response.responseJSON).length != 0 ) {

						var metaData = response.responseJSON.metadata;
						var aclData = response.responseJSON.acl;
						var statData = response.responseJSON.stat[0];

						$('.chksum', ctx).text(statData.chksum);
						$('.createTime', ctx).text((new Date(parseInt(statData.createTime) * 1000)).toLocaleString());
						$('.modifyTime', ctx).text((new Date(parseInt(statData.modifyTime) * 1000)).toLocaleString());
						$('.dataId', ctx).text(statData.dataId);
						$('.dataMode', ctx).text(statData.dataMode);
						$('.objSize', ctx).text(statData.objSize);
						$('.ownerName', ctx).text(statData.ownerName);
						$('.ownerZone', ctx).text(statData.ownerZone);


						if ( metaData.length > 0 ) {
							for ( var n = 0; metaData.length > n; n++ ) {
								metaTbl.append('<tr><td>' + escapeHtml(metaData[n].attribute) + '</td><td>' + escapeHtml(metaData[n].value) + '</td><td>' + escapeHtml(metaData[n].units) + '</td><td style="text-align:right;"><span class="meta-del glyphicon glyphicon-remove-circle"></span></td></tr>');
							}
						} else {
							metaTbl.append('<tr><td colspan="4" style="text-align:center;">No Metadata Found</td></tr>');
						}
						$(".meta-del", metaTbl).click(deleteMetaHandler(objname));

						for ( var n = 0; aclData.length > n; n++ ) {
							var userGroupSelect;

							if ( users.length == 0 ) {
								userGroupSelect = $('<input class="acl-name form-control" disabled value="' + escapeHtml(aclData[n].name) + '">');
							} else {
								userGroupSelect = $('<select class="acl-name form-control" data-prev="' + escapeHtml(aclData[n].name) + '"></select>');
								for ( var i in users ) {
									if ( users[i] == aclData[n].name ) {
										userGroupSelect.append('<option selected value="' + escapeHtml(users[i]) + '">User: ' + escapeHtml(users[i]) + '</option>');
									} else {
										userGroupSelect.append('<option value="' + escapeHtml(users[i]) + '">User: ' + escapeHtml(users[i]) + '</option>');
									}
								}
								for ( var i in groups ) {
									if ( groups[i] == aclData[n].name ) {
										userGroupSelect.append('<option selected value="' + escapeHtml(groups[i]) + '">Group: ' + escapeHtml(groups[i]) + '</option>');
									} else {
										userGroupSelect.append('<option value="' + escapeHtml(groups[i]) + '">Group: ' + escapeHtml(groups[i]) + '</option>');
									}
								}
							}
							

							var accessLevels = ['read', 'write', 'own', 'null'];
							var accessLevelSelect = $('<select class="access-level form-control"></select>');
							for ( var i in accessLevels ) {
								if ( accessLevels[i] == aclData[n].accessLevel ) {
									accessLevelSelect.append('<option selected value="' + escapeHtml(accessLevels[i]) + '">' + escapeHtml(accessLevels[i]) + '</option>');
								} else {
									accessLevelSelect.append('<option value="' + escapeHtml(accessLevels[i]) + '">' + escapeHtml(accessLevels[i]) + '</option>');
								}
							}

							aclTbl.append('<tr><td>' + userGroupSelect[0].outerHTML + '</td><td>' + escapeHtml(aclData[n].type) + '</td><td>' + accessLevelSelect[0].outerHTML+ '</td></tr>');


							$(".access-level", aclTbl).change(function() {
								var self = $(this);

								var tr = self.parent().parent();

								var formData = {
									name: $(".acl-name", tr).val(),
									access: self.val()
								};

								var f = $("form", tr.parent().parent().parent().parent());

								chmod(objname, formData, f);

							});

							$("select.acl-name", aclTbl).change(function() {
								var self = $(this);

								var tr = self.parent().parent();

								var formData = {
									name: self.val(),
									access: $(".access-level", tr).val()
								};

								var f = $("form", tr.parent().parent().parent().parent());

								chmod(objname, formData, f);

								if ( self.attr("data-prev") != self.val() ) {
									formData = {
										name: self.attr("data-prev"),
										access: "null"
									};

									chmod(objname, formData, f);
								}
							});
						}
					} else {
						aclTbl.append('<tr><td colspan="3" style="text-align:center;">Error Fetching ACLs</td></tr>');
						metaTbl.append('<tr><td colspan="4" style="text-align:center;">Error Fetching Metadata</td></tr>');
					}

					if ( done ) done();

					
				}
			});
		};

		$('.upload-btn').click(function() {

			var form = $(document.createElement('form'));
			var input = $(document.createElement('input'));

	        input.attr("type", "file").attr("name", "data");

	        form.append(input);

	        input.trigger('click');

	        input.change(function() {
	        	var ajaxPath = document.location.pathname + "?upload=1";

	        	var formData = new FormData(form[0]);

	        	var progBar = $(".prog-bar");
	        	progBar.show();

	        	$.ajax({
	                url: ajaxPath,
	                type: 'POST',
	                xhr: function() {
	                    myXhr = $.ajaxSettings.xhr();
	                    if ( myXhr.upload ) {
	                        myXhr.upload.addEventListener('progress', function(ev) {
	                        	var percentComplete = (ev.loaded / ev.total) * 100;

	                        	progBar.animate({
	                        		width: percentComplete + "%"
	                        	}, 100);

	                        }, false);
	                    }
	                    return myXhr;
	                },
	                success: function(response) {
	                    document.location.reload(true);
	                },
	                error: function(response) {
	                   	alert("An error has occured: " + response.Message);
	                },
	                data: formData,
	                cache: false,
	                contentType: false,
	                processData: false
            	}, 'json');

	        });
	        
		});

		$('.show-meta-modal').click(function() {
			var self = $(this);

			var objName = self.attr("data-objname");
			var ctx = self.parent().parent();

			refreshMetas(objName, ctx, function() {
				$('.modal', ctx).modal('show');
			});
			
			
		});

		$('.delete-obj').click(function() {
			var self = $(this);

			if ( !confirm("Are you sure you want to delete this iRODS object? This action cannot be undone.") ) {
				return;
			}

			var objName = self.attr("data-objname");

			var ajaxPath = document.location.pathname + objName + "?delete=1";

			$.post(ajaxPath, {}, function(response, status) {
				if ( response.Success == true ) {
					document.location.reload(true);
				} else {
					alert("An error has occured: " + response.Message);
				}
			});
		});

		$('.acl-form').submit(function(e) {
			var self = $(this);

			e.preventDefault();

			var objName = self.attr("data-objname");

			var formData = {
				name: $(".acl-name", self).val(),
				access: $(".acl-access", self).val()
			};

			chmod(objName, formData, self);
		});

		$('.avu-form').submit(function(e) {
			
			e.preventDefault();

			var self = $(this);

			var objName = self.attr("data-objname");
			var ajaxPath = document.location.pathname + objName + "?meta=1";

			var formData = {
				"attribute": $(".avu-attribute", self).val(),
				"value": $(".avu-value", self).val(),
				"units": $(".avu-units", self).val()
			};

			$.post(ajaxPath, formData, function(response, status) {
				if ( response.Success == true ) {
					var ctx = self.parent();

					$(".avu-attribute", self).val("");
					$(".avu-value", self).val("");
					$(".avu-units", self).val("");

					refreshMetas(objName, ctx);
				} else {
					alert("An error has occured: " + response.Message);
				}
			});

		});
		
	});

	</script>

</head>
<body>
	<div class="prog-bar" style="display: none;"></div>
	<nav class="navbar navbar-default navbar-fixed-top">
		<div class="container">
			<div class="navbar-header">
				<a class="navbar-brand" href="#">GoRODS HTTP File Server</a>
			</div>
			<div id="navbar" class="navbar-collapse collapse">
				<ul class="nav navbar-nav navbar-right">
					{{range headerLinks}}
						<li class="li-sep">&gt;</li>
						<li><a href="{{ index . "url" }}">{{ index . "name" }}</a></li>
					{{end}}
				</ul>
			</div><!--/.nav-collapse -->
		</div>
	</nav>

	<div class="container">
		<br /><br /><br />

		<div style="float:right;margin-top: 8px;" class="btn-group" role="group">
			<button type="button" class="btn btn-default" data-container="body" data-contentwrapper="#create-collection-cont" data-toggle="popover" data-placement="bottom">Create Collection</button>
			<button type="button" class="btn btn-default upload-btn">Upload Data Object</button>
		</div>

		<h4 style="margin-top:15px;">{{.Path}}</h4>

		<table class="table table-hover">
			<thead>
				<tr>
					<th>Name</th>
					<th>Size</th>
					<th>Type</th>
					<th style="width:8%;"></th>
				</tr>
			</thead>
			<tbody>
				{{ $length := len headerLinks }}{{ if ne $length 0 }}
					<tr>
						<th><a href="..">..</a></th>
						<td></td>
						<td>Collection</td>
						<td></td>
					</tr>
				{{ end }}
				{{range .Collections}}
					<tr>
						<th><a href="{{.Name}}/">{{.Name}}</a></th>
						<td>{{prettySize .Size}}</td>
						<td>Collection</td>
						<td>
							<div style="text-align:right;">
								<span style="cursor:pointer;color:#337ab7;margin-right:10px;" data-objname="{{.Name}}/" class="glyphicon glyphicon-th-list show-meta-modal"></span>
								<span class="glyphicon glyphicon-remove delete-obj" style="color:red;cursor:pointer;" data-objname="{{.Name}}/"></span>
							</div>

							<!-- Modal -->
							<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel">
								<div class="modal-dialog" role="document">
									<div class="modal-content">
										<div class="modal-header">
											<button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
											<h4 class="modal-title" class="myModalLabel">Collection "{{.Name}}"</h4>
										</div>
										<div class="modal-body">

											<ul class="nav nav-tabs">
												<li class="active"><a data-toggle="tab" data-target=".stat-cont">Stat</a></li>
												<li><a data-toggle="tab" data-target=".metadata-cont">Metadata</a></li>
												<li><a data-toggle="tab" data-target=".acl-cont">ACL</a></li>
											</ul>

											<div class="tab-content">
												<div class="stat-cont tab-pane fade in active">
													<br />
													<table class="table table-hover">
													<tbody>
														<tr>
															<td>
																Checksum:
															</td>
															<td class="chksum">
																
															</td>
														</tr>
														<tr>
															<td>
																Created At:
															</td>
															<td class="createTime">
																
															</td>
														</tr>
														<tr>
															<td>
																Modified At:
															</td>
															<td class="modifyTime">
																
															</td>
														</tr>
														<tr>
															<td>
																Data ID:
															</td>
															<td class="dataId">
																
															</td>
														</tr>
														<tr>
															<td>
																Data Mode:
															</td>
															<td class="dataMode">
																
															</td>
														</tr>
														<tr>
															<td>
																Object Size (bytes):
															</td>
															<td class="objSize">
																
															</td>
														</tr>
														<tr>
															<td>
																Owner:
															</td>
															<td class="ownerName">
																
															</td>
														</tr>
														<tr>
															<td>
																Zone:
															</td>
															<td class="ownerZone">
																
															</td>
														</tr>
													</tbody>
													</table>
												</div>
												<div class="metadata-cont tab-pane fade">
													<br />
													<table class="table table-hover meta-tbl">
													<thead>
														<tr>
															<th>Attribute</th>
															<th>Value</th>
															<th>Units</th>
															<th style="width:1%;"></th>
														</tr>
													</thead>
													<tbody>
													</tbody>
													</table>
													<form class="form-inline avu-form" data-objname="{{.Name}}/">
														<div class="form-group">
															<div class="input-group" style="width: 84%;">
																<div class="input-group-addon">A:</div>
																<input type="text" class="form-control avu-attribute" placeholder="Attribute" style="border-right: 0;">
																<div class="input-group-addon">V:</div>
																<input type="text" class="form-control avu-value" placeholder="Value" style="border-left: 0; border-right:0;">
																<div class="input-group-addon">U:</div>
																<input type="text" class="form-control avu-units" placeholder="Units" style="border-left: 0;">
															</div>
															<button type="submit" class="add-avu btn btn-primary">Add AVU</button>
														</div>
													</form>
												</div>
												<div class="acl-cont tab-pane fade">
													<br />
													<table class="table table-hover acl-tbl">
													<thead>
														<tr>
															<th>Name</th>
															<th>Type</th>
															<th style="min-width:16%;">Access Level</th>
														</tr>
													</thead>
													<tbody>
														
													</tbody>
													</table>
													<form class="form-inline acl-form" data-objname="{{.Name}}/">
														<div class="form-group" style="width:100%;">
															<div class="input-group" style="width: 77%;">
																<div class="input-group-addon">User/Group:</div>
																{{ if ne (len .Con.Users) 0 }}
																	<select class="form-control acl-name" style="margin-left:-3px;border-radius: 4px;margin-right:5px;">
																		<option></option>
																		
																		{{ range .Con.Users }}
																			<option value="{{ .Name }}">User: {{ .Name }}</option>
																		{{ end }}
																		{{ range .Con.Groups }}
																			<option value="{{ .Name }}">Group: {{ .Name }}</option>
																		{{ end }}
																		
																	</select>
																{{ else }}
																	<input type="text" class="form-control acl-name" style="margin-left:-3px;border-radius: 4px;margin-right:5px;">
																{{ end }}
																<div class="input-group-addon" style="border-radius: 4px;">Access:</div>
																<select class="form-control acl-access" style="margin-left:-6px;">
																	<option></option>
																	<option value="read">read</option>
																	<option value="write">write</option>
																	<option value="own">own</option>
																	<option value="null">revoke</option>
																</select>
															</div>
															<button type="submit" class="add-acl btn btn-primary">Modify Access</button>
														</div>
													</form>
												</div>
											</div>
										</div>
										<div class="modal-footer">
											<button type="button" class="btn btn-default" data-dismiss="modal">Close</button>
										</div>
									</div>
								</div>
							</div>
						</td>
					</tr>
				{{end}}
				{{range .DataObjs}}
					<tr>
						<th><a href="{{.Name}}">{{.Name}}</a></th>
						<td>{{prettySize .Size}}</td>
						<td>Data Object</td>
						<td>

							<div style="text-align:right;">
								<a href="{{.Name}}?download=1"><span style="margin-right:10px;" class="glyphicon glyphicon-download-alt"></span></a>
								<span style="cursor:pointer;color:#337ab7;margin-right:10px;" data-objname="{{.Name}}" class="glyphicon glyphicon-th-list show-meta-modal"></span>
								<span class="glyphicon glyphicon-remove delete-obj" style="color:red;cursor:pointer;" data-objname="{{.Name}}"></span>
							</div>

							<!-- Modal -->
							<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel">
								<div class="modal-dialog" role="document">
									<div class="modal-content">
										<div class="modal-header">
											<button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
											<h4 class="modal-title" id="myModalLabel">Data Object "{{.Name}}"</h4>
										</div>
										<div class="modal-body">
											<ul class="nav nav-tabs">
												<li class="active"><a data-toggle="tab" data-target=".stat-cont">Stat</a></li>
												<li><a data-toggle="tab" data-target=".metadata-cont">Metadata</a></li>
												<li><a data-toggle="tab" data-target=".acl-cont">ACL</a></li>
											</ul>

											<div class="tab-content">
												<div class="stat-cont tab-pane fade in active">
													<br />
													<table class="table table-hover">
													<tbody>
														<tr>
															<td>
																Checksum:
															</td>
															<td class="chksum">
																
															</td>
														</tr>
														<tr>
															<td>
																Created At:
															</td>
															<td class="createTime">
																
															</td>
														</tr>
														<tr>
															<td>
																Modified At:
															</td>
															<td class="modifyTime">
																
															</td>
														</tr>
														<tr>
															<td>
																Data ID:
															</td>
															<td class="dataId">
																
															</td>
														</tr>
														<tr>
															<td>
																Data Mode:
															</td>
															<td class="dataMode">
																
															</td>
														</tr>
														<tr>
															<td>
																Object Size (bytes):
															</td>
															<td class="objSize">
																
															</td>
														</tr>
														<tr>
															<td>
																Owner:
															</td>
															<td class="ownerName">
																
															</td>
														</tr>
														<tr>
															<td>
																Zone:
															</td>
															<td class="ownerZone">
																
															</td>
														</tr>
													</tbody>
													</table>
												</div>
												<div class="metadata-cont tab-pane fade">
													<br />
													<table class="table table-hover meta-tbl">
													<thead>
														<tr>
															<th>Attribute</th>
															<th>Value</th>
															<th>Units</th>
															<th style="width:1%;"></th>
														</tr>
													</thead>
													<tbody>
														
													</tbody>
													</table>
													<form class="form-inline avu-form" data-objname="{{.Name}}">
														<div class="form-group">
															<div class="input-group" style="width: 84%;">
																<div class="input-group-addon">A:</div>
																<input type="text" class="form-control avu-attribute" placeholder="Attribute" style="border-right: 0;">
																<div class="input-group-addon">V:</div>
																<input type="text" class="form-control avu-value" placeholder="Value" style="border-left: 0; border-right:0;">
																<div class="input-group-addon">U:</div>
																<input type="text" class="form-control avu-units" placeholder="Units" style="border-left: 0;">
															</div>
															<button type="submit" class="add-avu btn btn-primary">Add AVU</button>
														</div>
													</form>
												</div>
												<div class="acl-cont tab-pane fade">
													<br />
													<table class="table table-hover acl-tbl">
													<thead>
														<tr>
															<th>Name</th>
															<th>Type</th>
															<th style="min-width:16%;">Access Level</th>
														</tr>
													</thead>
													<tbody>
														
													</tbody>
													</table>
													<form class="form-inline acl-form" data-objname="{{.Name}}">
														<div class="form-group" style="width:100%;">
															<div class="input-group" style="width: 77%;">
																<div class="input-group-addon">User/Group:</div>
																{{ if ne (len .Con.Users) 0 }}
																	<select class="form-control acl-name" style="margin-left:-3px;border-radius: 4px;margin-right:5px;">
																		<option></option>
																		
																		{{ range .Con.Users }}
																			<option value="{{ .Name }}">User: {{ .Name }}</option>
																		{{ end }}
																		{{ range .Con.Groups }}
																			<option value="{{ .Name }}">Group: {{ .Name }}</option>
																		{{ end }}
																		
																	</select>
																{{ else }}
																	<input type="text" class="form-control acl-name" style="margin-left:-3px;border-radius: 4px;margin-right:5px;">
																{{ end }}
																<div class="input-group-addon" style="border-radius: 4px;">Access:</div>
																<select class="form-control acl-access" style="margin-left:-6px;">
																	<option></option>
																	<option value="read">read</option>
																	<option value="write">write</option>
																	<option value="own">own</option>
																	<option value="null">revoke</option>
																</select>
															</div>
															<button type="submit" class="add-acl btn btn-primary">Modify Access</button>
														</div>
													</form>	
												</div>
											</div>
										</div>
										<div class="modal-footer">
											<button type="button" class="btn btn-default" data-dismiss="modal">Close</button>
										</div>
									</div>
								</div>
							</div>
						</td></td>
					</tr>
				{{end}}
			</tbody>
		</table>
	</div>
	<div id="create-collection-cont" style="display:none;">
		<strong>Create Collection</strong>
		<div style='margin-top:10px;' class='input-group'>
			<input type='text' class='form-control collection-name' placeholder='Collection name...'>
			<span class='input-group-btn'><button class='btn create-collection btn-success' type='button'>Create</button></span>
		</div>
	</div>
</body>
</html>
`

type JSONMap map[string]string
type JSONArr []JSONMap

func (handler *HttpHandler) ServeJSONMeta(obj IRodsObj) {
	handler.response.Header().Set("Content-type", "application/json")

	var jsonResponse map[string]JSONArr = make(map[string]JSONArr)
	var metaResponse JSONArr = make(JSONArr, 0)
	var aclResponse JSONArr = make(JSONArr, 0)
	var statResponse JSONArr = make(JSONArr, 0)

	acls, aclErr := obj.ACL()
	stats, stErr := obj.Stat()

	if mc, err := obj.Meta(); err == nil && aclErr == nil && stErr == nil {
		mc.Each(func(m *Meta) {
			metaResponse = append(metaResponse, JSONMap{
				"attribute": m.Attribute,
				"value":     m.Value,
				"units":     m.Units,
			})
		})
		for _, acl := range acls {
			aclResponse = append(aclResponse, JSONMap{
				"name":        acl.AccessObject.Name(),
				"accessLevel": getTypeString(acl.AccessLevel),
				"type":        getTypeString(acl.Type),
			})
		}

		statResponse = append(statResponse, JSONMap{})
		for k, v := range stats {
			switch realVal := v.(type) {
			case int:
				statResponse[0][k] = strconv.Itoa(realVal)
			case string:
				statResponse[0][k] = realVal

			}
		}

		jsonResponse["metadata"] = metaResponse
		jsonResponse["acl"] = aclResponse
		jsonResponse["stat"] = statResponse
	}

	jsonBytes, _ := json.Marshal(jsonResponse)
	handler.response.Write(jsonBytes)
}

type RangeSegmentOutput struct {
	ContentRange string
	ContentType  string
	ByteContent  []byte
}

type RangeOutput []RangeSegmentOutput

func (ro RangeOutput) TotalLength() string {
	var sum int

	for _, seg := range ro {
		sum += len(seg.ByteContent)
	}

	return strconv.Itoa(sum)
}

func (handler *HttpHandler) ServeDataObj(obj *DataObj) {

	rng := handler.request.Header["Range"]
	objMime := handler.getObjMime(obj)

	lenStr := strconv.FormatInt(obj.Size(), 10)

	if len(rng) > 0 {
		rangeHeader := rng[0]

		byteRange := strings.Split(rangeHeader, "bytes=")

		if len(byteRange) == 2 && byteRange[0] == "" && byteRange[1] != "" {

			var outputBuffer RangeOutput

			byteRangeSplit := strings.Split(byteRange[1], ",")

			for _, rangeStr := range byteRangeSplit {

				rangeStrSplit := strings.Split(rangeStr, "-")
				if len(rangeStrSplit) == 2 {

					var convErr error
					var firstByteN int64
					var lastByteN int64

					firstByte := rangeStrSplit[0]
					lastByte := rangeStrSplit[1]

					// should we only get last bytes
					if firstByte == "" {
						if lastByteN, convErr = strconv.ParseInt(lastByte, 10, 64); convErr != nil {
							log.Print("Error parsing byte range")
							return
						}

						firstByteN = obj.Size() - lastByteN
						lastByteN = obj.Size() - 1
					} else if lastByte == "" {
						if firstByteN, convErr = strconv.ParseInt(firstByte, 10, 64); convErr != nil {
							log.Print("Error parsing byte range")
							return
						}

						lastByteN = obj.Size() - 1
					} else {
						if firstByteN, convErr = strconv.ParseInt(firstByte, 10, 64); convErr != nil {
							log.Print("Error parsing byte range")
							return
						}

						if lastByteN, convErr = strconv.ParseInt(lastByte, 10, 64); convErr != nil {
							log.Print("Error parsing byte range")
							return
						}
					}

					if byteData, err := obj.FastReadFree(firstByteN, int(lastByteN-firstByteN)+1); err == nil {

						defer byteData.Free()

						outputBuffer = append(outputBuffer, RangeSegmentOutput{
							ContentRange: "bytes " + strconv.FormatInt(firstByteN, 10) + "-" + strconv.FormatInt(lastByteN, 10) + "/" + lenStr,
							ContentType:  objMime,
							ByteContent:  byteData.Contents,
						})

					} else {
						log.Print(err)
					}

				} else {
					log.Print("Error parsing byte range")
				}
			}

			if len(outputBuffer) > 1 {

				handler.response.Header().Set("Accept-Ranges", "bytes")
				//handler.response.Header().Set("Content-Length", outputBuffer.TotalLength())

				mpWriter := multipart.NewWriter(handler.response)

				handler.response.Header().Set("Content-Type", "multipart/byteranges; boundary="+mpWriter.Boundary())

				handler.response.WriteHeader(http.StatusPartialContent)

				for _, outputSegment := range outputBuffer {

					var headers textproto.MIMEHeader = make(textproto.MIMEHeader)

					headers.Add("Content-Type", outputSegment.ContentType)
					headers.Add("Content-Range", outputSegment.ContentRange)

					if writer, err := mpWriter.CreatePart(headers); err != nil {
						log.Print(err)
						continue
					} else {
						writer.Write(outputSegment.ByteContent)
					}
				}

			} else if len(outputBuffer) == 1 {
				handler.response.Header().Set("Content-Range", outputBuffer[0].ContentRange)
				handler.response.Header().Set("Accept-Ranges", "bytes")
				handler.response.Header().Set("Content-Length", strconv.Itoa(len(outputBuffer[0].ByteContent)))
				handler.response.Header().Set("Content-Type", outputBuffer[0].ContentType)

				handler.response.WriteHeader(http.StatusPartialContent)

				handler.response.Write(outputBuffer[0].ByteContent)
			}
		}

	} else {

		if handler.opts.Download || handler.query.Get("download") != "" {
			handler.response.Header().Set("Content-Disposition", "attachment; filename="+obj.Name())
			handler.response.Header().Set("Content-type", "application/octet-stream")
		} else {
			handler.response.Header().Set("Content-type", objMime)
		}

		handler.response.Header().Set("Accept-Ranges", "bytes")
		handler.response.Header().Set("Content-Length", lenStr)

		outBuff := make(chan *ByteArr, 100)

		go func() {
			if readEr := obj.ReadChunkFree(10240000, func(chunk *ByteArr) {
				outBuff <- chunk
			}); readEr != nil {
				log.Print(readEr)

				handler.response.WriteHeader(http.StatusInternalServerError)
				handler.response.Write([]byte("Error: " + readEr.Error()))

			}

			close(outBuff)
		}()

		for b := range outBuff {
			handler.response.Write(b.Contents)
			b.Free()
		}
	}

}

func (handler *HttpHandler) getObjMime(obj *DataObj) string {
	var mimeType string

	ext := filepath.Ext(obj.Path())

	if ext != "" {
		mimeType = mime.TypeByExtension(ext)

		if mimeType == "" {
			log.Printf("Can't find mime type for %s extension", ext)
			mimeType = "application/octet-stream"
		}
	} else {
		mimeType = "application/octet-stream"
	}

	return mimeType
}

func (handler *HttpHandler) Serve404() {
	handler.response.Header().Set("Content-Type", "text/html")
	handler.response.WriteHeader(http.StatusNotFound)

	handler.response.Write([]byte("<h3>404 Not Found: " + handler.openPath + "</h3>"))
}

func (handler *HttpHandler) ServeCollectionView(col *Collection) {

	handler.response.Header().Set("Content-Type", "text/html")

	t, err := template.New("collectionList").Funcs(template.FuncMap{
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
		"preHeaderLinksLen": func() int {
			return len(strings.Split(handler.handlerPath, "/"))
		},
		"headerLinks": func() []map[string]string {
			headerLinks := make([]map[string]string, 0)

			preFrags := strings.Split(handler.handlerPath, "/")
			for i := range preFrags {

				headerLinks = append(headerLinks, map[string]string{
					"name": preFrags[i],
					"url":  handler.opts.StripPrefix,
				})

			}

			if handler.openPath == handler.handlerPath {
				return headerLinks
			}

			p := strings.TrimPrefix(handler.openPath, handler.handlerPath+"/")
			frags := strings.Split(p, "/")

			for i := range frags {
				var path string

				if i > 0 {
					path = strings.Join(frags[0:i], "/") + "/"
				} else {
					path = ""
				}

				headerLinks = append(headerLinks, map[string]string{
					"name": frags[i],
					"url":  (handler.opts.StripPrefix + path + frags[i] + "/"),
				})

			}

			return headerLinks
		},
		"usersJSON": func() []string {
			usrs := make([]string, 0)

			conUsers, err := col.Con().Users()
			check(err)

			for _, u := range conUsers {
				usrs = append(usrs, u.Name())
			}

			return usrs
		},
		"groupsJSON": func() []string {
			grps := make([]string, 0)

			conGroups, err := col.Con().Groups()
			check(err)

			for _, g := range conGroups {
				grps = append(grps, g.Name())
			}

			return grps
		},
	}).Parse(tpl)
	check(err)

	err = t.Execute(handler.response, col)
	check(err)
}

func (handler *HttpHandler) AddMetaAVU(obj IRodsObj) {
	handler.response.Header().Set("Content-type", "application/json")

	var response struct {
		Success bool
		Message string
	}

	req := handler.request

	req.ParseForm()

	a := strings.TrimSpace(req.PostForm.Get("attribute"))
	v := strings.TrimSpace(req.PostForm.Get("value"))
	u := strings.TrimSpace(req.PostForm.Get("units"))

	if _, err := obj.AddMeta(Meta{
		Attribute: a,
		Value:     v,
		Units:     u,
	}); err == nil {
		response.Success = true
		response.Message = "Added metadata successfully"
	} else {
		response.Message = err.Error()
	}

	jsonBytes, _ := json.Marshal(response)
	handler.response.Write(jsonBytes)

}

func (handler *HttpHandler) DeleteObj(obj IRodsObj) {
	handler.response.Header().Set("Content-type", "application/json")

	var response struct {
		Success bool
		Message string
	}

	if err := obj.Delete(true); err == nil {
		response.Success = true
		response.Message = "Object delete successfully"
	} else {
		response.Message = err.Error()
	}

	jsonBytes, _ := json.Marshal(response)
	handler.response.Write(jsonBytes)
}

func (handler *HttpHandler) DeleteMetaAVU(obj IRodsObj) {
	handler.response.Header().Set("Content-type", "application/json")

	var response struct {
		Success bool
		Message string
	}

	req := handler.request

	req.ParseForm()

	a := strings.TrimSpace(req.PostForm.Get("attribute"))
	v := strings.TrimSpace(req.PostForm.Get("value"))
	u := strings.TrimSpace(req.PostForm.Get("units"))

	if mc, err := obj.Meta(); err == nil {
		match := mc.Metas.MatchOne(&Meta{
			Attribute: a,
			Value:     v,
			Units:     u,
		})

		if match != nil {
			if _, er := match.Delete(); er == nil {
				response.Success = true
				response.Message = "Meta AVU deleted successfully"
			} else {
				response.Message = er.Error()
			}
		} else {
			response.Message = "Error finding meta AVU to delete"
		}
	} else {
		response.Message = err.Error()
	}

	jsonBytes, _ := json.Marshal(response)
	handler.response.Write(jsonBytes)

}

func (handler *HttpHandler) CreateCollection(col *Collection) {
	handler.response.Header().Set("Content-type", "application/json")

	var response struct {
		Success bool
		Message string
	}

	req := handler.request

	req.ParseForm()

	colName := strings.TrimSpace(req.PostForm.Get("colname"))

	if _, err := col.CreateSubCollection(colName); err == nil {
		response.Success = true
		response.Message = "Subcollection created successfully"
	} else {
		response.Message = err.Error()
	}

	jsonBytes, _ := json.Marshal(response)
	handler.response.Write(jsonBytes)

}

func (handler *HttpHandler) AddACL(obj IRodsObj) {
	handler.response.Header().Set("Content-type", "application/json")

	var response struct {
		Success bool
		Message string
	}

	req := handler.request

	req.ParseForm()

	name := strings.TrimSpace(req.PostForm.Get("name"))
	accessString := strings.TrimSpace(req.PostForm.Get("access"))

	var accessLevel int
	switch accessString {
	case "own":
		accessLevel = Own
	case "write":
		accessLevel = Write
	case "read":
		accessLevel = Read
	case "null":
		accessLevel = Null
	}

	if err := obj.Chmod(name, accessLevel, true); err == nil {
		response.Success = true
		response.Message = "Added metadata successfully"
	} else {
		response.Message = err.Error()
	}

	jsonBytes, _ := json.Marshal(response)
	handler.response.Write(jsonBytes)

}

func (handler *HttpHandler) Upload(col *Collection) {
	handler.response.Header().Set("Content-type", "application/json")

	var response struct {
		Success bool
		Message string
	}

	req := handler.request

	if mpReader, err := req.MultipartReader(); err == nil {

	MPLoop:
		for {
			part, pErr := mpReader.NextPart()
			if pErr == io.EOF {
				break MPLoop
			}

			if pErr != nil {
				response.Message = pErr.Error()
				log.Print(pErr)
				break MPLoop
			}

			if obj, cEr := col.CreateDataObj(DataObjOptions{
				Name: part.FileName(),
			}); cEr == nil {

				contents := make([]byte, 1024*200000)

				response.Success = true
				response.Message = "File upload success"

			ReadLoop:
				for {

					n, fErr := part.Read(contents)

					if fErr != nil && fErr != io.EOF {
						log.Print(fErr.Error())
						panic(fErr)
					}

					if n == 0 && io.EOF == fErr {
						break ReadLoop
					}

					if wEr := obj.WriteBytes(contents[:n]); wEr != nil {
						response.Message = wEr.Error()
						response.Success = false

						log.Print(wEr)

						break ReadLoop
					}
				}

				obj.Close()

			} else {
				log.Print(cEr)
				response.Message = cEr.Error()
			}
		}

	} else {
		log.Print(err)
		response.Message = err.Error()
	}

	if jsonBytes, jErr := json.Marshal(response); jErr == nil {
		if _, wErr := handler.response.Write(jsonBytes); wErr != nil {
			log.Print(wErr)
		}
	} else {
		log.Print(jErr)
	}
}

func (handler *HttpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {

	handler.response = response
	handler.request = request

	handler.handlerPath = strings.TrimRight(handler.path, "/")
	urlPath := strings.TrimRight(handler.request.URL.Path, "/")
	handler.openPath = strings.TrimRight(handler.handlerPath+"/"+urlPath, "/")

	handler.query = request.URL.Query()

	var handlerMain = func(con *Connection) {
		if objType, err := con.PathType(handler.openPath); err == nil {

			if objType == DataObjType {
				if obj, er := con.DataObject(handler.openPath); er == nil {

					switch q := handler.query; true {
					case q.Get("meta") != "":
						if request.Method == "GET" {
							handler.ServeJSONMeta(obj)
						} else if request.Method == "POST" {
							handler.AddMetaAVU(obj)
						}
					case q.Get("deletemeta") != "":
						if request.Method == "POST" {
							handler.DeleteMetaAVU(obj)
						}
					case q.Get("createacl") != "":
						if request.Method == "POST" {
							handler.AddACL(obj)
						}
					case q.Get("delete") != "":
						if request.Method == "POST" {
							handler.DeleteObj(obj)
						}
					default:
						handler.ServeDataObj(obj)
					}

					if cErr := obj.Close(); cErr != nil {
						log.Print(cErr)
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
					Path:      handler.openPath,
					Recursive: false,
					GetRepls:  false,
				}); er == nil {

					switch q := handler.query; true {

					case q.Get("meta") != "":
						if request.Method == "GET" {
							handler.ServeJSONMeta(col)
						} else if request.Method == "POST" {
							handler.AddMetaAVU(col)
						}
					case q.Get("deletemeta") != "":
						if request.Method == "POST" {
							handler.DeleteMetaAVU(col)
						}
					case q.Get("createcol") != "":
						if request.Method == "POST" {
							handler.CreateCollection(col)
						}
					case q.Get("createacl") != "":
						if request.Method == "POST" {
							handler.AddACL(col)
						}
					case q.Get("upload") != "":
						if request.Method == "POST" {
							handler.Upload(col)
						}
					case q.Get("delete") != "":
						if request.Method == "POST" {
							handler.DeleteObj(col)
						}
					default:
						handler.ServeCollectionView(col)
					}

					if cErr := col.Close(); cErr != nil {
						log.Print(cErr)
					}

				} else {
					log.Print(er)
				}
			}

		} else {

			handler.Serve404()

			log.Print(err)
		}

	}

	if handler.client != nil {
		if er := handler.client.OpenConnection(handlerMain); er != nil {
			log.Print(er)
			return
		}
	} else if handler.connection != nil {
		handlerMain(handler.connection)
	}

}
