package main

import (
	"log"
)

var blkKey []uint8 = []uint8{
	2,
	5, 7, 10, 12, 14,
	17, 19, 22, 24, 26,
	29, 31, 34, 36, 38,
	41, 43, 46, 48, 50,
	53, 55, 58, 60, 62,
	65, 67, 70, 72, 74,
	77, 79, 83, 85, 86,
}

// getPnoKey returns the number of the next key pressed on the piano.
// Key A0 = 1; key C8 = 88.
func getPnoKey() uint8 {
Drain:
	for {
		select {
		case <-pnoKeys:
			log.Print("pnoKeys undrained")
		default:
			break Drain
		}
	}
	issueDtaRq(request{hardw, hwKey, 0x0, 0x1, 0x0})
	k := <-pnoKeys // key 1 (A0) yields 21
	return k - 20
}
