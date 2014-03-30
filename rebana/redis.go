/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
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
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"time"
)

var rdp *redis.Pool

func setRedisServerNew(s *ServerInfo) (vid int64, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	if vid, err = getRedisServerId(); err != nil || vid == 0 {
		return vid, errors.New("Unable to retrieve new server ID")
	}

	var key = fmt.Sprintf("server:%v:id", s.Name)

	if err = checkRedisKeyExist(key); err == nil {
		return vid, errors.New(fmt.Sprintf("Server %v exists", s.Name))
	}

	key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err == nil {
		return vid, errors.New(fmt.Sprintf("Server %v exists", s.Name))
	}

	var t = time.Now().Format(time.RFC1123)

	rdb.Do("hmset", key, "id", vid, "name", s.Name, "alias", s.Alias,
		"descr", s.Descr, "admin", "disabled", "status", "inactive",
		"entity", s.Entity, "location", s.Location, "access",
		s.Access, "tunnel", s.Tunnel, "tunsrc", s.TunnelSrc,
		"url", s.Url, "ppprefix", s.PpPrefix, "rtprefix", s.RtPrefix,
		"activated", t)

	key = fmt.Sprintf("server:%v:id", s.Name)

	rdb.Do("set", key, s.Id)

	key = fmt.Sprintf("svid:%v:sid:next", s.Id)

	rdb.Do("set", key, 1)

	for i := 1; i <= 1000; i++ {
		var idx = strconv.FormatInt(int64(i), 16)
		var skey = fmt.Sprintf("svid:%v:sid:%v", s.Id, i)

		rdb.Do("hmset", skey, "id", i, "uid", "", "status", "inactive",
			"type", "6in4", "dst", "", "idx", idx, "lactiont", t)

		var asl = fmt.Sprintf("svid:%v:all-sessions-list", i)
		var usl = fmt.Sprintf("svid:%v:unassigned-sessions-list", i)

		rdb.Do("rpush", asl, i)
		rdb.Do("rpush", usl, i)

		rdb.Do("incr", key)
	}

	var all = "server:all-list"
	var enl = "server:enabled-list"
	var acl = "server:active-list"

	rdb.Do("rpush", all, s.Id)
	rdb.Do("rpush", enl, s.Id)
	rdb.Do("rpush", acl, s.Id)

	err = nil

	event(logdebug, li, "Tunnel server %v added: [%v]", s.Name, s.Id)
	return
}

func setRedisServerAttr(vid int64, field, value string) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	rdb.Do("hset", key, field, value)

	event(logdebug, li, "Server [%v] has new [%v]: [%v]", vid, field, value)
	return
}

func setRedisServerAdminStatus(vid int64, f bool) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var status string

	var enl = "server:enabled-list"
	var dil = "server:disabled-list"

	if f {
		status = "enabled"

		rdb.Do("hset", key, "admin", status)

		rdb.Do("lrem", dil, 0, vid)
		rdb.Do("lrem", enl, 0, vid)

		rdb.Do("rpush", enl, vid)
	} else {
		status = "disabled"

		rdb.Do("hset", key, "admin", status)

		rdb.Do("lrem", enl, 0, vid)
		rdb.Do("lrem", dil, 0, vid)

		rdb.Do("rpush", dil, vid)
	}

	event(logdebug, li, "Server [%v] admin status is now %v", vid, status)
	return
}

func setRedisServerStatus(vid int64, f bool) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var status string

	var acl = "server:active-list"
	var inl = "server:inactive-list"

	var t = time.Now().Format(time.RFC1123)

	if f {
		status = "active"

		rdb.Do("hmset", key, "status", status, "lactiont", t)

		rdb.Do("lrem", inl, 0, vid)
		rdb.Do("lrem", acl, 0, vid)

		rdb.Do("rpush", acl, vid)
	} else {
		status = "inactive"

		rdb.Do("hmset", key, "status", status, "lactiont", t)

		rdb.Do("lrem", acl, 0, vid)
		rdb.Do("lrem", inl, 0, vid)

		rdb.Do("rpush", inl, vid)
	}

	event(logdebug, li, "Server [%v] status is now %v", vid, status)
	return
}

