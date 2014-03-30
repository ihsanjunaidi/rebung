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
	"io/ioutil"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	chMsgFatal  = -1
	chMsgDebug  = 1
	chMsgInfo   = 2
	chMsgNotice = 3
)

type ChMsg struct {
	Type int
	Msg  string
}

var tlsc *tls.Config

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

	if app.AdminEmail == "" {
		fatal("Admin e-mail is empty")
	}

	if app.SMTPHost == "" {
		fatal("SMTP host is empty")
	}

	if app.SMTPUser == "" {
		fatal("SMTP user is empty")
	}

	if app.SMTPPw == "" {
		fatal("SMTP password is empty")
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
		stat.ReqErrData++
		return nil, errors.New("Error unmarshaling NameList struct")
	}

	if d.Id != 0 {
		if c == "set-server-attr" {
			if err = checkRedisServerId(d.Id); err != nil {
				stat.ReqErrServerId++
				return
			}
		} else {
			stat.ReqErrServerId++
			return nil, errors.New("Invalid tunnel server ID")
		}
	}

	if len(d.Entry) == 0 {
		stat.ReqErrData++
		return nil, errors.New("Invalid arg list")
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
		if c == "get-server-list" || c == "activate-session" ||
			c == "deactivate-session" || c == "check-session" {
			if err = checkRedisServerId(d.Id); err != nil {
				stat.ReqErrServerId++
				return
			}
		} else {
			stat.ReqErrServerId++
			return nil, errors.New("Invalid tunnel server ID")
		}
	}

	if len(d.Entry) == 0 {
		stat.ReqErrData++
		return nil, errors.New("Invalid arg list")
	}

	return
}

