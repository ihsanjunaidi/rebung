/*
* Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type RequestMsg struct {
	Id      int64
	UserId  int64
	MsgId   int64
	Command string
	Data    string
}

type Msg struct {
	Id    int64
	MsgId int64
	ErrNo int
	Data  string
}

type RebanaRequestMsg struct {
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

type Id struct {
	Id    int64
	ErrNo int
	Opt   string
}

type IdList struct {
	Id    int64
	Entry []Id
}

type ServerInfo struct {
	Id       int64
	TunSrc   string
	PpPrefix string
	RtPrefix string
	Session  []Session
}

type Session struct {
	Id  int64
	Dst string
	Idx int64
}

type BindInfo struct {
	Host string
	Port string
}

type AppConfig struct {
	HostName string `json:"ServerName"`
	ProgName string
	Version  string
	Pid      int

	Bind      []BindInfo
	RebanaUrl string
	LogUrl    string

	Secret    string
	TLSCACert []string `json:"TLSCACert"`

	SvInfo *ServerInfo
}

type AppStat struct {
	HostName         string
	ReqAll           int64
	ReqActivate      int64
	ReqDeactivate    int64
	ReqCheck         int64
	ReqStatus        int64
	ReqError         int64
	ReqErrUrl        int64
	ReqErrHeader     int64
	ReqErrPayload    int64
	ReqErrSignature  int64
	ReqErrServerId   int64
	ReqErrUserId     int64
	ReqErrMsgId      int64
	ReqErrCommand    int64
	ReqErrData       int64
	ReqErrActivate   int64
	ReqErrDeactivate int64
	ReqErrCheck      int64
	ReqErrStatus     int64
}

const (
	APPNAME  = "rebanats"
	APPVER   = "1.0.0"
	PIDFILE  = "/var/run/rebanats.pid"
	CONFFILE = "/usr/local/etc/rebanats.json"

	// error codes
	EOK    = 0
	EINVAL = 1
	EAGAIN = 2
	ENOENT = 3
)

var (
	app  *AppConfig
	stat *AppStat
	tlsc *tls.Config
)

func activate(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqActivate++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrActivate++
		return
	}

	var e = m.Entry[0]

	if e.Id == 0 {
		stat.ReqErrActivate++
		return
	}

	var ipf int

	if ipf, err = checkIPFamily(e.Opt); ipf != 4 {
		stat.ReqErrActivate++
		return
	}

	var pp = strings.Split(app.SvInfo.PpPrefix, "::/")
	var rt = strings.Split(app.SvInfo.RtPrefix, "::/")

	var ifname = fmt.Sprintf("gif%v", e.Id)
	var in6addr = fmt.Sprintf("%v:%v::1", pp[0], e.Id)
	var in6dest = fmt.Sprintf("%v:%v::2", pp[0], e.Id)
	var rt6dest = fmt.Sprintf("%v:%v::", rt[0], e.Id)

	var s = &CSession{Ifname: ifname, TunAddr: app.SvInfo.TunSrc,
		TunDest: e.Opt, In6Addr: in6addr, In6Dest: in6dest,
		In6Plen: uint(64), Rt6Dest: rt6dest, Rt6Plen: uint(64)}

	event(logdebug, li, "Session [%v] activation parameters: [Interface "+
		"name: %v, Tunnel source: %v, Tunnel destination: %v, Server "+
		"inet6 address: %v/64, Client inet6 address: %v/64, Routed "+
		"prefix: %v/64]", e.Id, ifname, app.SvInfo.TunSrc, e.Opt,
		in6addr, in6dest, rt6dest)

	var si []Id

	if err = activateSession(s); err != nil {
		si = []Id{Id{ErrNo: EINVAL}}
		event(logwarn, li, err.Error())
	} else {
		si = []Id{Id{}}
	}

	var buf, _ = json.Marshal(&IdList{Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return nil
}

func deactivate(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqDeactivate++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrDeactivate++
		return
	}

	var e = m.Entry[0]

	if e.Id == 0 {
		stat.ReqErrDeactivate++
		return
	}

	var ipf int

	if ipf, err = checkIPFamily(e.Opt); ipf != 4 {
		stat.ReqErrDeactivate++
		return
	}

	var pp = strings.Split(app.SvInfo.PpPrefix, "::/")
	var rt = strings.Split(app.SvInfo.RtPrefix, "::/")

	var ifname = fmt.Sprintf("gif%v", e.Id)
	var in6addr = fmt.Sprintf("%v:%v::1", pp[0], e.Id)
	var in6dest = fmt.Sprintf("%v:%v::2", pp[0], e.Id)
	var rt6dest = fmt.Sprintf("%v:%v::", rt[0], e.Id)

	var s = &CSession{Ifname: ifname, TunAddr: app.SvInfo.TunSrc,
		TunDest: e.Opt, In6Addr: in6addr, In6Dest: in6dest,
		In6Plen: uint(64), Rt6Dest: rt6dest, Rt6Plen: uint(64)}

	event(logdebug, li, "Session [%v] deactivation parameters: [Interface "+
		"name: %v, Tunnel source: %v, Tunnel destination: %v, Server "+
		"inet6 address: %v/64, Client inet6 address: %v/64, Routed "+
		"prefix: %v/64]", e.Id, ifname, app.SvInfo.TunSrc, e.Opt,
		in6addr, in6dest, rt6dest)

	var si []Id

	if err = deactivateSession(s); err != nil {
		si = []Id{Id{ErrNo: EINVAL}}
		event(logwarn, li, err.Error())
	} else {
		si = []Id{Id{}}
	}

	var buf, _ = json.Marshal(&IdList{Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return nil
}

func check(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqCheck++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrCheck++
		return
	}

	var e = m.Entry[0]

	if e.Id == 0 {
		stat.ReqErrCheck++
		return
	}

	var ipf int

	if ipf, err = checkIPFamily(e.Opt); ipf != 4 {
		stat.ReqErrCheck++
		return
	}

	var pp = strings.Split(app.SvInfo.PpPrefix, "::/")
	//var ip6s = fmt.Sprintf("%v:%v::1", pp[0], e.Id)
	var ip6c = fmt.Sprintf("%v:%v::2", pp[0], e.Id)

	var si = make([]Id, 2)

	for i := range si {
		var dst = e.Opt
		var udp = "udp4"

		if i == 1 {
			udp = "udp6"
			dst = ip6c
		}

		if rtt, tgt, err := pingSession(dst, udp); err != nil {
			si[i] = Id{ErrNo: EINVAL, Opt: tgt}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: int64(rtt), Opt: tgt}
		}
	}

	var buf, _ = json.Marshal(&IdList{Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return nil
}

func status(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqStatus++

	var buf, _ = json.Marshal(&stat)

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func serverInfo() (err error) {
	var data = &RebanaRequestMsg{UserId: 102, Command: "server-info",
		Data: app.HostName}
	var buf, _ = json.Marshal(data)

	var url = app.RebanaUrl + "/v/info"
	var rd = bytes.NewReader(buf)

	var req *http.Request

	if req, err = http.NewRequest("POST", url, rd); err != nil {
		return
	}

	var loc = &time.Location{}

	if loc, err = time.LoadLocation("Etc/GMT"); err != nil {
		return
	}

	req.Header.Add("Date", time.Now().In(loc).Format(time.RFC1123))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-N3-Service-Name", "rebana")
	req.Header.Add("X-N3-Tunnel-Server", app.HostName)
	req.Header.Add("X-N3-Signature", signRequest(buf, 0))

	var c = &http.Client{}

	c.Transport = &http.Transport{TLSClientConfig: tlsc}

	var res *http.Response

	if res, err = c.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	var t time.Time

	if t, err = checkResponseHeader(res); err != nil {
		return
	}

	var msg = &RebanaMsg{}

	if err = json.NewDecoder(res.Body).Decode(msg); err != nil {
		return
	}

	var sig = res.Header.Get("X-N3-Signature")

	buf, _ = json.Marshal(msg)

	if err = checkSignature(sig, buf); err != nil {
		stat.ReqErrSignature++
		return
	}

	if msg.ErrNo != EOK {
		return errors.New(msg.Data)
	}

	var sv = &ServerInfo{}

	if err = json.Unmarshal([]byte(msg.Data), sv); err != nil {
		return
	}

	var ipf int

	if ipf, err = checkIPFamily(sv.TunSrc); ipf != 4 {
		return
	}

	if !strings.Contains(sv.PpPrefix, "::/48") ||
		!strings.Contains(sv.RtPrefix, "::/48") {
		return
	}

	app.SvInfo = &ServerInfo{}
	app.SvInfo.Id = sv.Id
	app.SvInfo.TunSrc = sv.TunSrc
	app.SvInfo.PpPrefix = sv.PpPrefix
	app.SvInfo.RtPrefix = sv.RtPrefix
	app.SvInfo.Session = sv.Session

	var sl = "Active tunnel session(s): "

	for i := range sv.Session {
		var e = sv.Session[i]

		var sa = fmt.Sprintf("%v[%v]", e.Id, e.Dst)

		sl += sa

		if i < (len(sv.Session) - 1) {
			sl += ", "
		}
	}

	var ts = t.In(time.Local).Format(time.RFC1123)

	event(logdebug, li, "Server information retrieved: "+
		"[Timestamp: %v, Server name: %v[%v], "+
		"Tunnel source address: %v, "+
		"Tunnel point-to-point prefix: %v, "+
		"Tunnel routed prefix: %v, Active tunnel sessions: %v]",
		ts, app.HostName, sv.Id, sv.TunSrc, sv.PpPrefix, sv.RtPrefix,
		len(sv.Session))
	return
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	stat.ReqAll++

	var err error
	var str = "Invalid request"

	if err = checkUrl(r); err != nil {
		stat.ReqErrHeader++
		sendError(w, EINVAL, str, err)
		return
	}

	if err = checkHeader(r); err != nil {
		stat.ReqErrHeader++
		sendError(w, EINVAL, str, err)
		return
	}

	var d *RequestMsg

	if d, err = checkData(r); err != nil {
		stat.ReqErrPayload++
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Processing request [%v:%v]", d.Command, d.MsgId)

	switch d.Command {
	case "activate":
		err = activate(w, d)

	case "deactivate":
		err = deactivate(w, d)

	case "check":
		err = check(w, d)

	case "status":
		err = status(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Request [%v:%v] completed", d.Command, d.MsgId)
}

func main() {
	var help, debug bool
	var conf string

	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.BoolVar(&help, "h", false, "Display usage")
	flag.StringVar(&conf, "c", CONFFILE, "Configuration file")

	flag.Parse()

	if help {
		usage()
	}

	app = &AppConfig{ProgName: APPNAME, Version: APPVER, Pid: os.Getpid()}
	stat = &AppStat{HostName: app.HostName}

	var err error

	if err = parseConfig(conf); err != nil {
		fatal(err.Error())
	}

	if err = setupLog(debug); err != nil {
		fatal(err.Error())
	}

	go sigHandler()

	event(loginfo, li, "%v-%v server started: %v", app.ProgName, app.Version,
		app.HostName)

	if err = setupServer(); err != nil {
		fatal(err.Error())
	}

	if err = serverInfo(); err != nil {
		fatal(err.Error())
	}

	var pid = fmt.Sprintf("%v", app.Pid)

	if err = ioutil.WriteFile(PIDFILE, []byte(pid), 0644); err != nil {
		fatal(err.Error())
	}

	select {}
}

func sigHandler() {
	var quit bool

	var c = make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	var signal = <-c

	event(lognotice, li, "Signal received: "+signal.String())

	switch signal {
	case syscall.SIGINT, syscall.SIGTERM:
		quit = true
	}

	if quit {
		event(lognotice, li, "Terminating..")

		os.Remove(PIDFILE)
		os.Exit(0)
	}

}

func usage() {
	var str = fmt.Sprintf("%v-%v\nusage: %v [-d] [-h] [-c config file]\n",
		APPNAME, APPVER, APPNAME)

	fmt.Fprintf(os.Stderr, str)
	os.Exit(1)
}
