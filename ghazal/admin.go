/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SessionInfo struct {
	Id          string
	Uid         string
	AccessToken string
	Expire      string
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

	RegDate int64
	Idx     int64
	ErrNo   int
}

type UserInfoList struct {
	Id    int64
	Entry []UserInfo
}

func resolveUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqResolveUser++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrResolveUser++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if uid, err := getRedisUserIdFromLogin(e.Name); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{ErrNo: ENOENT, Opt: e.Name}
		} else {
			si[i] = Id{Id: uid, Opt: e.Name}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func resolveUserId(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqResolveUserId++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrResolveUserId++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if s, err := getRedisUserInfo(e.Id); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
		} else {
			si[i] = Id{Id: e.Id, Opt: s.Login}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func resetUserPw(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqResetUserPw++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrResetUserPw++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if s, err := getRedisUserInfo(e.Id); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{ErrNo: ENOENT}
		} else {
			pw := generateTempPassword()
			h := sha256.New()
			io.WriteString(h, pw)

			tpw := base64.StdEncoding.EncodeToString(h.Sum(nil))

			if err = setRedisUserAttr(s.Id, "password",
				tpw, d.Origin); err != nil {
				event(logwarn, li, err.Error())
				si[i] = Id{ErrNo: EINVAL}
			} else {
				si[i] = Id{}

				var rcpt = []string{s.Login}
				var subj = "Rebung.IO User Password Reset"
				var body = fmt.Sprintf("Greetings %v,\n\n"+
					"Your Rebung.IO account password has "+
					"been reset by the administrator.\n\n"+
					"Your new password is: %v\n\n"+
					"You may change your password in your "+
					"user page.\n\n"+
					"Regards,\nRebung.IO service robot",
					s.Name, pw)

				if err = sendMail(rcpt, subj, body,
					false); err != nil {
					event(logwarn, li, err.Error())
				}
			}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func addUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqAddUser++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrAddUser++
		return
	}

	var email, name, field string

	s := &UserInfo{}
	c := 2

	for i := range m.Entry {
		e := m.Entry[i]

		if field = e.Name; field == "login" {
			if s.Login = e.Opt; s.Login == "" {
				err = errors.New(field)
				break
			}

			email = e.Opt
			c--
		} else if field = e.Name; field == "name" {
			if s.Name = e.Opt; s.Name == "" {
				err = errors.New(field)
				break
			}

			name = e.Opt
			c--
		} else {
			err = errors.New(field)
			break
		}
	}

	if err != nil {
		stat.ReqErrAddUser++
		return errors.New("Invalid user parameter: " + field)
	}

	if c != 0 {
		stat.ReqErrAddUser++
		return errors.New("Insufficient user parameters")
	}

	s.Registered = time.Now().Format(time.RFC1123)

	si := make([]Id, 1)

	var uid int64
	var passwd string

	if uid, passwd, err = setRedisUserNew(s, d.Origin); err != nil {
		stat.ReqErrAddUser++
		return
	} else {
		si[0] = Id{Id: uid, Opt: email}

		var subj, body string
		var rcpt []string

		var t = time.Now().Format(time.RFC1123)

		rcpt = []string{app.AdminEmail}
		subj = fmt.Sprintf("Rebung.IO new user registration: %v", name)
		body = fmt.Sprintf("New user registration\n\n"+
			"Registered on %v\n"+
			"Registered from [%v]\n\n"+
			"Login name: %v\n"+
			"Name: %v", t, d.Origin, email, name)

		if err = sendMail([]string{}, subj, body, true); err != nil {
			event(logwarn, li, err.Error())
		}

		rcpt = []string{email}
		subj = "Rebung.IO user registration information"
		body = fmt.Sprintf("Greetings %v,\n\n"+
			"Thank you for registering for a Rebung.IO "+
			"account.\n\n"+
			"You may now login into your account using the "+
			"following information:\n\n"+
			"Login name: %v\n"+
			"Password: %v\n\n"+
			"You may change your password in your user "+
			"page.\n\n"+
			"Be sure to read and understand our AUP.\n\n"+
			"As a reminder, you will need to login into your "+
			"account within 48 hours for activation or the "+
			"account will be disabled.\n\n"+
			"Regards,\nRebung.IO service robot", name, email, passwd)

		if err = sendMail(rcpt, subj, body, false); err != nil {
			event(logwarn, li, err.Error())
		}
	}

	buf, _ := json.Marshal(&IdList{Id: m.Id, Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func setUserAttr(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqSetUserAttr++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrSetUserAttr++
		return
	}

	if err = checkRedisUserStatus(m.Id); err != nil {
		stat.ReqErrSetUserAttr++
		return
	}

	si := make([]Name, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if e.Name == "password" {
			var pw []byte

			if pw, err = bcrypt.GenerateFromPassword([]byte(e.Opt),
				10); err != nil {
				stat.ReqErrSetUserAttr++
				return errors.New("Error generating password")
			}

			e.Opt = string(pw)
		}

		if e.Name == "name" || e.Name == "password" {
			if err = setRedisUserAttr(m.Id, e.Name,
				e.Opt, d.Origin); err != nil {
				event(logwarn, li, err.Error())
				si[i] = Name{Name: e.Name, ErrNo: ENOENT}
			} else {
				si[i] = Name{Name: e.Name, Opt: e.Opt}
			}
		} else {
			stat.ReqErrSetUserAttr++
			return errors.New("Attribute not permitted: " + e.Name)
		}
	}

	buf, _ := json.Marshal(&NameList{Id: m.Id, Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func enableUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqEnableUser++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrEnableUser++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if err = setRedisUserAdminStatus(e.Id, true,
			d.Origin); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func disableUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqDisableUser++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrDisableUser++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if err = setRedisUserAdminStatus(e.Id, false,
			d.Origin); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func activateUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqActivateUser++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrActivateUser++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if err = setRedisUserStatus(e.Id, true,
			d.Origin); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func deactivateUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqDeactivateUser++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrDeactivateUser++
		return
	}

	si := make([]Id, len(m.Entry))

	for i := range m.Entry {
		e := m.Entry[i]

		if err = setRedisUserStatus(e.Id, false,
			d.Origin); err != nil {
			event(logwarn, li, err.Error())
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	buf, _ := json.Marshal(&IdList{Entry: si})
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func listUser(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqListUser++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrListUser++
		return
	}

	var si []UserInfo
	var uil *UserInfoList

	if m.Entry[0].Id == 0 {
		args := strings.Split(m.Entry[0].Opt, ":")

		if c := len(args); c != 4 {
			stat.ReqErrListUser++
			return errors.New("Invalid server list parameter count")
		}

                list := args[0]

		if list == "all" || list == "enabled" || list == "disabled" ||
			list == "active" || list == "inactive" ||
			list == "new" || list == "admin" {
		} else {
			stat.ReqErrListUser++
			return errors.New("Invalid server list: " + list)
		}

                page, _ := strconv.ParseInt(args[1], 0, 32)
                npage, _ := strconv.ParseInt(args[2], 0, 32)

		if page == 0 || npage < 10 {
			stat.ReqErrListUser++
			return errors.New("Invalid list page count")
		}

		// init sort closures
		var sfn UserSortBy

                if sby := args[3]; sby == "id" {
			sfn = func(s1, s2 *UserInfo) bool {
				return s1.Id < s2.Id
			}
		} else if sby == "id-r" {
			sfn = func(s1, s2 *UserInfo) bool {
				return s1.Id > s2.Id
			}
		} else if sby == "name" {
			sfn = func(s1, s2 *UserInfo) bool {
				return s1.Name < s2.Name
			}
		} else if sby == "name-r" {
			sfn = func(s1, s2 *UserInfo) bool {
				return s1.Name > s2.Name
			}
		} else if sby == "rdate" {
			sfn = func(s1, s2 *UserInfo) bool {
				return s1.RegDate < s2.RegDate
			}
		} else if sby == "rdate-r" {
			sfn = func(s1, s2 *UserInfo) bool {
				return s1.RegDate > s2.RegDate
			}
		} else {
			stat.ReqErrListUser++
			return errors.New("Invalid sort field: " + sby)
		}

		var v []string

		if v, err = getRedisUserList(list); err != nil {
			stat.ReqErrListUser++
			return
		}

		si = make([]UserInfo, len(v))

		for i := range v {
			uid, _ := strconv.ParseInt(v[i], 0, 32)

			if s, err := getRedisUserInfo(uid); err != nil {
				event(logwarn, li, err.Error())
				si[i] = UserInfo{Idx: uid, ErrNo: ENOENT}
			} else {
				si[i] = *s
				si[i].Idx = uid
			}
		}

		// sort data
		UserSortBy(sfn).Sort(si)

		// get page entries
                c := len(si)

		if int(npage) < c {
			var st []UserInfo

			pofs := int(page - 1)
			sofs := int(pofs * int(npage))
			eofs := int(pofs*int(npage) + (int(npage) - 1))

			if sofs >= c {
				stat.ReqErrListUser++
				return errors.New("Invalid page offset")
			}

			if eofs < c {
				st = make([]UserInfo, npage)
				copy(st, si[sofs:])
			} else {
				eofs = c
				st = make([]UserInfo, len(si[sofs:eofs]))
				copy(st, si[sofs:eofs])
			}

			uil = &UserInfoList{Id: int64(c), Entry: st}
		} else {
			uil = &UserInfoList{Id: int64(c), Entry: si}
		}
	} else {
		si = make([]UserInfo, len(m.Entry))

		for i := range m.Entry {
			e := m.Entry[i]

			if s, err := getRedisUserInfo(e.Id); err != nil {
				event(logwarn, li, err.Error())
				si[i] = UserInfo{Idx: e.Id, ErrNo: ENOENT}
			} else {
				si[i] = *s
				si[i].Idx = e.Id
			}
		}

		uil = &UserInfoList{Id: int64(len(si)), Entry: si}
	}

	buf, _ := json.Marshal(uil)
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func getUserList(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqGetUserList++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrGetUserList++
		return
	}

	args := strings.Split(m.Entry[0].Opt, ":")

	if c := len(args); c != 4 {
		stat.ReqErrGetUserList++
		return errors.New("Invalid server list parameter count")
	}

	list := args[0]

	if list == "activity" || list == "login" {
	} else {
		stat.ReqErrGetUserList++
		return errors.New("Invalid user list: " + list)
	}

	page, _ := strconv.ParseInt(args[1], 0, 32)
	npage, _ := strconv.ParseInt(args[2], 0, 32)

	if page == 0 || npage < 10 {
		stat.ReqErrListUser++
		return errors.New("Invalid list page count")
	}

	var v []string

	if v, err = getRedisUserUidList(m.Entry[0].Id, list); err != nil {
		stat.ReqErrGetUserList++
		return
	}

	si := make([]Name, len(v))

	for i := range v {
		si[i] = Name{Name: v[i]}
	}

	var nl *NameList

	// get page entries
	c := len(si)

	if int(npage) < c {
		var st []Name

		pofs := int(page - 1)
		sofs := int(pofs * int(npage))
		eofs := int(pofs*int(npage) + (int(npage) - 1))

		if sofs >= c {
			stat.ReqErrGetUserList++
			return errors.New("Invalid page offset")
		}

		if eofs < c {
			st = make([]Name, npage)
			copy(st, si[sofs:])
		} else {
			eofs = c
			st = make([]Name, len(si[sofs:eofs]))
			copy(st, si[sofs:eofs])
		}

		nl = &NameList{Id: int64(c), Entry: st}
	} else {
		nl = &NameList{Id: int64(c), Entry: si}
	}

	buf, _ := json.Marshal(nl)
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func defaultAdminUserHandler(w http.ResponseWriter, r *http.Request) {
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

	if err = checkUserAdmin(d.UserId); err != nil {
		stat.ReqErrUserId++
		sendError(w, EPERM, str, err)
		return
	}

	event(logdebug, li, "Processing request [%v:%v]", d.Command, li.Msgid)

	switch d.Command {
	case "resolve-user":
		err = resolveUser(w, d)

	case "resolve-user-id":
		err = resolveUserId(w, d)

	case "reset-user-pw":
		err = resetUserPw(w, d)

	case "add-user":
		err = addUser(w, d)

	case "set-user-attr":
		err = setUserAttr(w, d)

	case "enable-user":
		err = enableUser(w, d)

	case "disable-user":
		err = disableUser(w, d)

	case "activate-user":
		err = activateUser(w, d)

	case "deactivate-user":
		err = deactivateUser(w, d)

	case "list-user":
		err = listUser(w, d)

	case "get-user-list":
		err = getUserList(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Request [%v:%v] completed", d.Command, li.Msgid)
}
