package main

import (
	"fmt"
	"log"
	"time"
)

type uiKey [3]byte

// func input(){
// 	// all combinations available:
// 	// shift alt altgr (but not on all ttys)
// 	// ctl alt
// 	for {
// 		fmt.Print(getChar(), " ")
// 	}
// }
var (
	keyNil uiKey                     // zero value of type uiKey
	keyG   = uiKey{0x67, 0x00, 0x00} // plain g
	keyG1  = uiKey{0x07, 0x00, 0x00} // ctl-g
	keyG12 = uiKey{0x1B, 0x07, 0x00} // ctl-alt-g
	keyG2  = uiKey{0x1B, 0x67, 0x00} // alt-g
	keyK   = uiKey{0x6B, 0x00, 0x00} // etc.
	keyK1  = uiKey{0x0B, 0x00, 0x00}
	keyK12 = uiKey{0x1B, 0x0B, 0x00}
	keyK2  = uiKey{0x1B, 0x6B, 0x00}
	keyM   = uiKey{0x6D, 0x00, 0x00}
	keyM1  = uiKey{0x0D, 0x00, 0x00}
	keyM12 = uiKey{0x1B, 0x0D, 0x00}
	keyM2  = uiKey{0x1B, 0x6D, 0x00}
	keyP   = uiKey{0x70, 0x00, 0x00}
	keyP1  = uiKey{0x10, 0x00, 0x00}
	keyP12 = uiKey{0x1B, 0x10, 0x00}
	keyP2  = uiKey{0x1B, 0x70, 0x00}
	keyR   = uiKey{0x72, 0x00, 0x00}
	keyR1  = uiKey{0x12, 0x00, 0x00}
	keyR12 = uiKey{0x1B, 0x12, 0x00}
	keyR2  = uiKey{0x1B, 0x72, 0x00}
	keyS   = uiKey{0x73, 0x00, 0x00}
	keyS1  = uiKey{0x13, 0x00, 0x00}
	keyS12 = uiKey{0x1B, 0x13, 0x00}
	keyS2  = uiKey{0x1B, 0x73, 0x00}
	keyX   = uiKey{0x78, 0x00, 0x00}
	keyX1  = uiKey{0x18, 0x00, 0x00}
	keyX12 = uiKey{0x1B, 0x18, 0x00}
	keyX2  = uiKey{0x1B, 0x78, 0x00}
	// key1 = uiKey{0x31, 0x00, 0x00}
	// key2 = uiKey{0x32, 0x00, 0x00}
	// key3 = uiKey{0x33, 0x00, 0x00}
	// key4 = uiKey{0x34, 0x00, 0x00}
	// key5 = uiKey{0x35, 0x00, 0x00}
	// key6 = uiKey{0x36, 0x00, 0x00}
	// key7 = uiKey{0x37, 0x00, 0x00}
	// key8 = uiKey{0x38, 0x00, 0x00}
	keyESC = uiKey{0x1B, 0x00, 0x00}
)

func input() {
	for {
		var cmd uiKey
		copy(cmd[:], getChar())
		if mode, ok := mbStateItemOk("serviceMode"); ok && mode == coSvc {
			serviceModeInput(cmd)
		} else if _, ok := mbStateItemOk("toneGeneratorMode"); ok { // normal playing mode
			pianoModeInput(cmd)
		} else { // useful only during debugging
			miniInput(cmd)
		}
	}
}

