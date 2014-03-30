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
 * uid:[uid]
 * uid:activity-list
 * uid:login-list
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
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

var rdp *redis.Pool

func setRedisUserNew(s *UserInfo, ip string) (uid int64, pw string, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	if uid, err = getRedisUserId(); err != nil || uid == 0 {
		return uid, pw, errors.New("Unable to retrieve new user ID")
	}

	ukey := fmt.Sprintf("user:%v:id", s.Login)

	if err = checkRedisKeyExist(ukey); err == nil {
		return uid, pw, errors.New("User " + s.Login + " exists")
	}

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err == nil {
		return uid, pw, errors.New("User " + s.Login + " ID exists")
	}

	pw = generateTempPassword()

	var dgst []byte

	if dgst, err = bcrypt.GenerateFromPassword([]byte(pw), 10); err != nil {
		return uid, pw, errors.New("Error generating " + s.Login + " password")
	}

	rdb.Do("set", ukey, uid)
	rdb.Do("hmset", key, "id", uid, "name", s.Name, "login", s.Login,
		"password", string(dgst), "admin", "enabled", "status", "active",
		"registered", time.Now().Format(time.RFC1123))

	rdb.Do("rpush", "user:all-list", s.Id)
	rdb.Do("rpush", "user:enabled-list", s.Id)
	rdb.Do("rpush", "user:active-list", s.Id)
	rdb.Do("rpush", "user:new-list", s.Id)

	if err = setRedisUserActivityList(uid, "created", ip); err != nil {
		event(logwarn, li, err.Error())
	}

	err = nil

	event(logdebug, li, "User %v added: [%v]", s.Login, s.Id)
	return
}

func setRedisUserAttr(uid int64, field, value, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	rdb.Do("hset", key, field, value)

	action := fmt.Sprintf("attribute changed: %v", field)

	if err = setRedisUserActivityList(uid, action, ip); err != nil {
		event(logwarn, li, err.Error())
	}

	event(logdebug, li, "User [%v] has new [%v]: [%v]", uid, field, value)
	return
}

func setRedisUserFirstLogin(uid int64, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	rdb.Do("hset", key, "flogin", time.Now().Format(time.RFC1123))
	rdb.Do("lrem", "user:new:list", 0, uid)
	return
}

func setRedisUserActivityList(uid int64, act, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:activity-list", uid)
	val := fmt.Sprintf("%v;%v;%v", act, ip, time.Now().Format(time.RFC1123))

	rdb.Do("lpush", key, val)

	var cnt int64

	if cnt, err = redis.Int64(rdb.Do("llen", key)); err != nil {
		return errors.New("Error retrieving Redis key " + key)
	}

	if cnt > 1000 {
		rdb.Do("rpop", key)
	}

	return
}

func setRedisUserLoginList(uid int64, ip, act string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:login-list", uid)
	val := fmt.Sprintf("%v;%v;%v", act, ip, time.Now().Format(time.RFC1123))

	rdb.Do("lpush", key, val)

	var cnt int64

	if cnt, err = redis.Int64(rdb.Do("llen", key)); err != nil {
		return errors.New("Error retrieving Redis key " + key)
	}

	if cnt > 1000 {
		rdb.Do("rpop", key)
	}

	return
}

func setRedisUserAdminStatus(uid int64, f bool, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var status string

	enl := "user:enabled-list"
	dil := "user:disabled-list"

	if f {
		status = "enabled"

		rdb.Do("hset", key, "admin", status)
		rdb.Do("lrem", dil, 0, uid)
		rdb.Do("lrem", enl, 0, uid)
		rdb.Do("rpush", enl, uid)
	} else {
		status = "disabled"

		rdb.Do("hset", key, "admin", status)
		rdb.Do("lrem", enl, 0, uid)
		rdb.Do("lrem", dil, 0, uid)
		rdb.Do("rpush", dil, uid)
	}

	action := fmt.Sprintf("status changed: %v", status)

	if err = setRedisUserActivityList(uid, action, ip); err != nil {
		event(logwarn, li, err.Error())
	}

	event(logdebug, li, "User [%v] admin status is now %v", uid, status)
	return
}

func setRedisUserStatus(uid int64, f bool, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var status string

	acl := "user:active-list"
	inl := "user:inactive-list"

	if f {
		status = "active"

		rdb.Do("hset", key, "status", status)
		rdb.Do("lrem", inl, 0, uid)
		rdb.Do("lrem", acl, 0, uid)
		rdb.Do("rpush", acl, uid)
	} else {
		status = "inactive"

		rdb.Do("hset", key, "status", status)
		rdb.Do("lrem", acl, 0, uid)
		rdb.Do("lrem", inl, 0, uid)
		rdb.Do("rpush", inl, uid)
	}

	action := fmt.Sprintf("status changed: %v", status)

	if err = setRedisUserActivityList(uid, action, ip); err != nil {
		event(logwarn, li, err.Error())
	}

	event(logdebug, li, "User [%v] is now %v", uid, status)
	return
}

