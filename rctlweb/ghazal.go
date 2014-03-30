/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type GhazalRequest struct {
	UserId  int64
	Origin  string
	Command string
	Data    string
}

type GhazalMsg struct {
	HostName string
	UserId   int64
	MsgId    int64
	ErrNo    int
	Data     string
}

type UserInfo struct {
	Id         int64
	Name       string
	Login      string
	Password   string
	Admin      string
	Status     string
	Registered string
	FirstLogin string

	AdminFlag      bool
	StatusFlag     bool
	FirstLoginFlag bool
	Idx            int64
	ErrNo          int
}

type UserInfoList struct {
	Id    int64
	Entry []UserInfo
}

type UserUidList struct {
	Action string
	IP     string
	Time   string
}

func wsResolveUser(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	var uid int64

	if uid, err = resolveUserLogin(s.UserId, d.Data); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Id int64
	}{Id: uid}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsSetUserAttr(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	if d.Cmd == "name" {
	} else {
		sendWSResponse(w, EINVAL, "Invalid user attribute")
		return
	}

	var nl *NameList

	if nl, err = setUserAttr(s.UserId, d.Uid, []string{d.Cmd},
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

func wsSetUserStatus(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	if d.Cmd == "enable-user" || d.Cmd == "disable-user" ||
		d.Cmd == "activate-user" || d.Cmd == "deactivate-user" {
	} else {
		sendWSResponse(w, EINVAL, "Invalid command")
		return
	}

	var idl *IdList

	if idl, err = setUserStatus(s.UserId, d.Uid, d.Cmd); err != nil {
		sendWSResponse(w, EINVAL, "Invalid server response")
		return
	}

	msg := struct {
		Id int64
	}{Id: idl.Entry[0].Id}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsResetUserPw(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest
	var s *Session

	if s, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	if _, err = resetUserPw(s.UserId, d.Uid); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	sendWSResponse(w, EOK, "")
	return
}

func wsListUser(w http.ResponseWriter, r *http.Request) (err error) {
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

		if list = args[0]; list == "all" || list == "enabled" ||
			list == "disabled" || list == "active" ||
			list == "inactive" || list == "new" || list == "admin" {
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

	var uil *UserInfoList

	if uil, err = listUser(s.UserId, []int64{d.Uid}, data); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Total int64
		Entry []UserInfo
	}{Total: uil.Id, Entry: uil.Entry}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsGetUserList(w http.ResponseWriter, r *http.Request) (err error) {
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

	if list = args[0]; list == "login" || list == "activity" {
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

	var nl *NameList

	if nl, err = getUserList(s.UserId, d.Uid, data); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	ul := make([]UserUidList, len(nl.Entry))

	for i := range nl.Entry {
		st := strings.Split(nl.Entry[i].Name, ";")
		ul[i] = UserUidList{Action: st[0], IP: st[1], Time: st[2]}
	}

	msg := struct {
		Total int64
		Entry []UserUidList
	}{Total: nl.Id, Entry: ul}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func wsGetUserSessions(w http.ResponseWriter, r *http.Request) (err error) {
	var d *WSRequest

	if _, d, err = wsCheck(w, r, false); err != nil {
		return
	}

	var usl *UserSessionInfoList

	if usl, err = listUserSession(d.Uid); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	msg := struct {
		Total int64
		Entry []UserSessionInfo
	}{Total: usl.Id, Entry: usl.Entry}

	buf, _ := json.Marshal(msg)
	sendWSResponse(w, EOK, string(buf))
	return
}

func resolveUserLogin(id int64, s string) (uid int64, err error) {
	url := app.GhazalUrl + "/s/resolve"
	cmd := "resolve-user"

	e := []Name{Name{Name: s}}

	buf, _ := json.Marshal(&NameList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	var idl *IdList

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	if idl.Entry[0].ErrNo != EOK {
		return 0, errors.New("User " + s + " not found")
	} else {
		uid = idl.Entry[0].Id
	}

	return
}

func addUser(id int64, l, n, ip string) (idl *IdList, err error) {
	url := app.GhazalUrl + "/s/add"
	cmd := "add-user"

	e := make([]Name, 3)

	e[0] = Name{Name: "login", ErrNo: 1, Opt: l}
	e[1] = Name{Name: "name", Opt: n}
	e[2] = Name{Name: "ip", Opt: ip}

	buf, _ := json.Marshal(&NameList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func setUserAttr(id, uid int64, k, v []string) (nl *NameList, err error) {
	url := app.GhazalUrl + "/s/set"
	cmd := "set-user-attr"

	e := make([]Name, len(k))

	for i := range k {
		e[i] = Name{Name: k[i], Opt: v[i]}
	}

	buf, _ := json.Marshal(&NameList{Id: uid, Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if nl, err = getNameList(res.Data, cmd); err != nil {
		return
	}

	return
}

func setUserStatus(id, uid int64, cmd string) (idl *IdList, err error) {
	url := app.GhazalUrl + "/s/set"

	e := []Id{Id{Id: uid}}
	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func resetUserPw(id, uid int64) (idl *IdList, err error) {
	url := app.GhazalUrl + "/s/reset"
	cmd := "reset-user-pw"

	e := []Id{Id{Id: uid}}
	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func listUser(id int64, ids []int64, opt string) (uil *UserInfoList, err error) {
	url := app.GhazalUrl + "/s/list"
	cmd := "list-user"

	e := make([]Id, len(ids))

	for i := range ids {
		e[i] = Id{Id: ids[i], Opt: opt}
	}

	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if uil, err = getUserInfoList(res.Data, cmd); err != nil {
		return
	}

	return
}

func getUserList(id, uid int64, opt string) (nl *NameList, err error) {
	url := app.GhazalUrl + "/s/list"
	cmd := "get-user-list"

	e := []Id{Id{Id: uid, Opt: opt}}

	buf, _ := json.Marshal(&IdList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if nl, err = getNameList(res.Data, cmd); err != nil {
		return
	}

	return
}

func userLogin(login, pw string) (idl *IdList, err error) {
	url := app.GhazalUrl + "/u/login"
	cmd := "login"

	e := make([]Name, 2)

	e[0] = Name{Name: "login", Opt: login}
	e[1] = Name{Name: "password", Opt: pw}

	buf, _ := json.Marshal(&NameList{Entry: e})

	data := string(buf)
	req := &RequestOpt{Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}

func userLogout(id int64, key, ip string) (idl *IdList, err error) {
	url := app.GhazalUrl + "/u/logout"
	cmd := "logout"

	e := []Name{Name{Name: key}}

	buf, _ := json.Marshal(&NameList{Id: id, Entry: e})

	data := string(buf)
	req := &RequestOpt{Uid: id, Cmd: cmd, Data: data, Url: url}

	var res *GhazalMsg

	if res, err = sendGhazalRequest(req); err != nil {
		return
	}

	if idl, err = getIdList(res.Data, cmd); err != nil {
		return
	}

	return
}
