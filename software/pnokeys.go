package main

import (
	"log"
	"time"
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

// getPnoKey returns the number of the next key double-pressed on the piano.
// Key A0 = 1; key C8 = 88.
func getPnoKey() (k uint8, ok bool) {
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
	var key0, key1 uint8
Key0:
	for {
		select {
		case key0 = <-pnoKeys: // key 1 (A0) yields 21
			break Key0
		default:
			seg14.brth <- struct{}{}
			time.Sleep(16 * time.Millisecond)
		}
	}
	t0 := time.Now()
	ok = true
	issueDtaRq(request{hardw, hwKey, 0x0, 0x1, 0x0})
Key1:
	for {
		select {
		case key1 = <-pnoKeys: // key 1 (A0) yields 21
			break Key1
		default:
			seg14.brth <- struct{}{}
			time.Sleep(16 * time.Millisecond)
			if time.Since(t0) > 300*time.Millisecond {
				ok = false
				break Key1
			}
		}
	}
	if key0 != key1 {
		ok = false
	}
	k = key0 - 20
	return
}
