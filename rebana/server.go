/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TSReqMsg struct {
	Id      int64
	UserId  int64
	MsgId   int64
	Command string
	Data    string
}

type TSMsg struct {
	Id    int64
	MsgId int64
	ErrNo int
	Data  string
}

type TSInfo struct {
	Id       int64
	TunSrc   string
	PpPrefix string
	RtPrefix string
	Session  []TSInfoSession
}

type TSInfoSession struct {
	Id   int64
	Type string
	Dst  string
	Idx  int64
}

type ServerInfo struct {
	Id        int64
	Name      string
	Alias     string
	Descr     string
	Entity    string
	Location  string
	Access    string
	Tunnel    string
	TunnelSrc string
	Url       string
	PpPrefix  string
	RtPrefix  string
	Admin     string
	Status    string
	Activated string

	RegDate int64
	Idx     int64
	ErrNo   int
}

type ServerInfoList struct {
	Id    int64
	Entry []ServerInfo
}

type UserServerInfo struct {
	Id        int64
	Name      string
	Alias     string
	Descr     string
	Entity    string
	Location  string
	Access    string
	Tunnel    string
	TunnelSrc string
	PpPrefix  string
	RtPrefix  string

	Idx   int64
	ErrNo int
}

type UserServerInfoList struct {
	Id    int64
	Entry []UserServerInfo
}

func resolveServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqResolveServer++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrResolveServer++
		return
	}

	var si = make([]Id, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if vid, err := getRedisServerIdFromName(e.Name); err != nil {
			si[i] = Id{ErrNo: ENOENT, Opt: e.Name}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: vid, Opt: e.Name}
		}
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func resolveServerId(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqResolveServerId++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrResolveServerId++
		return
	}

	var si = make([]Id, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if s, err := getRedisServerInfo(e.Id); err != nil {
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: e.Id, Opt: s.Name}
		}
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func addServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqAddServer++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrAddServer++
		return
	}

	var p string
	var c = 3

	var s = &ServerInfo{}

	for i := range m.Entry {
		var e = m.Entry[i]

		if p = e.Name; e.Name == "name" {
			if s.Name = e.Opt; s.Name == "" {
				err = errors.New(p)
				break
			}

			c--
		} else if p = e.Name; e.Name == "ppprefix" {
			if s.PpPrefix = e.Opt; s.PpPrefix == "" {
				err = errors.New(p)
				break
			} else {
				if !strings.Contains(s.PpPrefix, "::/48") {
					err = errors.New(p)
					break
				}

				c--
			}
		} else if p = e.Name; e.Name == "rtprefix" {
			if s.RtPrefix = e.Opt; s.RtPrefix == "" {
				err = errors.New(p)
				break
			} else {
				if !strings.Contains(s.RtPrefix, "::/48") {
					err = errors.New(p)
					break
				}

				c--
			}
		} else {
			err = errors.New(p)
			break
		}
	}

	if err != nil {
		stat.ReqErrAddServer++
		return errors.New("Invalid tunnel server parameter: " + p)
	}

	if c != 0 {
		stat.ReqErrAddServer++
		err = errors.New("Insufficient tunnel server parameters")
		return
	}

	var si = make([]Id, 1)

	s.Alias = "Central Node"
	s.Descr = "N3 Labs Tunnel Server"
	s.Entity = "N3 Labs"
	s.Location = "Kuala Lumpur"
	s.Access = "public"
	s.Tunnel = "6in4"
	s.TunnelSrc = ""
	s.Url = fmt.Sprintf("https://%v:443", s.Name)

	var vid int64

	if vid, err = setRedisServerNew(s); err != nil {
		stat.ReqErrAddServer++
		return
	} else {
		si[0] = Id{Id: vid, Opt: s.Name}
	}

	var t = time.Now().Format(time.RFC1123)

	var rcpt = []string{app.AdminEmail}
	var subj = fmt.Sprintf("Rebung.IO tunnel server registration notice: "+
		"%v", s.Name)
	var body = fmt.Sprintf("New tunnel server registered\n\n"+
		"Registered on %v\n\n"+
		"Server: %v\n"+
		"Point-to-Point Prefix: %v\n"+
		"Routed Prefix: %v\n"+
		"Management URL: %v", t, s.Name, s.PpPrefix, s.RtPrefix, s.Url)

	if err = sendMail(rcpt, subj, body); err != nil {
		event(logwarn, li, err.Error())
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func setServerAttr(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqSetServerAttr++

	var m *NameList

	if m, err = getNameList(d.Data, d.Command); err != nil {
		stat.ReqErrSetServerAttr++
		return
	}

	var si = make([]Name, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if e.Name == "alias" || e.Name == "descr" ||
			e.Name == "entity" || e.Name == "location" ||
			e.Name == "access" || e.Name == "tunnel" ||
			e.Name == "tunsrc" || e.Name == "ppprefix" ||
                        e.Name == "rtprefix" {

			if err = setRedisServerAttr(m.Id, e.Name,
				e.Opt); err != nil {
				si[i] = Name{Name: e.Name, ErrNo: ENOENT}
				event(logwarn, li, err.Error())
			} else {
				si[i] = Name{Name: e.Name, Opt: e.Opt}
			}
		} else {
			stat.ReqErrSetServerAttr++
			return errors.New("Attribute not permitted: " + e.Name)
		}
	}

	var buf, _ = json.Marshal(&NameList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func enableServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqEnableServer++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrEnableServer++
		return
	}

	var si = make([]Id, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if err = setRedisServerAdminStatus(e.Id, true); err != nil {
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func disableServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqDisableServer++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrDisableServer++
		return
	}

	var si = make([]Id, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if err = setRedisServerAdminStatus(e.Id, false); err != nil {
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func activateServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqActivateServer++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrActivateServer++
		return
	}

	var si = make([]Id, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if err = setRedisServerStatus(e.Id, true); err != nil {
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func deactivateServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqDeactivateServer++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrDeactivateServer++
		return
	}

	var si = make([]Id, len(m.Entry))

	for i := range m.Entry {
		var e = m.Entry[i]

		if err = setRedisServerStatus(e.Id, false); err != nil {
			si[i] = Id{Id: e.Id, ErrNo: ENOENT}
			event(logwarn, li, err.Error())
		} else {
			si[i] = Id{Id: e.Id}
		}
	}

	var buf, _ = json.Marshal(&IdList{Id: m.Id, Entry: si})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func listServer(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqListServer++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrListServer++
		return
	}

	var si []ServerInfo
	var sil *ServerInfoList

	if m.Entry[0].Id == 0 {
		var args = strings.Split(m.Entry[0].Opt, ":")

		if c := len(args); c != 4 {
			stat.ReqErrListServer++
			return errors.New("Invalid server list parameter count")
		}

		var list = args[0]

		if list == "all" || list == "enabled" || list == "disabled" ||
			list == "active" || list == "inactive" {
		} else {
			stat.ReqErrListServer++
			return errors.New("Invalid server list: " + list)
		}

		var page, _ = strconv.ParseInt(args[1], 0, 32)
		var npage, _ = strconv.ParseInt(args[2], 0, 32)

		if page == 0 || npage < 10 {
			stat.ReqErrListServer++
			return errors.New("Invalid list page count")
		}

		// init sort closures
		var sfn ServerSortBy
		var sby = args[3]

		if sby == "id" {
			sfn = func(s1, s2 *ServerInfo) bool {
				return s1.Id < s2.Id
			}
		} else if sby == "id-r" {
			sfn = func(s1, s2 *ServerInfo) bool {
				return s1.Id > s2.Id
			}
		} else if sby == "name" {
			sfn = func(s1, s2 *ServerInfo) bool {
				return s1.Name < s2.Name
			}
		} else if sby == "name-r" {
			sfn = func(s1, s2 *ServerInfo) bool {
				return s1.Name > s2.Name
			}
		} else if sby == "rdate" {
			sfn = func(s1, s2 *ServerInfo) bool {
				return s1.RegDate < s2.RegDate
			}
		} else if sby == "rdate-r" {
			sfn = func(s1, s2 *ServerInfo) bool {
				return s1.RegDate > s2.RegDate
			}
		} else {
			stat.ReqErrListServer++
			return errors.New("Invalid sort field: " + sby)
		}

		var v []string

		if v, err = getRedisServerList(list); err != nil {
			stat.ReqErrListServer++
			return
		}

		si = make([]ServerInfo, len(v))

		for i := range v {
			var id, _ = strconv.ParseInt(v[i], 0, 64)

			if s, err := getRedisServerInfo(id); err != nil {
				si[i] = ServerInfo{Idx: id, ErrNo: ENOENT}
				event(logwarn, li, err.Error())
			} else {
				si[i] = *s
				si[i].Idx = id
			}
		}

		// sort data
		ServerSortBy(sfn).Sort(si)

		// get page entries
		var c = len(si)

		if int(npage) < c {
			var st []ServerInfo

			var pofs = int(page - 1)
			var sofs = int(pofs * int(npage))
			var eofs = int(pofs*int(npage) + (int(npage) - 1))

			if sofs >= c {
				stat.ReqErrListServer++
				return errors.New("Invalid page offset")
			}

			if eofs < c {
				st = make([]ServerInfo, npage)
				copy(st, si[sofs:])
			} else {
				eofs = c
				st = make([]ServerInfo, len(si[sofs:eofs]))
				copy(st, si[sofs:eofs])
			}

			sil = &ServerInfoList{Id: int64(c), Entry: st}
		} else {
			sil = &ServerInfoList{Id: int64(c), Entry: si}
		}
	} else {
		si = make([]ServerInfo, len(m.Entry))

		for i := range m.Entry {
			var e = m.Entry[i]

			if s, err := getRedisServerInfo(e.Id); err != nil {
				si[i] = ServerInfo{Idx: e.Id, ErrNo: ENOENT}
				event(logwarn, li, err.Error())
			} else {
				si[i] = *s
			}
		}

		sil = &ServerInfoList{Id: int64(len(si)), Entry: si}
	}

	var buf, _ = json.Marshal(sil)

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func getServerList(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqGetServerList++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrGetServerList++
		return
	}

	var args = strings.Split(m.Entry[0].Opt, ":")

	if c := len(args); c != 4 {
		stat.ReqErrGetServerList++
		return errors.New("Invalid server list parameter count")
	}

	var list = args[0]

	if list == "all-users" || list == "all-sessions" ||
		list == "active-sessions" || list == "assigned-sessions" ||
		list == "unassigned-sessions" || list == "session-activity" {
	} else {
		stat.ReqErrGetServerList++
		return errors.New("Invalid server list: " + list)
	}

	var page, _ = strconv.ParseInt(args[1], 0, 32)
	var npage, _ = strconv.ParseInt(args[2], 0, 32)

	if page == 0 || npage < 10 {
		stat.ReqErrGetServerList++
		return errors.New("Invalid list page count")
	}

	var v []string

	if v, err = getRedisServerSvidList(m.Entry[0].Id, list); err != nil {
		stat.ReqErrGetServerList++
		return
	}

	var buf []byte

	if list == "session-activity" || list == "all-users" {
		var nl *NameList
		var si = make([]Name, len(v))

		for i := range v {
			si[i] = Name{Name: v[i]}
		}

		// get page entries
		var c = len(si)

		if int(npage) < c {
			var st []Name

			var pofs = int(page - 1)
			var sofs = int(pofs * int(npage))
			var eofs = int(pofs*int(npage) + (int(npage) - 1))

			if sofs >= c {
				stat.ReqErrGetServerList++
				return errors.New("Invalid page offset")
			}

			if eofs < c {
				st = make([]Name, npage)
				copy(st, si[sofs:])
			} else {
				eofs = c
				st = make([]Name, len(si[sofs:eofs]))
				copy(st, si[sofs:eofs])
			}

			nl = &NameList{Id: int64(c), Entry: st}
		} else {
			nl = &NameList{Id: int64(c), Entry: si}
		}

		buf, _ = json.Marshal(nl)
	} else {
		var e = m.Entry[0]
		var sv *ServerInfo
		var sil *SessionInfoList

		if sv, err = getRedisServerInfo(e.Id); err != nil {
			stat.ReqErrGetServerList++
			return
		}

		var si = make([]SessionInfo, len(v))

		for i := range v {
			var sid, _ = strconv.ParseInt(v[i], 0, 64)

			var s *SessionInfo

			if s, err = getRedisSessionInfo(e.Id, sid); err != nil {
				si[i] = SessionInfo{Sid: sid, ErrNo: ENOENT}
				event(logwarn, li, err.Error())
			}

			var pp = strings.Split(sv.PpPrefix, "::/")
			var rt = strings.Split(sv.RtPrefix, "::/")

			si[i] = *s
			si[i].ServerName = sv.Name
			si[i].TunSrc = sv.TunnelSrc
			si[i].Src = pp[0] + ":" + s.Idx + "::1"
			si[i].Dst = pp[0] + ":" + s.Idx + "::2"
			si[i].Rt = rt[0] + ":" + s.Idx + "::/64"
			si[i].Sid = sid
		}

		// get page entries
		var c = len(si)

		if int(npage) < c {
			var st []SessionInfo

			var pofs = int(page - 1)
			var sofs = int(pofs * int(npage))
			var eofs = int(pofs*int(npage) + (int(npage) - 1))

			if sofs >= c {
				stat.ReqErrGetServerList++
				return errors.New("Invalid page offset")
			}

			if eofs < c {
				st = make([]SessionInfo, npage)
				copy(st, si[sofs:])
			} else {
				eofs = c
				st = make([]SessionInfo, len(si[sofs:eofs]))
				copy(st, si[sofs:eofs])
			}

			sil = &SessionInfoList{Id: int64(c), Entry: st}
		} else {
			sil = &SessionInfoList{Id: int64(c), Entry: si}
		}

		buf, _ = json.Marshal(sil)
	}

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func serverStatus(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqServerStatus++

	var m *IdList

	if m, err = getIdList(d.Data, d.Command); err != nil {
		stat.ReqErrServerStatus++
		return
	}

	var e = m.Entry[0]

	var url string

	if url, err = getRedisServerUrl(e.Id); err != nil {
		stat.ReqErrServerStatus++
		return
	}

	var req = &TSReqMsg{Id: e.Id, UserId: li.Uid, MsgId: li.Msgid,
		Command: "status"}

	url += "/status"

	var res *TSMsg

	if res, err = sendTSRequest(url, req); err != nil {
		stat.ReqErrServerStatus++
		return
	}

	sendResponse(w, &Msg{Data: res.Data})
	return
}

func serverInfo(w http.ResponseWriter, d *RequestMsg) (err error) {
	stat.ReqServerInfo++

	var vid int64

	if vid, err = getRedisServerIdFromName(d.Data); err != nil {
		stat.ReqErrServerInfo++
		return
	}

	var si = &ServerInfo{}

	if si, err = getRedisServerInfo(vid); err != nil {
		stat.ReqErrServerInfo++
		return
	}

	var list []string
	var sa []TSInfoSession

	if list, err = getRedisServerSvidList(vid, "active-sessions"); err != nil {
		err = nil
		event(lognotice, li, "No active sessions for server [%v]", vid)
	} else {
		sa = make([]TSInfoSession, len(list))

		for i := range list {
			var sid, _ = strconv.ParseInt(list[i], 0, 64)

			if s, err := getRedisSessionInfo(vid, sid); err != nil {
				sa[i] = TSInfoSession{}
				event(logwarn, li, err.Error())
			} else {
				var idx, _ = strconv.ParseInt(s.Idx, 0, 64)

				sa[i] = TSInfoSession{Id: s.Id, Type: s.Type,
					Dst: s.TunDst, Idx: idx}
			}
		}
	}

	var buf, _ = json.Marshal(&TSInfo{Id: vid, TunSrc: si.TunnelSrc,
		PpPrefix: si.PpPrefix, RtPrefix: si.RtPrefix, Session: sa})

	sendResponse(w, &Msg{Data: string(buf)})
	return
}

func defaultServerHandler(w http.ResponseWriter, r *http.Request) {
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

	if err = checkUserAdmin(d.UserId); err != nil {
		stat.ReqErrUserId++
		sendError(w, EPERM, str, err)
		return
	}

	event(logdebug, li, "Processing request [%v:%v]", d.Command, li.Msgid)

	switch d.Command {
	case "resolve-server":
		err = resolveServer(w, d)

	case "resolve-server-id":
		err = resolveServerId(w, d)

	case "add-server":
		err = addServer(w, d)

	case "set-server-attr":
		err = setServerAttr(w, d)

	case "enable-server":
		err = enableServer(w, d)

	case "disable-server":
		err = disableServer(w, d)

	case "activate-server":
		err = activateServer(w, d)

	case "deactivate-server":
		err = deactivateServer(w, d)

	case "list-server":
		err = listServer(w, d)

	case "get-server-list":
		err = getServerList(w, d)

	case "tunnel-server-status":
		err = serverStatus(w, d)

	case "server-info":
		err = serverInfo(w, d)
	}

	if err != nil {
		str += ": " + d.Command
		sendError(w, EINVAL, str, err)
		return
	}

	event(logdebug, li, "Request [%v:%v] completed", d.Command, li.Msgid)
}
