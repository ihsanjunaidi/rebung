/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type RequestOpt struct {
	Uid  int64
	Cmd  string
	Data string
	Url  string
}

type Name struct {
	Name  string
	ErrNo int
	Opt   string
}

type NameList struct {
	Id    int64
	Entry []Name
}

type Id struct {
	Id    int64
	ErrNo int
	Opt   string
}

type IdList struct {
	Id    int64
	Entry []Id
}

type RenderVar struct {
	Title     string
	SessionId string
	UserId    int64
	Username  string
	Name      string
	Flash     map[string]string
	Data      map[string]string

	User    *UserInfo
	Users   bool
	Server  *ServerInfo
	Servers bool

	Sessions     []SessionInfo
	UserSessions []UserSessionInfo
}

type WSRequest struct {
	Sid  string
	Uid  int64
	Cmd  string
	Data string
}

type WSResponse struct {
	Uid   int64
	ErrNo int
	Data  string
}

var gmt  *time.Location
var tlsc *tls.Config
var tpl *template.Template

func parseConfig(f string) (err error) {
	if _, err = os.Stat(f); err != nil {
		return errors.New("Configuration file does not exist")
	}

	var buf []byte

	if buf, err = ioutil.ReadFile(f); err != nil {
		return errors.New("Unable to read configuration file")
	}

	if err = json.Unmarshal(buf, app); err != nil {
		return errors.New("Unable to unmarshal AppConfig struct")
	}

	if app.HostName == "" {
		warn("Hostname is empty")
	}

	if len(app.Bind) == 0 {
		fatal("Invalid bind parameters")
	}

	if app.LogUrl == "" {
		warn("Log URL is empty")
	}

	if app.GhazalUrl == "" {
		fatal("Ghazal URL is empty")
	}

	if app.RebanaUrl == "" {
		fatal("Rebana URL is empty")
	}

	if app.SessionSecret == "" {
		fatal("Session secret is empty")
	}

	if app.GhazalSecret == "" {
		fatal("Ghazal secret is empty")
	}

	if app.RebanaSecret == "" {
		fatal("Rebana secret is empty")
	}

	if app.AppRoot == "" {
		fatal("Application directory is empty")
	} else {
		app.TemplateDir = app.AppRoot + TPLDIR
	}

	if len(app.TLSCACert) == 0 {
		fatal("Invalid TLS CA cert parameters")
	} else {
		for i := range app.TLSCACert {
			e := app.TLSCACert[i]
			if _, err = os.Stat(e); err != nil {
				fatal("CA TLS cert file not found: %v", e)
			}
		}
	}

	if app.RedisUrl == "" {
		fatal("Redis URL is empty")
	}

	if app.RedisPw == "" {
		fatal("Redis password is empty")
	}

	if app.RedisDb == "" {
		fatal("Redis database index is empty")
	}

	return
}

func getNameList(s string, c string) (d *NameList, err error) {
	d = &NameList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		return d, errors.New("Error unmarshaling NameList struct")
	}

	return
}

func getIdList(s string, c string) (d *IdList, err error) {
	d = &IdList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		return d, errors.New("Error unmarshaling IdList struct")
	}

	return
}

func getUserInfoList(s string, c string) (d *UserInfoList, err error) {
	d = &UserInfoList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		return d, errors.New("Error unmarshaling UserInfoList struct")
	}

	return
}

func getServerInfoList(s string, c string) (d *ServerInfoList, err error) {
	d = &ServerInfoList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		return d, errors.New("Error unmarshaling ServerInfoList struct")
	}

	return
}

func getSessionInfoList(s string, c string) (d *SessionInfoList, err error) {
	d = &SessionInfoList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		return d, errors.New("Error unmarshaling SessionInfoList struct")
	}

	return
}

func getUserSessionInfoList(s string, c string) (d *UserSessionInfoList,
	err error) {
	d = &UserSessionInfoList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		return d, errors.New("Error unmarshaling UserSessionInfoList struct")
	}

	return
}

func signRequest(m []byte, key string) string {
	dgst := hmac.New(sha256.New, []byte(key))
	dgst.Write(m)

	return base64.StdEncoding.EncodeToString(dgst.Sum(nil))
}

func checkSignature(sig, key string, m []byte) (err error) {
	dgst := hmac.New(sha256.New, []byte(key))
	dgst.Write(m)

	var s []byte

	if s, err = base64.StdEncoding.DecodeString(sig); err != nil {
		return errors.New("Error decoding signature string")
	}

	if !hmac.Equal(s, dgst.Sum(nil)) {
		return errors.New("Message signature does not match")
	}

	return
}

func checkMsgExpiry(t time.Time) (err error) {
	if time.Now().Sub(t).Seconds() > 10.0 {
		return errors.New("Message has expired")
	}

	return
}

