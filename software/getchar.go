package main

import (
	"github.com/pkg/term"
	"log"
)

var t *term.Term

func init() {
	t, _ = term.Open("/dev/tty")
	t.SetRaw()
	t.Flush()
}

func getChar() []byte {
	bytes := make([]byte, 3)
	_, err := t.Read(bytes)
	if err != nil {
		log.Print(err)
	}
	return bytes
}
