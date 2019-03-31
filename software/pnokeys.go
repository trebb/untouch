package main

import (
	"log"
	"time"
)

var blkKey []int8 = []int8{
	2,
	5, 7, 10, 12, 14,
	17, 19, 22, 24, 26,
	29, 31, 34, 36, 38,
	41, 43, 46, 48, 50,
	53, 55, 58, 60, 62,
	65, 67, 70, 72, 74,
	77, 79, 82, 84, 86,
}

// getPnoKey returns the number of the next key double-pressed on the piano.
// Key A0 = 1; key C8 = 88.
func getPnoKey() (k int8, ok bool) {
Drain:
	for {
		select {
		case <-pnoKeys:
			log.Print("pnoKeys undrained")
		default:
			break Drain
		}
	}
	var key0, key1 int8
	issueDtaRq(request{hardw, hwKey, 0x0, 0x1, 0x0})
	key0 = <-pnoKeys // key 1 (A0) yields 21
	issueDtaRq(request{hardw, hwKey, 0x0, 0x1, 0x0})
	key1Timeout := time.NewTimer(300 * time.Millisecond)
	for {
		select {
		case key1 = <-pnoKeys: // key 1 (A0) yields 21
			ok = key0 == key1
			k = key0 - 20
			return
		case <-key1Timeout.C:
			ok = false
			return
		}
	}

}
