/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"sort"
)

type ServerSortBy func(s1, s2 *ServerInfo) bool

func (by ServerSortBy) Sort(sv []ServerInfo) {
	svsort := &ServerInfoSorter{
		sv: sv,
		by: by,
	}

	sort.Sort(svsort)
}

type ServerInfoSorter struct {
	sv []ServerInfo
	by func(p1, p2 *ServerInfo) bool
}

func (s *ServerInfoSorter) Len() int {
	return len(s.sv)
}

func (s *ServerInfoSorter) Swap(i, j int) {
	s.sv[i], s.sv[j] = s.sv[j], s.sv[i]
}

func (s *ServerInfoSorter) Less(i, j int) bool {
	return s.by(&s.sv[i], &s.sv[j])
}
