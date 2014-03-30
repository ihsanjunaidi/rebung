/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

//
// #cgo CFLAGS: -I/home/ihsan/c/librebana
// #cgo LDFLAGS: /usr/local/lib/librebana.a
// #include <stdlib.h>
// #include "rebana.h"
//
import "C"
import (
	"errors"
	"net"
	"time"
	"unsafe"
)

type CSession struct {
	Ifname  string
	TunAddr string
	TunDest string
	In6Addr string
	In6Dest string
	In6Plen uint
	Rt6Dest string
	Rt6Plen uint
}

type PingVar struct {
	Dst string
	Rtt time.Duration
}

type PingReq struct {
	Src4 *net.UDPAddr
	Src6 *net.UDPAddr
	Dst4 *net.UDPAddr
	Dst6 *net.UDPAddr
}

func activateSession(s *CSession) (err error) {
	var cs = C.struct_session{
		ifname:  C.CString(s.Ifname),
		tunaddr: C.CString(s.TunAddr),
		tundest: C.CString(s.TunDest),
		in6addr: C.CString(s.In6Addr),
		in6dest: C.CString(s.In6Dest),
		in6plen: C.uint8_t(s.In6Plen),
		rt6dest: C.CString(s.Rt6Dest),
		rt6plen: C.uint8_t(s.Rt6Plen),
	}

	// create tunnel interface
	if C.createInterface(cs.ifname, cs.tunaddr, cs.tundest) == -1 {
		return errors.New("Error creating interface")
	}

	// create tunnel inet6 addresses
	if C.createAddress(cs.ifname, cs.in6addr, cs.in6plen) == -1 {
		return errors.New("Error assigning tunnel address")
	}

	// create route
	if C.createRoute(cs.rt6dest, cs.rt6plen, cs.in6dest) == -1 {
		return errors.New("Error creating route")
	}

	C.free(unsafe.Pointer(cs.ifname))
	C.free(unsafe.Pointer(cs.tunaddr))
	C.free(unsafe.Pointer(cs.tundest))
	C.free(unsafe.Pointer(cs.in6addr))
	C.free(unsafe.Pointer(cs.in6dest))
	C.free(unsafe.Pointer(cs.rt6dest))

	event(loginfo, li, "Tunnel session to %v established", s.TunDest)
	return
}

func deactivateSession(s *CSession) (err error) {
	var cs = C.struct_session{
		ifname:  C.CString(s.Ifname),
		tunaddr: C.CString(s.TunAddr),
		tundest: C.CString(s.TunDest),
		in6addr: C.CString(s.In6Addr),
		in6dest: C.CString(s.In6Dest),
		in6plen: C.uint8_t(s.In6Plen),
		rt6dest: C.CString(s.Rt6Dest),
		rt6plen: C.uint8_t(s.Rt6Plen),
	}

	// delete route
	if C.deleteRoute(cs.rt6dest, cs.rt6plen, cs.in6dest) == -1 {
		return errors.New("Error deleting route")
	}

	if C.deleteInterface(cs.ifname) == -1 {
		return errors.New("Error deleting interface")
	}

	C.free(unsafe.Pointer(cs.ifname))
	C.free(unsafe.Pointer(cs.tunaddr))
	C.free(unsafe.Pointer(cs.tundest))
	C.free(unsafe.Pointer(cs.in6addr))
	C.free(unsafe.Pointer(cs.in6dest))
	C.free(unsafe.Pointer(cs.rt6dest))

	event(loginfo, li, "Tunnel session to %v deleted", s.TunDest)
	return
}

func pingSession(dst, udp string) (rtt time.Duration, tgt string, err error) {
	tgt = dst

        var laddr *net.UDPAddr
        var ip = net.ParseIP(dst)

	var addr = &net.UDPAddr{IP: ip, Port: 33450}

	if udp == "udp4" {
		laddr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	} else {
		laddr = &net.UDPAddr{IP: net.IPv6zero, Port: 0}
	}

	var t = time.Now()
	var con *net.UDPConn

	if con, err = net.ListenUDP(udp, laddr); err != nil {
		return rtt, tgt, err.(net.Error)
	}
	defer con.Close()

	con.SetDeadline(t.Add(100 * time.Millisecond))

	if _, err = con.WriteToUDP([]byte("PING"), addr); err != nil {
		return rtt, tgt, err.(net.Error)
	}

	var buf = make([]byte, 4)

	if _, _, err = con.ReadFromUDP(buf); err != nil {
		return rtt, tgt, err.(net.Error)
	}

	rtt = time.Since(t)

	return
}
