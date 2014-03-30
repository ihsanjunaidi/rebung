/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
        "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type RebanaRequest struct {
	UserId  int64
	Command string
	Data    string
}

type RebanaMsg struct {
	HostName string
	UserId   int64
	MsgId    int64
	ErrNo    int
	Data     string
}

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

	AdminFlag  bool
	StatusFlag bool
	RegDate    int64
	Idx        int64
	ErrNo      int
}

type ServerInfoList struct {
	Id    int64
	Entry []ServerInfo
}

type SessionInfo struct {
	Id          int64
	Uid         string
	Type        string
	Status      string
	ServerName  string
	TunSrc      string
	TunDst      string
	Src         string
	Dst         string
	Rt          string
	Idx         string
	LastActionT string

	StatusFlag bool
	Sid        int64
	ErrNo      int
}

type SessionInfoList struct {
	Id    int64
	Entry []SessionInfo
}

type UserSessionInfo struct {
	Id         int64
	ServerId   string
	Type       string
	Status     string
	ServerName string
	TunSrc     string
	TunDst     string
	Src        string
	Dst        string
	Rt         string

	StatusFlag bool
	Sid        int64
	ErrNo      int
}

type UserSessionInfoList struct {
	Id    int64
	Entry []UserSessionInfo
}

type SessionActivityL struct {
	Action string
	Uid    string
	Sid    string
	Time   string
}

