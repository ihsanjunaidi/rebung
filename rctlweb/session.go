/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

type Cookie struct {
	Id    string
	Flash map[string]string
}

type Session struct {
	Id       string
	Hash     string
	UserId   int64
	Username string
	Name     string
	Key      string
	Flash    map[string]string
}

func setFlashMessage(w http.ResponseWriter, s *Session) {
	buf, _ := json.Marshal(s)
	str := base64.StdEncoding.EncodeToString(buf)
	c := &http.Cookie{Name: COOKIE, Domain: CDOMAIN, Path: CPATH, Value: str}

	http.SetCookie(w, c)
}

func setSession(w http.ResponseWriter, s *Session) (sid string, err error) {
	cs := &Cookie{Id: generateSessionId()}
	buf, _ := json.Marshal(cs)
	str := base64.StdEncoding.EncodeToString(buf)

	s.Hash = signRequest([]byte(buf), app.SessionSecret)

	c := &http.Cookie{Name: COOKIE, Domain: CDOMAIN, Path: CPATH, Value: str}

	http.SetCookie(w, c)

	if err = setRedisSession(cs.Id, s); err != nil {
		return
	}

	return cs.Id, nil
}

func getSession(r *http.Request) (s *Session, err error) {
	c := &http.Cookie{}

	if c, err = r.Cookie(COOKIE); err != nil {
		return s, errors.New("Error retrieving session cookie")
	}

	var str []byte

	if str, err = base64.StdEncoding.DecodeString(c.Value); err != nil {
		return s, errors.New("Error decoding session value")
	}

	cs := &Cookie{}

	if err = json.Unmarshal(str, cs); err != nil {
		return s, errors.New("Corrupted session data")
	}

	if cs.Id == "" {
		return s, errors.New("No session ID retrieved")
	} else if cs.Id == "0" {
		s = &Session{Id: cs.Id, Flash: cs.Flash}
	} else {
		if s, err = getRedisSession(cs.Id); err != nil || s == nil {
			return
		}

		if checkSignature(c.Value, app.SessionSecret,
                        []byte(s.Hash)) != nil {
			return
		}
	}

	return
}

func getWSSession(sid string) (s *Session, err error) {
	if sid == "" {
		return s, errors.New("Invalid session ID")
	}

	if s, err = getRedisSession(sid); err != nil {
		return
	}

	return
}

func deleteSession(w http.ResponseWriter, sid string) (err error) {
	c := &http.Cookie{Name: COOKIE, Domain: CDOMAIN, Path: CPATH,
		Expires: time.Unix(0, 0)}

	http.SetCookie(w, c)

	if sid != "" {
		if err = deleteRedisSessionId(sid); err != nil {
			return
		}
	}

	return
}

func generateSessionId() string {
	const c string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" +
		"1234567890"

	rand.Seed(time.Now().UnixNano())

	p := make([]byte, 40)

	for i := range p {
		r := rand.Intn(len(c))
		p[i] = c[r]
	}

	return string(p)
}
