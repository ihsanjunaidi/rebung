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
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

type RebanaRequest struct {
	UserId  int64
	Command string
	Data    string
}

type RebanaMsg struct {
	HostName string
	UserId   int64
	MsgId    int64
	ErrNo    int
	Data     string
}

type GhazalRequest struct {
	UserId  int64
	Origin  string
	Command string
	Data    string
}

type GhazalMsg struct {
	HostName    string
	UserId      int64
	RebanaMsgId int64
	ErrNo       int
	Data        string
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

func setTLSConfig() (t *tls.Config, err error) {
	var cert *x509.Certificate
	var data []byte
	var asn1 *pem.Block

	var cacert = []string{"/Users/ihsan/.ssl/cacert.pem"}
	var opts = x509.VerifyOptions{Roots: x509.NewCertPool()}

	for i := range cacert {

		if data, err = ioutil.ReadFile(cacert[i]); err != nil {
			return
		}

		if asn1, data = pem.Decode(data); asn1 == nil {
			return
		}

		if cert, err = x509.ParseCertificate(asn1.Bytes); err != nil {
			return
		}

		opts.Roots.AddCert(cert)
	}

	if _, err = cert.Verify(opts); err != nil {
		return
	}

	t = &tls.Config{RootCAs: opts.Roots}
	return
}

func sendRebanaRequest(s, url string) (d *RebanaMsg, err error) {
	var m = &RebanaRequest{UserId: app.Cmd.UserId, Command: app.Cmd.Command,
		Data: s}
	var buf, _ = json.Marshal(m)

	var req *http.Request

	if req, err = http.NewRequest("POST", url, bytes.NewReader(buf)); err != nil {
		return
	}

	var loc, _ = time.LoadLocation("Etc/GMT")

	req.Header.Add("Date", time.Now().In(loc).Format(time.RFC1123))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-N3-Service-Name", "rebana")
	req.Header.Add("X-N3-Signature", signRequest(buf))

	dumpRequest(req)

	var tlsc *tls.Config

	if tlsc, err = setTLSConfig(); err != nil {
		return
	}

	var con = &http.Client{}

	con.Transport = &http.Transport{TLSClientConfig: tlsc}

	var res *http.Response

	if res, err = con.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	dumpResponse(res)

	if err = checkMsgExpiry(res); err != nil {
		return
	}

	if err = json.NewDecoder(res.Body).Decode(&d); err != nil {
		return
	}

	if d.ErrNo != EOK {
		err = errors.New(d.Data)
		return
	}

	return
}

func sendGhazalRequest(s, url string) (d *GhazalMsg, err error) {
	var m = &GhazalRequest{UserId: app.Cmd.UserId, Origin: "localhost",
		Command: app.Cmd.Command, Data: s}

	var buf, _ = json.Marshal(m)

	var req *http.Request

	if req, err = http.NewRequest("POST", url, bytes.NewReader(buf)); err != nil {
		return
	}

	var loc, _ = time.LoadLocation("Etc/GMT")

	req.Header.Add("Date", time.Now().In(loc).Format(time.RFC1123))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-N3-Service-Name", "ghazal")
	req.Header.Add("X-N3-Signature", signRequest(buf))

	dumpRequest(req)

	var tlsc *tls.Config

	if tlsc, err = setTLSConfig(); err != nil {
		fatal(err.Error())
	}

	var con = &http.Client{}

	con.Transport = &http.Transport{TLSClientConfig: tlsc}

	var res *http.Response

	if res, err = con.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	dumpResponse(res)

	if err = checkMsgExpiry(res); err != nil {
		return
	}

	if err = json.NewDecoder(res.Body).Decode(&d); err != nil {
		return
	}

	if d.ErrNo != EOK {
		err = errors.New(d.Data)
		return
	}

	return
}

func signRequest(m []byte) (s string) {
	var dgst = hmac.New(sha256.New, []byte(app.Key))

	dgst.Write(m)

	return base64.StdEncoding.EncodeToString(dgst.Sum(nil))
}

func checkSignature(sig string, m []byte) (err error) {
	var dgst = hmac.New(sha256.New, []byte(app.Key))

	dgst.Write(m)

	var msig []byte

	if msig, err = base64.StdEncoding.DecodeString(sig); err != nil {
		return
	}

	if !hmac.Equal(msig, dgst.Sum(nil)) {
		return errors.New("Message signature does not match")
	}

	return
}

func setIdParam(args []string) (s []Id, err error) {
	s = make([]Id, len(args))

	for i := range args {
		var id int64

		if id, err = strconv.ParseInt(args[i], 0, 64); err != nil {
			return
		} else {
			s[i] = Id{Id: id}
		}
	}

	if len(s) == 0 {
		return s, errors.New("Invalid argument format")
	}

	return
}

func setNameParam(args []string) (s []Name, err error) {
	s = make([]Name, len(args))

	for i := range args {
		s[i] = Name{Name: args[i]}
	}

	if len(s) == 0 {
		return s, errors.New("Invalid argument format")
	}

	return
}

func setNameArgParam(args []string) (s []Name, err error) {
	s = make([]Name, len(args))

	for i := range args {
		var a = strings.Split(args[i], "=")

		if len(a) != 2 {
			return s, errors.New("Invalid argument format")
		}

		s[i] = Name{Name: a[0], Opt: a[1]}
	}

	if len(s) == 0 {
		return s, errors.New("Invalid argument format")
	}

	return
}

func checkMsgExpiry(r *http.Response) (err error) {
	var t time.Time

	if t, err = time.Parse(time.RFC1123, r.Header.Get("Date")); err != nil {
		return
	}

	if time.Now().Sub(t).Seconds() > 10.0 {
		return errors.New("Message has expired")
	}

	if debug {
		var ts = t.In(time.Local).Format(time.RFC1123)
		event("Response timestamped at: " + ts)
	}

	return nil
}

func checkIP(s string) (ip net.IP) {
	if ip = net.ParseIP(s); ip == nil {
		return
	}

	if !strings.Contains(ip.String(), ".") {
		return nil
	}

	return
}

func dumpRequest(r *http.Request) {
	if debug {
		var d, _ = httputil.DumpRequest(r, true)

		event(string(d))
	}
}

func dumpResponse(r *http.Response) {
	if debug {
		var d, _ = httputil.DumpResponse(r, true)

		event(string(d))
	}

}
