/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

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

	if len(app.Bind) == 0 {
		fatal("Invalid bind parameters")
	}

	if app.HostName == "" {
		warn("Rebana URL is empty")
	}

	if app.RebanaUrl == "" {
		warn("Rebana URL is empty")
	}

	if app.LogUrl == "" {
		warn("Log URL is empty")
	}

	if app.Secret == "" {
		fatal("Secret is empty")
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

	return
}

func getIdList(s string, c string) (d *IdList, err error) {
	d = &IdList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		stat.ReqErrData++
		return nil, errors.New("Error unmarshaling IdList struct")
	}

	if d.Id != 0 {
		if c == "activate" || c == "deactivate" || c == "check" {
			if err = checkServerId(d.Id); err != nil {
				stat.ReqErrServerId++
				return
			}
		} else {
			stat.ReqErrServerId++
			return d, errors.New("Invalid tunnel server ID for " + c)
		}
	}

	return
}

func signRequest(m []byte, id int64) string {
	var dgst = hmac.New(sha256.New, []byte(app.Secret))

	dgst.Write(m)

	return base64.StdEncoding.EncodeToString(dgst.Sum(nil))
}

func checkSignature(sig string, m []byte) (err error) {
	var dgst = hmac.New(sha256.New, []byte(app.Secret))

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

func checkUrl(r *http.Request) (err error) {
	var p = r.URL.Path[1:]

	switch p {
	case "activate":
	case "deactivate":
	case "check":
	case "status":
		break

	default:
		return errors.New("Invalid URL: " + r.URL.Path)
	}

	if ip := r.Header.Get("X-Forwarded-For"); ip == "" {
		li.Src = r.RemoteAddr
	} else {
		li.Src = ip
	}

	event(logdebug, li, "New connection from %v to %v", li.Src, r.URL.Path)
	return
}

func checkHeader(r *http.Request) (err error) {
	if r.Method != "POST" {
		return errors.New("Invalid method: " + r.Method)
	}

	if c := r.Header.Get("Content-Type"); c != "application/json" {
		return errors.New("Invalid Content-Type header: " + c)
	}

	if s := r.Header.Get("X-N3-Service-Name"); s != "rebana" {
		return errors.New("Invalid Service-Name header: " + s)
	}

	return
}

func checkData(r *http.Request) (d *RequestMsg, err error) {
	if err = json.NewDecoder(r.Body).Decode(&d); err != nil {
		return d, errors.New("Invalid JSON payload")
	}

	var t, _ = time.Parse(time.RFC1123, r.Header.Get("Date"))

	if err = checkMsgExpiry(t); err != nil {
		return
	}

	var sig = r.Header.Get("X-N3-Signature")
	var buf, _ = json.Marshal(d)

	if err = checkSignature(sig, buf); err != nil {
		stat.ReqErrSignature++
		return
	}

	if err = checkServerId(d.Id); err != nil {
		stat.ReqErrServerId++
		return
	}

	if err = checkCommand(d.Command, r.RemoteAddr); err != nil {
		stat.ReqErrCommand++
		return
	}

	if d.UserId == 0 {
		stat.ReqErrUserId++
		return d, errors.New("Invalid user ID")
	}

	li.Uid = d.UserId
	li.Msgid = d.MsgId

	var ts = t.In(time.Local).Format(time.RFC1123)

	event(logdebug, li, "Request verified: [Timestamp: %v, Server ID: %v, "+
		"Message ID: %v, User ID: %v, Command: %v]", ts, d.Id, d.MsgId,
		d.UserId, d.Command)
	return
}

func checkResponseHeader(r *http.Response) (t time.Time, err error) {
	t, _ = time.Parse(time.RFC1123, r.Header.Get("Date"))

	if err = checkMsgExpiry(t); err != nil {
		return
	}

	if c := r.Header.Get("Content-Type"); c != "application/json" {
		return t, errors.New("Invalid Content-Type header: " + c)
	}

	if s := r.Header.Get("X-N3-Service-Name"); s != "rebana" {
		return t, errors.New("Invalid Service-Name header: " + s)
	}

	return
}

func checkServerId(id int64) (err error) {
	if id != app.SvInfo.Id {
		return errors.New("Invalid tunnel server ID")
	}

	return
}

func checkCommand(c string, a string) (err error) {
	switch c {
	case "activate":
	case "deactivate":
	case "check":
	case "status":
		break

	default:
		return errors.New("Invalid command: " + c)
	}

	return
}

func checkIPFamily(s string) (f int, err error) {
	var ip = net.IP{}

	if ip = net.ParseIP(s); ip == nil {
		return f, errors.New("Invalid IP address: " + s)
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

	return f, errors.New("Invalid IP address: " + s)
}

func sendResponse(w http.ResponseWriter, m *Msg) {
	var data = &Msg{Id: app.SvInfo.Id, MsgId: li.Msgid, Data: m.Data}
	var buf, _ = json.Marshal(data)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-N3-Service-Name", "rebana")
	w.Header().Add("X-N3-Signature", signRequest(buf, data.MsgId))

	fmt.Fprintf(w, "%s", buf)

	event(logdebug, li, "Server response sent: [Hostname: %v, User ID: "+
		"%v, Message ID: %v, Error code: %v]", app.HostName, li.Uid,
		li.Msgid, m.ErrNo)
}

func sendError(w http.ResponseWriter, errno int, estr string, err error) {
	event(logwarn, li, err.Error())
	stat.ReqError++

	var data = &Msg{Id: app.SvInfo.Id, MsgId: li.Msgid, ErrNo: errno,
		Data: estr}
	var buf, _ = json.Marshal(data)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-N3-Service-Name", "rebana")
	w.Header().Add("X-N3-Signature", signRequest(buf, data.MsgId))

	fmt.Fprintf(w, "%s", buf)

	event(logdebug, li, "Server error response sent: [Hostname: %v, "+
		"User ID: %v, Message ID: %v, Error code: %v]", app.HostName,
		li.Uid, li.Msgid, errno)
}

func setupServer() (err error) {
	http.HandleFunc("/", defaultHandler)

	for i := range app.Bind {
		var b = net.JoinHostPort(app.Bind[i].Host, app.Bind[i].Port)

		go http.ListenAndServe(b, nil)
		event(loginfo, li, "Listening on %v", b)
	}

	if tlsc, err = setTLSConfig(); err != nil {
		return
	}

	return
}

func setTLSConfig() (p *tls.Config, err error) {
	var cert *x509.Certificate
	var data []byte
	var asn1 *pem.Block

	var opts = x509.VerifyOptions{Roots: x509.NewCertPool()}

	for i := range app.TLSCACert {
		if data, err = ioutil.ReadFile(app.TLSCACert[i]); err != nil {
			return p, errors.New("Error reading TLS CA cert")
		}

		if asn1, data = pem.Decode(data); asn1 == nil {
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
