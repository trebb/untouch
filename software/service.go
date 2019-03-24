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
		serviceScreen()
	case coVer:
		displayVersionScreenContent()
	case coMUd:
		notify(serviceNames["mbFirmwareUpdate"], 2, 1500*time.Millisecond)
	case coUUd:
		notify(serviceNames["uiUpdateNotImplemented"], 0, 5*time.Hour)
	default:
		log.Print("unknown serviceMode", mode)
	}
}

func displayVersionScreenContent() {
	for {
		notify(serviceNames["mbFirmwareVersionMode"], 0, 3*time.Second)
		time.Sleep(3 * time.Second)
		notify(fmt.Sprint(mbStateItem("romName")), 0, 2*time.Second)
		time.Sleep(2*time.Second)
		notify(fmt.Sprint(mbStateItem("romVersion")), 0, 2*time.Second)
		time.Sleep(2*time.Second)
		notify(name("pianoModel", mbStateItem("pianoModel")), 0, 2*time.Second)
		time.Sleep(2*time.Second)
		notify(name("marketDestination", mbStateItem("marketDestination")), 0, 2*time.Second)
		time.Sleep(2*time.Second)
		notify(fmt.Sprint(mbStateItem("romChecksum")), 0, 2*time.Second)
		time.Sleep(2*time.Second)
	}
}

func serviceScreen() { notify(serviceNames["serviceMode"], 10, 3*time.Second) }

type serviceModeObservation struct {
	device byte
	state  []byte
}

var (
	serviceMode1Observations = make(chan serviceModeObservation)
	serviceMode6Observations = make(chan serviceModeObservation)
	serviceMode9Observations = make(chan serviceModeObservation)
)

func observeServiceMode1(device byte, state []byte) {
	serviceMode1Observations <- serviceModeObservation{device, state}
}

func observeServiceMode6(device byte, state []byte) {
	serviceMode6Observations <- serviceModeObservation{device, state}
}

func observeServiceMode9(switchDevice byte, switchState []byte) {
	serviceMode9Observations <- serviceModeObservation{switchDevice, switchState}
}

func observeServiceMode1Monitor() {
	var currentKey byte
	var currentOnVelocity byte
	var currentOffVelocity byte
	var output string
	for {
		o := <-serviceMode1Observations
		switch o.device {
		case 0:
			currentKey = 0
			output = fmt.Sprintf("Ped3 %3d", o.state[0])
		case 1:
			currentKey = 0
			output = fmt.Sprintf("Ped2 %3d", o.state[0])
		case 2:
			currentKey = 0
			output = fmt.Sprintf("Ped1 %3d", o.state[0])
		case 3:
			currentKey = 0
			output = fmt.Sprintf("MAIN %3d", o.state[0])
		case 4:
			currentKey = 0
			output = fmt.Sprintf("L-IN%3d", o.state[0])
		case 5:
			currentKey = o.state[0]
			currentOffVelocity = 0
			currentOnVelocity = o.state[1]
			output = fmt.Sprintf("%2d.%3d.%3d.", currentKey-20, currentOnVelocity, currentOffVelocity)
		case 6:
			if currentKey != o.state[0] {
				currentKey = o.state[0]
				currentOnVelocity = 0
			}
			currentOffVelocity = o.state[1]
			output = fmt.Sprintf("%2d.%3d.%3d.", currentKey-20, currentOnVelocity, currentOffVelocity)
		}
		notify(output, 0, 5*time.Second)
	}
}

func observeServiceMode6Monitor() {
	for {
		var output string
		o := <-serviceMode6Observations
		switch o.device {
		case 1:
			switch o.state[0] {
			case 0:
				output = fmt.Sprint("NO USB")
			case 1:
				output = fmt.Sprint("USB OK")
			default:
				output = fmt.Sprint("NOT IMPL")
			}
		}
		notify(output, 0, 5*time.Second)
	}
}

func observeServiceMode9Monitor() {
	var currentKey byte
	var currentKeySwitchState [3]string
	var output string
	for {
		o := <-serviceMode9Observations
		switch o.device {
		case 0, 1, 2:
			if currentKey != o.state[0] {
				currentKey = o.state[0]
				currentKeySwitchState[0] = ""
				currentKeySwitchState[1] = ""
				currentKeySwitchState[2] = ""
			}
			if o.state[1] == 0 {
				currentKeySwitchState[o.device] = ""
			} else {
				currentKeySwitchState[o.device] = fmt.Sprintf("s%d", o.device+1)
			}
			output = fmt.Sprintf("%2d.%2s.%2s.%2s.", currentKey-20, currentKeySwitchState[0], currentKeySwitchState[1], currentKeySwitchState[2])
		case 3:
			currentKey = 0
			output = fmt.Sprintf("Ped3 %3d", o.state[0])
		case 4:
			currentKey = 0
			output = fmt.Sprintf("Ped2 %3d", o.state[0])
		case 5:
			currentKey = 0
			output = fmt.Sprintf("Ped1 %3d", o.state[0])
		case 6:
			currentKey = 0
			output = fmt.Sprintf("MAIN %3d", o.state[0])
		case 7:
			currentKey = 0
			output = fmt.Sprintf("L-IN %3d", o.state[0])
		}
		notify(output, 0, 5*time.Second)
	}
}

func init() {
	go observeServiceMode1Monitor()
	go observeServiceMode6Monitor()
	go observeServiceMode9Monitor()
}