func getUserSession(uid, vid int64) (sid int64, err error) {
	var list []string

	if list, err = getRedisUserUidList(uid, "sessions"); err != nil {
		return
	}

	for i := range list {
		var tok = strings.Split(list[i], ":")
		var v, _ = strconv.ParseInt(tok[0], 0, 64)

		if len(tok) != 2 {
			return sid, errors.New("Invalid user session list")
		}

		if v == vid {
			sid, _ = strconv.ParseInt(tok[1], 0, 64)
		}
	}

	if sid == 0 {
		return sid, errors.New("Invalid session ID")
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

func checkUserAdmin(id int64) (err error) {
	var list []string
	var m bool

	if list, err = getRedisUserList("admin"); err != nil {
		return
	}

	var uid = fmt.Sprintf("%v", id)

	for i := range list {
		if uid == list[i] {
			m = true
			break
		}
	}

	if !m {
		return errors.New(fmt.Sprintf("User [%v] is not an admin", uid))
	}

	return
}

func checkUrl(r *http.Request) (err error) {
	switch r.URL.Path {
	case "/v/resolve":
	case "/v/add":
	case "/v/set":
	case "/v/list":
	case "/v/status":
	case "/v/info":

	case "/s/set":
	case "/s/assign":
	case "/s/list":

	case "/status":

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

        t, _ := time.Parse(time.RFC1123, r.Header.Get("Date"))

        /*
	if err = checkMsgExpiry(t); err != nil {
		return
	}*/

	var sig = r.Header.Get("X-N3-Signature")
	var buf, _ = json.Marshal(d)

	if err = checkSignature(sig, buf); err != nil {
		stat.ReqErrSignature++
		return
	}

	if err = checkCommand(d.Command); err != nil {
		stat.ReqErrCommand++
		return
	}

	if d.UserId == 0 {
		stat.ReqErrUserId++
		return d, errors.New("Invalid user ID")
	}

	li.Uid = d.UserId

	var ts = t.In(time.Local).Format(time.RFC1123)

	event(logdebug, li, "Request verified: [Timestamp: %v, User ID: %v, "+
		"Request IP: %v, Command: %v]", ts, d.UserId, li.Src, d.Command)
	return
}

func checkCommand(c string) (err error) {
	switch c {
	case "status":
		break

	case "resolve-server":
	case "resolve-server-id":
	case "add-server":
	case "set-server-attr":
	case "enable-server":
	case "disable-server":
	case "activate-server":
	case "deactivate-server":
	case "list-server":
	case "get-server-list":
	case "tunnel-server-status":
	case "server-status":
	case "server-info":
		break

	case "activate-session":
	case "deactivate-session":
	case "check-session":
	case "assign-session":
	case "reassign-session":
		break

	case "activate-user-session":
	case "deactivate-user-session":
	case "check-user-session":
	case "list-user-sessions":
	case "list-user-servers":
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
	var data = &Msg{HostName: app.HostName, UserId: li.Uid, MsgId: li.Msgid,
		Data: m.Data}
	var buf, _ = json.Marshal(data)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-N3-Service-Name", "rebana")
	w.Header().Add("X-N3-Signature", signRequest(buf, data.MsgId))

	fmt.Fprintf(w, "%s", buf)

	event(logdebug, li, "Server response sent: [Hostname: %v, User ID: "+
		"%v, Message ID: %v, Error code: %v]", data.HostName,
		data.UserId, data.MsgId, data.ErrNo)
}

func sendError(w http.ResponseWriter, errno int, estr string, err error) {
	event(logwarn, li, err.Error())
	stat.ReqError++

	var data = &Msg{HostName: app.HostName, UserId: li.Uid, MsgId: li.Msgid,
		ErrNo: errno, Data: estr}
	var buf, _ = json.Marshal(data)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-N3-Service-Name", "rebana")
	w.Header().Add("X-N3-Signature", signRequest(buf, data.MsgId))

	fmt.Fprintf(w, "%s", buf)

	event(logdebug, li, "Server error response sent: [Hostname: %v, "+
		"User ID: %v, Message ID: %v, Error code: %v]", data.HostName,
		data.UserId, data.MsgId, data.ErrNo)
}

func sendTSRequest(url string, m *TSReqMsg) (d *TSMsg, err error) {
	var buf, _ = json.Marshal(m)
	var rd = bytes.NewReader(buf)

	var req *http.Request

	if req, err = http.NewRequest("POST", url, rd); err != nil {
		return d, errors.New("Unable to craft request message")
	}

	event(logdebug, li, "Sending tunnel server request to %v: [Server ID: "+
		"%v, User ID: %v, "+"Message ID: %v, Command: %v]", url, m.Id,
		m.UserId, m.MsgId, m.Command)

	var loc = &time.Location{}

	if loc, err = time.LoadLocation("Etc/GMT"); err != nil {
		return
	}

	req.Header.Add("Date", time.Now().In(loc).Format(time.RFC1123))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-N3-Service-Name", "rebana")
	req.Header.Add("X-N3-Signature", signRequest(buf, m.MsgId))

	var con = &http.Client{}

	con.Transport = &http.Transport{TLSClientConfig: tlsc}

	var res *http.Response

	if res, err = con.Do(req); err != nil {
		return d, errors.New("Unable to send request to tunnel server")
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&d); err != nil {
		return d, errors.New("Invalid server response data")
	}

	event(logdebug, li, "Tunnel server response: [Server ID: %v, Message "+
		"ID: %v, Error code: %v]", d.Id, d.MsgId, d.ErrNo)

	var sig = res.Header.Get("X-N3-Signature")

	buf, _ = json.Marshal(d)

	if err = checkSignature(sig, buf); err != nil {
		stat.ReqErrSignature++
		return
	}

	var estr string

	if d.ErrNo != EOK {
		estr = fmt.Sprintf("Tunnel server [%v] reported error", d.Id)
		return d, errors.New(estr)
	}

	event(logdebug, li, "Reply received from tunnel server [%v]", d.Id)
	return
}

func sendMail(r []string, s, b string) (err error) {
	var rebana = mail.Address{"Rebana Web Service", "rebana@s.rebung.io"}
	var auth = smtp.PlainAuth("", app.SMTPUser, app.SMTPPw, app.SMTPHost)

	var url = net.JoinHostPort(app.SMTPHost, "587")

	var con = &smtp.Client{}

	if con, err = smtp.Dial(url); err != nil {
		err = errors.New("SMTP server not available")
		return
	}

	if err = con.StartTLS(tlsc); err != nil {
		return errors.New("StartTLS negotiation failed")
	}

	if err = con.Auth(auth); err != nil {
		return errors.New("SMTP Plain authentication failed")
	}

	header := make(map[string]string)
	header["From"] = rebana.String()
	header["Subject"] = s
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	con.Mail(rebana.String())

	for i := range r {
		con.Rcpt(r[i])
		header["To"] = r[i]
	}

	wc, err := con.Data()
	if err != nil {
		return errors.New("Unable to open SMTP DATA request")
	}
	defer wc.Close()

	var msg string

	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(b))

	var buf = bytes.NewBufferString(msg)

	if _, err = buf.WriteTo(wc); err != nil {
		return errors.New("Unable write message body")
	}

	return
}

func setupServer(ch chan<- ChMsg) {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/s/", defaultSessionHandler)
	http.HandleFunc("/v/", defaultServerHandler)

	for i := range app.Bind {
		var b = net.JoinHostPort(app.Bind[i].Host, app.Bind[i].Port)

		go func() {
			if err := http.ListenAndServe(b, nil); err != nil {
				ch <- ChMsg{Type: chMsgFatal, Msg: err.Error()}
			}
		}()

                msg := fmt.Sprintf("Listening on %v", b)
		ch <- ChMsg{Type: chMsgNotice, Msg: msg}
	}

	var err error
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
