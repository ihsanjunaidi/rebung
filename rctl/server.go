/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ServerInfo struct {
	Id        int64
	Name      string
	Alias     string
	Descr     string
	Entity    string
	Location  string
	Access    string
	Tunnel    string
	TunnelSrc string
	Url       string
	PpPrefix  string
	RtPrefix  string
	Admin     string
	Status    string
	Activated string

	Idx   int64
	ErrNo int
}

type ServerInfoList struct {
	Idx   int64
	Entry []ServerInfo
}

type RebanaTSInfo struct {
	Id       int
	TunSrc   string
	PpPrefix string
	RtPrefix string
	Session  []RebanaTSSession
}

type RebanaTSSession struct {
	Id   int
	Type string
	Dst  string
	Idx  int
}

type RebanaTSStat struct {
	HostName          string
	ReqAll            int64
	ReqActivate       int64
	ReqDeactivate     int64
	ReqCheck          int64
	ReqStatus         int64
	ReqError          int64
	ReqErrUrl         int64
	ReqErrHeader      int64
	ReqErrPayload     int64
	ReqErrSignature   int64
	ReqErrServerId    int64
	ReqErrUserId      int64
	ReqErrRebanaMsgId int64
	ReqErrCommand     int64
	ReqErrData        int64
	ReqErrActivate    int64
	ReqErrDeactivate  int64
	ReqErrCheck       int64
	ReqErrStatus      int64
}

type RebanaStat struct {
	HostName                    string
	ReqAll                      int64
	ReqActivateSession          int64
	ReqDeactivateSession        int64
	ReqCheckSession             int64
	ReqAssignSession            int64
	ReqReassignSession          int64
	ReqListSession              int64
	ReqActivateUserSession      int64
	ReqDeactivateUserSession    int64
	ReqCheckUserSession         int64
	ReqListUserSession          int64
	ReqListUserServer           int64
	ReqResolveServer            int64
	ReqResolveServerId          int64
	ReqAddServer                int64
	ReqSetServerAttr            int64
	ReqEnableServer             int64
	ReqDisableServer            int64
	ReqActivateServer           int64
	ReqDeactivateServer         int64
	ReqListServer               int64
	ReqGetServerList            int64
	ReqGetUserList              int64
	ReqServerStatus             int64
	ReqServerInfo               int64
	ReqStatus                   int64
	ReqError                    int64
	ReqErrUrl                   int64
	ReqErrHeader                int64
	ReqErrRedis                 int64
	ReqErrPayload               int64
	ReqErrSignature             int64
	ReqErrUserId                int64
	ReqErrServerId              int64
	ReqErrSessionId             int64
	ReqErrMsgId                 int64
	ReqErrCommand               int64
	ReqErrData                  int64
	ReqErrActivateSession       int64
	ReqErrDeactivateSession     int64
	ReqErrCheckSession          int64
	ReqErrAssignSession         int64
	ReqErrReassignSession       int64
	ReqErrListSession           int64
	ReqErrActivateUserSession   int64
	ReqErrDeactivateUserSession int64
	ReqErrCheckUserSession      int64
	ReqErrListUserSession       int64
	ReqErrListUserServer        int64
	ReqErrResolveServer         int64
	ReqErrResolveServerId       int64
	ReqErrAddServer             int64
	ReqErrSetServerAttr         int64
	ReqErrEnableServer          int64
	ReqErrDisableServer         int64
	ReqErrActivateServer        int64
	ReqErrDeactivateServer      int64
	ReqErrListServer            int64
	ReqErrGetServerList         int64
	ReqErrGetUserList           int64
	ReqErrServerStatus          int64
	ReqErrServerInfo            int64
	ReqErrStatus                int64
}

func resolveServerName() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Name

	if list, err = setNameParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&NameList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Tunnel server %v has ID %v", e.Opt, e.Id)
		} else {
			event("Tunnel server %v not found", e.Id)
		}
	}

	return
}

