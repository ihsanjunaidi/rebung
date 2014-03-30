/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type UserInfo struct {
	Id         string
	Name       string
	Login      string
	Password   string
	Admin      string
	Status     string
	Registered string
	FirstLogin string

	Idx   int64
	ErrNo int
}

type UserInfoList struct {
	Id    int
	Entry []UserInfo
}

func resolveUserLogin() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Name

	if list, err = setNameParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&NameList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Tunnel user %v has ID %v", e.Opt, e.Id)
		} else {
			event("Tunnel user %v not found", e.Id)
		}
	}

	return
}
func resolveUserId(args []string) (s []string, err error) {
	var url = GHAZALBASEURL + "s/resolve"

	app.Cmd.Command = "resolve-user-id"
	app.Cmd.UserId = 103

	var list []Id

	if list, err = setIdParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), url); err != nil {
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

func resolveUserIds() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Id

	if list, err = setIdParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Tunnel user [%v] name is %v", e.Id, e.Opt)
		} else {
			event("Tunnel user [%v] not found", e.Id)
		}
	}

	return
}

func addUser() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Name

	if list, err = setNameArgParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&NameList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("User %v created: ID[%v]", e.Opt, e.Id)
		} else {
			event("User %v cannot be created", e.Opt)
		}
	}

	return
}

func setUserAttr() (err error) {
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

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *NameList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	var id = fmt.Sprintf("%v", m.Id)

	var n []string

	if n, err = resolveUserId([]string{id}); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("User %v has new %v: %v", n[i], e.Name, e.Opt)
		} else {
			event("Parameter %v cannot be changed", e.Name)
		}
	}

	return
}

func setUserStatus() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ",")

	var list []Id

	if list, err = setIdParam(args); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
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

	if app.Cmd.Command == "enable-user" {
		v = "enabled"
	} else if app.Cmd.Command == "disable-user" {
		v = "disabled"
	} else if app.Cmd.Command == "activate-user" {
		v = "activated"
	} else if app.Cmd.Command == "deactivate-user" {
		v = "deactivated"
	}

	var n []string

	if n, err = resolveUserId(a); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("User %v is now %v", n[i], v)
		} else {
			event("User %v cannot be %v", n[i], v)
		}
	}

	return
}

func listUser() (err error) {
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

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *UserInfoList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			var tr, _ = time.Parse(time.RFC1123, e.Registered)
			var tf, _ = time.Parse(time.RFC1123, e.FirstLogin)
			var trs = tr.Format(time.RFC1123)
			var tfs = tf.Format(time.RFC1123)

			event("User ID [%v] information:\n"+
				"------------------------------\n"+
				"Name: %v\n"+
				"Login: %v\n"+
				"Admin status: %v\n"+
				"Status: %v\n"+
				"Registration date: %v\n"+
				"First login: %v\n", e.Id, e.Name,
				e.Login, e.Admin, e.Status, trs, tfs)
		} else {
			event("User ID [%v] not found\n"+
				"---------------------------\n", e.Id)
		}
	}

	return
}

func getUserList() (err error) {
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

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *NameList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("%v", e)
		}
	}

	return
}

func userLogin() (err error) {
	var admin bool = false

	if app.Cmd.Command == "admin-login" {
		admin = true
	}

	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

	var args = strings.Split(app.Cmd.Args[0], ":")

	var h = sha256.New()

	io.WriteString(h, args[1])

	var dgst = base64.StdEncoding.EncodeToString(h.Sum(nil))

	var list = make([]Name, 2)

	list[0] = Name{Name: "login", Opt: args[0]}
	list[1] = Name{Name: "password", Opt: dgst}

	if admin {
		list[0].ErrNo = 1
	}

	var d, _ = json.Marshal(&NameList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), m); err != nil {
		return
	}

	var uid = fmt.Sprintf("%v", m.Id)

	var n []string

	if n, err = resolveUserId([]string{uid}); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("%v has been assigned session key %v", n[i], e.Opt)
		} else {
			event("User %v not found", n[i])
		}
	}

	return
}

func userLogout() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Incorrect number of arguments")
	}

        var list = []Name{Name{Name: app.Cmd.Args[0]}}

	var d, _ = json.Marshal(&NameList{Entry: list})

	var msg *GhazalMsg

	if msg, err = sendGhazalRequest(string(d), app.GhazalUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), m); err != nil {
		return
	}

	var uid = fmt.Sprintf("%v", m.Id)

	var n []string

	if n, err = resolveUserId([]string{uid}); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("%v has been logged out", n[i])
		} else {
			event("User %v cannot be logged out", n[i])
		}
	}

	return
}
