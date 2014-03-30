/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

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

type RequestMsg struct {
	UserId  int64
	Command string
	Data    string
}

type Msg struct {
	HostName string
	UserId   int64
	MsgId    int64
	ErrNo    int
	Data     string
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

	Bind   []BindInfo
	LogUrl string

	AdminEmail string
	SMTPHost   string
	SMTPUser   string
	SMTPPw     string

	Secret    string
	TLSCACert []string `json:"TLSCACert"`

	RedisUrl string
	RedisPw  string
	RedisDb  string
}

type AppStat struct {
	HostName                    string
	ReqAll                      int64
	ReqActivateSession          int64
	ReqDeactivateSession        int64
	ReqCheckSession             int64
	ReqAssignSession            int64
	ReqReassignSession          int64
	ReqListSession              int64
	ReqActivateUserSession      int64
	ReqDeactivateUserSession    int64
	ReqCheckUserSession         int64
	ReqListUserSession          int64
	ReqListUserServer           int64
	ReqResolveServer            int64
	ReqResolveServerId          int64
	ReqAddServer                int64
	ReqSetServerAttr            int64
	ReqEnableServer             int64
	ReqDisableServer            int64
	ReqActivateServer           int64
	ReqDeactivateServer         int64
	ReqListServer               int64
	ReqGetServerList            int64
	ReqGetUserList              int64
	ReqServerStatus             int64
	ReqServerInfo               int64
	ReqStatus                   int64
	ReqError                    int64
	ReqErrUrl                   int64
	ReqErrHeader                int64
	ReqErrRedis                 int64
	ReqErrPayload               int64
	ReqErrSignature             int64
	ReqErrUserId                int64
	ReqErrServerId              int64
	ReqErrSessionId             int64
	ReqErrMsgId                 int64
	ReqErrCommand               int64
	ReqErrData                  int64
	ReqErrActivateSession       int64
	ReqErrDeactivateSession     int64
	ReqErrCheckSession          int64
	ReqErrAssignSession         int64
	ReqErrReassignSession       int64
	ReqErrListSession           int64
	ReqErrActivateUserSession   int64
	ReqErrDeactivateUserSession int64
	ReqErrCheckUserSession      int64
	ReqErrListUserSession       int64
	ReqErrListUserServer        int64
	ReqErrResolveServer         int64
	ReqErrResolveServerId       int64
	ReqErrAddServer             int64
	ReqErrSetServerAttr         int64
	ReqErrEnableServer          int64
	ReqErrDisableServer         int64
	ReqErrActivateServer        int64
	ReqErrDeactivateServer      int64
	ReqErrListServer            int64
	ReqErrGetServerList         int64
	ReqErrGetUserList           int64
	ReqErrServerStatus          int64
	ReqErrServerInfo            int64
	ReqErrStatus                int64
}

const (
	APPNAME  = "rebana"
	APPVER   = "1.0.0"
	PIDFILE  = "/var/run/rebung/rebana.pid"
	CONFFILE = "/usr/local/etc/rebung/rebana.json"

	// result codes
	EOK    = 0
	EINVAL = 1
	EAGAIN = 2
	ENOENT = 3
	EPERM  = 4
)

var (
	app  *AppConfig
	stat *AppStat
)

func status(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqStatus++

	var buf, _ = json.Marshal(&stat)

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	stat.ReqAll++
	li.Msgid = 0

	var err error
	var str = "Invalid request"

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
	case "server-status":
		err = status(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Request [%v:%v] completed", d.Command, li.Msgid)
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

	if err := parseConfig(conf); err != nil {
		fatal(err.Error())
	}

	if err := setupLog(debug); err != nil {
		fatal(err.Error())
	}

	go sigHandler()

	var ch = channelHandler()

	event(loginfo, li, "%v-%v server started: %v", app.ProgName, app.Version,
		app.HostName)

	setupServer(ch)

	var pid = fmt.Sprintf("%v", app.Pid)

	if err := ioutil.WriteFile(PIDFILE, []byte(pid), 0644); err != nil {
		fatal(err.Error())
	}

	select {}
}

func channelHandler() chan<- ChMsg {
	var c = make(chan ChMsg)

	go func() {
		for {
			select {
			case s := <-c:
				if s.Type == chMsgFatal {
					fatal(s.Msg)
				} else {
					event(loginfo, li, s.Msg)
				}
			}
		}
	}()

	return c
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
