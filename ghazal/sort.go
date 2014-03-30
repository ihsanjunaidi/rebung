/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"sort"
)

type UserSortBy func(s1, s2 *UserInfo) bool

func (by UserSortBy) Sort(us []UserInfo) {
	ussort := &UserInfoSorter{
		us: us,
		by: by,
	}

	sort.Sort(ussort)
}

type UserInfoSorter struct {
	us []UserInfo
	by func(p1, p2 *UserInfo) bool
}

func (s *UserInfoSorter) Len() int {
	return len(s.us)
}

func (s *UserInfoSorter) Swap(i, j int) {
	s.us[i], s.us[j] = s.us[j], s.us[i]
}

func (s *UserInfoSorter) Less(i, j int) bool {
	return s.by(&s.us[i], &s.us[j])
}
