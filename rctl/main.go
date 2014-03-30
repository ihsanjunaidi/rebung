/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"flag"
	"fmt"
	"os"
)

type Command struct {
	UserId  int64
	ReqIp   string
	Command string
	Args    []string
}

type AppVar struct {
	Key       string
	RebanaUrl string
	GhazalUrl string
	Cmd       *Command
}

const (
	APPNAME       = "rebungctl"
	APPVER        = "1.0.0"
	REBANABASEURL = "https://rebana.rebung.io/"
	GHAZALBASEURL = "https://ghazal.rebung.io/"

	// error codes
	EOK    = 0
	EINVAL = 1
	EAGAIN = 2
	ENOENT = 3
	EPERM  = 4
)

var (
	debug bool
	app   *AppVar
)

func main() {
	var help bool
	var err error
	var svc string

	app = &AppVar{}
	app.Cmd = &Command{}
	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.BoolVar(&help, "h", false, "Display usage")
	flag.Int64Var(&app.Cmd.UserId, "i", 0, "User ID")
	flag.StringVar(&app.Cmd.Command, "c", "", "Command to issue")
	flag.StringVar(&svc, "s", "rebana", "Service to configure")

	flag.Parse()

	if help {
		usage()
	}

	app.Cmd.Args = flag.Args()
	app.Key = "secret"

	if svc == "rebana" {
		switch app.Cmd.Command {
		case "activate-session", "deactivate-session", "check-session":
			app.RebanaUrl = REBANABASEURL + "s/set"
			err = setSession()

		case "assign-session", "reassign-session":
			app.RebanaUrl = REBANABASEURL + "s/assign"
			err = setSessionOwner()

		case "list-user-sessions":
			app.RebanaUrl = REBANABASEURL + "s/list"
			err = listUserSessions()

		case "list-user-servers":
			app.RebanaUrl = REBANABASEURL + "s/list"
			err = listUserServers()

		case "resolve-server":
			app.RebanaUrl = REBANABASEURL + "v/resolve"
			err = resolveServerName()

		case "resolve-server-id":
			app.RebanaUrl = REBANABASEURL + "v/resolve"
			err = resolveServerIds()

		case "add-server":
			app.RebanaUrl = REBANABASEURL + "v/add"
			err = addServer()

		case "set-server-attr":
			app.RebanaUrl = REBANABASEURL + "v/set"
			err = setServerAttr()

		case "enable-server", "disable-server", "activate-server",
		        "deactivate-server":
			app.RebanaUrl = REBANABASEURL + "v/set"
			err = setServerStatus()

		case "list-server":
			app.RebanaUrl = REBANABASEURL + "v/list"
			err = listServer()

		case "get-server-list":
			app.RebanaUrl = REBANABASEURL + "v/list"
			err = getServerList()

		case "tunnel-server-status":
			app.RebanaUrl = REBANABASEURL + "v/status"
			err = serverStatus()

		case "server-status":
			app.RebanaUrl = REBANABASEURL + "status"
			err = status()

		default:
			usage()

		}
	} else if svc == "ghazal" {
		switch app.Cmd.Command {
		case "resolve-user":
			app.GhazalUrl = GHAZALBASEURL + "s/resolve"
			err = resolveUserLogin()

		case "resolve-user-id":
			app.GhazalUrl = GHAZALBASEURL + "s/resolve"
			err = resolveUserIds()

		case "reset-user-pw":
			app.GhazalUrl = GHAZALBASEURL + "s/reset"
			err = resolveUserIds()

		case "add-user":
			app.GhazalUrl = GHAZALBASEURL + "s/add"
			err = addUser()

		case "set-user-attr":
			app.GhazalUrl = GHAZALBASEURL + "s/set"
			err = setUserAttr()

		case "disable-user", "enable-user", "activate-user",
		        "deactivate-user":
			app.GhazalUrl = GHAZALBASEURL + "s/set"
			err = setUserStatus()

		case "list-user":
			app.GhazalUrl = GHAZALBASEURL + "s/list"
			err = listUser()

		case "get-user-list":
			app.GhazalUrl = GHAZALBASEURL + "s/list"
			err = getUserList()

		case "register":
			app.GhazalUrl = GHAZALBASEURL + "u/register"
			err = addUser()

		case "login":
			app.GhazalUrl = GHAZALBASEURL + "u/login"
			err = userLogin()

		case "logout":
			app.GhazalUrl = GHAZALBASEURL + "u/logout"
			err = userLogout()

		case "server-status":
			app.GhazalUrl = GHAZALBASEURL + "status"
			err = status()

		default:
			usage()

		}
	} else {
		usage()
	}

	if err != nil {
		fatal(err.Error())
	}
}

func usage() {
	var str = fmt.Sprintf("%v-%v\nBase usage: %v [-d] [-h] [-c config file] "+
		"[-s service]\n\n", APPNAME, APPVER, APPNAME)

	str += fmt.Sprintf("Rebana usage\n" +
		"------------\n" +
		"-c resolve-server -i [auid] [name]\n" +
		"-c resolve-server-id -i [auid] [svid1],[svid2],..\n" +
		"-c add-server -i [auid] [attr1=val],[attr2=val],..\n" +
		"-c set-server-attr -i [auid] [attr1=val],[attr2=val],..\n" +
		"-c enable-server -i [auid] [svid1],[svid2],..\n" +
		"-c disable-server -i [auid] [svid1],[svid2],..\n" +
		"-c activate-server -i [auid] [svid1],[svid2],..\n" +
		"-c deactivate-server -i [auid] [svid1],[svid2],..\n" +
		"-c list-server -i [auid] [svid1],[svid2],..\n" +
                "-c list-server -i [auid] [0:list-name,page,entries,sort-field]\n" +
		"-c list-user-servers -i [uid]\n" +
		"-c list-user-sessions -i [uid]\n" +
		"-c get-server-list -i [auid] [svid:list-name,page,entries,sort-field]\n" +
                "-c activate-session -i [uid] [svid] [ip]\n" +
                "-c deactivate-session -i [uid] [svid] [ip]\n" +
                "-c check-session -i [uid] [svid] [ip]\n" +
                "-c assign-session -i [uid] [svid]\n" +
                "-c reassign-session -i [uid] [svid]\n" +
		"-c tunnel-server-status -i [auid] [svid]\n" +
		"-c server-status -i [auid]\n\n")

	str += fmt.Sprintf("Ghazal usage\n" +
		"------------\n" +
		"-c resolve-user -i [auid] [name]\n" +
		"-c resolve-user-id -i [auid] [uid1],[uid2],..\n" +
		"-c add-user -i [auid] [attr1=val],[attr2=val],..\n" +
		"-c set-user-attr -i [auid] [uid] [attr1=val],[attr2=val],..\n" +
		"-c enable-user -i [auid] [uid1],[uid2],..\n" +
		"-c disable-user -i [auid] [uid1],[uid2],..\n" +
		"-c activate-user -i [auid] [uid1],[uid2],..\n" +
		"-c deactivate-user -i [auid] [uid1],[uid2],..\n" +
		"-c list-user -i [auid] [uid1],[uid2],..\n" +
                "-c list-user -i [auid] [0:list-name,page,entries,sort-field]\n" +
		"-c get-user-list -i [auid] [uid:list-name,page,entries,sort-field]\n" +
		"-c login [login=val],[password=val]\n" +
		"-c logout -i [uid] [session-key]\n" +
		"-c server-status -i [auid]\n\n")

	fmt.Fprintf(os.Stderr, str)
	os.Exit(1)
}
