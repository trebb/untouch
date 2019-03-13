package main

import (
	"fmt"
	"log"
	"time"
)

func service(mode byte) {
	fmt.Println("serviceMode", mode)
	switch mode {
	case coSvc:
	case coVer:
		for {
			notify(fmt.Sprint(mbStateItem("romName")), 0, 1500*time.Millisecond)
			time.Sleep(1500 * time.Millisecond)
			notify(fmt.Sprint(mbStateItem("romVersion")), 0, 1500*time.Millisecond)
			time.Sleep(1500 * time.Millisecond)
			notify(name("pianoModel", mbStateItem("pianoModel")), 0, 1500*time.Millisecond)
			time.Sleep(1500 * time.Millisecond)
			notify(name("marketDestination", mbStateItem("marketDestination")), 0, 1500*time.Millisecond)
			time.Sleep(1500 * time.Millisecond)
			notify(fmt.Sprint(mbStateItem("romChecksum")), 0, 1500*time.Millisecond)
			time.Sleep(1500 * time.Millisecond)
		}
	case coMUd:
	case coUUd:
	default:
		log.Print("unknown serviceMode", mode)
	}
}