func setRedisSessionOwner(uid, vid int64, f bool) (sid int64, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var st []string

	var sil = fmt.Sprintf("uid:%v:sessions-list", uid)

	// one session per server
	if st, err = redis.Strings(rdb.Do("lrange", sil, 0, -1)); err != nil {
		event(lognotice, li, "User [%v] has no tunnel session", uid)
	}

	err = errors.New(fmt.Sprintf("User [%v] already has tunnel session on "+
		"server [%v]", uid, vid))

	if len(st) > 0 {
		for i := range st {
			var tok = strings.Split(st[i], ":")
			var v, _ = strconv.ParseInt(tok[0], 0, 64)
			var s, _ = strconv.ParseInt(tok[1], 0, 64)

			if v == vid {
				if f {
					return
				} else {
					sid = s
				}
			}
		}
	}

	var uil = fmt.Sprintf("svid:%v:all-users-list", vid)
	var ail = fmt.Sprintf("svid:%v:assigned-sessions-list", vid)
	var ril = fmt.Sprintf("svid:%v:unassigned-sessions-list", vid)

	if f {
		var s string

		if s, err = redis.String(rdb.Do("lpop", ril)); err != nil ||
			s == "" {
			return sid, errors.New(fmt.Sprintf("Error retrieving "+
				"Redis key [%v]", ril))
		}

		sid, _ = strconv.ParseInt(s, 0, 64)
	}

	var action string

	var key = fmt.Sprintf("svid:%v:sid:%v", vid, sid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var us = fmt.Sprintf("%v:%v", vid, sid)

	if f {
		action = "assignment"

		rdb.Do("hset", key, "uid", uid)

		rdb.Do("lrem", sil, 0, us)
		rdb.Do("lrem", uil, 0, uid)
		rdb.Do("lrem", ail, 0, sid)

		rdb.Do("rpush", sil, us)
		rdb.Do("rpush", uil, uid)
		rdb.Do("rpush", ail, sid)
	} else {
		action = "reassignment"

		rdb.Do("hset", key, "uid", "-1")

		rdb.Do("lrem", sil, 0, us)
		rdb.Do("lrem", uil, 0, uid)
		rdb.Do("lrem", ail, 0, sid)
		rdb.Do("lrem", ril, 0, sid)

		rdb.Do("rpush", ril, sid)
	}

	if err = setRedisSessionActivityList(uid, vid, sid, action); err != nil {
		event(logwarn, li, err.Error())
	}

	event(logdebug, li, "Session [%v:%v] is now assigned to user [%v]", vid,
		sid, uid)
	return
}

func setRedisSessionStatus(vid, sid, uid int64, dst string, f bool) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v:sid:%v", vid, sid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var status, action string

	var acl = fmt.Sprintf("svid:%v:active-sessions-list", vid)

	if f {
		status = "active"
		action = "activation"

		rdb.Do("hmset", key, "dst", dst, "status", status)
		rdb.Do("lrem", acl, 0, sid)
		rdb.Do("rpush", acl, sid)
	} else {
		status = "inactive"
		action = "deactivation"

		rdb.Do("hmset", key, "dst", "", "status", status)
		rdb.Do("lrem", acl, 0, sid)
	}

	if err = setRedisSessionActivityList(uid, vid, sid, action); err != nil {
		event(logwarn, li, err.Error())
	}

	event(logdebug, li, "Session [%v:%v] is now %v", vid, sid, status)
	return
}

func setRedisSessionActivityList(uid, vid, sid int64, act string) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	const max = 1000

	var t = time.Now().Format(time.RFC1123)

	var key = fmt.Sprintf("svid:%v:session-activity-list", vid)
	var val = fmt.Sprintf("%v;%v;%v;%v", act, uid, sid, t)

	rdb.Do("lpush", key, val)

	var cnt int64

	if cnt, err = redis.Int64(rdb.Do("llen", key)); err != nil {
		return errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	if cnt > max {
		rdb.Do("rpop", key)
	}

	event(logdebug, li, "User [%v] session activity record updated", uid)
	return
}

