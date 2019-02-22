package main

import (
	"fmt"
	"log"
)

// func input(){
// 	// all combinations available:
// 	// shift alt altgr
// 	// ctl alt
// 	for {
// 		fmt.Print(getChar(), " ")
// 	}
// }
var (
	keyA   = [...]byte{0x61, 0x00, 0x00} // plain a
	keyAs  = [...]byte{0x41, 0x00, 0x00} // shift-A
	keyAa  = [...]byte{0x1B, 0x61, 0x00} // alt-a
	keyAsa = [...]byte{0x1B, 0x41, 0x00} // shift-alt-A
	keyF   = [...]byte{0x66, 0x00, 0x00} // etc.
	keyFs  = [...]byte{0x46, 0x00, 0x00}
	keyFa  = [...]byte{0x1B, 0x66, 0x00}
	keyFsa = [...]byte{0x1B, 0x46, 0x00}
	keyK   = [...]byte{0x6B, 0x00, 0x00}
	keyKs  = [...]byte{0x4B, 0x00, 0x00}
	keyKa  = [...]byte{0x1B, 0x6B, 0x00}
	keyKsa = [...]byte{0x1B, 0x4B, 0x00}
	keyM   = [...]byte{0x6D, 0x00, 0x00}
	keyMs  = [...]byte{0x4D, 0x00, 0x00}
	keyMa  = [...]byte{0x1B, 0x6D, 0x00}
	keyMsa = [...]byte{0x1B, 0x4D, 0x00}
	keyP   = [...]byte{0x70, 0x00, 0x00}
	keyPs  = [...]byte{0x50, 0x00, 0x00}
	keyPa  = [...]byte{0x1B, 0x70, 0x00}
	keyPsa = [...]byte{0x1B, 0x50, 0x00}
	keyR   = [...]byte{0x72, 0x00, 0x00}
	keyRs  = [...]byte{0x52, 0x00, 0x00}
	keyRa  = [...]byte{0x1B, 0x72, 0x00}
	keyRsa = [...]byte{0x1B, 0x52, 0x00}
	keyS   = [...]byte{0x73, 0x00, 0x00}
	keySs  = [...]byte{0x53, 0x00, 0x00}
	keySa  = [...]byte{0x1B, 0x73, 0x00}
	keySsa = [...]byte{0x1B, 0x53, 0x00}

	key1 = [...]byte{0x31, 0x00, 0x00}
	key2 = [...]byte{0x32, 0x00, 0x00}
	key3 = [...]byte{0x33, 0x00, 0x00}
	key4 = [...]byte{0x34, 0x00, 0x00}
	key5 = [...]byte{0x35, 0x00, 0x00}
	key6 = [...]byte{0x36, 0x00, 0x00}
	key7 = [...]byte{0x37, 0x00, 0x00}
	key8 = [...]byte{0x38, 0x00, 0x00}
)