func wsSetServerAttr(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	if d.Cmd == "alias" || d.Cmd == "descr" || d.Cmd == "entity" ||
		d.Cmd == "location" || d.Cmd == "access" || d.Cmd == "tunnel" ||
		d.Cmd == "tunsrc" || d.Cmd == "ppprefix" || d.Cmd == "rtprefix" {
	} else {
		sendWSResponse(w, EINVAL, "Invalid user attribute")
		return
	}

	var nl *NameList

	if nl, err = setServerAttr(s.UserId, d.Uid, []string{d.Cmd},
		[]string{d.Data}); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Name  string
		Value string
	}{Name: nl.Entry[0].Name, Value: nl.Entry[0].Opt}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsSetServerStatus(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	if d.Cmd == "enable-server" || d.Cmd == "disable-server" ||
		d.Cmd == "activate-server" || d.Cmd == "deactivate-server" {
	} else {
		sendWSResponse(w, EINVAL, "Invalid command")
		return
	}

	var idl *IdList

	if idl, err = setServerStatus(s.UserId, d.Uid, d.Cmd); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Id int64
	}{Id: idl.Entry[0].Id}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsSetSessionOwner(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest

	if _, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	if d.Cmd == "assign-session" || d.Cmd == "reassign-session" {
	} else {
		sendWSResponse(w, EINVAL, "Invalid command")
		return
	}

        uid, _ := strconv.ParseInt(d.Data, 0, 64)

        if uid == 0 {
		sendWSResponse(w, EINVAL, "Invalid user ID")
		return
        }

	var idl *IdList

	if idl, err = setSessionOwner(uid, d.Uid, d.Cmd); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Vid int64
		Sid int64
	}{Vid: idl.Id, Sid: idl.Entry[0].Id}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsSetUserSession(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest

	if _, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	vid := d.Uid

	if vid == 0 {
		sendWSResponse(w, EINVAL, "Invalid server ID")
		return
	}

	var uid int64
	var ip string

	if d.Cmd == "activate-session" || d.Cmd == "deactivate-session" ||
		d.Cmd == "check-session" {
		var args = strings.Split(d.Data, ":")

		if len(args) != 2 {
			sendWSResponse(w, EINVAL, "Invalid request data")
			return
		}

		if ip = args[0]; checkIPFamily(ip) != 4 {
			sendWSResponse(w, EINVAL, "Invalid IP address format")
			return
		}

		if uid, err = strconv.ParseInt(args[1], 0, 64); err != nil {
			sendWSResponse(w, EINVAL, "Invalid user ID")
			return
		}
	} else {
		sendWSResponse(w, EINVAL, "Invalid command")
		return
	}

	var idl *IdList

	if idl, err = setUserSession(uid, vid, d.Cmd, ip); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Vid int64
		Sid int64
		IP  string
	}{Vid: idl.Id, Sid: idl.Entry[0].Id, IP: idl.Entry[0].Opt}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsListServer(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, true); err != nil {
		return
	}

	var list, page, cnt, order, data string

	if d.Uid == 0 {
		args := strings.Split(d.Data, ":")

		if c := len(args); c != 4 {
			sendWSResponse(w, EINVAL, "Invalid list parameter")
			return
		}

		// just verify list name
		if list = args[0]; list == "all" || list == "enabled" ||
			list == "disabled" || list == "active" ||
			list == "inactive" {
			if page = args[1]; page == "" {
				sendWSResponse(w, EINVAL, "Invalid list parameter")
				return
			}

			if cnt = args[2]; cnt == "" {
				sendWSResponse(w, EINVAL, "Invalid list parameter")
				return
			}

			if order = args[3]; order == "" {
				sendWSResponse(w, EINVAL, "Invalid list parameter")
				return
			}

			data = fmt.Sprintf("%v:%v:%v:%v", list, page, cnt, order)
		} else {
			sendWSResponse(w, EINVAL, "Invalid list name")
			return
		}
	}

	var vil *ServerInfoList

	if vil, err = listServer(s.UserId, []int64{d.Uid}, data); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Total int64
		Entry []ServerInfo
	}{Total: vil.Id, Entry: vil.Entry}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsGetServerList(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	var list, page, cnt, data string

	args := strings.Split(d.Data, ":")

	if c := len(args); c != 4 {
		sendWSResponse(w, EINVAL, "Invalid list parameter")
		return
	}

	if list = args[0]; list == "all-users" || list == "all-sessions" ||
		list == "active-sessions" || list == "assigned-sessions" ||
		list == "unassigned-sessions" || list == "session-activity" {
		if page = args[1]; page == "" {
			sendWSResponse(w, EINVAL, "Invalid list parameter")
			return
		}

		if cnt = args[2]; cnt == "" {
			sendWSResponse(w, EINVAL, "Invalid list parameter")
			return
		}

		data = fmt.Sprintf("%v:%v:%v:", list, page, cnt)
	} else {
		sendWSResponse(w, EINVAL, "Invalid list name")
		return
	}

	var buf []byte

	if list == "all-users" {
		var nl *NameList

		if nl, err = getServerNameList(s.UserId, d.Uid, data); err != nil {
			sendWSResponse(w, EINVAL, "Invalid server response")
			return
		}

		ids := make([]int64, len(nl.Entry))

		for i := range nl.Entry {
			if ids[i], err = strconv.ParseInt(nl.Entry[i].Name, 0,
				64); err != nil {
				sendWSResponse(w, EINVAL, "Invalid result payload")
			}
		}

		var uil *UserInfoList

		if uil, err = listUser(s.UserId, ids, ""); err != nil {
                        sendWSResponse(w, EINVAL, err.Error())
			return
		}

		msg := struct {
			Total int64
			Entry []UserInfo
		}{Total: uil.Id, Entry: uil.Entry}

		buf, _ = json.Marshal(msg)
	} else if list == "session-activity" {
		var nl *NameList

		if nl, err = getServerNameList(s.UserId, d.Uid,
			d.Data); err != nil {
                        sendWSResponse(w, EINVAL, err.Error())
			return
		}

		sl := make([]SessionActivityL, len(nl.Entry))

		for i := range nl.Entry {
			st := strings.Split(nl.Entry[i].Name, ";")
			sl[i] = SessionActivityL{Action: st[0], Uid: st[1],
				Sid: st[2], Time: st[3]}
		}

                msg := struct {
			Total int64
			Entry []SessionActivityL
		}{Total: nl.Id, Entry: sl}

		buf, _ = json.Marshal(msg)
	} else {
		var sil *SessionInfoList

		if sil, err = getServerSessionList(s.UserId, d.Uid,
			d.Data); err != nil {
                        sendWSResponse(w, EINVAL, err.Error())
			return
		}

                msg := struct {
			Total int64
			Entry []SessionInfo
		}{Total: sil.Id, Entry: sil.Entry}

		buf, _ = json.Marshal(msg)
	}

	sendWSResponse(w, EOK, string(buf))
	return
}

