/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"bytes"
	"code.google.com/p/go.crypto/bcrypt"
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
	"math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"time"
)

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
		return d, errors.New("Error unmarshaling NameList struct")
	}

	if d.Id != 0 {
		if c == "set-user-attr" || c == "logout" {
			if err = checkRedisUserId(d.Id); err != nil {
				stat.ReqErrUserId++
				return
			}
		} else {
			stat.ReqErrUserId++
			return d, errors.New("Invalid user ID")
		}
	}

	if len(d.Entry) == 0 {
		stat.ReqErrData++
		return d, errors.New("Invalid arg list")
	}

	return
}

func getIdList(s string, c string) (d *IdList, err error) {
	d = &IdList{}

	if err = json.Unmarshal([]byte(s), d); err != nil {
		stat.ReqErrData++
		return d, errors.New("Error unmarshaling IdList struct")
	}

	if d.Id != 0 {
		if c == "get-user-list" {
			if err = checkRedisUserId(d.Id); err != nil {
				stat.ReqErrUserId++
				return
			}
		} else {
			stat.ReqErrUserId++
			return d, errors.New("Invalid user ID")
		}
	}

	if len(d.Entry) == 0 {
		stat.ReqErrData++
		return d, errors.New("Invalid arg list for")
	}

	return
}

func generateTempPassword() string {
	const c string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" +
		"1234567890"

	rand.Seed(time.Now().UnixNano())

	p := make([]byte, 8)

	for i := range p {
		r := rand.Intn(len(c))
		p[i] = c[r]
	}

	return string(p)
}

func setUserSession(uid int64, login, pw, ip string) (tok string, err error) {
	const c string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" +
		"1234567890"

	var s *UserInfo

	if s, err = getRedisUserInfo(uid); err != nil {
		return
	}

	if login != s.Login {
		return tok, errors.New("User login did not match: " + login)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(s.Password),
		[]byte(pw)); err != nil {
		return tok, errors.New("User password did not match")
	}

	// record first-time login
	if s.FirstLogin == "" {
		if err = setRedisUserFirstLogin(uid, ip); err != nil {
			return
		}
	}

	rand.Seed(time.Now().UnixNano())

	p := make([]byte, 24)

	for i := range p {
		r := rand.Intn(len(c))
		p[i] = c[r]
	}

	tok = base64.StdEncoding.EncodeToString(p)

	if err = setRedisUserSession(uid, tok, ip); err != nil {
		return
	}

	return
}

func signRequest(m []byte, id int64) string {
	dgst := hmac.New(sha256.New, []byte(app.Secret))
	dgst.Write(m)

	return base64.StdEncoding.EncodeToString(dgst.Sum(nil))
}