func serviceModeInput(cmd uiKey) {
	switch cmd {
	case keyG: // mode 4
		issueCmd(servi, sMSel, 0x0, sML_R)
		notify(serviceNames["sm4L/R"], 1, 3*time.Second)
		go func() {
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyR: // L
					issueCmd(servi, srL_R, 0x0, byte(0x7F))
					issueCmd(servi, srL_R, 0x1, byte(0x0))
					notify("L", 0, 5*time.Second)
				case keyP: // R
					issueCmd(servi, srL_R, 0x1, byte(0x7F))
					issueCmd(servi, srL_R, 0x0, byte(0x0))
					notify("R", 0, 5*time.Second)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyG1: // mode 2
		notify(serviceNames["sm2EffectReverb"], 1, 3*time.Second)
		issueCmd(servi, sMSel, 0x0, sMEfR)
		go func() {
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyR:
					issueCmd(servi, srEfR, 0x1, byte(0x7F))
					notify("REVERB", 0, 5*time.Second)
				case keyP:
					issueCmd(servi, srEfR, 0x0, byte(0x7F))
					notify("EFFECTS", 0, 5*time.Second)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyG2: // mode 5
		notify(serviceNames["sm5EqLevel"], 1, 3*time.Second)
		issueCmd(servi, sMSel, 0x0, sMEqL)
		go func() {
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyM:
					issueCmd(servi, srEqL, 0x0, byte(0x7F))
					issueCmd(servi, srEqL, 0x1, byte(0x0))
					issueCmd(servi, srEqL, 0x2, byte(0x0))
					issueCmd(servi, srEqL, 0x3, byte(0x0))
					notify("1 ON", 0, 5*time.Second)
				case keyM1:
					issueCmd(servi, srEqL, 0x1, byte(0x7F))
					issueCmd(servi, srEqL, 0x0, byte(0x0))
					issueCmd(servi, srEqL, 0x2, byte(0x0))
					issueCmd(servi, srEqL, 0x3, byte(0x0))
					notify("2 ON", 0, 5*time.Second)
				case keyM2:
					issueCmd(servi, srEqL, 0x2, byte(0x7F))
					issueCmd(servi, srEqL, 0x0, byte(0x0))
					issueCmd(servi, srEqL, 0x1, byte(0x0))
					issueCmd(servi, srEqL, 0x3, byte(0x0))
					notify("3 ON", 0, 5*time.Second)
				case keyM12:
					issueCmd(servi, srEqL, 0x3, byte(0x7F))
					issueCmd(servi, srEqL, 0x0, byte(0x0))
					issueCmd(servi, srEqL, 0x1, byte(0x0))
					issueCmd(servi, srEqL, 0x2, byte(0x0))
					notify("4 ON", 0, 5*time.Second)
				case keyP:
					issueCmd(servi, srEqL, 0x9, byte(0x1))
					notify("PLAY", 0, 5*time.Second)
				case keyP1:
					issueCmd(servi, srEqL, 0x9, byte(0x0))
					notify("MUTE", 0, 5*time.Second)
				case keyR:
					issueCmd(servi, srEqL, 0x8, byte(0x1))
					notify("SP.EQ ON", 0, 5*time.Second)
				case keyR1:
					issueCmd(servi, srEqL, 0x8, byte(0x0))
					notify("SP.EQ OFF", 0, 5*time.Second)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyG12: // mode 13
		issueCmd(servi, sMSel, 0x0, sMTcS)
		notify(serviceNames["sm13TouchSelect"], 1, 3*time.Second)
		go func() {
			var model byte
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyM:
					model = byte(0x0)
					issueCmd(servi, srTcS, 0x0, model)
					notify(name("touchSelectModel", model), 0, 5*time.Second)
				case keyM1:
					model = byte(0x1)
					issueCmd(servi, srTcS, 0x0, model)
					notify(name("touchSelectModel", model), 0, 5*time.Second)
				case keyM2:
					model = byte(0x2)
					issueCmd(servi, srTcS, 0x0, model)
					notify(name("touchSelectModel", model), 0, 5*time.Second)
				case keyM12:
					model = byte(0x3)
					issueCmd(servi, srTcS, 0x0, model)
					notify(name("touchSelectModel", model), 0, 5*time.Second)
				case keyP:
					issueCmd(servi, srTcS, 0x10, model)
					notify("SAVED", 10, 5*time.Second)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyK: // mode 9
		issueCmd(servi, sMSel, 0x0, sMKRw)
		notify(serviceNames["sm9KeyboardS1S2S3AdRaw"], 1, 3*time.Second)
	case keyK1: // mode 1
		issueCmd(servi, sMSel, 0x0, sMPdV)
		notify(serviceNames["sm1PedalVolumeKeyboardMidi"], 1, 3*time.Second)
	case keyK2: // mode 11
		issueCmd(servi, sMSel, 0x0, sMAlK)
		notify(serviceNames["sm112llKeyOn"], 1, 3*time.Second)
		go func() {
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyP:
					issueCmd(servi, srAlK, 0x0, byte(0))
					closeSmControlKeys()
					break Loop
				case keyNil:
					break Loop
				}
			}
		}()
	case keyK12: // mode 7
		issueCmd(servi, sMSel, 0x0, sMMTc)
		notify(serviceNames["sm7MaxTouch"], 1, 3*time.Second)
		go func() {
			controlKeys := openSmControlKeys()
			var sound uint8 = 0
			issueCmd(servi, srMTc, 0x1, sound)
			notify(fmt.Sprintf("Sound.%3d", sound), 0, 5*time.Second)
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyR: // sound_number--
					sound--
					issueCmd(servi, srMTc, 0x1, sound)
					notify(fmt.Sprintf("Sound.%3d", sound), 0, 5*time.Second)
				case keyP: // sound_number++
					sound++
					issueCmd(servi, srMTc, 0x1, sound)
					notify(fmt.Sprintf("Sound.%3d", sound), 0, 5*time.Second)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyM, keyM1, keyM2, keyM12, keyP, keyP1, keyP2, keyP12, keyR, keyR1, keyR2, keyR12:
		smControlKeys <- cmd
	case keyS:
	case keyS1: // mode 6
		issueCmd(servi, sMSel, 0x0, sMUBt)
		issueCmd(servi, srUBt, 0x0, byte(0x0))
	case keyS2: // mode 14
		issueCmd(servi, sMSel, 0x0, sMVer)
		notify(buildDate, 1, 5*time.Second)
	case keyS12: // mode 10
		issueCmd(servi, sMSel, 0x0, sMWCk)
		notify(serviceNames["sm10WaveChecksum"], 1, 3*time.Second)
		go func() {
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyP:
					issueCmd(servi, srWCk, 0x0, byte(0))
					closeSmControlKeys()
					notify(errorName("cancelled"), 0, 1500*time.Millisecond)
					break Loop
				case keyNil:
					break Loop
				}
			}
		}()
	case keyX: // mode 8
		issueCmd(servi, sMSel, 0x0, sMTCk)
		notify(serviceNames["sm8ToneCheck"], 1, 3*time.Second)
		go func() {
			controlKeys := openSmControlKeys()
			var sound uint8 = 0
			issueCmd(servi, srTCk, 0x0, byte(0x0), sound)
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyR: // sound_number--
					sound--
					issueCmd(servi, srTCk, 0x0, byte(0x0), sound)
				case keyP: // sound_number++
					sound++
					issueCmd(servi, srTCk, 0x0, byte(0x0), sound)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyX1: // mode 3
		issueCmd(servi, sMSel, 0x0, sMTgA)
		notify(serviceNames["sm3TgAllChannel"], 1, 3*time.Second)
		go func() {
			controlKeys := openSmControlKeys()
		Loop:
			for {
				k := <-controlKeys
				switch k {
				case keyR:
					issueCmd(servi, srTgA, 0x0, byte(1))
					notify("RUNNING", 0, 5*time.Second)
				case keyP:
					issueCmd(servi, srTgA, 0x0, byte(0))
					notify("STOPPED", 0, 5*time.Second)
				case keyNil:
					break Loop
				}
			}
		}()
	case keyX2: // mode 12
		issueCmd(servi, sMSel, 0x0, sMKAd)
		notify(serviceNames["sm12KeyAdjust"], 1, 3*time.Second)
	case keyX12:
	case keyESC:
		close(exit) // for debugging
	default:
		log.Printf("Svc[%X %X %X] ", cmd[0], cmd[1], cmd[2])
	}
}

var (
	smControlKeys           = make(chan uiKey)
	newSmControlKeyListener = make(chan (chan uiKey))
)

func openSmControlKeys() <-chan uiKey {
	var c = make(chan uiKey)
	newSmControlKeyListener <- c
	return c
}

func closeSmControlKeys() { newSmControlKeyListener <- nil }

func smControlKeyMonitor() {
	var listener chan<- uiKey
	for {
		select {
		case l := <-newSmControlKeyListener:
			if listener != nil {
				close(listener)
			}
			listener = l
		case cmd := <-smControlKeys:
			if listener == nil {
				log.Printf("Ctl[%X %X %X] ", cmd[0], cmd[1], cmd[2])
			} else {
				listener <- cmd
			}
		}
	}
}

func init() { go smControlKeyMonitor() }

func pianoModeInput(cmd uiKey) {
	switch cmd {
	case keyG:
		loadRegistration()
	case keyG1:
		storeRegistration()
	case keyG2:
		storeToSound()
	case keyG12:
	case keyK: // aggregate mode: pianist mode, all sound mode splits
		if mbStateItem("toneGeneratorMode") == tgSnd {
			notifyLock(name("keyboardMode", mbStateItem("keyboardMode")))
		} else {
			notifyLock("PIANIST")
		}
		k, ok := getPnoKey()
		switch {
		case !ok:
			notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
		case k == 1: // pianist mode
			notifyUnlock("PIANIST", 0, 1500*time.Millisecond)
			keepMbState("toneGeneratorMode", tgPia)
			issueCmd(tgMod, tgMod, 0x0, tgPia)
		case k <= 5: // one of the sound mode keyboard modes
			notifyUnlock(name("keyboardMode", int(k-2)), 0, 1500*time.Millisecond)
			issueCmd(kbSpl, kbSpM, 0x0, k-2) // KB split mode 0..3
			keepMbState("toneGeneratorMode", tgSnd)
			issueCmd(tgMod, tgMod, 0x0, tgSnd)
			requestAllVTSettings()
		default:
			notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
		}
	case keyK1: // aggregate sound: rendering of pianist mode or first/only sound of sound mode
		switch mbStateItem("toneGeneratorMode") {
		case tgPia:
			notifyLock(name("renderingCharacter", mbStateItem("renderingCharacter")))
			k, ok := getPnoKey()
			if ok && k <= 10 {
				notifyUnlock(name("renderingCharacter", int(k-1)), 0, 1500*time.Millisecond)
				issueCmd(pmSet, pmRen, 0x0, k-1)
				issueCmd(tgMod, tgMod, 0x0, tgPia) // triggers confirmation of the changes
			} else {
				notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
			}
		case tgSnd:
			switch mbStateItem("keyboardMode") {
			case kbSpMSingle:
				notifyLock(name("instrumentSound", mbStateItem("single")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, iSing, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			case kbSpMDual:
				notifyLock(name("instrumentSound", mbStateItem("dual1")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, iDua1, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			case kbSpMSplit:
				notifyLock(name("instrumentSound", mbStateItem("split1")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, iSpl1, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			case kbSpM4hands:
				notifyLock(name("instrumentSound", mbStateItem("4hands1")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, i4Hd1, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			default:
				log.Print("bad keyboardMode")
			}
		default:
			log.Print("bad toneGeneratorMode")
		}
	case keyK2: // aggregate sound: second sound of sound mode
		switch mbStateItem("toneGeneratorMode") {
		case tgPia:
			notifyUnlock("NO 2ND", 0, 1500*time.Millisecond)
		case tgSnd:
			switch mbStateItem("keyboardMode") {
			case kbSpMSingle:
				notifyUnlock("NO 2ND", 0, 1500*time.Millisecond)
			case kbSpMDual:
				notifyLock(name("instrumentSound", mbStateItem("dual2")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, iDua2, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			case kbSpMSplit:
				notifyLock(name("instrumentSound", mbStateItem("split2")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, iSpl2, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			case kbSpM4hands:
				notifyLock(name("instrumentSound", mbStateItem("4hands2")))
				k, ok := getPnoKey()
				if ok {
					notifyUnlock(name("instrumentSound", int(k-1)), 0, 1500*time.Millisecond)
					issueCmd(instr, i4Hd2, 0x0, k-1)
					issueCmd(tgMod, tgMod, 0x0, tgSnd)
					requestAllVTSettings()
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			default:
				log.Print("bad keyboardMode")
			}
		default:
			log.Print("bad toneGeneratorMode")
		}
	case keyK12:
	case keyM:
		issueCmd(metro, mVolu, 0x0, mbStateItem("metronomeVolume"))
		keepMbState("metronomeOnOff", issueTglCmd("metronomeOnOff", metro, mOnOf, 0x0))
	case keyM1:
		notifyLock(fmt.Sprint(mbStateItem("metronomeTempo"), "/min"))
		k, ok := getPnoKey()
		if ok {
			tempo := scaleVal(10, 400, 88, k)
			notifyUnlock(fmt.Sprint(tempo, "/min"), 0, 1500*time.Millisecond)
			issueCmd(metro, mTmpo, 0x0, uint16(tempo))
			issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode"))
		} else {
			notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
		}
	case keyM2:
		notifyLock(name("rhythmPattern", mbStateItem("rhythmPattern")))
		k, ok := getPnoKey()
		if ok && int(k)-1 < len(rhythmGroupIndex) {
			notifyLock(name("rhythmGroup", int(k-1)))
			k2, ok2 := getPnoKey()
			if ok2 && k2 >= 42 { // middle-D = begin of rhythmGroup k
				pat := byte(int(k2) - 42 + rhythmGroupIndex[k-1])
				if pat <= 109 { // topmost rhythm pattern
					notifyUnlock(name("rhythmPattern", pat), 0, 1500*time.Millisecond)
					issueCmd(metro, mSign, 0x0, pat)
					issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode"))
				} else {
					notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
				}
			} else {
				notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
			}
		} else {
			notifyUnlock(errorName("cancelled"), -10, 1500*time.Millisecond)
		}
	case keyM12:
		tapTempo()
	case keyP:
		play()
	case keyP1:
		selectPlaySong(0)
	case keyP2:
		selectPlaySong(1)
	case keyP12:
		selectPlaySong(2)
	case keyR:
		record()
	case keyR1:
		selectRecordSong(0)
	case keyR2:
		selectRecordSong(1)
	case keyR12:
		eraseSongParts()
	case keyS:
		settings()
	case keyS1:
		virtualTechnician()
	case keyS2:
	case keyS12:
	case keyX:
		immediateActions()
	case keyX1:
	case keyX2:
	case keyX12:
	case keyESC:
		select {
		case <-exit:
		default:
			close(exit) // for debugging
		}
	default:
		log.Printf("Pno[%X %X %X] ", cmd[0], cmd[1], cmd[2])
	}
}

func miniInput(cmd uiKey) {
	switch cmd {
	case keyESC:
		close(exit) // for debugging
	default:
		log.Printf("Min[%X %X %X] ", cmd[0], cmd[1], cmd[2])
	}
}
