/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan@n3labs.my>
 */

/*
 * Server keys
 * -----------
 * server:[name]:id
 * server:all-list
 * server:enabled-list
 * server:disabled-list
 * server:active-list
 * server:inactive-list
 *
 * svid:next
 * svid:[svid]
 * svid:[svid]:sid:next
 * svid:[svid]:sid[sid]
 * svid:[svid]:all-users-list
 * svid:[svid]:all-sessions-list
 * svid:[svid]:assigned-sessions-list
 * svid:[svid]:unassigned-sessions-list
 * svid:[svid]:active-sessions-list
 * svid:[svid]:session-activity-list
 *
 * Messaging keys
 * --------------
 * msgid:next
 *
 * User keys
 * ---------
 * user:admin-list
 * uid:[uid]:sessions-list
 */

package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"strconv"
	"time"
)

func ip4() (s string) {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%v.%v.%v.%v", rand.Intn(255), rand.Intn(255),
		rand.Intn(255), rand.Intn(255))
}

func ip6() (pp, rt string) {
	rand.Seed(time.Now().UnixNano())
	n1 := rand.Intn(1279) + 8193
	n2 := rand.Intn(65535)
	n3 := rand.Intn(254)

	pp = fmt.Sprintf("%x:%x:%x::/48", n1, n2, n3)
	rt = fmt.Sprintf("%x:%x:%x::/48", n1, n2, n3+1)
	return
}

func resetSession(db int) (err error) {
	const servermax = 100
	const sessionmax = 1000

	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	rdb.Do("multi")

	rdb.Do("set", "msgid:next", 1)
	rdb.Do("set", "svid:next", 1)

	loc := "AIMS, Kuala Lumpur"
	org := "N3 Labs"
	alias := "Central Node"
	t := time.Now().Format(time.RFC1123)

	for i := 1; i <= servermax; i++ {
		name := fmt.Sprintf("tun%v.s.rebung.io", i)
		descr := fmt.Sprintf("N3 Labs Tunnel Server #%v", i)
		url := fmt.Sprintf("https://tun%v.s.rebung.io:443", i)

		if i < 10 {
			name = fmt.Sprintf("tun0%v.s.rebung.io", i)
			url = fmt.Sprintf("https://tun0%v.s.rebung.io:443", i)
		}

		key := fmt.Sprintf("server:%v:id", name)
		rdb.Do("set", key, i)

		pp, rt := ip6()

		key = fmt.Sprintf("svid:%v", i)
		rdb.Do("hmset", key, "id", i, "name", name, "alias", alias,
			"descr", descr, "admin", "enabled", "status", "active",
			"entity", org, "location", loc, "access", "Public",
			"tunnel", "6in4", "tunsrc", ip4(), "url", url, "ppprefix",
			pp, "rtprefix", rt, "activated", t)

		rdb.Do("rpush", "server:all-list", i)
		rdb.Do("rpush", "server:enabled-list", i)
		rdb.Do("rpush", "server:active-list", i)

		key = fmt.Sprintf("svid:%v:sid:next", i)
		rdb.Do("set", key, 1)

		rdb.Do("incr", "svid:next")

		for j := 1; j <= sessionmax; j++ {
			var idx = strconv.FormatInt(int64(j), 16)

			skey := fmt.Sprintf("svid:%v:sid:%v", i, j)
			rdb.Do("hmset", skey, "id", j, "uid", "", "status",
				"inactive", "type", "6in4", "dst", "", "idx",
				idx, "lactiont", t)

			rdb.Do("incr", key)

			skey = fmt.Sprintf("svid:%v:all-sessions-list", i)
			rdb.Do("rpush", skey, j)

			skey = fmt.Sprintf("svid:%v:unassigned-sessions-list", i)
			rdb.Do("rpush", skey, j)
		}
	}

	s := 500
	for i := 60; i < 85; i++ {
		key := fmt.Sprintf("svid:%v", i)
		rdb.Do("hset", key, "status", "inactive")

		rdb.Do("rpush", "server:inactive-list", i)
		rdb.Do("lrem", "server:active-list", 0, i)

		for j := 500; j < 650; j++ {
			ip := fmt.Sprintf("211.24.200.%v", i)

			skey := fmt.Sprintf("svid:%v:sid:%v", i, j)
			rdb.Do("hmset", skey, "id", j, "uid", i, "status",
				"active", "dst", ip)

			skey = fmt.Sprintf("svid:%v:active-sessions-list", i)
			rdb.Do("rpush", skey, j)
		}

		key = fmt.Sprintf("svid:%v:all-users-list", i)
		rdb.Do("rpush", key, i)

		key = fmt.Sprintf("svid:%v:assigned-sessions-list", i)
		rdb.Do("rpush", key, i)

		key = fmt.Sprintf("svid:%v:unassigned-sessions-list", i)
		rdb.Do("lrem", key, 0, i)

		key = fmt.Sprintf("uid:%v:sessions-list", i)
		val := fmt.Sprintf("%v:%v", i, s)
		rdb.Do("rpush", key, val)
		s++
	}

	rdb.Do("rpush", "user:admin-list", "101", "102", "103", "501", "502")

	// assign sessions
	uid := 125
	svid := 50
	sid := 250

	for ; uid < 175; uid++ {
		us := fmt.Sprintf("%v:%v", svid, sid)

		var sil = fmt.Sprintf("uid:%v:sessions-list", uid)
		rdb.Do("rpush", sil, us)

		var uil = fmt.Sprintf("svid:%v:all-users-list", svid)
		var ail = fmt.Sprintf("svid:%v:assigned-sessions-list", svid)
		var ril = fmt.Sprintf("svid:%v:unassigned-sessions-list", svid)

		rdb.Do("lrem", ril, 0, sid)
		rdb.Do("rpush", ail, sid)
		rdb.Do("rpush", uil, uid)

		sid++
	}

	rdb.Do("hmset", "svid:1", "tunsrc", "124.108.16.93", "ppprefix",
		"2400:3700:80::/48", "rtprefix", "2400:3700:81::/48")

	rdb.Do("hmset", "svid:2", "tunsrc", "124.108.16.94", "ppprefix",
		"2400:3700:82::/48", "rtprefix", "2400:3700:83::/48")

	var res []interface{}

	if res, err = redis.Values(rdb.Do("exec")); err != nil || res == nil {
		return
	}

	return
}