func setRedisUserSession(uid int64, tok, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:session", uid)

	if err = checkRedisKeyExist(key); err == nil {
		event(logwarn, li, "Redis key %v exists", key)
	}

	rdb.Do("hset", key, "session-key", tok)

	if err = setRedisUserLoginList(uid, ip, "login"); err != nil {
		return
	}

	event(logdebug, li, "User [%v] session info updated", uid)
	return nil
}

func getRedisMsgId(c string) (id int64, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := "msgid:next"

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if id, err = redis.Int64(rdb.Do("get", key)); err != nil {
		return id, errors.New("Error retrieving Redis key " + key)
	}

	rdb.Do("incr", key)
	return
}

func getRedisUserId() (uid int64, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := "uid:next"

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if uid, err = redis.Int64(rdb.Do("get", key)); err != nil {
		return uid, errors.New("Error retrieving Redis key " + key)
	}

	rdb.Do("incr", key)
	return
}

func getRedisUserIdFromLogin(login string) (uid int64, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("user:%v:id", login)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if uid, err = redis.Int64(rdb.Do("get", key)); err != nil || uid == 0 {
		return uid, errors.New("Error retrieving Redis key " + key)
	}

	return
}

func getRedisUserInfo(uid int64) (s *UserInfo, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var r []string

	if r, err = redis.Strings(rdb.Do("hmget", key, "id", "name", "login",
		"password", "admin", "status", "registered",
		"flogin")); err != nil || len(r) == 0 {
		return s, errors.New("Error retrieving Redis key " + key)
	}

	id, _ := strconv.ParseInt(r[0], 0, 64)

	s = &UserInfo{Id: id, Name: r[1], Login: r[2], Password: r[3],
		Admin: r[4], Status: r[5], Registered: r[6], FirstLogin: r[7]}

	if s.FirstLogin != "" {
		rdb.Do("hset", "flogin", time.Now().Format(time.RFC1123))
	}

	return
}

func getRedisUserList(s string) (l []string, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("user:%v-list", s)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if l, err = redis.Strings(rdb.Do("lrange", key, 0, -1)); err != nil ||
		len(l) == 0 {
		return l, errors.New("Error retrieving Redis key " + key)
	}

	return
}

func getRedisUserUidList(uid int64, s string) (l []string, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:%v-list", uid, s)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if l, err = redis.Strings(rdb.Do("lrange", key, 0, -1)); err != nil ||
		len(l) == 0 {
		return l, errors.New("Error retrieving Redis key " + key)
	}

	return
}

func getRedisUserSessionKey(uid int64) (s string, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:session", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if s, err = redis.String(rdb.Do("hget", key,
		"session-key")); err != nil || s == "" {
		return s, errors.New("Error retrieving Redis key " + key)
	}

	return
}

func deleteRedisUserSession(uid int64, ip string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:session", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if _, err = rdb.Do("del", key); err != nil {
		return errors.New("Error deleting Redis key " + key)
	}

	if err = setRedisUserLoginList(uid, ip, "logout"); err != nil {
		return
	}

	return
}

func checkRedisUserStatus(uid int64) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	if s := fmt.Sprintf("%v", uid); uid == 0 {
		return errors.New("Invalid user ID: " + s)
	}

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var s *UserInfo

	if s, err = getRedisUserInfo(uid); err != nil {
		return
	}

	if s.Admin != "enabled" {
		return errors.New("User is disabled: " + s.Name)
	}

	if s.Status != "active" {
		return errors.New("User is inactive: " + s.Name)
	}

	return
}

func checkRedisUserSession(uid int64) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:session", uid)

	if err = checkRedisKeyExist(key); err == nil {
		return errors.New("Redis key " + key + " exists")
	}

	return
}

func checkRedisKeyExist(key string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	var exist bool

	if exist, err = redis.Bool(rdb.Do("exists", key)); err != nil || !exist {
		return errors.New("Error retrieving Redis key " + key)
	}

	return
}

func checkRedisUserId(uid int64) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v", uid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	return
}

func checkRedisMsgId(id int64) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("msgid:%v", id)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	return
}

func checkRedis() (err error) {
	if rdp != nil {
		return
	}

	rdp = &redis.Pool{MaxIdle: 5, IdleTimeout: 300 * time.Second,
		Dial: func() (rdb redis.Conn, err error) {
			if rdb, err = redis.Dial("tcp", app.RedisUrl); err != nil {
				return
			}

			if _, err = rdb.Do("auth", app.RedisPw); err != nil {
				return
			}

			if _, err = rdb.Do("select", app.RedisDb); err != nil {
				return
			}

			return
		}}

	event(loginfo, li, "Connected to Redis: %v", app.RedisUrl)
	return
}
