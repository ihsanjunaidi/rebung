/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type SessionInfo struct {
	Id         int64
	Uid        string
	Type       string
	Status     string
	ServerName string
	TunSrc     string
	TunDst     string
	Src        string
	Dst        string
	Rt         string

	Sid   int64
	ErrNo int
}

type SessionInfoList struct {
	Id    int64
	Entry []SessionInfo
}

type UserSessionInfo struct {
	Id         int64
	ServerId   string
	Type       string
	Status     string
	ServerName string
	TunSrc     string
	TunDst     string
	Src        string
	Dst        string
	Rt         string

	Sid   int64
	ErrNo int
}

type UserSessionInfoList struct {
	Id    int64
	Entry []UserSessionInfo
}

type UserServerInfo struct {
	Id         int64
	Name       string
	Alias      string
	Descr      string
	Entity     string
	Location   string
	AccessType string
	TunnelType string
	Addr       string
	PpPrefix   string
	RtPrefix   string

	Sid   int64
	ErrNo int
}

type UserServerInfoList struct {
	Id    int64
	Entry []UserServerInfo
}

func setSession() (err error) {
	if len(app.Cmd.Args) != 2 {
		return errors.New("Invalid argument format")
	}

	var svid, _ = strconv.ParseInt(app.Cmd.Args[0], 0, 64)

	var si = []Id{Id{Id: svid, Opt: app.Cmd.Args[1]}}
	var d, _ = json.Marshal(&IdList{Entry: si})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

        var cmd = app.Cmd.Command
	var id = fmt.Sprintf("%v", m.Id)
	var n []string

	if n, err = resolveServerId([]string{id}); err != nil {
		return
	}

	if cmd == "check-session" {
		if len(m.Entry) != 2 {
			return errors.New("Invalid result format")
		}

		for i := range m.Entry {
			var e = m.Entry[i]

			var rtt = time.Duration(e.Id)

			if e.ErrNo == EOK {
				event("Destination [%v] is alive: %v", e.Opt, rtt)
			} else {
				event("Destination [%v] is down", e.Opt)
			}
		}
	} else {
		var v string

		if cmd == "activate-session" {
			v = "activated"
		} else if cmd == "deactivate-session" {
			v = "deactivated"
		}

		var e = m.Entry[0]

		if e.ErrNo == EOK {
			event("Session [%v] at %v has been %v: %v", e.Id, n[0],
				v, e.Opt)
		} else {
			event("Session [%v] at %v cannot be %v", e.Id, n[0], v)
		}

	}
	return
}

func setSessionOwner() (err error) {
	if len(app.Cmd.Args) != 1 {
		return errors.New("Invalid argument format")
	}

	var list []Id

	if list, err = setIdParam([]string{app.Cmd.Args[0]}); err != nil {
		return
	}

	var d, _ = json.Marshal(&IdList{Entry: list})

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest(string(d), app.RebanaUrl); err != nil {
		return
	}

	var m *IdList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	var cmd = app.Cmd.Command
	var uid = fmt.Sprintf("%v", app.Cmd.UserId)

	var n, u []string

	if u, err = resolveUserId([]string{uid}); err != nil {
		return
	}

	if cmd == "reassign-session" {
		u[0] = "[-1]"
	}

	var id = fmt.Sprintf("%v", m.Id)

	if n, err = resolveServerId([]string{id}); err != nil {
		return
	}

	var e = m.Entry[0]

	if e.ErrNo == EOK {
		event("Session [%v] at %v is now assigned to %v", e.Id, n[0], u[0])
	} else {
		event("Tunnel session at %v cannot be assigned to %v", n[0], u[0])
	}

	return
}

func listUserSessions() (err error) {
	if len(app.Cmd.Args) != 0 {
		return errors.New("Incorrect number of arguments")
	}

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest("", app.RebanaUrl); err != nil {
		return
	}

	var m *UserSessionInfoList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.TunDst == "" {
			e.TunDst = "<Inactive>"
		}

		if e.ErrNo == EOK {
			event("Session [%v:%v] information:\n"+
				"------------------------\n"+
				"Session type: %v\n"+
				"Status: %v\n"+
				"Server: %v\n"+
				"Tunnel source: %v\n"+
				"Tunnel destination: %v\n"+
				"Tunnel inet6 source: %v\n"+
				"Tunnel inet6 destination: %v\n"+
				"Routed inet6 destination: %v\n", e.ServerId,
				e.Sid, e.Type, e.Status, e.ServerName, e.TunSrc,
				e.TunDst, e.Src, e.Dst, e.Rt)
		} else {
			event("[Session [%v:%v] not found]\n"+
				"---------------------\n", e.ServerId, e.Sid)
		}
	}

	return
}

func listUserServers() (err error) {
	if len(app.Cmd.Args) != 0 {
		return errors.New("Incorrect number of arguments")
	}

	var msg *RebanaMsg

	if msg, err = sendRebanaRequest("", app.RebanaUrl); err != nil {
		return
	}

	var m *UserServerInfoList

	if err = json.Unmarshal([]byte(msg.Data), &m); err != nil {
		return
	}

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.ErrNo == EOK {
			event("Server %v[%v] information:\n"+
				"------------------------\n"+
				"Alias: %v\n"+
				"Description: %v\n"+
				"Entity: %v\n"+
				"Location: %v\n"+
				"Access type: %v\n"+
				"Tunnel supported: %v\n"+
				"Tunnel source address: %v\n"+
				"Point-to-Point prefix: %v\n"+
				"Routed prefix: %v\n", e.Name, e.Id, e.Alias,
				e.Descr, e.Entity, e.Location, e.AccessType,
				e.TunnelType, e.Addr, e.PpPrefix, e.RtPrefix)
		}
	}

	return
}