func resolveServerId(args []string) (s []string, err error) {
	var url = REBANABASEURL + "v/resolve"

	app.Cmd.Command = "resolve-server-id"
	app.Cmd.UserId = 103

	var list []Id

	if list, err = setIdParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), url); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	s = make([]string, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			s[i] = e.Opt
		} else {
			s[i] = fmt.Sprintf("[%v]", e.Opt)
		}
	}

	return
}

func resolveServerIds() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Id

	if list, err = setIdParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Tunnel server [%v] name is %v", e.Id, e.Opt)
		} else {
			event("Tunnel server [%v] not found", e.Id)
		}
	}

	return
}

func addServer() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Name

	if list, err = setNameArgParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&NameList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Server %v created: ID[%v]", e.Opt, e.Id)
		} else {
			event("Server %v cannot be created", e.Opt)
		}
	}

	return
}

func setServerAttr() (err error) {
	if len(app.Cmd.Args) != 2 {
		return errors.New("Incorrect number of arguments")
	}

	var svid, _ = strconv.ParseInt(app.Cmd.Args[0], 0, 64)
	var args = strings.Split(app.Cmd.Args[1], ",")

	var list []Name

	if list, err = setNameArgParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&NameList{Id: svid, Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *NameList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	var id = fmt.Sprintf("%v", m.Id)

	var n []string

	if n, err = resolveServerId([]string{id}); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Server %v has new %v: %v", n[i], e.Name, e.Opt)
		} else {
			event("Parameter %v cannot be changed", e.Name)
		}
	}

	return
}

func setServerStatus() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Id

	if list, err = setIdParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	var a = make([]string, 1)

	for i := range m.Entry {
		var e = m.Entry[i]

		if i == len(m.Entry)-1 {
			a[0] += fmt.Sprintf("%v", e.Id)
		} else {
			a[0] += fmt.Sprintf("%v,", e.Id)
		}

	}

	var v string

	if app.Cmd.Command == "enable-server" {
		v = "enabled"
	} else if app.Cmd.Command == "disable-server" {
		v = "disabled"
	} else if app.Cmd.Command == "activate-server" {
		v = "activated"
	} else if app.Cmd.Command == "deactivate-server" {
		v = "deactivated"
	}

	var n []string

	if n, err = resolveServerId(a); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Server %v is now %v", n[i], v)
		} else {
			event("Server %v cannot be %v", n[i], v)
		}
	}

	return
}

func listServer() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var list []Id

	var args = strings.Split(app.Cmd.Args[0], ",")

	if len(args) == 1 {
		var a = strings.Split(args[0], ":")

		if len(a) == 5 {
			var opt = fmt.Sprintf("%v:%v:%v:%v", a[1], a[2], a[3], a[4])

			list = []Id{Id{Opt: opt}}
		} else {
			var id, _ = strconv.ParseInt(args[0], 0, 64)

			list = []Id{Id{Id: id}}
		}
	} else {
		if list, err = setIdParam(args); err != nil {
			return
		}
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *ServerInfoList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	event("Total tunnel servers: %v\n", len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == ENOENT {
			event("Tunnel server [%v] not found\n"+
				"---------------------------\n", e.Idx)
		} else {
			var t, _ = time.Parse(time.RFC1123, e.Activated)
			var ts = t.Format(time.RFC1123)

			event("Tunnel server [%v] information:\n"+
				"------------------------------\n"+
				"Server name: %v\n"+
				"Server alias: %v\n"+
				"Description: %v\n"+
				"Entity: %v\n"+
				"Location: %v\n"+
				"Type: %v\n"+
				"Tunnel supported: %v\n"+
				"Tunnel source address: %v\n"+
				"Tunnel management URL: %v\n"+
				"Point-to-Point prefix: %v\n"+
				"Routed prefix: %v\n"+
				"Admin status: %v\n"+
				"Status: %v\n"+
				"Activation date: %v\n", e.Id, e.Name,
				e.Alias, e.Descr, e.Entity, e.Location, e.Access,
				e.Tunnel, e.TunnelSrc, e.Url, e.PpPrefix,
				e.RtPrefix, e.Admin, e.Status, ts)
		}
	}

	return
}

