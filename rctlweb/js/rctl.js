!function ($) {

    $(function() {
        var $window = $(window);
        var $body = $(document.body);
        var $form = $('form', $body);
        var	$wrap = $('.main-container', $body);
		var $modal = '<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="ModalLabel" aria-hidden="true">' +
						'<div class="modal-dialog">' +
							'<div class="modal-content">' +
								'<div class="modal-header">' +
									'<button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>' +
									'<h4 class="modal-title" id="ModalLabel">Modal title</h4>' +
								'</div>' +
								'<div class="modal-body"></div>' +
							'</div>' +
						'</div>' +
					'</div>';

        var changeUserAdminStatus = function (obj) {
            $('.notify').remove();

            self = obj;

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#auid').val()),
                Cmd: $(self).hasClass('disabled') ? 'enable-user' : 'disable-user' }

            $.ajax({
                type: "POST",
                url: "/set-user-status",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function(data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        var clr = $(self).hasClass('disabled') ? 'glyphicon-ok' : 'glyphicon-remove';
                        var dsc = $(self).hasClass('disabled') ? 'User enabled' : 'User disabled';
                        var val = $(self).hasClass('disabled') ? 'enabled' : 'disabled';

                        $(self).html('<span class="glyphicon ' + clr + '"></span>&nbsp; ' + dsc);

						if ( $(self).hasClass('disabled')) {
							$(self).removeClass('disabled');
							$(self).addClass('enabled');
						} else {
							$(self).removeClass('enabled');
							$(self).addClass('disabled');
						}

                        getUserUidList('Activity History', 'activity', '1', '10');

                        var msg = $form.find('#login').val() + ' is now ' + val;
                        $wrap.append('<div class="notify"><div class="alert alert-success">' + msg + '</div></div>');
                    }
                }
            });
        };

        var changeUserStatus = function (obj) {
            $('.notify').remove();

            self = obj;

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#auid').val()),
                Cmd: $(self).hasClass('inactive') ? 'activate-user' : 'deactivate-user' }

            $.ajax({
                type: "POST",
                url: "/set-user-status",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        var clr = $(self).hasClass('inactive') ? 'glyphicon-ok' : 'glyphicon-remove';
                        var dsc = $(self).hasClass('inactive') ? 'User active' : 'User inactive';
                        var val = $(self).hasClass('inactive') ? 'activated' : 'deactivated';

                        $(self).html('<span class="glyphicon ' + clr + '"></span>&nbsp ' + dsc);

						if ( $(self).hasClass('inactive')) {
							$(self).removeClass('inactive');
							$(self).addClass('active');
						} else {
							$(self).removeClass('active');
							$(self).addClass('inactive');
						}

                        getUserUidList('Activity History', 'activity', '1', '10');

                        var msg = $form.find('#login').val() + ' is now ' + val;
                        $wrap.append('<div class="notify"><div class="alert alert-success">' + msg + '</div></div>');
                    }
                }
            });
        };

        var changeUserAttr = function (attr) {
            $('.notify').remove();

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#auid').val()),
                Cmd: attr,
                Data: $form.find('#' + attr).val() }

            $.ajax({
                type: "POST",
                url: "/set-user-attr",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        getUserUidList('Activity History', 'activity', '1', '10');

                        var e = JSON.parse(data.Data);
                        var msg = $('#login', $form).val() + ' ' + attr + ' is now ' + e.Value;
                        $wrap.append('<div class="notify"><div class="alert alert-success">' + msg + '</div></div>');
                    }
                }
            });
        };

        var getUserList = function (title, list, page, entries, sort) {
            var opt = list + ':' + page + ':' + entries + ':' + sort;
            var param = { Sid: $form.find('#sid').val(),
                Uid: 0,
                Cmd: '',
                Data: opt };

            $.ajax({
                type: "POST",
                url: "/list-user",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    $('div.user-list').empty();
                    $("div.user-list").removeClass("text-center");
                    $("div.user-list").removeClass("col-md-12");

                    if(data.ErrNo != 0) {
                        $("div.user-list").addClass("text-center");
                        $("div.user-list").append('Empty ' + list + ' user list');
                    } else {
                        var e = JSON.parse(data.Data);

                        $("div.user-list").addClass("col-md-12");
                        $("div.user-list").attr("id", "user-cont");

                        $con = $('#user-cont');

                        $('<h3 class="panel-title grey">' + title + '</h3>').appendTo($con);
                        $('<table class="table table-striped table-condensed table-bordered list-user">').appendTo($con);

                        $tbl = $('table', $con);
                        $tbl.append('<thead><tr><td>ID</td>' +
                            '<td>Username</td>' +
                            '<td>Name</td>' +
                            '<td>Registered</td>' +
                            '<td>Admin</td>' +
                            '<td>Status</td></tr></thead>');

                        $.each(e.Entry, function (k, v) {
                            var at, st;

                            if (v.Admin == 'enabled') {
                                at = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                            } else {
                                at = '<span class="glyphicon glyphicon-remove-sign red"></span>'
                            }

                            if (v.Status == 'active') {
                                st = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                            } else {
                                st = '<span class="glyphicon glyphicon-remove-sign red"></span>'
                            }

                            $('<tr><td class="text-center">' + v.Id + '</td>' +
                                '<td><a href="/list?uid=' + v.Id + '">' + v.Login + '</a></td>' +
                                '<td>' + v.Name + '</td>' +
                                '<td>' + v.Registered + '</td>' +
                                '<td class="text-center">' + at + '</td>' +
                                '<td class="text-center">' + st + '</td></tr>').appendTo($tbl);
                        });

						scrollTable(page, entries, e.Total, $con, function (page) {
							getUserList(title, list, page, entries, sort)
						});
                    }
                }
            });
        };

        var getUserUidList = function (title, list, page, entries, sort) {
            var opt = list + ':' + page + ':' + entries + ':' + sort;
			var	param = { Sid: $form.find('#sid').val(),
                    Uid: parseInt($form.find('#auid').val()),
                    Cmd: '',
                    Data: opt };

            $.ajax({
                type: "POST",
                url: "/get-user-list",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
					$('.uid-list').empty();

					if(data.ErrNo != 0) {
                        var user = $form.find('#login').val();

                        $("div.uid-list").addClass("text-center");
                        $("div.uid-list").append('Empty ' + list + ' user list');
					} else {
						var e = JSON.parse(data.Data);

						if ($('div.uid-list').hasClass('text-center'))
							$('div.uid-list').removeClass('text-center');

                        $("div.uid-list").addClass("col-md-12");

                        $con = $('.uid-list');

                        $('<h3 class="panel-title grey">' + title + '</h3>').appendTo($con);
                        $('<table class="table table-striped table-condensed table-bordered list-user">').appendTo($con);

                        $tbl = $('table', $con);
                        $tbl.append('<thead><tr>' +
                            '<td>Action</td>' +
                            '<td>IP</td>' +
                            '<td>Timestamp</td></tr></thead>');

                        $.each(e.Entry, function(k, v) {
                            $('<tr><td>' + v.Action + '</td>' +
                                '<td>' + v.IP + '</td>' +
                                '<td>' + v.Time + '</td></tr>').appendTo($tbl);
                        });

						scrollTable(page, entries, e.Total, $con, function (page) {
							getUserUidList(title, list, page, entries, sort)
						});
                    }
                }
            });
        };

        var getUserSessionList = function () {
            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#auid').val()),
                Cmd: '',
                Data: '' };

            $.ajax({
                type: "POST",
                url: "/get-user-sessions",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    $('.uid-list').empty();

					if(data.ErrNo != 0) {
                        var user = $form.find('#login').val();

                        $("div.uid-list").addClass("text-center");
                        $("div.uid-list").append('Empty tunnel session list');
                    } else {
                        var e = JSON.parse(data.Data);

						if ($('div.uid-list').hasClass('text-center'))
							$('div.uid-list').removeClass('text-center');

                        $("div.uid-list").addClass("col-md-12");

                        $con = $('.uid-list');

                        $('<h3 class="panel-title grey">Tunnel Sessions</h3>').appendTo($con);
                        $('<table class="table table-striped table-condensed table-bordered list-user">').appendTo($con);
                        $('h3', $con).append('<span class="label label-info blue"> [' + e.Total + ']</span>');

                        $tbl = $('table', $con);
                        $tbl.append('<thead><tr>' +
                            '<td>Server Name</td>' +
                            '<td>Type</td>' +
                            '<td>Tunnel Source</td>' +
                            '<td>Tunnel Destination</td>' +
                            '<td>Source Address</td>' +
                            '<td>Destination Address</td>' +
                            '<td>Routed Prefix</td>' +
                            '<td>Status</td></tr></thead><tbody></tbody>');

                        $.each(e.Entry, function(k, v) {
                            var st = '<span class="glyphicon glyphicon-remove-sign red inactive" style="cursor: pointer"></span>'

                            if(v.Status == 'active') {
                                st = '<span class="glyphicon glyphicon-ok-sign blue active" style="cursor: pointer"></span>'
                            }

                            $('<tr><td class="svinfo"><a href="/list?vid=' + v.Id + '">' + v.ServerName +
                                '<input type="hidden" id="svsid" value="' + v.Id + '"/>' +
                                '<input type="hidden" id="vid" value="' + v.ServerId + '"/>' +
                                '</a></td>' +
                                '<td>' + v.Type + '</td>' +
                                '<td>' + v.TunSrc + '</td>' +
                                '<td class="tun-dst">' + v.TunDst + '</td>' +
                                '<td>' + v.Src + '</td>' +
                                '<td>' + v.Dst + '</td>' +
                                '<td>' + v.Rt + '</td>' +
                                '<td class="trigger text-center">' + st + '</td></tr>').appendTo($tbl);
                        });
                    }
                }
            });
        };

        var resetUserPw = function (e) {
            $('.notify').remove();

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#auid').val()),
                Cmd: "",
                Data: "" }

            $.ajax({
                type: "POST",
                url: "/reset-user-pw",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        getUserActivityList();

                        $wrap.append('<div class="notify"><div class="alert alert-success">' + $form.find('#login').val() + ' password has been reset</div></div>');
                    }
                }
            });
        };

        var changeServerAdminStatus = function (obj) {
            $('.notify').remove();

            self = obj;

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#vid').val()),
                Cmd: $(self).hasClass('disabled') ? 'enable-server' : 'disable-server' }

            $.ajax({
                type: "POST",
                url: "/set-server-status",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function(data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        var clr = $(self).hasClass('disabled') ? 'glyphicon-ok' : 'glyphicon-remove';
                        var dsc = $(self).hasClass('disabled') ? 'Server enabled' : 'Server disabled';
                        var val = $(self).hasClass('disabled') ? 'enabled' : 'disabled';

                        $(self).html('<span class="glyphicon ' + clr + '"></span>&nbsp; ' + dsc);

						if ( $(self).hasClass('disabled')) {
							$(self).removeClass('disabled');
							$(self).addClass('enabled');
						} else {
							$(self).removeClass('enabled');
							$(self).addClass('disabled');
						}

                        getServerSvidList('Assigned Tunnel Sessions', 'assigned-sessions', '1', '10');

                        var msg = $form.find('#svname').val() + ' is now ' + val;
                        $wrap.append('<div class="notify"><div class="alert alert-success">' + msg + '</div></div>');
                    }
                }
            });
        };

        var changeServerStatus = function (obj) {
            $('.notify').remove();

            self = obj;

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#vid').val()),
                Cmd: $(self).hasClass('inactive') ? 'activate-server' : 'deactivate-server' }

            $.ajax({
                type: "POST",
                url: "/set-server-status",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        var clr = $(self).hasClass('inactive') ? 'glyphicon-ok' : 'glyphicon-remove';
                        var dsc = $(self).hasClass('inactive') ? 'Server active' : 'Server inactive';
                        var val = $(self).hasClass('inactive') ? 'activated' : 'deactivated';

                        $(self).html('<span class="glyphicon ' + clr + '"></span>&nbsp; ' + dsc);

						if ( $(self).hasClass('inactive')) {
							$(self).removeClass('inactive');
							$(self).addClass('active');
						} else {
							$(self).removeClass('active');
							$(self).addClass('inactive');
						}

                        getServerSvidList('Assigned Tunnel Sessions', 'assigned-sessions', '1', '10');

                        var msg = $form.find('#svname').val() + ' is now ' + val;
                        $wrap.append('<div class="notify"><div class="alert alert-success">' + msg + '</div></div>');
                    }
                }
            });
        };

        var changeServerAttr = function (attr) {
            $('.notify').remove();

            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#vid').val()),
                Cmd: attr,
                Data: $form.find('#' + attr).val() }

            $.ajax({
                type: "POST",
                url: "/set-server-attr",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    if(data.ErrNo != 0) {
                        $wrap.append('<div class="notify"><div class="alert alert-danger">' + data.Data + '</div></div>');
                    } else {
                        getUserActivityList(1);

                        var e = JSON.parse(data.Data);
                        var msg = $('#svname', $form).val() + ' ' + attr + ' is now ' + e.Value;
                        $wrap.append('<div class="notify"><div class="alert alert-success">' + msg + '</div></div>');
                    }
                }
            });
        };

        var getServerList = function (title, list, page, entries, sort) {
            var opt = list + ':' + page + ':' + entries + ':' + sort
            var param = { Sid: $form.find('#sid').val(),
                Uid: 0,
                Cmd: '',
                Data: opt };

            $.ajax({
                type: "POST",
                url: "/list-server",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    $('div.server-list').empty();
                    $("div.server-list").removeClass("text-center");
                    $("div.server-list").removeClass("col-md-12");

                    if(data.ErrNo != 0) {
                        $("div.server-list").addClass("text-center");
                        $("div.server-list").append('Empty ' + list + ' server list');
                    } else {
                        var e = JSON.parse(data.Data);

                        $("div.server-list").addClass("col-md-12");
                        $("div.server-list").attr("id", "server-cont");

                        $con = $('#server-cont');

                        $('<h3 class="panel-title grey">' + title + '</h3>').appendTo($con);
                        $('<table class="table table-striped table-condensed table-bordered list-user">').appendTo($con);

                        $tbl = $('table', $con);
                        $tbl.append('<thead><tr><td>ID</td>' +
                            '<td>Server Name</td>' +
                            '<td>Entity</td>' +
                            '<td>Access</td>' +
                            '<td>Tunnel</td>' +
                            '<td>Alias</td>' +
                            '<td>Description</td>' +
                            '<td>Location</td>' +
                            '<td>Admin</td>' +
                            '<td>Status</td></tr></thead>');

                        $.each(e.Entry, function (k, v) {
                            var at, st;

                            if (v.Admin == 'enabled') {
                                at = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                            } else {
                                at = '<span class="glyphicon glyphicon-remove-sign red"></span>'
                            }

                            if (v.Status == 'active') {
                                st = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                            } else {
                                st = '<span class="glyphicon glyphicon-remove-sign red"></span>'
                            }

                            $('<tr><td class="text-center">' + v.Id + '</td>' +
                                '<td><a href="/list?vid=' + v.Id + '">' + v.Name + '</a></td>' +
                                '<td>' + v.Entity + '</td>' +
                                '<td>' + v.Access + '</td>' +
                                '<td>' + v.Tunnel + '</td>' +
                                '<td>' + v.Alias + '</td>' +
                                '<td>' + v.Descr + '</td>' +
                                '<td>' + v.Location + '</td>' +
                                '<td class="text-center">' + at + '</td>' +
                                '<td class="text-center">' + st + '</td></tr>').appendTo($tbl);
                        });

						scrollTable(page, entries, e.Total, $con, function (cpage) {
							getServerList(title, list, cpage, entries, sort)
						});
                    }
                }
            });
        };

        var getServerSvidList = function (title, list, page, entries) {
            var opt = list + ':' + page + ':' + entries + ':'
            var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#vid').val()),
                Cmd: '',
                Data: opt };

            $.ajax({
                type: "POST",
                url: "/get-server-list",
                data: JSON.stringify(param),
                dataType: "json",
                contentType: "application/json; charset=utf-8",
                traditional: true,
                success: function (data) {
                    $('.svid-list').empty();

                    if(data.ErrNo != 0) {
                        var sv = $form.find('#svname').val()

                        $("div.svid-list").addClass("text-center");
                        $("div.svid-list").append('<span class="text-light-bold grey">' + sv + '</span> has no assigned tunnel sessions');
                    } else {
                        var e = JSON.parse(data.Data);

						if ($('div.svid-list').hasClass('text-center'))
							$('div.svid-list').removeClass('text-center');

                        $("div.svid-list").addClass("col-md-12");
                        $("div.svid-list").attr("id", "svid-cont");

                        $con = $('#svid-cont');

                        $('<h3 class="panel-title grey">' + title + '</h3>').appendTo($con);
                        $('<table class="table table-striped table-condensed table-bordered list-user">').appendTo($con);

                        $tbl = $('table', $con);

                        if (list == "session-activity") {
                            $tbl.append('<thead><tr><td>ID</td>' +
                                '<td>User ID</td>' +
                                '<td>Action</td>' +
                                '<td>TImestamp</td></tr></thead>');

                            $.each(e.Entry, function (k, v) {
                                $('<tr><td class="text-center">' + v.Sid + '</td>' +
                                    '<td class="text-center">' + v.Uid + '</td>' +
                                    '<td>' + v.Action + '</td>' +
                                    '<td>' + v.Time + '</td></tr>').appendTo($tbl);
                            });
                        } else if (list == "all-users") {
                            $tbl.append('<thead><tr><td>User ID</td>' +
                                '<td>Username</td>' +
                                '<td>Name</td>' +
                                '<td>Registered</td>' +
                                '<td>Admin</td>' +
                                '<td>Status</td></tr></thead>');

                            $.each(e.Entry, function (k, v) {
                                var at, st;

                                if (v.Admin == 'enabled') {
                                    at = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                                } else {
                                    at = '<span class="glyphicon glyphicon-remove-sign red"></span>'
                                }

                                if (v.Status == 'active') {
                                    st = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                                } else {
                                    st = '<span class="glyphicon glyphicon-remove-sign red"></span>'
                                }

                                $('<tr><td class="text-center">' + v.Id + '</td>' +
                                    '<td>' + v.Login + '</td>' +
                                    '<td>' + v.Name + '</td>' +
                                    '<td>' + v.Registered + '</td>' +
                                    '<td class="text-center">' + at + '</td>' +
                                    '<td class="text-center">' + st + '</td></tr>').appendTo($tbl);
                            });
                        } else {
                            $tbl.append('<thead><tr><td>ID</td>' +
                                '<td>User ID</td>' +
                                '<td>Tunnel Destination</td>' +
                                '<td>Source Address</td>' +
                                '<td>Destination Address</td>' +
                                '<td>Routed Prefix</td>' +
                                '<td>Status</td></tr></thead>');

                            $.each(e.Entry, function (k, v) {
                                var st = '<span class="glyphicon glyphicon-remove-sign red"></span>';

                                if (v.Status == 'active') {
                                    st = '<span class="glyphicon glyphicon-ok-sign blue"></span>'
                                }

                                $('<tr><td class="text-center">' + v.Id + '</td>' +
                                    '<td class="text-center"><a href="/list?uid=' + v.Uid + '">' + v.Uid + '</td>' +
                                    '<td>' + v.TunDst + '</td>' +
                                    '<td>' + v.Src + '</td>' +
                                    '<td>' + v.Dst + '</td>' +
                                    '<td>' + v.Rt + '</td>' +
                                    '<td class="text-center">' + st + '</td></tr>').appendTo($tbl);
                            });
                        }

						scrollTable(page, entries, e.Total, $con, function (cpage) {
							getServerSvidList(title, list, cpage, entries)
						});
                    }
                }
            });
        };

        var setTunnelSession = function (obj) {
            $('.notify', $wrap).remove();

            var self = obj;

			if (self.hasClass('active')) {
                ip = self.parent().siblings('.tun-dst').text();
				sendTunnelSessionRequest(self, ip);
			} else {
				$($modal).appendTo($body);
				$('.modal', $body).attr('id', 'dst');

				var container = $('#dst', $body);
				var loading = '<span class="loading" style="background: url(/s/img/wait.gif) no-repeat 0 0; line-height: 1; width: 13px; height: 11px; position: absolute; top: 13px; right: 7px;display: block;z-index: 3"></span>';

				$('#ModalLabel', container).text('Tunnel Destination');
				$('.modal-body', container).append('<form class="form-inline" role="form">' +
                    '<div class="form-group"><input type="text" required id="dst" class="form-control" autocomplete="off"></div>' +
                    '&nbsp;&nbsp;<button class="btn btn-primary">Activate</button><span class="help-block"></span></form>');

				var form = $('form', container);
				var helpBlock = $('span.help-block', form);
				var inputText = $('input#dst', form);
				var button = $('button.btn-primary', form);

				container.modal();

				container.on('shown.bs.modal', function () {
					inputText.focus();
				});

				container.on('hidden.bs.modal', function () {
					this.remove();
				});

				container.on('click', button, function(event) {
					event.preventDefault();
					sendTunnelSessionRequest(self, inputText.val());
					container.modal('hide');
				});

				container.on('keypress', inputText, function(event) {
					var key = (event.keyCode ? event.keyCode : event.which);
					if ( key == 13 && !inputText.val() ) {
						event.preventDefault();
						sendTunnelSessionRequest(self, inputText.val());
						container.modal('hide');
					}
				});
			}
        };

		var sendTunnelSessionRequest = function (context, ip) {
			var cmd = context.hasClass('active') ? 'deactivate-session' : 'activate-session';
			var param = { Sid: $form.find('#sid').val(),
				Uid: parseInt(context.parent().siblings('.svinfo').find('input#vid').val()),
				Cmd: cmd,
				Data: ip + ':' + $form.find('#auid').val() };

			$.ajax({
				type: "POST",
				url: "/set-user-session",
				data: JSON.stringify(param),
				dataType: "json",
				contentType: "application/json; charset=utf-8",
				traditional: true,
				success: function (data) {
					var svname = context.parent().siblings('.svinfo').text();
					var sid = context.parent().siblings('.svinfo').find('input#svsid').val();

					if(data.ErrNo != 0) {
						var msg = context.hasClass('active') ? 'Session [' + svname + ':' + sid + '] cannot be deactivated' : 'Session [' + svname + ':' + sid + '] cannot be activated';
						$wrap.append('<div class="notify"><div class="alert ' + 'alert-danger">' + msg + '</div></div>');
					} else {
						var msg = context.hasClass('active') ? 'Session [' + svname + ':' + sid + '] deactivated' : 'Session [' + svname + ':' + sid + '] activated';
						$wrap.append('<div class="notify"><div class="alert ' + 'alert-success">' + msg + '</div></div>');

						var e = JSON.parse(data.Data);

						if (cmd == 'activate-session')
							context.parent().siblings('td.tun-dst').text(e.IP);
						else
							context.parent().siblings('td.tun-dst').text('');
						if (context.hasClass('active')) {
							context.removeClass('glyphicon glyphicon-ok-sign blue active');
							context.addClass('glyphicon glyphicon-remove-sign red inactive');
						} else {
							context.removeClass('glyphicon glyphicon-remove-sign red inactive');
							context.addClass('glyphicon glyphicon-ok-sign blue active');
						}
					}
				}
			});
		};

        var resolveUserLogin = function (cmd) {
			$($modal).appendTo($body);
			$('.modal', $body).attr('id', 'session-cont');

			var container = $('#session-cont', $body);
			var loading = '<span class="loading" style="background: url(/s/img/wait.gif) no-repeat 0 0; line-height: 1; width: 13px; height: 11px; position: absolute; top: 13px; right: 7px;display: block;z-index: 3"></span>';
			var found   = '<span class="found glyphicon glyphicon-ok-sign" style="line-height: 1; width: 16px; height: 11px; position: absolute; top: 11px; right: 7px;display: block;z-index: 3"></span>';

			$('#ModalLabel', container).text('Assign User');
			$('.modal-body', container).append('<form class="form-inline" role="form">' +
                '<div class="form-group"><div class="input-group"><input type="email" id="login" class="form-control" autocomplete="off"></div></div>' +
                '&nbsp;&nbsp;<button class="btn btn-primary">Assign</button><span class="help-block"></span></form>');

			var inputGroup = $('.input-group', container);
			var helpBlock  = $('span.help-block', container);
			var inputText  = $('input#login', container);
			var button = $('button.btn-primary', container);

			container.modal();

			container.on('shown.bs.modal', function () {
				inputText.focus();
			});

			container.on('hidden.bs.modal', function () {
				this.remove();
			});

			container.on('click', '.btn-primary', function (event) {
				var param = { Sid: $form.find('#sid').val(),
                    Uid: parseInt($form.find('#uid').val()),
                    Cmd: 'resolve-user',
                    Data: inputText.val()
				};

				inputGroup.append(loading);

				$.ajax({
					type: "POST",
					url: "/resolve-user",
					data: JSON.stringify(param),
					dataType: "json",
					contentType: "application/json; charset=utf-8",
					traditional: true,
					success: function (data) {
						if(data.ErrNo != 0) {
							$("span.loading", inputGroup).remove();
							$("span.found", inputGroup).remove();
							helpBlock.text(data.Data);
							helpBlock.addClass('red');
							inputText.select();
						} else {
							if (helpBlock.hasClass('red'))
								helpBlock.removeClass('red');

							$("span.loading", inputGroup).remove();
							$("span.found", inputGroup).remove();
							helpBlock.text('');
							inputGroup.append(found);

							var e = JSON.parse(data.Data);

							setSessionOwner(cmd, inputText.val(), e.Id);
							container.modal('hide');
						}
					}
				});

				event.preventDefault();
			});
        };

		var setSessionOwner = function (cmd, login, uid) {
			var param = { Sid: $form.find('#sid').val(),
                Uid: parseInt($form.find('#vid').val()),
                Cmd: cmd,
                Data: uid.toString()
			};

			$.ajax({
				type: "POST",
				url: "/set-session-owner",
				data: JSON.stringify(param),
				dataType: "json",
				contentType: "application/json; charset=utf-8",
				traditional: true,
				success: function (data) {
					if(data.ErrNo != 0) {
						$('.notify', $wrap).append('<div class="alert alert-danger">' +
                            '<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' +
                            'Unable to assign tunnel server session for ' + login + '</div>');
					} else {
                        var e = JSON.parse(data.Data);

						$('.notify', $wrap).append('<div class="alert alert-success">' +
                            '<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' +
                            login + ' assigned tunnel session ID ' + e.Sid + '</div>');

						if (cmd == 'assign-session')
							getServerSvidList('Assigned Tunnel Sessions', 'assigned-sessions', '1', '10');
						else if (cmd == 'reassign-session')
							getServerSvidList('Unassigned Tunnel Sessions', 'unassigned-sessions', '1', '10');
					}
				}
			});
		};

        var getParameter = function (id) {
            var url = window.location.search.substring(1);
            var q = url.split('&');

            for (i = 0; i < q.length; i++) {
                arg = q[i].split('=');

                if (arg[0] == id) {
                    return arg[1];
                }
            }
        };

        var setMenuList = function (url) {
            // user list
            $('a#all-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=all&page=1&cnt=25&order=rdate-r";
            });

            $('a#new-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=new&page=1&cnt=25&order=rdate-r";
            });

            $('a#admin-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=admin&page=1&cnt=25&order=rdate-r";
            });

            $('a#enabled-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=enabled&page=1&cnt=25&order=rdate-r";
            });

            $('a#disabled-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=disabled&page=1&cnt=25&order=rdate-r";
            });

            $('a#active-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=active&page=1&cnt=25&order=rdate-r";
            });

            $('a#inactive-users', 'li.user-menu').click(function () {
                url.href = "/list?uid=0&list=inactive&page=1&cnt=25&order=rdate-r";
            });

            // server list
            $('a#all-servers', 'li.server-menu').click(function () {
                url.href = "/list?vid=0&list=all&page=1&cnt=25&order=rdate-r";
            });

            $('a#enabled-servers', 'li.server-menu').click(function () {
                url.href = "/list?vid=0&list=enabled&page=1&cnt=25&order=rdate-r";
            });

            $('a#disabled-servers', 'li.server-menu').click(function () {
                url.href = "/list?vid=0&list=disabled&page=1&cnt=25&order=rdate-r";
            });

            $('a#active-servers', 'li.server-menu').click(function () {
                url.href = "/list?vid=0&list=active&page=1&cnt=25&order=rdate-r";
            });

            $('a#inactive-servers', 'li.server-menu').click(function () {
                url.href = "/list?vid=0&list=inactive&page=1&cnt=25&order=rdate-r";
            });
        };

        var url = window.location.pathname.split('/');

        switch (url[1]) {
        case '':
            setMenuList(window.location);
            break;

        case 'profile':
            setMenuList(window.location);
            break;

        case 'home':
            setMenuList(window.location);
            getUserList('New Users', 'new', '1', '10', 'rdate-r');
            break;

        case 'list':
            setMenuList(window.location);

            var uid = getParameter('uid');
            var vid = getParameter('vid');

            if (uid) {
                if (uid != 0) {
                    getUserSessionList();

                    $('.has-trigger').on('click', '.trigger > span', function () {
                        setTunnelSession($(this));
                    });

                    $('#ghazal-form').submit(function (e) {
                        e.preventDefault();
                    });

                    $('input#name', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeUserAttr('name');
                            e.preventDefault();
                        }
                    });

                    $('a#tunnel-sessions', 'li').click(function () {
                        getUserSessionList();
                    });

                    $('a#login-hist', 'li').click(function () {
                        getUserUidList('Login History', 'login', '1', '10');
                    });

                    $('a#activity-hist', 'li').click(function () {
                        getUserUidList('Activity History', 'activity', '1', '10');
                    });

                    $('a#admin', 'li').click(function () {
                        changeUserAdminStatus(this)
                    });

                    $('a#status', 'li').click(function () {
                        changeUserStatus(this);
                    });

                    $('a#reset-pw', 'li').click(function () {
                        resetUserPw();
                    });
                } else {
                    var list = getParameter('list');
                    var title;

                    if (list == 'all') {
                        title = 'All Users'
                    } else if (list == 'new') {
                        title = 'New Users'
                    } else if (list == 'admin') {
                        title = 'Admin Users'
                    } else if (list == 'enabled') {
                        title = 'Enabled Users'
                    } else if (list == 'disabled') {
                        title = 'Disabled Users'
                    } else if (list == 'active') {
                        title = 'Active Users'
                    } else if (list == 'inactive') {
                        title = 'Inactive Users'
                    }

                    var page = getParameter('page');
                    var count = getParameter('cnt');
                    var order = getParameter('order');

                    getUserList(title, list, page, count, order);
                }
            }

            if (vid) {
                if (vid != 0) {
                    getServerSvidList('Assigned Tunnel Sessions', 'assigned-sessions', '1', '10');

                    $('#rebana-form').submit(function (e) {
                        e.preventDefault();
                    });

                    $('input#alias', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('alias');
                            e.preventDefault();
                        }
                    });

                    $('input#descr', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('descr');
                            e.preventDefault();
                        }
                    });

                    $('input#location', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('location');
                            e.preventDefault();
                        }
                    });

                    $('input#access', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('access');
                            e.preventDefault();
                        }
                    });

                    $('input#tunnel', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('tunnel');
                            e.preventDefault();
                        }
                    });

                    $('input#url', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('url');
                            e.preventDefault();
                        }
                    });

                    $('input#tun-src', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('tun-src');
                            e.preventDefault();
                        }
                    });

                    $('input#pppfx', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeServerAttr('pppfx');
                            e.preventDefault();
                        }
                    });

                    $('input#rtpfx', $form).keypress(function (e) {
                        if (e.which == 13) {
                            changeSeverAttr('rtpfx');
                            e.preventDefault();
                        }
                    });

                    $('a#all-sessions', 'li').click(function () {
                        getServerSvidList('All Tunnel Sessions', 'all-sessions', '1', '10');
                    });

                    $('a#assigned-sessions', 'li').click(function () {
                        getServerSvidList('Assigned Tunnel Sessions', 'assigned-sessions', '1', '10');
                    });

                    $('a#unassigned-sessions', 'li').click(function () {
                        getServerSvidList('Unassigned Tunnel Sessions', 'unassigned-sessions', '1', '10');
                    });

                    $('a#active-sessions', 'li').click(function () {
                        getServerSvidList('Active Tunnel Sessions', 'active-sessions', '1', '10');
                    });

                    $('a#session-activity', 'li').click(function () {
                        getServerSvidList('Session Activity Logs', 'session-activity', '1', '10');
                    });

                    $('a#all-users', 'li').click(function () {
                        getServerSvidList('Tunnel Session Users', 'all-users', '1', '10');
                    });

                    $('a#assign-session', 'li').click(function () {
                        resolveUserLogin('assign-session');
                    });

                    $('a#reassign-sessions', 'li').click(function () {
                        resolveUserLogin('reassign-session');
                    });

                    $('a#admin', 'li').click(function () {
                        changeServerAdminStatus(this)
                    });

                    $('a#status', 'li').click(function () {
                        changeServerStatus(this);
                    });
                } else {
                    var list = getParameter('list');
                    var title;

                    if (list == 'all') {
                        title = 'All Tunnel Servers'
                    } else if (list == 'enabled') {
                        title = 'Enabled Tunnel Servers'
                    } else if (list == 'disabled') {
                        title = 'Disabled Tunnel Servers'
                    } else if (list == 'active') {
                        title = 'Active Tunnel Servers'
                    } else if (list == 'inactive') {
                        title = 'Inactive Tunnel Servers'
                    }

                    var page = getParameter('page');
                    var count = getParameter('cnt');
                    var order = getParameter('order');

                    getServerList(title, list, page, count, order);
                }
            }
        }

		var scrollTable = function(cpage, npage, TotalRecords, domScope, callback) {
			var ViewedRecords = cpage * npage,
				lastPage      = Math.ceil(TotalRecords / npage),
				TotalViewRecords;
			if(ViewedRecords > TotalRecords)
				TotalViewRecords = TotalRecords;
			else
				TotalViewRecords = ViewedRecords;
			$('h3', domScope).append('<span class="label label-info blue"> [' + TotalViewRecords + '/' + TotalRecords + ']</span>');
			if (TotalRecords > npage) {
				domScope.prepend('<div class="pull-right page-cont"></div>');
				var container = $('.page-cont', domScope);
				container.append('<ul class="pagination pagination-sm"><li><a class="prev"><span class="glyphicon glyphicon-chevron-left"></span> Prev</a></li><li><a class="next">Next <span class="glyphicon glyphicon-chevron-right"></span></a></li></ul>');
				var pager = $('ul', container);
				// If we're not on the first page, enable the "Previous" link.
				if (cpage != 1) {
					$('.prev').parent().removeClass('disabled');
					$('.prev').attr('href', '#');
					$('ul', container).on('click', '.prev', function(e) {
						// Prevent the browser from navigating needlessly to #.
						e.preventDefault();
						// Load and render the next page of results, and
						// increment the current page number.
						callback(--cpage);
					});
				} else {
					$('.prev').parent().addClass('disabled');
				}

				// If we're not on the last page, enable the "Next" link.
				if (cpage != lastPage) {
					$('.next').attr('href', '#');
					$('ul', container).on('click', '.next', function(e) {
						e.preventDefault();
						callback(++cpage);
					});
					$('.next').parent().removeClass('disabled');
				} else
					$('.next').parent().addClass('disabled');
			}
		};
    })
}(window.jQuery)