func checkIPFamily(s string) (f int) {
	ip := net.IP{}

	if ip = net.ParseIP(s); ip == nil {
		return
	}

	if strings.Count(ip.String(), ".") == 3 {
		f = 4
		return
	} else {
		if strings.Contains(ip.String(), "::") {
			f = 6
			return
		}
	}

	return
}

func urlCheck(w http.ResponseWriter, r *http.Request, t string) (s *Session,
	v *RenderVar, err error) {
	if s, err = getSession(r); err != nil {
		redirectLogin(w, r, "URL access restricted", "", err)
		return
	}

	v = &RenderVar{Title: t, SessionId: s.Id, UserId: s.UserId,
		Username: s.Username, Name: s.Name}

	var fs, ft string

	if fs, ft, err = getRedisFlashMessage(s.UserId); err != nil {
		event(lognotice, li, err.Error())
	}

	if ft != "" {
		v.Flash = make(map[string]string, 1)
		v.Flash[ft] = fs
	}

	if err = deleteRedisFlashMessage(s.UserId); err != nil {
		event(lognotice, li, err.Error())
		err = nil
	}

	return
}

func wsCheck(w http.ResponseWriter, r *http.Request, flag bool) (s *Session,
	d *WSRequest, err error) {
	d = &WSRequest{}

	if err = json.NewDecoder(r.Body).Decode(&d); err != nil {
		sendWSResponse(w, EINVAL, "Invalid JSON payload")
		return
	}

	if s, err = getWSSession(d.Sid); err != nil {
		sendWSResponse(w, EINVAL, err.Error())
		return
	}

	// do not check for user or server ID
	if flag {
		return
	} else {
		if d.Uid == 0 {
			sendWSResponse(w, EINVAL, "Invalid ID field")
			return
		}
	}

	return
}

func render(w http.ResponseWriter, r *RenderVar, t string) error {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	return tpl.ExecuteTemplate(w, t+".html", r)
}

func renderError(w http.ResponseWriter, r *http.Request, errno int,
        estr string) error {
	event(logwarn, li, estr)

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

        s, _ := getSession(r)
	v := &RenderVar{}

        if s != nil {
            v.UserId = s.UserId
        }

	if errno == http.StatusNotFound {
		v.Title = "Page not found"
	} else if errno == http.StatusInternalServerError {
		v.Title = "Internal server error"
	}

	v.Data = make(map[string]string, 1)
	v.Data["estr"] = estr

	w.WriteHeader(errno)

	return tpl.ExecuteTemplate(w, "errors.html", v)
}

