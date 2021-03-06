/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	logdebug  string = "debug"
	loginfo   string = "info"
	lognotice string = "notice"
	logwarn   string = "warn"
	logcrit   string = "critical"
)

type LogInfo struct {
	Src string
	Uid int64
}

type RebungLog struct {
	Timestamp int64
	HostName  string
	ProgName  string
	Pid       int
	Priority  string
	Src       string
	UserId    int64
	Message   string
}

var (
	debug bool
	logfp *os.File
	li    *LogInfo
)

func setupLog(d bool) (err error) {
	debug = d

        li = &LogInfo{Src: "::1", Uid: 0}

	if debug {
		logfp = os.Stderr
		fmt.Fprintf(logfp, "Debugging enabled - redirect to stderr\n")
	}

	return
}

func event(p string, li *LogInfo, a string, v ...interface{}) {
	var n string

	if p == "logdebug" {
		if !debug {
			return
		}
	}

	if len(v) != 0 {
		n = fmt.Sprintf(a, v...)
	} else {
		n = a
	}

	writeLog(p, li, n)
	return
}

func writeLog(p string, li *LogInfo, str string) {
	if debug {
		fmt.Fprintf(logfp, "%v: %v %v[%v] %v[%v] [%v]%v\n",
			time.Now().Format(time.RFC1123), app.HostName,
			app.ProgName, app.Pid, p, li.Src, li.Uid, str)
	} else {
		buf := &RebungLog{Timestamp: time.Now().UnixNano(),
			HostName: app.HostName, ProgName: app.ProgName,
			Pid: app.Pid, Priority: p, Src: li.Src, UserId: li.Uid,
			Message: str}

		j, _ := json.Marshal(buf)
		fmt.Fprintf(logfp, "%s\n", j)
	}

	return
}

func fatal(a string, v ...interface{}) {
	var n string

	if len(v) != 0 {
		n = fmt.Sprintf(a, v...)
	} else {
		n = a
	}

	fmt.Fprintf(os.Stderr, "Fatal: "+n+"\n")
	os.Exit(1)
}

func warn(a string, v ...interface{}) {
	var n string

	if len(v) != 0 {
		n = fmt.Sprintf(a, v...)
	} else {
		n = a
	}

	fmt.Fprintf(os.Stderr, "Warn: "+n+"\n")
	return
}
