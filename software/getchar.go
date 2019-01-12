package main

import (
	"github.com/pkg/term"
	"log"
)

func getChar() []byte {
	t, _ := term.Open("/dev/tty")
	defer t.Restore()
	defer t.Close()
	t.SetRaw()
	t.Flush()
	bytes := make([]byte, 3)
	_, err := t.Read(bytes)
	if err != nil {
		log.Print(err)
	}
	return bytes
}