func redirectUrl(w http.ResponseWriter, r *http.Request, s *Session,
	url, fstr string, err error) {
	c := "success"

	if err != nil {
		event(logwarn, li, err.Error())
		c = "error"
	}

	if err = setRedisFlashMessage(s.UserId, c, fstr); err != nil {
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func redirectLogin(w http.ResponseWriter, r *http.Request, fstr, sid string,
	err error) {
	if sid != "" {
		if err := deleteSession(w, sid); err != nil {
			event(logwarn, li, err.Error())
		}
	}

	sn := &Session{Id: "0"}

	if err != nil {
		sn.Flash = map[string]string{"error": fstr}
	} else {
		sn.Flash = map[string]string{"success": fstr}
	}

	setFlashMessage(w, sn)
	http.Redirect(w, r, "/", http.StatusFound)
}

func sendWSResponse(w http.ResponseWriter, e int, s string) {
	buf, _ := json.Marshal(&WSResponse{Uid: li.Uid, ErrNo: e, Data: s})

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", buf)
}

func parseTemplates() {
        tpl = template.Must(template.ParseGlob(app.TemplateDir + "/*"))
}

func setupServer() (err error) {
	gmt = &time.Location{}

	if gmt, err = time.LoadLocation("Etc/GMT"); err != nil {
		return
	}

	http.HandleFunc("/", mainUrlHandler)

	for i := range app.Bind {
		go http.ListenAndServe(net.JoinHostPort(app.Bind[i].Host,
			app.Bind[i].Port), nil)
		event(loginfo, li, "listening on %v:%v", app.Bind[i].Host,
			app.Bind[i].Port)
	}

	if tlsc, err = setTLSConfig(); err != nil {
		return
	}

	return
}

func setTLSConfig() (p *tls.Config, err error) {
	var cert *x509.Certificate

	opts := x509.VerifyOptions{Roots: x509.NewCertPool()}

	for i := range app.TLSCACert {
		var data []byte

		if data, err = ioutil.ReadFile(app.TLSCACert[i]); err != nil {
			return p, errors.New("Error reading TLS CA cert")
		}

		var asn1 *pem.Block

		if asn1, _ = pem.Decode(data); asn1 == nil {
			return p, errors.New("Error decoding TLS CA cert")
		}

		if cert, err = x509.ParseCertificate(asn1.Bytes); err != nil {
			return p, errors.New("Error parsing TLS CA cert")
		}

		opts.Roots.AddCert(cert)
	}

	if _, err = cert.Verify(opts); err != nil {
		return p, errors.New("Error verifying new root CA certs")
	}

	p = &tls.Config{RootCAs: opts.Roots}
	return
}

func sendGhazalRequest(r *RequestOpt) (msg *GhazalMsg, err error) {
	m := &GhazalRequest{UserId: r.Uid, Origin: li.Src, Command: r.Cmd,
		Data: r.Data}
	buf, _ := json.Marshal(m)
	rd := bytes.NewReader(buf)

	var req *http.Request

	if req, err = http.NewRequest("POST", r.Url, rd); err != nil {
		return
	}

	req.Header.Add("Date", time.Now().In(gmt).Format(time.RFC1123))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-N3-Service-Name", "ghazal")
	req.Header.Add("X-N3-Signature", signRequest(buf, app.GhazalSecret))

	con := &http.Client{}
	con.Transport = &http.Transport{TLSClientConfig: tlsc}

	var res *http.Response

	if res, err = con.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	event(logdebug, li, "Ghazal request sent to %v: [User ID: %v, "+
		"Origin: %v, Command: %v]", r.Url, m.UserId, m.Origin, m.Command)

	t, _ := time.Parse(time.RFC1123, res.Header.Get("Date"))

	if err = checkMsgExpiry(t); err != nil {
		return
	}

	if c := res.Header.Get("Content-Type"); c != "application/json" {
		return msg, errors.New("Invalid Content-Type header: " + c)
	}

	if s := res.Header.Get("X-N3-Service-Name"); s != "ghazal" {
		return msg, errors.New("Invalid Service-Name header: " + s)
	}

	msg = &GhazalMsg{}

	if err = json.NewDecoder(res.Body).Decode(&msg); err != nil {
		return msg, errors.New("Invalid JSON payload")
	}

	buf, _ = json.Marshal(msg)

	sig := res.Header.Get("X-N3-Signature")

	if err = checkSignature(sig, app.GhazalSecret, buf); err != nil {
		return
	}

	li.Uid = msg.UserId

	ts := t.In(time.Local).Format(time.RFC1123)

	event(logdebug, li, "Ghazal response verified: [Timestamp: %v, "+
		"Hostname: %v, User ID: %v, ErrNo: %v]", ts, msg.HostName,
		msg.UserId, msg.ErrNo)

	if msg.ErrNo != EOK {
		err = errors.New(msg.Data)
	}

	return
}

func sendRebanaRequest(r *RequestOpt) (msg *RebanaMsg, err error) {
	m := &RebanaRequest{UserId: r.Uid, Command: r.Cmd, Data: r.Data}
        buf, _ := json.Marshal(m)
        rd := bytes.NewReader(buf)

	var req *http.Request

	if req, err = http.NewRequest("POST", r.Url, rd); err != nil {
		return
	}

	req.Header.Add("Date", time.Now().In(gmt).Format(time.RFC1123))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-N3-Service-Name", "rebana")
	req.Header.Add("X-N3-Signature", signRequest(buf, app.RebanaSecret))

        con := &http.Client{}
	con.Transport = &http.Transport{TLSClientConfig: tlsc}

	var res *http.Response

	if res, err = con.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	event(logdebug, li, "Rebana request sent to %v: [User ID: %v, "+
		"Command: %v]", r.Url, m.UserId, m.Command)

        t, _ := time.Parse(time.RFC1123, res.Header.Get("Date"))

	if err = checkMsgExpiry(t); err != nil {
		return
	}

	if c := res.Header.Get("Content-Type"); c != "application/json" {
		return msg, errors.New("Invalid Content-Type header: " + c)
	}

	if s := res.Header.Get("X-N3-Service-Name"); s != "rebana" {
		return msg, errors.New("Invalid Service-Name header: " + s)
	}

	msg = &RebanaMsg{}

	if err = json.NewDecoder(res.Body).Decode(&msg); err != nil {
		return msg, errors.New("Invalid JSON payload")
	}

	buf, _ = json.Marshal(msg)

        sig := res.Header.Get("X-N3-Signature")

	if err = checkSignature(sig, app.RebanaSecret, buf); err != nil {
		return
	}

	li.Uid = msg.UserId

        ts := t.In(time.Local).Format(time.RFC1123)

	event(logdebug, li, "Rebana response verified: [Timestamp: %v, "+
		"Hostname: %v, User ID: %v, ErrNo: %v]", ts, msg.HostName,
		msg.UserId, msg.ErrNo)

	if msg.ErrNo != EOK {
		err = errors.New(msg.Data)
	}

	return
}
