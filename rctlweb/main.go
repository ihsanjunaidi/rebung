/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

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
	LogUrl    string
	GhazalUrl string
	RebanaUrl string

	SessionSecret string
	GhazalSecret  string
	RebanaSecret  string

	AppRoot     string
	TemplateDir string

	TLSCACert []string `json:"TLSCACert"`

	RedisUrl string
	RedisPw  string
	RedisDb  string
}

type Msg struct {
	UserId int64
	ErrNo  int
	Data   string
}

const (
	APPNAME  string = "rctlweb"
	APPVER   string = "1.0.0"
	PIDFILE  string = "/var/run/rebung/rctlweb.pid"
	CONFFILE string = "/usr/local/etc/rebung/rctlweb.json"

	// result codes
	EOK    int = 0
	EINVAL int = 1
	EAGAIN int = 2
	ENOENT int = 3

	// application resources
	TPLDIR  string = "templates/"
	COOKIE  string = "RCTL_SESSION"
	CDOMAIN string = "ctl.s.rebung.io"
	CPATH   string = "/"
)

var app *AppConfig

func urlHome(w http.ResponseWriter, r *http.Request) (err error) {
	var v *RenderVar

	if _, v, err = urlCheck(w, r, "Rebung.IO Control Panel"); err != nil {
		return
	}

	return render(w, v, "home")
}

func urlProfile(w http.ResponseWriter, r *http.Request) (err error) {
	var v *RenderVar

	if _, v, err = urlCheck(w, r, "Rebumg.IO User Settings"); err != nil {
		return
	}

	return render(w, v, "profile")
}

func urlLogin(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, err = getSession(r); err != nil {
		event(logwarn, li, err.Error())
	}

	v := &RenderVar{Title: "Rebung.IO"}

	if s != nil && s.Id == "0" {
		_, exist := s.Flash["success"]

		if exist {
			v.Flash = make(map[string]string, 1)
			v.Flash["success"] = s.Flash["success"]
		} else {
			v.Flash = make(map[string]string, 1)
			v.Flash["error"] = s.Flash["error"]
		}

		deleteSession(w, "")
	}

	if s != nil {
		v.SessionId = s.Id
		v.UserId = s.UserId
		v.Username = s.Username
		v.Name = s.Name
	}

	return render(w, v, "index")
}

func urlLogout(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session

	if s, _, err = urlCheck(w, r, ""); err != nil {
		return
	}

	var idl *IdList

	if idl, err = userLogout(s.UserId, s.Key, li.Src); err != nil {
		redirectLogin(w, r, "Error logging out "+s.Username, s.Id, err)
		return
	}

	if idl.Entry[0].ErrNo == EOK {
		redirectLogin(w, r, "You have been logged out", s.Id, nil)
	} else {
		redirectLogin(w, r, "You cannot be logged out, forcing", s.Id,
			errors.New("Error logging out: "+s.Username))
	}

	return
}

func urlList(w http.ResponseWriter, r *http.Request) (err error) {
	var s *Session
	var v *RenderVar

	if s, v, err = urlCheck(w, r, "Rebung.IO Control Panel"); err != nil {
		return
	}

	var exist bool
	var id int64
	var uil *UserInfoList
	var vil *ServerInfoList

	estr := "Invalid list parameter"
	q := r.URL.Query()

	if _, exist = q["uid"]; exist {
		if id, err = strconv.ParseInt(q["uid"][0], 0, 64); err != nil {
			redirectUrl(w, r, s, "/home", estr, err)
		}

		if id == 0 {
			v.Users = true
		} else {
			if uil, err = listUser(s.UserId, []int64{id}, ""); err != nil {
				redirectUrl(w, r, s, "/home", estr, err)
				return
			}

			if len(uil.Entry) == 1 {
				var e = &uil.Entry[0]

				if e.Admin == "enabled" {
					e.AdminFlag = true
				}

				if e.Status == "active" {
					e.StatusFlag = true
				}

				if e.FirstLogin != "" {
					e.FirstLoginFlag = true
				}

				t, _ := time.Parse(time.RFC1123, e.Registered)
				y, m, d := t.Date()

				e.Registered = fmt.Sprintf("%v %v, %v",
					m.String(), d, y)
				v.User = e
			}
		}
	} else if _, exist = q["vid"]; exist {
		if id, err = strconv.ParseInt(q["vid"][0], 0, 64); err != nil {
			redirectUrl(w, r, s, "/home", estr, err)
		}

		if id == 0 {
			v.Servers = true
		} else {
			if vil, err = listServer(s.UserId, []int64{id},
				""); err != nil {
				redirectUrl(w, r, s, "/home", estr, err)
				return
			} else {
				e := &vil.Entry[0]

				if e.Admin == "enabled" {
					e.AdminFlag = true
				}

				if e.Status == "active" {
					e.StatusFlag = true
				}

				t, _ := time.Parse(time.RFC1123, e.Activated)
				y, m, d := t.Date()

				e.Activated = fmt.Sprintf("%v %v, %v", d,
					m.String(), y)
				v.Server = e
			}
		}
	} else {
		redirectUrl(w, r, s, "/home", estr, errors.New(estr))
		return
	}

	return render(w, v, "list")
}

func mainUrlHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if ip := r.Header.Get("X-Forwarded-For"); ip == "" {
		li.Src = r.RemoteAddr
	} else {
		li.Src = ip
	}

	if err = checkRedis(); err != nil {
		renderError(w, r, 500, err.Error())
		return
	}

	event(logdebug, li, "New connection from %v to %v", li.Src, r.URL.Path)

	switch r.URL.Path {
	case "/":
		err = urlLogin(w, r)

	case "/home":
		err = urlHome(w, r)

	case "/list":
		err = urlList(w, r)

	case "/search":
		err = formSearch(w, r)

	case "/add-server":
		err = formAddServer(w, r)

	case "/set-server-attr":
		err = wsSetServerAttr(w, r)

	case "/set-server-status":
		err = wsSetServerStatus(w, r)

	case "/list-server":
		err = wsListServer(w, r)

	case "/get-server-list":
		err = wsGetServerList(w, r)

	case "/add-user":
		err = formAddUser(w, r)

	case "/resolve-user":
		err = wsResolveUser(w, r)

	case "/set-user-attr":
		err = wsSetUserAttr(w, r)

	case "/set-user-status":
		err = wsSetUserStatus(w, r)

	case "/set-session-owner":
		err = wsSetSessionOwner(w, r)

	case "/set-user-session":
		err = wsSetUserSession(w, r)

	case "/reset-user-pw":
		err = wsResetUserPw(w, r)

	case "/list-user":
		err = wsListUser(w, r)

	case "/get-user-sessions":
		err = wsGetUserSessions(w, r)

	case "/get-user-list":
		err = wsGetUserList(w, r)

	case "/logout":
		err = urlLogout(w, r)

	case "/login":
		err = formLogin(w, r)

	case "/profile":
		err = urlProfile(w, r)

	case "/change-name":
		err = formChangeName(w, r)

	case "/change-pw":
		err = formChangePw(w, r)

	default:
		renderError(w, r, 404, "URL "+r.URL.Path+" does not exist")

	}

	if err != nil {
		event(logwarn, li, err.Error())
		return
	}
}

func main() {
	var help, debug bool
	var conf string

	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.BoolVar(&help, "h", false, "Display usage")
	flag.StringVar(&conf, "c", CONFFILE, "configuration file")

	flag.Parse()

	if help {
		usage()
	}

	app = &AppConfig{ProgName: APPNAME, Version: APPVER, Pid: os.Getpid()}

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

	parseTemplates()

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
