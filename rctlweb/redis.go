/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

/*
 * Session keys
 * ------------
 * sid:[id]
 */

package main

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

var rdp *redis.Pool

func setRedisSession(sid string, s *Session) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key1 := fmt.Sprintf("uid:%v:session", s.UserId)
	key2 := fmt.Sprintf("sid:%v:uid", sid)

	if _, err = rdb.Do("hmset", key1, "id", sid, "hash", s.Hash, "uid",
		s.UserId, "username", s.Username, "name", s.Name, "key",
		s.Key); err != nil {
		return errors.New(fmt.Sprintf("Error saving user [%v] session",
			s.UserId))
	}

	if _, err = rdb.Do("set", key2, s.UserId); err != nil {
		return errors.New(fmt.Sprintf("Error saving user [%v] session",
			s.UserId))
	}

	if _, err = rdb.Do("expire", key2, 3600); err != nil {
		return errors.New(fmt.Sprintf("Error saving user [%v] session",
			s.UserId))
	}

	return
}

func setRedisSessionName(id int64, name string) error {
	rdb := rdp.Get()
	defer rdb.Close()

	if _, err := rdb.Do("hset", fmt.Sprintf("uid:%v:session", id), "name",
		name); err != nil {
		return errors.New(fmt.Sprintf("Error updating user [%v] name", id))
	}

	return nil
}

func setRedisFlashMessage(id int64, c, msg string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	rdb.Do("hset", fmt.Sprintf("uid:%v:session", id), c, msg)
	return
}

func deleteRedisFlashMessage(id int64) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	rdb.Do("hdel", fmt.Sprintf("uid:%v:session", id), "error", "success")
	return
}

func deleteRedisSessionId(sid string) (err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	var key = fmt.Sprintf("sid:%v:uid", sid)

	var uid string

	if uid, err = redis.String(rdb.Do("get", key)); err != nil {
		return errors.New("Error retrieving Redis key " + key)
	}

	key1 := fmt.Sprintf("uid:%v:session", uid)
	key2 := fmt.Sprintf("sid:%v:uid", sid)

	rdb.Do("del", key1, key2)
	return
}

func getRedisFlashMessage(id int64) (s, t string, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("uid:%v:session", id)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var st []string

	if st, err = redis.Strings(rdb.Do("hmget", key, "success",
		"error")); err != nil {
		return s, t, errors.New("Error retrieving Redis key " + key)
	}

	if len(st) == 2 {
		if st[0] != "" {
			s = st[0]
			t = "success"
		} else {
			s = st[1]
			t = "error"
		}
	}

	return
}

func getRedisSession(sid string) (s *Session, err error) {
	rdb := rdp.Get()
	defer rdb.Close()

	key := fmt.Sprintf("sid:%v:uid", sid)

	if err = checkRedisKeyExist(key); err != nil {
		return
	}

	var uid string

	if uid, err = redis.String(rdb.Do("get", key)); err != nil || uid == "" {
		return s, errors.New("Error retrieving Redis key " + key)
	}

	var st []string

	key = fmt.Sprintf("uid:%v:session", uid)

	if st, err = redis.Strings(rdb.Do("hmget", key, "id", "hash", "uid",
                "username", "name", "key")); err != nil || len(st) == 0 {
		return s, errors.New("Error retrieving Redis key " + key)
	}

	id, _ := strconv.ParseInt(st[2], 0, 64)

        s = &Session{Id: st[0], Hash: st[1], UserId: id, Username: st[3],
                Name: st[4], Key: st[5]}

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