func getServerList() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var a = strings.Split(app.Cmd.Args[0], ":")

	if len(a) != 5 {
		return errors.New("Invalid argument format")
	}

	var opt = fmt.Sprintf("%v:%v:%v:%v", a[1], a[2], a[3], a[4])
	var id, _ = strconv.ParseInt(a[0], 0, 64)
	var list = []Id{Id{Id: id, Opt: opt}}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	if a[1] == "all-users" || a[1] == "session-activity" {
		var m *NameList

		if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
			return
		}

		event("Total entries: %v\n", len(m.Entry))

		for i := range m.Entry {
			var e = m.Entry[i]

			if e.ErrNo == EOK {
				event("%v", e)
			}
		}
	} else {
		var m *SessionInfoList

		if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
			return
		}

		event("Total entries: %v\n", len(m.Entry))

		for i := range m.Entry {
			var e = m.Entry[i]

			if e.ErrNo == EOK {
				event("%v", e)
			}
		}
	}

	return
}

func serverStatus() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var id, _ = strconv.ParseInt(app.Cmd.Args[0], 0, 64)
	var list = []Id{Id{Id: id}}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var rs *RebanaTSStat

	if err = json.Unmarshal([]byte(msg.Data), &rs); err != nil {
		return
	}

	event("%v tunnel server stats\n"+
		"--------------------\n"+
		"Total requests: %v\n"+
		"Session activation requests: %v\n"+
		"Session deactivation requests: %v\n"+
		"Session check requests: %v\n"+
		"Server status requests: %v\n"+
		"--------------------\n"+
		"Total errors: %v\n"+
		"URL errors: %v\n"+
		"Invalid headers: %v\n"+
		"Invalid payload: %v\n"+
		"Invalid message digest: %v\n"+
		"Invalid server ID: %v\n"+
		"Invalid user ID: %v\n"+
		"Invalid message ID: %v\n"+
		"Invalid command: %v\n"+
		"Invalid command arguments: %v\n"+
		"Session activation errors: %v\n"+
		"Session deactivation errors: %v\n"+
		"Session check errors: %v\n"+
		"Server status errors: %v", rs.HostName, rs.ReqAll,
		rs.ReqActivate, rs.ReqDeactivate, rs.ReqCheck, rs.ReqStatus,
		rs.ReqError, rs.ReqErrUrl, rs.ReqErrHeader, rs.ReqErrPayload,
		rs.ReqErrSignature, rs.ReqErrServerId, rs.ReqErrUserId,
		rs.ReqErrRebanaMsgId, rs.ReqErrCommand, rs.ReqErrData,
		rs.ReqErrActivate, rs.ReqErrDeactivate, rs.ReqErrCheck,
		rs.ReqErrStatus)

	return
}