func resolveServerName(uid int64, s string) (id int64, err error) {
	url := app.RebanaUrl + "/v/resolve"
	cmd := "resolve-server"

	list := []Name{Name{Name: s}}
	buf, _ := json.Marshal(&NameList{Entry: list})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	var idl *IdList

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	if idl.Entry[0].ErrNo != EOK {
		return 0, errors.New("Server " + s + " not found")
	} else {
		id = idl.Entry[0].Id
	}

	return
}

func addServer(uid int64, h, pp, rt string) (idl *IdList, err error) {
	url := app.RebanaUrl + "/v/add"
	cmd := "add-server"

	e := make([]Name, 3)

	e[0] = Name{Name: "name", Opt: h}
	e[1] = Name{Name: "ppprefix", Opt: pp}
	e[2] = Name{Name: "rtprefix", Opt: rt}

	buf, _ := json.Marshal(&NameList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func setServerAttr(id, vid int64, k, v []string) (nl *NameList, err error) {
	url := app.RebanaUrl + "/v/set"
	cmd := "set-server-attr"

	e := make([]Name, len(k))

	for i := range k {
		e[i] = Name{Name: k[i], Opt: v[i]}
	}

	buf, _ := json.Marshal(&NameList{Id: vid, Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if nl, err = getNameList(res.Data, cmd); err != nil {
		return
	}

	return
}

func setServerStatus(uid, vid int64, cmd string) (idl *IdList, err error) {
	url := app.RebanaUrl + "/v/set"

	e := []Id{Id{Id: vid}}
	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func setSessionOwner(uid, vid int64, cmd string) (idl *IdList, err error) {
	url := app.RebanaUrl + "/s/assign"

	e := []Id{Id{Id: vid}}
	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func setUserSession(uid, vid int64, cmd, ip string) (idl *IdList, err error) {
	url := app.RebanaUrl + "/s/set"

	e := []Id{Id{Id: vid, Opt: ip}}
	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func listServer(uid int64, ids []int64, opt string) (uil *ServerInfoList,
	err error) {
	url := app.RebanaUrl + "/v/list"
	cmd := "list-server"

	e := make([]Id, len(ids))

	for i := range ids {
		e[i] = Id{Id: ids[i], Opt: opt}
	}

	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if uil, err = getServerInfoList(res.Data, cmd); err != nil {
		return
	}

	return
}

func getServerSessionList(uid, vid int64, opt string) (sil *SessionInfoList,
	err error) {
	url := app.RebanaUrl + "/v/list"
	cmd := "get-server-list"

	e := []Id{Id{Id: vid, Opt: opt}}

	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if sil, err = getSessionInfoList(res.Data, cmd); err != nil {
		return
	}

	return
}

func getServerNameList(uid, vid int64, opt string) (nl *NameList, err error) {
	url := app.RebanaUrl + "/v/list"
	cmd := "get-server-list"

	e := []Id{Id{Id: vid, Opt: opt}}

	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: uid, Cmd: cmd, Data: data, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if nl, err = getNameList(res.Data, cmd); err != nil {
		return
	}

	return
}

func listUserSession(id int64) (sil *UserSessionInfoList, err error) {
	url := app.RebanaUrl + "/s/list"
	cmd := "list-user-sessions"

	req := &RequestOpt{Uid: id, Cmd: cmd, Url: url}

	var res *RebanaMsg

	if res, err = sendRebanaRequest(req); err != nil {
		return
	}

	if sil, err = getUserSessionInfoList(res.Data, cmd); err != nil {
		return
	}

	return
}
