/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type SessionInfo struct {
	Id          int64
	Uid         string
	Type        string
	Status      string
	ServerName  string
	TunSrc      string
	TunDst      string
	Src         string
	Dst         string
	Rt          string
	Idx         string
	LastActionT string

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

func activateSession(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqActivateSession++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrActivateSession++
		return
	}

        var ipf int
	var sid int64
	var e = m.Entry[0]

        if ipf, err = checkIPFamily(e.Opt); ipf != 4 {
		stat.ReqErrActivateSession++
		return
        }

	if err = checkRedisServerStatus(e.Id); err != nil {
		stat.ReqErrActivateSession++
		return
	}

	if sid, err = getUserSession(d.UserId, e.Id); err != nil {
		stat.ReqErrActivateSession++
		return
	}

	var s *SessionInfo

	if s, err = getRedisSessionInfo(e.Id, sid); err != nil {
		stat.ReqErrActivateSession++
		return
	}

	var idx, _ = strconv.ParseInt(s.Idx, 16, 64)
	var uid, _ = strconv.ParseInt(s.Uid, 0, 64)

	if err = setRedisSessionStatus(e.Id, sid, uid, e.Opt, true); err != nil {
		stat.ReqErrActivateSession++
		return
	}

	var data = &IdList{Entry: []Id{Id{Id: idx, Opt: e.Opt}}}
	var buf, _ = json.Marshal(data)

	var req = &TSReqMsg{Id: e.Id, UserId: li.Uid, MsgId: li.Msgid,
		Command: "activate", Data: string(buf)}

	var url string

	if url, err = getRedisServerUrl(e.Id); err != nil {
		stat.ReqErrActivateSession++
		return
	}

	url += "/activate"

	if _, err = sendTSRequest(url, req); err != nil {
		stat.ReqErrActivateSession++
		return
	}

	buf, _ = json.Marshal(&IdList{Id: e.Id, Entry: []Id{Id{Id: sid,
                Opt: e.Opt}}})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func deactivateSession(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqDeactivateSession++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

        var ipf int
	var sid int64
	var e = m.Entry[0]

        if ipf, err = checkIPFamily(e.Opt); ipf != 4 {
		stat.ReqErrActivateSession++
		return
        }

	if err = checkRedisServerStatus(e.Id); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	if sid, err = getUserSession(d.UserId, e.Id); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	var s *SessionInfo

	if s, err = getRedisSessionInfo(e.Id, sid); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	var idx, _ = strconv.ParseInt(s.Idx, 16, 64)
	var uid, _ = strconv.ParseInt(s.Uid, 0, 64)

	if err = setRedisSessionStatus(e.Id, sid, uid, e.Opt, false); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	var data = &IdList{Entry: []Id{Id{Id: idx, Opt: e.Opt}}}
	var buf, _ = json.Marshal(data)

	var req = &TSReqMsg{Id: e.Id, UserId: li.Uid, MsgId: li.Msgid,
		Command: "deactivate", Data: string(buf)}

	var url string

	if url, err = getRedisServerUrl(e.Id); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	url += "/deactivate"

	if _, err = sendTSRequest(url, req); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	buf, _ = json.Marshal(&IdList{Id: e.Id, Entry: []Id{Id{Id: sid,
                Opt: e.Opt}}})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func checkSession(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqCheckSession++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrCheckSession++
		return
	}

        var ipf int
	var sid int64
	var e = m.Entry[0]

        if ipf, err = checkIPFamily(e.Opt); ipf != 4 {
		stat.ReqErrActivateSession++
		return
        }

	if err = checkRedisServerStatus(e.Id); err != nil {
		stat.ReqErrCheckSession++
		return
	}

	if sid, err = getUserSession(d.UserId, e.Id); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	var s *SessionInfo

	if s, err = getRedisSessionInfo(e.Id, sid); err != nil {
		stat.ReqErrDeactivateSession++
		return
	}

	var idx, _ = strconv.ParseInt(s.Idx, 16, 64)

	var data = &IdList{Entry: []Id{Id{Id: idx, Opt: e.Opt}}}
	var buf, _ = json.Marshal(data)

	var req = &TSReqMsg{Id: e.Id, UserId: li.Uid, MsgId: li.Msgid,
		Command: "check", Data: string(buf)}

	var url string

	if url, err = getRedisServerUrl(e.Id); err != nil {
		stat.ReqErrCheckSession++
		return
	}

	url += "/check"

	var res *TSMsg

	if res, err = sendTSRequest(url, req); err != nil {
		stat.ReqErrCheckSession++
		return
	}

	var idl *IdList

	if idl, err = getIdList(res.Data, "ts-check-session"); err != nil {
		stat.ReqErrCheckSession++
		return
	}

	var si = make([]Id, len(idl.Entry))

	for i := range idl.Entry {
		if idl.Entry[i].ErrNo == EOK {
			si[i] = Id{Id: idl.Entry[i].Id, Opt: idl.Entry[i].Opt}
		} else {
			si[i] = Id{ErrNo: idl.Entry[i].ErrNo, Opt: idl.Entry[i].Opt}
		}
	}

	buf, _ = json.Marshal(&IdList{Id: e.Id, Entry: si})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func assignSession(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqAssignSession++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrAssignSession++
		return
	}

	var sid int64
	var e = m.Entry[0]

	if sid, err = setRedisSessionOwner(d.UserId, e.Id, true); err != nil {
		return
	}

	var buf, _ = json.Marshal(&IdList{Id: e.Id, Entry: []Id{Id{Id: sid}}})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func reassignSession(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqReassignSession++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrReassignSession++
		return
	}

	var sid int64
	var e = m.Entry[0]

	if sid, err = setRedisSessionOwner(d.UserId, e.Id, false); err != nil {
		return
	}

	var buf, _ = json.Marshal(&IdList{Id: e.Id, Entry: []Id{Id{Id: sid}}})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func listUserSessions(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqListUserSession++

	var m []string

	if m, err = getRedisUserUidList(d.UserId, "sessions"); err != nil {
		stat.ReqErrListUserSession++
		return
	}

	var si = make([]UserSessionInfo, len(m))

	for i := range m {
		var e = strings.Split(m[i], ":")

		if len(e) != 2 {
			continue
		}

		var vid, _ = strconv.ParseInt(e[0], 0, 64)
		var sid, _ = strconv.ParseInt(e[1], 0, 64)

		var v *ServerInfo

		if v, err = getRedisServerInfo(vid); err != nil {
			event(logwarn, li, err.Error())
		}

		var s *SessionInfo

		if s, err = getRedisSessionInfo(vid, sid); err != nil {
			si[i] = UserSessionInfo{Sid: sid, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		}

		var pp = strings.Split(v.PpPrefix, "::/")
		var rt = strings.Split(v.RtPrefix, "::/")

		si[i] = UserSessionInfo{Id: s.Id, ServerId: e[0], Type: s.Type,
			Status: s.Status, ServerName: v.Name, TunSrc: v.TunnelSrc,
			TunDst: s.TunDst}

		si[i].Src = pp[0] + ":" + s.Idx + "::1"
		si[i].Dst = pp[0] + ":" + s.Idx + "::2"
		si[i].Rt = rt[0] + ":" + s.Idx + "::/64"
		si[i].Sid = sid
	}

	var buf, _ = json.Marshal(&UserSessionInfoList{Id: int64(len(si)),
                Entry: si})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func listUserServers(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqListUserServer++

	var m []string

	if m, err = getRedisUserUidList(d.UserId, "sessions"); err != nil {
		event(lognotice, li, err.Error())
	}

	var v []string

	if v, err = getRedisServerList("all"); err != nil {
		stat.ReqErrListUserServer++
		return
	}

	var si = make([]UserServerInfo, len(v))

	for i := range v {
		var vid, _ = strconv.ParseInt(v[i], 0, 64)

		if len(m) != 0 {
			var tok []string

			for j := range m {
				tok = strings.Split(m[j], ":")

				if len(tok) != 2 {
					continue
				}
			}

			if v[i] == tok[0] {
				si[i] = UserServerInfo{Idx: vid, ErrNo: ENOENT}
				continue
			}
		}

		if s, err := getRedisServerInfo(vid); err != nil {
			si[i] = UserServerInfo{Idx: vid, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		} else {
			si[i] = UserServerInfo{Id: s.Id, Name: s.Name,
				Alias: s.Alias, Descr: s.Descr, Entity: s.Entity,
				Location: s.Location, Access: s.Access,
				Tunnel: s.Tunnel, TunnelSrc: s.TunnelSrc,
				PpPrefix: s.PpPrefix, RtPrefix: s.RtPrefix,
				Idx: vid}
		}
	}

	var buf, _ = json.Marshal(&UserServerInfoList{Id: int64(len(si)),
                Entry: si})

        sendResponse(w, &Msg{Data: string(buf)})
	return
}

func defaultSessionHandler(w http.ResponseWriter, r *http.Request) {
	stat.ReqAll++
        li.Msgid = 0

	var err error
	var str = "Invalid request"

	if err = checkUrl(r); err != nil {
		stat.ReqErrUrl++
		sendError(w, EINVAL, str, err)
		return
	}

	if err = checkHeader(r); err != nil {
		stat.ReqErrHeader++
		sendError(w, EINVAL, str, err)
		return
	}

	if err = checkRedis(); err != nil {
		stat.ReqErrRedis++
		sendError(w, EINVAL, str, err)
		return
	}

	var d *RequestMsg

	if d, err = checkData(r); err != nil {
		stat.ReqErrPayload++
		sendError(w, EINVAL, str, err)
		return
	}

	if li.Msgid, err = getRedisMsgId(d.Command); err != nil {
		stat.ReqErrMsgId++
		sendError(w, EINVAL, str, err)
		return
	}

        event(logdebug, li, "Processing request [%v:%v]", d.Command, li.Msgid)

	switch d.Command {
	case "activate-session":
		err = activateSession(w, d)

	case "deactivate-session":
		err = deactivateSession(w, d)

	case "check-session":
		err = checkSession(w, d)

	case "assign-session":
		err = assignSession(w, d)

	case "reassign-session":
		err = reassignSession(w, d)

	case "list-user-sessions":
		err = listUserSessions(w, d)

	case "list-user-servers":
		err = listUserServers(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

        event(logdebug, li, "Request [%v:%v] completed", d.Command, li.Msgid)
}
