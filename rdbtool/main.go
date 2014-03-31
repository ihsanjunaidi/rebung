/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"os"
)

type SessionKey struct {
	MsgIdNext    string
	ServerIdNext string

	NServerAll      int
	NServerEnabled  int
	NServerDisabled int
	NServerActive   int
	NServerInactive int
}

const (
	APPNAME = "rdbtool"
	APPVER  = "0.1"
)

var rdb redis.Conn

func flush(db int) (err error) {
	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	if _, err = rdb.Do("flushdb"); err != nil {
		return
	}

	return
}

func checkRedis(url, pw string) (err error) {
	if rdb == nil {
		rdb, err = redis.Dial("tcp", url)
		if err != nil {
			return
		}

		if _, err = rdb.Do("auth", pw); err != nil {
			return
		}
	}

	return
}

func main() {
	var (
		err      error
		debug    bool
		help     bool
		redisUrl = "localhost:6379"
		redisPw  = "password"
		redisDb  int
	)

	flag.BoolVar(&debug, "d", true, "Debug mode")
	flag.BoolVar(&help, "h", false, "Display usage")
	flag.Parse()

	if help {
		usage()
	}

	if len(flag.Args()) > 3 {
		usage()
	}

	err = checkRedis(redisUrl, redisPw)
	if err != nil {
		fatal(err.Error())
	}
	defer rdb.Close()

	f := flag.Args()
	if f[0] == "session" {
		redisDb = 1
		switch f[1] {
		case "flush":
			err = flush(redisDb)
		case "reset":
			err = resetSession(redisDb)
		case "dump":
			err = dumpSession(redisDb)
		case "stat":
			err = statSession(redisDb)

		default:
			fatal("Invalid command")
		}
	} else if f[0] == "user" {
		redisDb = 2
		switch f[1] {
		case "flush":
			err = flush(redisDb)
		case "reset":
			err = resetUser(redisDb)
		case "dump":
			err = dumpUser(redisDb)
		case "exist":
			err = existUser(redisDb, f[2])

		default:
			fatal("Invalid command")
		}
	} else {
		fatal("Invalid operation")
	}

	if err != nil {
		fatal(err.Error())
	}
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

func usage() {
	str := fmt.Sprintf("%v-%v\nusage: %v [-d] [-h] op cmd arg\n", APPNAME,
		APPVER, APPNAME)

	fmt.Fprintf(os.Stderr, str)
	os.Exit(1)
}
