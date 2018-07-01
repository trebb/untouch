package main

import (
	"github.com/pkg/term"
	"log"
)

func getChar() int {
	t, _ := term.Open("/dev/tty")
	defer t.Restore()
	defer t.Close()
	term.RawMode(t)
	bytes := make([]byte, 3)
	_, err := t.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return int(bytes[0])
}
