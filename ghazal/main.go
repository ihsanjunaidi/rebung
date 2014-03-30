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
	Origin  string
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
	HostName             string
	ReqAll               int64
	ReqResolveUser       int64
	ReqResolveUserId     int64
	ReqResetUserPw       int64
	ReqGetAccessToken    int64
	ReqAddUser           int64
	ReqSetUserAttr       int64
	ReqEnableUser        int64
	ReqDisableUser       int64
	ReqActivateUser      int64
	ReqDeactivateUser    int64
	ReqListUser          int64
	ReqGetUserList       int64
	ReqUserLogin         int64
	ReqUserLogout        int64
	ReqStatus            int64
	ReqError             int64
	ReqErrUrl            int64
	ReqErrHeader         int64
	ReqErrRedis          int64
	ReqErrPayload        int64
	ReqErrSignature      int64
	ReqErrPassword       int64
	ReqErrAccessToken    int64
	ReqErrUserId         int64
	ReqErrMsgId          int64
	ReqErrCommand        int64
	ReqErrData           int64
	ReqErrResolveUser    int64
	ReqErrResolveUserId  int64
	ReqErrResetUserPw    int64
	ReqErrGetAccessToken int64
	ReqErrAddUser        int64
	ReqErrSetUserAttr    int64
	ReqErrEnableUser     int64
	ReqErrDisableUser    int64
	ReqErrActivateUser   int64
	ReqErrDeactivateUser int64
	ReqErrListUser       int64
	ReqErrGetUserList    int64
	ReqErrUserLogin      int64
	ReqErrUserLogout     int64
	ReqErrStatus         int64
}

const (
	APPNAME  = "ghazal"
	APPVER   = "1.0.0"
	PIDFILE  = "/var/run/rebung/ghazal.pid"
	CONFFILE = "/usr/local/etc/rebung/ghazal.json"

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

func serverStatus(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqStatus++

	buf, _ := json.Marshal(&stat)
	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	stat.ReqAll++

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

	event(logdebug, li, "Processing command [%v]", d.Command)

	switch d.Command {
	case "server-status":
		err = serverStatus(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Command [%v] processing completed", d.Command)
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

	event(loginfo, li, "%v-%v server started: %v", app.ProgName, app.Version,
		app.HostName)

	if err := setupServer(); err != nil {
		fatal(err.Error())
	}

	pid := fmt.Sprintf("%v", app.Pid)

        if err := ioutil.WriteFile(PIDFILE, []byte(pid), 0644); err != nil {
		fatal(err.Error())
	}

	select {}
}

func sigHandler() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	signal := <-c

	event(lognotice, li, "Signal received: "+signal.String())

	switch signal {
	case syscall.SIGINT, syscall.SIGTERM:
		event(lognotice, li, "Terminating..")
		os.Remove(PIDFILE)
		os.Exit(0)
	}
}

func usage() {
	str := fmt.Sprintf("%v-%v\nusage: %v [-d] [-h] [-c config file]\n",
		APPNAME, APPVER, APPNAME)

	fmt.Fprintf(os.Stderr, str)
	os.Exit(1)
}