func status() (err error) {
	if len(app.Cmd.Args) != 0 {
		return errors.New("Incorrect number of arguments")
	}

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest("", app.RebanaUrl); err != nil {
		return
	}

	var rs *RebanaStat

	if err = json.Unmarshal([]byte(msg.Data), &rs); err != nil {
		return
	}

	event("%v Rebana stats\n"+
		"--------------------\n"+
		"Total requests: %v\n"+
		"Session activation requests: %v\n"+
		"Session deactivation requests: %v\n"+
		"Session check requests: %v\n"+
		"Session assignment requests: %v\n"+
		"Session reassignment requests: %v\n"+
		"Session listing requests: %v\n"+
		"User session activation requests: %v\n"+
		"User session deactivation requests: %v\n"+
		"User session check requests: %v\n"+
		"User session listing requests: %v\n"+
		"User server listing requests: %v\n"+
		"User session listing requests: %v\n"+
		"Tunnel server name resolution requests: %v\n"+
		"Tunnel server ID resolution requests: %v\n"+
		"Tunnel server registration requests: %v\n"+
		"Tunnel server parameter change requests: %v\n"+
		"Tunnel server enable requests: %v\n"+
		"Tunnel server disable requests: %v\n"+
		"Tunnel server activation requests: %v\n"+
		"Tunnel server deactivation requests: %v\n"+
		"Tunnel server listing requests: %v\n"+
		"Tunnel server list requests: %v\n"+
		"Tunnel server status requests: %v\n"+
		"Tunnel server info requests: %v\n"+
		"Rebana status requests: %v\n"+
		"--------------------\n"+
		"Total errors: %v\n"+
		"URL errors: %v\n"+
		"Invalid headers: %v\n"+
		"Redis errors: %v\n"+
		"Invalid payload: %v\n"+
		"Invalid message digest: %v\n"+
		"Invalid user ID: %v\n"+
		"Invalid server ID: %v\n"+
		"Invalid session ID: %v\n"+
		"Invalid message ID: %v\n"+
		"Invalid command: %v\n"+
		"Invalid command arguments: %v\n"+
		"Session activation errors: %v\n"+
		"Session deactivation errors: %v\n"+
		"Session check errors: %v\n"+
		"Session assignment errors: %v\n"+
		"Session reassignment errors: %v\n"+
		"Session listing errors: %v\n"+
		"User session activation errors: %v\n"+
		"User session deactivation errors: %v\n"+
		"User session check errors: %v\n"+
		"User session listing errors: %v\n"+
		"User server listing errors: %v\n"+
		"User session listing errors: %v\n"+
		"Tunnel server name resolution errors: %v\n"+
		"Tunnel server ID resolution errors: %v\n"+
		"Tunnel server registration errors: %v\n"+
		"Tunnel server parameter change errors: %v\n"+
		"Tunnel server enable errors: %v\n"+
		"Tunnel server disable errors: %v\n"+
		"Tunnel server activation errors: %v\n"+
		"Tunnel server deactivation errors: %v\n"+
		"Tunnel server listing errors: %v\n"+
		"Tunnel server list errors: %v\n"+
		"Tunnel server status errors: %v\n"+
		"Tunnel server info errors: %v\n"+
		"Rebana status errors: %v", rs.HostName, rs.ReqAll,
		rs.ReqActivateSession, rs.ReqDeactivateSession,
		rs.ReqCheckSession, rs.ReqAssignSession, rs.ReqReassignSession,
		rs.ReqListSession, rs.ReqActivateUserSession,
		rs.ReqDeactivateUserSession, rs.ReqCheckUserSession,
		rs.ReqListUserSession, rs.ReqListUserServer, rs.ReqResolveServer,
		rs.ReqResolveServerId, rs.ReqAddServer, rs.ReqSetServerAttr,
		rs.ReqEnableServer, rs.ReqDisableServer, rs.ReqActivateServer,
		rs.ReqDeactivateServer, rs.ReqListServer, rs.ReqGetServerList,
		rs.ReqGetUserList, rs.ReqServerStatus, rs.ReqServerInfo,
		rs.ReqStatus, rs.ReqError, rs.ReqErrUrl, rs.ReqErrHeader,
		rs.ReqErrRedis, rs.ReqErrPayload, rs.ReqErrSignature,
		rs.ReqErrUserId, rs.ReqErrServerId, rs.ReqErrSessionId,
		rs.ReqErrMsgId, rs.ReqErrCommand, rs.ReqErrData,
		rs.ReqErrActivateSession, rs.ReqErrDeactivateSession,
		rs.ReqErrCheckSession, rs.ReqErrAssignSession,
		rs.ReqErrReassignSession, rs.ReqErrListSession,
		rs.ReqErrActivateUserSession, rs.ReqErrDeactivateUserSession,
		rs.ReqErrCheckUserSession, rs.ReqErrListUserSession,
		rs.ReqErrListUserServer, rs.ReqErrResolveServer,
		rs.ReqErrResolveServerId, rs.ReqErrAddServer,
		rs.ReqErrSetServerAttr, rs.ReqErrEnableServer,
		rs.ReqErrDisableServer, rs.ReqErrActivateServer,
		rs.ReqErrDeactivateServer, rs.ReqErrListServer,
		rs.ReqErrGetServerList, rs.ReqErrGetUserList,
		rs.ReqErrServerStatus, rs.ReqErrServerInfo, rs.ReqErrStatus)

	return
}
