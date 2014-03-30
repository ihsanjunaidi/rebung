/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func formLogin(w http.ResponseWriter, r *http.Request) (err error) {
	login := r.FormValue("login")
	passwd := r.FormValue("password")

	var idl *IdList

	if idl, err = userLogin(login, passwd); err != nil {
		redirectLogin(w, r, "Incorrect login information", "", err)
		return
	}

	e := idl.Entry[0]

	var uil *UserInfoList

	if uil, err = listUser(e.Id, []int64{e.Id}, ""); err != nil {
		redirectLogin(w, r, "Error retrieving user information", "", err)
		return
	}

	deleteSession(w, "")

	u := uil.Entry[0]
	s := &Session{Key: e.Opt, UserId: e.Id, Username: u.Login, Name: u.Name}

	if s.Id, err = setSession(w, s); err != nil {
		event(logwarn, li, err.Error())
	}

	redirectUrl(w, r, s, "/home", "You are logged in as "+u.Login, nil)
	return
}

func formSearch(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, err = getSession(r); err != nil {
		redirectLogin(w, r, "URL access is restricted", "", err)
		return
	}

	uid, _ := strconv.ParseInt(r.FormValue("uid"), 0, 64)

	if s.UserId != uid {
		redirectUrl(w, r, s, "/home", "Invalid search parameter",
			errors.New("Mismatched user ID"))
		return
	}

	var user bool

	query := r.FormValue("query")

	if strings.Contains(query, "@") {
		user = true
	}

	var id int64

	v := url.Values{}

	if user {
		if id, err = resolveUserLogin(s.UserId, query); err != nil || id == 0 {
			redirectUrl(w, r, s, "/home", "User ["+query+
				"] not found", err)
			return
		}

		v.Set("uid", fmt.Sprintf("%v", id))
	} else {
		if id, err = resolveServerName(s.UserId, query); err != nil ||
			id == 0 {
			redirectUrl(w, r, s, "/home", "Server ["+query+
				"] not found", err)
			return
		}

		v.Set("vid", fmt.Sprintf("%v", id))
	}

	redirectUrl(w, r, s, "/list?"+v.Encode(), "", nil)
	return
}

func formAddServer(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, err = getSession(r); err != nil {
		redirectLogin(w, r, "URL access is restricted", "", err)
		return
	}

	uid, _ := strconv.ParseInt(r.FormValue("uid"), 0, 64)

	if s.UserId != uid {
		redirectUrl(w, r, s, "/home", "Invalid request parameter",
			errors.New("Mismatched user ID"))
		return
	}

	name := r.FormValue("name")
	pp := r.FormValue("ppprefix")
	rt := r.FormValue("rtprefix")

	if name == "" || pp == "" || rt == "" {
		redirectUrl(w, r, s, "/home", "Invalid request parameter", err)
		return
	}

	var idl *IdList

	if idl, err = addServer(s.UserId, name, pp, rt); err != nil {
		redirectUrl(w, r, s, "/home", "Unable to add tunnel server "+
			name, err)
		return
	}

	redirectUrl(w, r, s, "/home", fmt.Sprintf("Server %v[%v] successfully "+
		"added", idl.Entry[0].Opt, idl.Entry[0].Id), nil)
	return
}

func formAddUser(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, err = getSession(r); err != nil {
		redirectLogin(w, r, "URL access is restricted", "", err)
		return
	}

	uid, _ := strconv.ParseInt(r.FormValue("uid"), 0, 64)

	if s.UserId != uid {
		redirectUrl(w, r, s, "/home", "Invalid request parameter",
			errors.New("Mismatched user ID"))
		return
	}

	login := r.FormValue("email")
	name := r.FormValue("name")

	if login == "" || name == "" {
		redirectUrl(w, r, s, "/home", "Invalid request parameter", err)
		return
	}

	var idl *IdList

	if idl, err = addUser(s.UserId, login, name, li.Src); err != nil {
		redirectUrl(w, r, s, "/home", "Unable to add user "+login, err)
		return
	}

	redirectUrl(w, r, s, "/home", fmt.Sprintf("User %v[%v] successfully "+
		"added", idl.Entry[0].Opt, idl.Entry[0].Id), nil)
	return
}

func formChangeName(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, err = getSession(r); err != nil {
		redirectLogin(w, r, "URL access is restricted", "", err)
		return
	}

	uid, _ := strconv.ParseInt(r.FormValue("uid"), 0, 64)

	if s.UserId != uid {
		redirectUrl(w, r, s, "/profile", "Invalid request parameter",
			errors.New("Mismatched user ID"))
		return
	}

	name := r.FormValue("name")

	if name == "" {
		redirectUrl(w, r, s, "/profile", "Invalid request parameter", err)
		return
	}

	var nl *NameList

	k := []string{"name"}
	v := []string{name}

	if nl, err = setUserAttr(s.UserId, s.UserId, k, v); err != nil {
		redirectUrl(w, r, s, "/profile", "Error changing "+s.Username+
			" name", err)
		return
	}

	setRedisSessionName(s.UserId, nl.Entry[0].Opt)
	redirectUrl(w, r, s, "/home", s.Username+" name changed", nil)
	return
}

func formChangePw(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, err = getSession(r); err != nil {
		redirectLogin(w, r, "URL access is restricted", "", err)
		return
	}

	uid, _ := strconv.ParseInt(r.FormValue("uid"), 0, 64)

	if s.UserId != uid {
		redirectUrl(w, r, s, "/profile", "Invalid request parameter",
			errors.New("Mismatched user ID"))
		return
	}

	p1 := r.FormValue("pass1")
	p2 := r.FormValue("pass2")

	if p1 == "" || p2 == "" {
		redirectUrl(w, r, s, "/profile", "Invalid request parameter", err)
		return
	}

	k := []string{"password"}
	v := []string{p1}

	if _, err = setUserAttr(s.UserId, s.UserId, k, v); err != nil {
		redirectUrl(w, r, s, "/profile", "Error changing "+s.Username+
			" name", err)
		return
	}

	redirectUrl(w, r, s, "/home", s.Username+" password changed", nil)
	return
}