func checkSignature(sig string, m []byte) (err error) {
	dgst := hmac.New(sha256.New, []byte(app.Secret))
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

func checkUserSession(login string) (uid int64, err error) {
	if uid, err = getRedisUserIdFromLogin(login); err != nil {
		return
	}

	// ignore if user already has a session
	if err = checkRedisUserSession(uid); err != nil {
		event(logwarn, li, err.Error())
		return uid, nil
	}

	return
}

func checkUserAdmin(uid int64) (err error) {
	var m bool
	var list []string

	if list, err = getRedisUserList("admin"); err != nil {
		return
	}

	var uids = fmt.Sprintf("%v", uid)

	for i := range list {
		if uids == list[i] {
			m = true
			break
		}
	}

	if !m {
		return errors.New("User [" + uids + "] is not an admin")
	}

	return
}

func checkUrl(r *http.Request) (err error) {
	switch r.URL.Path {
	case "/s/resolve":
	case "/s/add":
	case "/s/set":
	case "/s/reset":
	case "/s/list":

	case "/u/register":
	case "/u/login":
	case "/u/logout":

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

	if s := r.Header.Get("X-N3-Service-Name"); s != "ghazal" {
		return errors.New("Invalid Service-Name header: " + s)
	}

	return
}

func checkData(r *http.Request) (d *RequestMsg, err error) {
	if err = json.NewDecoder(r.Body).Decode(&d); err != nil {
		return d, errors.New("Invalid JSON payload")
	}

	t, _ := time.Parse(time.RFC1123, r.Header.Get("Date"))

	if err = checkMsgExpiry(t); err != nil {
		return
	}

	sig := r.Header.Get("X-N3-Signature")
	buf, _ := json.Marshal(d)

	if err = checkSignature(sig, buf); err != nil {
		stat.ReqErrSignature++
		return
	}

	if err = checkCommand(d.Command); err != nil {
		stat.ReqErrCommand++
		return
	}

	li.Uid = d.UserId

	ts := t.In(time.Local).Format(time.RFC1123)

	event(logdebug, li, "Request verified: [Timestamp: %v, User ID: %v, "+
		"Origin: %v, Command: %v]", ts, d.UserId, d.Origin, d.Command)
	return
}

func checkCommand(c string) (err error) {
	switch c {
	case "server-status":

	case "resolve-user":
	case "resolve-user-id":
	case "reset-user-pw":
	case "add-user":
	case "set-user-attr":
	case "enable-user":
	case "disable-user":
	case "activate-user":
	case "deactivate-user":
	case "list-user":
	case "get-user-list":

	case "register":
	case "login":
	case "logout":

	default:
		return errors.New("Invalid command: " + c)
	}

	return
}

func sendResponse(w http.ResponseWriter, m *Msg) {
	data := &Msg{HostName: app.HostName, UserId: li.Uid, MsgId: li.Msgid,
		Data: m.Data}
	buf, _ := json.Marshal(data)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-N3-Service-Name", "ghazal")
	w.Header().Add("X-N3-Signature", signRequest(buf, data.MsgId))

	fmt.Fprintf(w, "%s", buf)

	event(logdebug, li, "Server response sent: [Hostname: %v, User ID: "+
		"%v, Message ID: %v, Error code: %v]", data.HostName,
		data.UserId, data.MsgId, data.ErrNo)

}

func sendError(w http.ResponseWriter, i int, s string, e error) {
	event(logwarn, li, e.Error())
	stat.ReqError++

	data := &Msg{HostName: app.HostName, UserId: li.Uid, MsgId: li.Msgid,
		ErrNo: i, Data: s}
	buf, _ := json.Marshal(data)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-N3-Service-Name", "ghazal")
	w.Header().Add("X-N3-Signature", signRequest(buf, data.MsgId))

	fmt.Fprintf(w, "%s", buf)

	event(logdebug, li, "Server error response sent: [Hostname: %v, "+
		"User ID: %v, Message ID: %v, Error code: %v]", data.HostName,
		data.UserId, data.MsgId, data.ErrNo)
}

func sendMail(r []string, s, b string, f bool) (err error) {
	ghazal := mail.Address{"Ghazal Web Service", "ghazal@s.rebung.io"}
	admin := mail.Address{"Rebung.IO Administrator", app.AdminEmail}

	auth := smtp.PlainAuth("", app.SMTPUser, app.SMTPPw, app.SMTPHost)
	url := net.JoinHostPort(app.SMTPHost, "587")
	con := &smtp.Client{}

	if con, err = smtp.Dial(url); err != nil {
		return errors.New("SMTP server not available")
	}

	if err = con.StartTLS(tlsc); err != nil {
		return errors.New("StartTLS negotiation failed")
	}

	if err = con.Auth(auth); err != nil {
		return errors.New("SMTP Plain authentication failed")
	}

	header := make(map[string]string)

	if f {
		header["From"] = ghazal.String()
	} else {
		header["From"] = admin.String()
	}

	header["Subject"] = s
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	r = append(r, admin.String())

	if f {
		con.Mail(ghazal.String())
	} else {
		con.Mail(admin.String())
	}

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

	buf := bytes.NewBufferString(msg)

	if _, err = buf.WriteTo(wc); err != nil {
		return errors.New("Unable write message body")
	}

	return
}

func setupServer() (err error) {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/s/", defaultAdminUserHandler)
	http.HandleFunc("/u/", defaultUserHandler)

	for i := range app.Bind {
		b := net.JoinHostPort(app.Bind[i].Host, app.Bind[i].Port)

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
