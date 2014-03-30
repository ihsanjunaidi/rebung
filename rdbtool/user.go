/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

/*
 * User keys
 * -----------
 * user:[name]:id
 * user:all-list
 * user:enabled-list
 * user:disabled-list
 * user:active-list
 * user:inactive-list
 * user:admin-list
 * user:new-list
 *
 * uid:next
 * uid:[uid]:activity-list
 * uid:[uid]:login-list
 *
 * Messaging keys
 * --------------
 * msgid:next
 *
 * Session keys
 * ------------
 * uid:[uid]:session
 */

package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

func resetUser(db int) (err error) {
	const usermax = 2000

	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	rdb.Do("multi")

	rdb.Do("set", "msgid:next", 1)
	rdb.Do("set", "uid:next", 1001)

	t := time.Now().Format(time.RFC1123)

	var dgst []byte

	if dgst, err = bcrypt.GenerateFromPassword([]byte("password"), 10); err != nil {
		return
	}

	for i := 1001; i <= usermax; i++ {
		name := fmt.Sprintf("Rebung User %v", i)
		login := fmt.Sprintf("user%v@rebung.io", i)

		key := fmt.Sprintf("user:%v:id", login)
		rdb.Do("set", key, i)

		key = fmt.Sprintf("uid:%v", i)
		rdb.Do("hmset", key, "id", i, "name", name, "login", login,
			"password", string(dgst), "admin", "enabled", "status", "active",
			"registered", t)

		rdb.Do("rpush", "user:all-list", i)
		rdb.Do("rpush", "user:enabled-list", i)
		rdb.Do("rpush", "user:active-list", i)
		rdb.Do("rpush", "user:new-list", i)

		rdb.Do("incr", "uid:next")
	}

	for i := 1050; i < 1100; i++ {
		key := fmt.Sprintf("uid:%v", i)
		rdb.Do("hset", key, "status", "inactive")

		rdb.Do("rpush", "user:inactive-list", i)
		rdb.Do("lrem", key, "user:active-list", i)
	}

	for i := 1200; i < 1250; i++ {
		key := fmt.Sprintf("uid:%v", i)
		rdb.Do("hset", key, "admin", "disabled")

		rdb.Do("rpush", "user:disabled-list", i)
		rdb.Do("lrem", key, "user:enabled-list", i)
	}

	// create admin users
	rdb.Do("set", "user:ghazal@rebung.io:id", 101)
	rdb.Do("set", "user:rebanats@rebung.io:id", 102)
	rdb.Do("set", "user:rctl@rebung.io:id", 103)
	rdb.Do("set", "user:admin@rebung.io:id", 501)
	rdb.Do("set", "user:operator@rebung.io:id", 502)

	rdb.Do("hmset", "uid:101", "id", 101, "name", "Ghazal Web Service",
		"login", "ghazal@rebung.io", "password", string(dgst), "admin",
		"enabled", "status", "active", "registered", t)

	rdb.Do("hmset", "uid:102", "id", 102, "name", "Rebana Tunnel Server",
		"login", "rebanats@rebung.io", "password", string(dgst), "admin",
		"enabled", "status", "active", "registered", t)

	rdb.Do("hmset", "uid:103", "id", 103, "name", "Rebung Controller",
		"login", "rctl@rebung.io", "password", string(dgst), "admin",
		"enabled", "status", "active", "registered", t)

	rdb.Do("hmset", "uid:501", "id", 501, "name", "Rebung.IO Administrator",
		"login", "admin@rebung.io", "password", string(dgst), "admin",
		"enabled", "status", "active", "registered", t)

	rdb.Do("hmset", "uid:502", "id", 502, "name", "Rebung.IO Operator",
		"login", "operator@rebung.io", "password", string(dgst), "admin",
		"enabled", "status", "active", "registered", t)

	rdb.Do("rpush", "user:admin-list", "101", "102", "103", "501", "502")
	rdb.Do("rpush", "user:enabled-list", "101", "102", "103", "501", "502")
	rdb.Do("rpush", "user:active-list", "101", "102", "103", "501", "502")

	var res []interface{}

	if res, err = redis.Values(rdb.Do("exec")); err != nil || res == nil {
		return
	}

	return
}

func dumpUser(db int) (err error) {
	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	return
}

func existUser(db int, key string) (err error) {
	if _, err = rdb.Do("select", db); err != nil {
		return
	}

	var exist bool

	if exist, err = redis.Bool(rdb.Do("exists", key)); err != nil {
		return
	}

	if exist {
		fmt.Printf(key+" exists: %v\n", exist)
	} else {
		fmt.Printf(key+" does not exist: %v\n", exist)
	}

	return
}