func input() {
	for {
		var cmd [3]byte
		copy(cmd[:], getChar())
		switch cmd {
		case keyA:
			immediateActions()
		case keyAs:
		case keyAa:
		case keyAsa:
		case keyF:
			loadRegistration()
		case keyFs:
			storeRegistration()
		case keyFa:
			storeToSound()
		case keyFsa:
		case keyK: // aggregate mode: pianist mode, all sound mode splits
			if mbStateItem("toneGeneratorMode") == tgSnd {
				seg14.w <- name("keyboardMode", mbStateItem("keyboardMode"))
			} else {
				seg14.w <- "PIANIST"
			}
			k, ok := getPnoKey()
			switch {
			case !ok:
				writeError("cancelled")
			case k == 1: // pianist mode
				notify <- "PIANIST"
				keepMbState("toneGeneratorMode", tgPia)
				issueCmd(tgMod, tgMod, 0x0, tgPia)
			case k <= 5: // one of the sound mode keyboard modes
				notify <- name("keyboardMode", int(k-2))
				issueCmd(kbSpl, kbSpM, 0x0, byte(k-2)) // KB split mode 0..3
				keepMbState("toneGeneratorMode", tgSnd)
				issueCmd(tgMod, tgMod, 0x0, tgSnd)
				requestAllVTSettings()
			default:
				writeError("cancelled")
			}
		case keyKs: // aggregate sound: rendering of pianist mode or first/only sound of sound mode
			switch mbStateItem("toneGeneratorMode") {
			case tgPia:
				seg14.w <- name("renderingCharacter", mbStateItem("renderingCharacter"))
				k, ok := getPnoKey()
				if ok && k <= 10 {
					notify <- name("renderingCharacter", int(k-1))
					issueCmd(pmSet, pmRen, 0x0, byte(k-1))
					issueCmd(tgMod, tgMod, 0x0, tgPia) // triggers confirmation of the changes
				} else {
					writeError("cancelled")
				}
			case tgSnd:
				switch mbStateItem("keyboardMode") {
				case kbSpMSingle:
					seg14.w <- name("instrumentSound", mbStateItem("single"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, iSing, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				case kbSpMDual:
					seg14.w <- name("instrumentSound", mbStateItem("dual1"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, iDua1, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				case kbSpMSplit:
					seg14.w <- name("instrumentSound", mbStateItem("split1"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, iSpl1, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				case kbSpM4hands:
					seg14.w <- name("instrumentSound", mbStateItem("4hands1"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, i4Hd1, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				default:
					log.Print("bad keyboardMode")
				}
			default:
				log.Print("bad toneGeneratorMode")
			}
		case keyKa: // aggregate sound: second sound of sound mode
			switch mbStateItem("toneGeneratorMode") {
			case tgPia:
				notify <- "NO 2ND"
			case tgSnd:
				switch mbStateItem("keyboardMode") {
				case kbSpMSingle:
					notify <- "NO 2ND"
				case kbSpMDual:
					seg14.w <- name("instrumentSound", mbStateItem("dual2"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, iDua2, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				case kbSpMSplit:
					seg14.w <- name("instrumentSound", mbStateItem("split2"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, iSpl2, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				case kbSpM4hands:
					seg14.w <- name("instrumentSound", mbStateItem("4hands2"))
					k, ok := getPnoKey()
					if ok {
						notify <- name("instrumentSound", int(k-1))
						issueCmd(instr, i4Hd2, 0x0, byte(k-1))
						issueCmd(tgMod, tgMod, 0x0, tgSnd)
						requestAllVTSettings()
					} else {
						writeError("cancelled")
					}
				default:
					log.Print("bad keyboardMode")
				}
			default:
				log.Print("bad toneGeneratorMode")
			}
		case keyKsa:
		case keyM:
			issueCmd(metro, mVolu, 0x0, byte(mbStateItem("metronomeVolume")))
			keepMbState("metronomeOnOff", int(issueTglCmd("metronomeOnOff", metro, mOnOf, 0x0)))
		case keyMs:
			seg14.w <- fmt.Sprint(mbStateItem("metronomeTempo"), "/min")
			k, ok := getPnoKey()
			if ok {
				tempo := scaleVal(10, 400, 88, int(k))
				seg14.w <- fmt.Sprint(tempo, "/min")
				issueCmd(metro, mTmpo, 0x0, uint16(tempo))
				issueCmd(tgMod, tgMod, 0x0, byte(mbStateItem("toneGeneratorMode")))
			} else {
				writeError("cancelled")
			}
		case keyMa:
			seg14.w <- name("rhythmPattern", mbStateItem("rhythmPattern"))
			k, ok := getPnoKey()
			if ok && int(k) < len(rhythmGroupIndex) {
				seg14.w <- name("rhythmGroup", int(k-1))
				k2, ok2 := getPnoKey()
				if ok2 && k2 >= 42 { // middle-D = begin of rhythmGroup k
					pat := int(k2) - 42 + rhythmGroupIndex[k-1]
					notify <- name("rhythmPattern", int(pat))
					issueCmd(metro, mSign, 0x0, int(pat))
					issueCmd(tgMod, tgMod, 0x0, byte(mbStateItem("toneGeneratorMode")))
				} else {
					writeError("cancelled")
				}
			} else {
				writeError("cancelled")
			}
		case keyMsa:
			tapTempo()
		case keyP:
			play()
		case keyPs:
			selectPlaySong(0)
		case keyPa:
			selectPlaySong(1)
		case keyPsa:
			selectPlaySong(2)
		case keyR:
			record()
		case keyRs:
			selectRecordSong(0)
		case keyRa:
			selectRecordSong(1)
		case keyRsa:
			eraseSongParts()
		case keyS:
			settings()
		case keySs:
			virtualTechnician()
		case keySa:
		case keySsa:
		default:
			fmt.Printf("[%X %X %X] ", cmd[0], cmd[1], cmd[2])
		}
	}
}
