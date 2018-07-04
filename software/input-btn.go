package main

import (
	"fmt"
)

func input() {
	for  {
		cmd := string(getChar())
		fmt.Print("CMD-> ")
		switch cmd {
		case "H": // hi 
			issueCmdAc(commu, 0x7F, 0x0, 0x0)
		case "k":	// piano key
			issueDtaRq(request{hardw, hwKey, 0x0, 0x1, 0x0})
		default:
			fmt.Println("???", cmd)
		}
	}
}