func dumpSession(db int) (err error) {
	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	return
}

func statSession(db int) (err error) {
	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	s := &SessionKey{}

	key := "msgid:next"
	s.MsgIdNext, _ = redis.String(rdb.Do("get", key))

	key = "svid:next"
	s.ServerIdNext, _ = redis.String(rdb.Do("get", key))

	key = "server-all-list"
	s.NServerAll, _ = redis.Int(rdb.Do("llen", key))

	key = "server-enabled-list"
	s.NServerEnabled, _ = redis.Int(rdb.Do("llen", key))

	key = "server-disabled-list"
	s.NServerDisabled, _ = redis.Int(rdb.Do("llen", key))

	key = "server-active-list"
	s.NServerActive, _ = redis.Int(rdb.Do("llen", key))

	key = "server-inactive-list"
	s.NServerInactive, _ = redis.Int(rdb.Do("llen", key))

	fmt.Printf("Redis DB statistics:\n"+
		"--------------------\n"+
		"Next message ID: %v\n"+
		"Next server ID: %v\n"+
		"Total servers: %v\n"+
		"Total enabled servers: %v\n"+
		"Total disabled servers: %v\n"+
		"Total active servers: %v\n"+
		"Total inactive servers: %v\n", s.MsgIdNext, s.ServerIdNext,
		s.NServerAll, s.NServerEnabled, s.NServerDisabled,
		s.NServerActive, s.NServerInactive)

	return
}
