/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

func userLogin(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqUserLogin++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrUserLogin++
		return
	}

	var login, pw string
	var admin bool

	si := make([]Id, 1)
	c := 2

	for i := range m.Entry {
		e := m.Entry[i]

		if e.Name == "login" {
			login = e.Opt

			if e.ErrNo == 1 {
				admin = true
			}

			c--
		} else if e.Name == "password" {
			pw = e.Opt
			c--
		}
	}

	if c != 0 {
		stat.ReqErrUserLogin++
		return errors.New("Insufficient user login parameters")
	}

	var uid int64

	if uid, err = checkUserSession(login); err != nil {
		stat.ReqErrUserLogin++
		return
	}

	if err = checkRedisUserStatus(uid); err != nil {
		stat.ReqErrUserLogin++
		return
	}

	if admin {
		if err = checkUserAdmin(uid); err != nil {
			stat.ReqErrUserLogin++
			return
		}
	}

	var tok string

	if tok, err = setUserSession(uid, login, pw, d.Origin); err != nil {
		stat.ReqErrUserLogin++
		return
	}

	si[0] = Id{Id: uid, Opt: tok}

	buf, _ := json.Marshal(&IdList{Id: uid, Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func userLogout(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqUserLogout++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrUserLogout++
		return
	}

	var si []Id
	var tok string

	if tok, err = getRedisUserSessionKey(m.Id); err != nil {
		stat.ReqErrUserLogout++
		return
	} else {
		e := m.Entry[0]

		if tok != e.Name {
			stat.ReqErrUserLogout++
			event(logwarn, li, "Mismatched user [%v] session key", m.Id)
		}

		if err = deleteRedisUserSession(m.Id, d.Origin); err != nil {
			stat.ReqErrUserLogout++
			return
		} else {
			si = []Id{Id{Id: m.Id}}
		}

	}

	data := &IdList{Id: m.Id, Entry: si}
	buf, _ := json.Marshal(data)
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func defaultUserHandler(w http.ResponseWriter, r *http.Request) {
	stat.ReqAll++
	li.Msgid = 0

	var err error

	str := "Invalid request"

	if err = checkUrl(r); err != nil {
		stat.ReqErrUrl++
		sendError(w, EINVAL, str, err)
		return
	}

	if err = checkHeader(r); err != nil {
		stat.ReqErrHeader++
		sendError(w, EINVAL, str, err)
		return
	}

	if err = checkRedis(); err != nil {
		stat.ReqErrRedis++
		sendError(w, EINVAL, str, err)
		return
	}

	var d *RequestMsg

	if d, err = checkData(r); err != nil {
		stat.ReqErrPayload++
		sendError(w, EINVAL, str, err)
		return
	}

	if li.Msgid, err = getRedisMsgId(d.Command); err != nil {
		stat.ReqErrMsgId++
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Processing request [%v:%v]", d.Command, li.Msgid)

	switch d.Command {
	case "register":
		err = addUser(w, d)

	case "login":
		err = userLogin(w, d)

	case "logout":
		err = userLogout(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Request [%v:%v] completed", d.Command, li.Msgid)
}