func getRedisMsgId(c string) (id int64, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = "msgid:next"

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if id, err = redis.Int64(rdb.Do("get", key)); err != nil {
		return id, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	rdb.Do("incr", key)
	return
}

func getRedisServerId() (vid int64, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = "svid:next"

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if vid, err = redis.Int64(rdb.Do("get", key)); err != nil || vid == 0 {
		return vid, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	rdb.Do("incr", key)
	return
}

func getRedisServerIdFromName(host string) (vid int64, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("server:%v:id", host)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if vid, err = redis.Int64(rdb.Do("get", key)); err != nil || vid == 0 {
		return vid, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	return
}

func getRedisServerUrl(vid int64) (url string, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if url, err = redis.String(rdb.Do("hget", key, "url")); err != nil ||
		url == "" {
		return url, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	return
}

func getRedisServerInfo(vid int64) (s *ServerInfo, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var r []string

	r, err = redis.Strings(rdb.Do("hmget", key, "id", "name", "alias",
		"descr", "admin", "status", "entity", "location", "access",
		"tunnel", "tunsrc", "url", "ppprefix", "rtprefix", "activated"))
	if err != nil {
		return s, errors.New(fmt.Sprintf("Error retrieving server [%v] "+
			"info", vid))
	}

	if len(r) == 0 {
		return s, errors.New(fmt.Sprintf("Empty server [%v] data", vid))
	}

	var id, _ = strconv.ParseInt(r[0], 0, 64)
	var t, _ = time.Parse(time.RFC1123, r[14])

	s = &ServerInfo{Id: id, Name: r[1], Alias: r[2], Descr: r[3],
		Admin: r[4], Status: r[5], Entity: r[6], Location: r[7],
		Access: r[8], Tunnel: r[9], TunnelSrc: r[10], Url: r[11],
		PpPrefix: r[12], RtPrefix: r[13], Activated: r[14],
		RegDate: t.Unix()}

	return
}

func getRedisServerList(s string) (l []string, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("server:%v-list", s)

	if err = checkRedisKeyExist(key); err != nil {
		event(lognotice, li, "Empty server [%v] list", s)
		return
	}

	if l, err = redis.Strings(rdb.Do("lrange", key, 0,
		-1)); err != nil || len(l) == 0 {
		return l, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	if len(l) == 0 {
		event(lognotice, li, "Empty server list: %v", key)
	}

	return
}

func getRedisServerSvidList(vid int64, s string) (l []string, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v:%v-list", vid, s)

	if err = checkRedisKeyExist(key); err != nil {
		event(lognotice, li, "Empty server [%v] list", s)
		return
	}

	if l, err = redis.Strings(rdb.Do("lrange", key, 0,
		-1)); err != nil || len(l) == 0 {
		return l, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	return
}

func getRedisUserUidList(uid int64, s string) (l []string, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("uid:%v:%v-list", uid, s)

	if err = checkRedisKeyExist(key); err != nil {
		event(lognotice, li, "Empty user [%v] list", s)
		return
	}

	if l, err = redis.Strings(rdb.Do("lrange", key, 0,
		-1)); err != nil || len(l) == 0 {
		return l, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	return
}

func getRedisUserList(s string) (l []string, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("user:%v-list", s)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if l, err = redis.Strings(rdb.Do("lrange", key, 0,
		-1)); err != nil || len(l) == 0 {
		return l, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	return
}

func getRedisSessionId(vid int64) (sid int64, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v:sid:next", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	if sid, err = redis.Int64(rdb.Do("get", key)); err != nil {
		return sid, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	rdb.Do("incr", key)
	return
}

func getRedisSessionInfo(vid, sid int64) (s *SessionInfo, err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v:sid:%v", vid, sid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var r []string

	if r, err = redis.Strings(rdb.Do("hmget", key, "id", "uid", "type",
		"status", "dst", "idx")); err != nil || len(r) == 0 {
		return s, errors.New(fmt.Sprintf("Error retrieving Redis key "+
			"[%v]", key))
	}

	var id, _ = strconv.ParseInt(r[0], 0, 64)

	s = &SessionInfo{Id: id, Uid: r[1], Type: r[2], Status: r[3],
		TunDst: r[4], Idx: r[5]}

	return
}

func checkRedisKeyExist(key string) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var exist bool

	if exist, err = redis.Bool(rdb.Do("exists", key)); err != nil || !exist {
		return errors.New(fmt.Sprintf("Error retrieving Redis key [%v]",
			key))
	}

	return
}

func checkRedisServerId(vid int64) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	return
}

func checkRedisServerStatus(vid int64) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	if vid == 0 {
		return errors.New(fmt.Sprintf("Server [%v] is invalid", vid))
	}

	var key = fmt.Sprintf("svid:%v", vid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var s = &ServerInfo{}

	if s, err = getRedisServerInfo(vid); err != nil {
		return
	}

	if s.Admin != "enabled" {
		return errors.New(fmt.Sprintf("Server [%v] is disabled", vid))
	}

	if s.Status != "active" {
		return errors.New(fmt.Sprintf("Server [%v] is inactive", vid))
	}

	return
}

func checkRedisSessionId(vid, sid int64) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("svid:%v:sid:%v", vid, sid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	return
}

func checkRedisMsgId(id int64) (err error) {
	var rdb = rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("msgid:%v", id)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	return
}

func checkRedis() (err error) {
	if rdp == nil {
		rdp = &redis.Pool{MaxIdle: 5, IdleTimeout: 300 * time.Second,
			Dial: func() (rdb redis.Conn, err error) {
				if rdb, err = redis.Dial("tcp",
					app.RedisUrl); err != nil {
					return
				}

				if _, err = rdb.Do("auth",
					app.RedisPw); err != nil {
					return
				}

				if _, err = rdb.Do("select",
					app.RedisDb); err != nil {
					return
				}

				return
			}}

		event(loginfo, li, "Connected to Redis: %v", app.RedisUrl)
	}

	return
}
