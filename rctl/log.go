/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

package main

import (
	"fmt"
	"os"
)

func fatal(a string, v ...interface{}) {
	var n string

	if len(v) != 0 {
		n = fmt.Sprintf(a, v...)
	} else {
		n = a
	}

	fmt.Fprintf(os.Stderr, "Fatal: "+n+"\n")
	os.Exit(1)
}

func event(a string, v ...interface{}) {
	var n string

	if len(v) != 0 {
		n = fmt.Sprintf(a, v...)
	} else {
		n = a
	}

	fmt.Fprintf(os.Stderr, "%v\n", n)
}
