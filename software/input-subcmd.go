package main

import (
	"fmt"
	"log"
	"time"
)

func collectPianistSongs() {
	for i := 0; i < 0x3; i++ {
		keepPianistSongsSeen(i, false)
	}
	for i := 0; i < 0x3; i++ {
		issueDtaRq(request{pmRec, pmEmp, byte(i), 0x0})
	}
	allSeen := false
	for !allSeen {
		allSeen = true
		for i := 0; i < 3; i++ {
			s := pianistSongsItem(i)
			if !s.seen {
				time.Sleep(10 * time.Millisecond)
				allSeen = false
				break
			}
		}
	}
}

func collectSoundSongs() {
	for i := 0; i < 0xA; i++ {
		keepSoundSongsSeen(i, false)
	}
	for i := 0; i < 0xA; i++ {
		issueDtaRq(request{smRec, smEmp, byte(i), 0x0})
	}
	allSeen := false
	for !allSeen {
		allSeen = true
		for i := 0; i < 10; i++ {
			s := soundSongsItem(i)
			if !s.seen {
				time.Sleep(10 * time.Millisecond)
				allSeen = false
				break
			}
		}
	}
	for i := 0; i < 10; i++ {
		s := soundSongsItem(i)
		if s.data {
			keepMbState("soundSongPart1Seen", 0)
			keepMbState("soundSongPart2Seen", 0)
			issueCmd(smRec, smSel, 0x0, byte(i))
			for mbStateItem("soundSongPart1Seen") == 0 || mbStateItem("soundSongPart2Seen") == 0 {
				time.Sleep(10 * time.Millisecond)
			}
			keepSoundSongsPart1(i, mbStateItem("soundSongPart1") == 1)
			keepSoundSongsPart2(i, mbStateItem("soundSongPart2") == 1)
		}
	}
}

const (
	idle = iota
	standby
	recording
	playing
)

func selectPlaySong(partSet int) {
	currentRecorderState := <-getCurrentRecorderState
	switch currentRecorderState {
	case idle:
		if mbStateItem("usbThumbDrivePresence") == 1 {
			// TODO: check
		} else if mbStateItem("toneGeneratorMode") == tgPia {
			collectPianistSongs()
			seg14.w <- pianistSongName(false, mbStateItem("currentPianistSong"))
			k, ok := getPnoKey()
			if ok && k <= 3 {
				keepMbState("currentPianistSong", int(k-1))
			}
			notify <- pianistSongName(false, mbStateItem("currentPianistSong"))
			issueCmd(pmRec, pmSel, 0x0, byte(mbStateItem("currentPianistSong")))
		} else if mbStateItem("toneGeneratorMode") == tgSnd {
			collectSoundSongs()
			seg14.w <- soundSongName(false, mbStateItem("currentSoundSong"), partSet)
			k, ok := getPnoKey()
			fmt.Println("got Pno Key")
			if ok && k <= 10 {
				keepMbState("currentSoundSong", int(k-1))
				keepMbState("currentSoundSongPartSet", partSet)
			}
			notify <- soundSongName(false, mbStateItem("currentSoundSong"), partSet)
			issueCmd(smRec, smSel, 0x0, byte(mbStateItem("currentSoundSong")))
			issueCmd(smRec, smPlP, 0x0, byte(partSet))
		}
	case standby:
		notify <- "STANDBY"
	case recording:
		notify <- "RECORDNG"
	case playing:
		notify <- "PLAYING"
	default:
		log.Print("funny currentRecorderState")
		storeCurrentRecorderState <- idle
	}
}

func selectRecordSong(part int) {
	currentRecorderState := <-getCurrentRecorderState
	switch currentRecorderState {
	case idle:
		if mbStateItem("usbThumbDrivePresence") == 1 {
			// TODO: check
		} else if mbStateItem("toneGeneratorMode") == tgPia {
			collectPianistSongs()
			seg14.w <- pianistSongName(true, mbStateItem("currentPianistSong"))
			k, ok := getPnoKey()
			if ok && k <= 3 {
				keepMbState("currentPianistSong", int(k-1))
			}
			notify <- pianistSongName(true, mbStateItem("currentPianistSong"))
			issueCmd(pmRec, pmSel, 0x0, byte(mbStateItem("currentPianistSong")))
		} else if mbStateItem("toneGeneratorMode") == tgSnd {
			collectSoundSongs()
			seg14.w <- soundSongName(true, mbStateItem("currentSoundSong"), part)
			k, ok := getPnoKey()
			if ok && k <= 10 {
				keepMbState("currentSoundSong", int(k-1))
				keepMbState("currentSoundSongPartSet", part)
			}
			notify <- soundSongName(true, mbStateItem("currentSoundSong"), part)
			issueCmd(smRec, smSel, 0x0, byte(mbStateItem("currentSoundSong")))
			issueCmd(smRec, smRcP, 0x0, byte(part))
		}
	case standby:
		notify <- "STANDBY"
	case recording:
		notify <- "RECORDNG"
	case playing:
		notify <- "PLAYING"
	default:
		log.Print("funny currentRecorderState")
		storeCurrentRecorderState <- idle
	}
}

func songPartSymbol(rec bool, data bool, selected bool) string {
	if rec {
		if data {
			if selected {
				return "X"
			} else {
				return "*"
			}
		} else {
			if selected {
				return "+"
			} else {
				return "_"
			}
		}
	} else { // play
		if data {
			if selected {
				return "X"
			} else {
				return "*"
			}
		} else {
			if selected {
				return "-"
			} else {
				return "_"
			}
		}
	}
}

func pianistSongName(recording bool, n int) string {
	emptiness := songPartSymbol(recording, pianistSongsItem(n).data, true)
	return fmt.Sprintf("P_%d %s", n, emptiness)
}

func soundSongName(recording bool, n int, partSet int) string {
	part1 := songPartSymbol(recording, soundSongsItem(n).part1, partSet&0x1 == 0x0)
	part2 := songPartSymbol(recording, soundSongsItem(n).part2, partSet > 0x0)
	return fmt.Sprintf("S_%d %s.%s", n, part1, part2)
}

func play() {
	currentRecorderState := <-getCurrentRecorderState
	switch currentRecorderState {
	case idle:
		if mbStateItem("usbThumbDrivePresence") == 1 {
			// TODO: check
			issueCmd(pFace, pFUsb, 0x0, 0x0)
			issueCmd(auRec, auSel, 0x0, byte(mbStateItem("currentUsbSong")))
			issueCmd(auRec, au_20, 0x0, 0x0)
		} else if mbStateItem("toneGeneratorMode") == tgPia {
			issueCmd(pFace, pFInt, 0x0, 0x0)
			issueCmd(pmRec, pmSel, 0x0, byte(mbStateItem("currentPianistSong")))
		} else if mbStateItem("toneGeneratorMode") == tgSnd {
			issueCmd(pFace, pFInt, 0x0, 0x0)
			issueCmd(smRec, smSel, 0x0, byte(mbStateItem("currentSoundSong")))
			// issueCmd(smRec, smPlP, 0x0, 0x0)
		}
		storeCurrentRecorderState <- playing
		storePlayerMsg <- "PLAY"
		issueCmdAc(playr, plPla, 0x0, 0x0)
	case playing:
		storeCurrentRecorderState <- idle
		notify <- "END"
		issueCmd(playr, plSto, 0x0, 0x0)
		issueCmd(pFace, pFClo, 0x0, 0x0)
	case recording:
		notify <- "RECORDNG"
	default:
		log.Print("funny currentRecorderState")
		storeCurrentRecorderState <- idle
	}
}

func noticeRecording() {
	storeCurrentRecorderState <- recording
	if mbStateItem("usbThumbDrivePresence") == 1 {
		storePlayerMsg <- "REC"
	} else if mbStateItem("toneGeneratorMode") == tgPia {
		storePlayerMsg <- "REC" // pianist mode adds a seconds count
	} else if mbStateItem("toneGeneratorMode") == tgSnd {
		seg14.w <- "RECORDNG" // sound mode provides no information during recording
	}
}

func usbSongName(n int) string {
	return fmt.Sprintf("NOVUS_%02d", n)
}

func record() {
	currentRecorderState := <-getCurrentRecorderState
	fmt.Println("CALLING RECORD(); STATE=", currentRecorderState)
	switch currentRecorderState {
	case idle:
		fmt.Print("IDLE ")
		if mbStateItem("usbThumbDrivePresence") == 1 {
			storeCurrentRecorderState <- standby
			storePlayerMsg <- "STBY"
			seg14.w <- "STANDBY"
			issueCmd(pFace, pFUsb, 0x0, 0x0)
			issueCmd(playr, plSby, 0x0, 0x1)
		} else if mbStateItem("toneGeneratorMode") == tgPia {
			storeCurrentRecorderState <- standby
			storePlayerMsg <- "STBY"
			issueCmd(pFace, pFInt, 0x0, 0x1)
			issueCmd(playr, plSby, 0x0, 0x1)
		} else if mbStateItem("toneGeneratorMode") == tgSnd {
			storeCurrentRecorderState <- standby
			seg14.w <- "STANDBY"
			issueCmd(pFace, pFInt, 0x0, 0x1)
			issueCmd(playr, plSby, 0x0, 0x1)
		}
	case standby:
		noticeRecording()
		issueCmd(playr, plRec, 0x0, 0x0)
		storeConfirmedUsbSong <- ""
	case recording:
		if mbStateItem("usbThumbDrivePresence") == 1 {
			storeCurrentRecorderState <- idle
			seg14.w <- "STOP"
			storePlayerMsg <- "STOP"
			issueCmd(playr, plSto, 0x0, 0x0)
			issueCmd(auRec, auTyp, 0x0, byte(mbStateItem("currentUsbSongType")))
			issueCmdAc(auRec, auNam, 0xFF, usbSongName(mbStateItem("currentUsbSong")))
			for {
				confirmed := <-getConfirmedUsbSong
				if usbSongName(mbStateItem("currentUsbSong")) == confirmed {
					break
				}
				time.Sleep(500 * time.Millisecond)
				fmt.Println("Waiting for confirmation of", usbSongName(mbStateItem("currentUsbSong")), "=", confirmed, "truth=", usbSongName(mbStateItem("currentUsbSong")) == confirmed)
			}
			issueCmd(pFace, pFClo, 0x0, 0x1)
		} else if mbStateItem("toneGeneratorMode") == tgPia {
			storeCurrentRecorderState <- idle
			seg14.w <- "STOP"
			storePlayerMsg <- "STOP"
			issueCmd(playr, plSto, 0x0, 0x0)
			issueCmd(pFace, pFClo, 0x0, 0x1)
		} else if mbStateItem("toneGeneratorMode") == tgSnd {
			storeCurrentRecorderState <- idle
			notify <- "STOPPED"
			issueCmd(playr, plSto, 0x0, 0x0)
			issueCmd(pFace, pFClo, 0x0, 0x1)
		}
	case playing:
		fmt.Print("PLAYING")
		notify <- "PLAYING"
	default:
		log.Print("funny currentRecorderState")
		storeCurrentRecorderState <- idle
	}
}

func eraseSongParts() {
	seg14.w <- "ERA PRTS"
	k, ok := getPnoKey()
	if ok && k == 1 {
		notify <- "DONE"

		if mbStateItem("usbThumbDrivePresence") == 1 {
			// TODO: check
		} else if mbStateItem("toneGeneratorMode") == tgPia {
			issueCmdAc(pmRec, pmEra, byte(mbStateItem("currentPianistSong")))
		} else if mbStateItem("toneGeneratorMode") == tgSnd {
			issueCmdAc(smRec, smEra, byte(mbStateItem("currentSoundSong")), byte(mbStateItem("currentSoundSongPartSet")))
		}
	} else {
		notify <- "CANCELLD"
	}
}

func loadRegistration() {
	seg14.w <- fmt.Sprintf("REG %02d", mbStateItem("currentRegistration"))
	k, ok := getPnoKey()
	if ok && k <= 16 {
		notify <- fmt.Sprintf("REG %02d", k-1)
		issueCmd(regst, rgLoa, 0x0, k-1)
		issueDtaRq(
			request{regst, rgMod, k - 1, 0x1, 0x0},
			// request{regst, rgNam, k - 1, 0x0},
		)
	} else {
		writeError("cancelled")
	}
}

func storeRegistration() {
	seg14.w <- "TO REGST"
	k, ok := getPnoKey()
	if ok && k <= 16 {
		notify <- fmt.Sprintf("REG %02d", k-1)
		// issueCmd(regst, rgNam, k-1, textarg2)
		issueCmd(regst, rgSto, 0x0, k-1)
	} else {
		writeError("cancelled")
	}
}

func storeToSound() {
	seg14.w <- "TO SOUND"
	k, ok := getPnoKey()
	if ok && k == 1 {
		notify <- "DONE"
		issueCmd(vTech, toSnd, 0x0, 0x21)
	} else {
		notify <- "CANCELLD"
	}
}

func immediateAction(id string, cmd byte, subCmd byte, expectedKey int) {
	seg14.w <- immediateActionNames[id]
	k, ok := getPnoKey()
	kMiD := int(k) - 42 // middle-D = 0
	if ok && kMiD == expectedKey {
		notify <- name(id, kMiD)
		issueCmd(cmd, subCmd, 0x0, 0x0)
	} else {
		notify <- "CANCELLD"
	}
}

// // not settings, but actions
// renameFile
// deleteFile
// sendPgmNumber

func immediateActions() {
	seg14.w <- "COMMAND"
	k, ok := getPnoKey()
	if ok {
		switch k {
		case blkKey[0]:
			immediateAction("factoryReset", mainF, mFact, 12)
		case blkKey[1]:
			immediateAction("usbFormat", files, fFmat, 12)
		case blkKey[2]:
			seg14.w <- "ERASE"
			k, ok := getPnoKey()
			kMiD := int(k) - 42 // middle-D = 0
			if ok && kMiD == 12 {
				notify <- "ERASED"
				issueCmdAc(pmRec, pmEra, 0xFF)
				issueCmdAc(smRec, smEra, 0xFF, 0x2)
			} else {
				notify <- "CANCELLD"
			}
		default:
			writeError("cancelled")
		}
	} else {
		writeError("cancelled")
	}
}

func requestInitialMbData() {
	issueDtaRq(
		request{mainF, mTran, 0x0, 0x1, 0x0},
		request{mainF, mWall, 0x0, 0x1, 0x0},
		request{mainF, mAOff, 0x0, 0x1, 0x0},
		request{bluet, btAuV, 0x0, 0x1, 0x0},
		request{effct, eOnOf, 0x0, 0x1, 0x0},
		request{revrb, rOnOf, 0x0, 0x1, 0x0},
		request{revrb, rType, 0x0, 0x1, 0x0},
		request{revrb, rDpth, 0x0, 0x1, 0x0},
		request{revrb, rTime, 0x0, 0x1, 0x0},
		request{pmSet, pmAmb, 0x0, 0x1, 0x0},
		request{pmSet, pmAmD, 0x0, 0x1, 0x0},
	)
	requestAllVTSettings()
	metro, _ := mbStateItemOk("metronomeOnOff")
	keepMbState("metronomeOnOff", metro) // create or leave unchanged
}

func requestAllVTSettings() {
	issueDtaRq(
		// pianist mode and sound mode parameters
		request{vTech, tCurv, 0x0, 0x1, 0x0},
		request{vTech, voicg, 0x0, 0x1, 0x0},
		request{vTech, dmpNs, 0x0, 0x1, 0x0},
		request{vTech, fBkNs, 0x0, 0x1, 0x0},
		request{vTech, hmDly, 0x0, 0x1, 0x0},
		request{vTech, miTch, 0x0, 0x1, 0x0},
		request{vTech, keyVo, 0x0, 0x1, 0x0},
		request{vTech, hfPdl, 0x0, 0x1, 0x0},
		request{vTech, sfPdl, 0x0, 0x1, 0x0},
		// sound mode-only parameters
		request{vTech, dmpRs, 0x0, 0x1, 0x0},
		request{vTech, strRs, 0x0, 0x1, 0x0},
		request{vTech, uStRs, 0x0, 0x1, 0x0},
		request{vTech, cabRs, 0x0, 0x1, 0x0},
		request{vTech, koEff, 0x0, 0x1, 0x0},
		request{vTech, topBd, 0x0, 0x1, 0x0},
		request{vTech, dcayT, 0x0, 0x1, 0x0},
		request{vTech, streT, 0x0, 0x1, 0x0},
		request{vTech, tmpmt, 0x0, 0x1, 0x0},
		request{vTech, tmKey, 0x0, 0x1, 0x0},
		// smart setting state
		request{vTech, smart, 0x0, 0x1, 0x0},
	)
}

func writeError(errorsItem string) {
	m, ok := errors[errorsItem]
	if ok {
		notify <- m
	} else {
		notify <- "* * * *"
	}
}

func inputSettingsValue(id string, cmd byte, subCmd byte, lowerBound int, upperBound int) {
	seg14.w <- settingTopics[id]
	time.Sleep(1500 * time.Millisecond)
	seg14.w <- name(id, mbStateItem(id))
	k, ok := getPnoKey()
	kMiD := int(k) - 42 // middle-D = 0
	if ok && kMiD >= lowerBound && kMiD <= upperBound {
		notify <- name(id, kMiD)
		keepMbState(id, int(kMiD))
		issueCmd(cmd, subCmd, 0x0, 0x0, kMiD)
		issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode")) // necessary only in a few cases
	} else {
		writeError("cancelled")
	}
}

func inputKeySpecificSettingsValue(id string, cmd byte, subCmd byte, lowerBound int, upperBound int) {
	seg14.w <- settingTopics[id]
	k, ok := getPnoKey()
	if ok {
		clearUserKeySetting <- struct{}{}
		issueDtaRq(request{cmd, subCmd, byte(k - 1), 1, 0x0})
		var s int
		for {
			if seen := <-userKeySettingSeen; seen {
				s = <-userKeySetting
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
		seg14.w <- fmt.Sprintf("%d %d", k-1, s) // TODO: translate k into note name
		k2, ok2 := getPnoKey()
		kMiD := int(k2) - 42 // middle-D = 0
		if ok2 && kMiD >= lowerBound && kMiD <= upperBound {
			notify <- fmt.Sprintf("%d %d", k-1, kMiD) // TODO: translate k into note name
			issueCmd(cmd, subCmd, byte(k-1), byte(kMiD))
		} else {
			writeError("cancelled")
		}
	}
}

func scaledValue(i int, zero int, step float64, unit string) string {
	return fmt.Sprintf("%d %s", float64(zero)+float64(i)*step, unit)
}

func settings() {
	seg14.w <- "SETTINGS"
	k, ok := getPnoKey()
	if !ok {
		writeError("cancelled")
		return
	}
	switch k {
	case blkKey[0]:
		if mbStateItem("toneGeneratorMode") == tgSnd {
			writeError("notInSoundMode")
		} else {
			inputSettingsValue("ambienceType", pmSet, pmAmb, 0, 9)
		}
	case blkKey[1]:
		if mbStateItem("toneGeneratorMode") == tgSnd {
			writeError("notInSoundMode")
		} else {
			inputSettingsValue("ambienceDepth", pmSet, pmAmD, 0, 10)
		}
	case blkKey[2]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("reverbType", revrb, rType, 0, 5)
		}
	case blkKey[3]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("reverbDepth", revrb, rDpth, 1, 10)
		}
	case blkKey[4]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("effectsType", effct, eType, 0, 23)
		}
	case blkKey[5]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("effectsParam1", effct, ePar1, 1, 10)
		}
	case blkKey[6]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("effectsParam2", effct, ePar2, 1, 10)
		}
	case blkKey[7]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("transpose", mainF, mTran, -12, 12)
		}

	case blkKey[8]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			switch mbStateItem("keyboardMode") {
			case kbSpMSingle:
				writeError("onlyIn2SoundModes")
			case kbSpMDual:
				inputSettingsValue("balance", dlSet, dlBal, 0, 16)
			case kbSpMSplit:
				inputSettingsValue("balance", spSet, spBal, 0, 16)
			case kbSpM4hands:
				inputSettingsValue("balance", h4Set, h4Bal, 0, 16)
			}
		}
	case blkKey[9]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			switch mbStateItem("keyboardMode") {
			case kbSpMSingle:
				writeError("onlyIn2SoundModes")
			case kbSpMDual:
				inputSettingsValue("layerOctaveShift", dlSet, dlOcS, -2, 2)
			case kbSpMSplit:
				inputSettingsValue("lowerOctaveShift", spSet, spOcS, 0, 3)
			case kbSpM4hands:
				inputSettingsValue("leftOctaveShift", h4Set, h4LOS, 0, 3)
			}
		}
	case blkKey[10]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			switch mbStateItem("keyboardMode") {
			case kbSpMSingle:
				writeError("onlyIn2SoundModes")
			case kbSpMDual:
				inputSettingsValue("dynamics", dlSet, dlDyn, 1, 10)
			case kbSpMSplit:
				inputSettingsValue("lowerPedal", spSet, spPed, 0, 1)
			case kbSpM4hands:
				inputSettingsValue("rightOctaveShift", h4Set, h4ROS, -3, 0)
			}
		}
	case blkKey[11]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			switch mbStateItem("keyboardMode") {
			case kbSpMSingle:
				writeError("onlyIn2SoundModes")
			case kbSpMDual:
				writeError("onlyInSplitModes")
			case kbSpMSplit:
				seg14.w <- "SPLIT PT"
				k, ok := getPnoKey()
				if ok {
					issueCmd(spSet, spSpP, 0, byte(k+20))
					notify <- fmt.Sprintf("KEY %d", k)
				} else {
					writeError("cancelled")
				}
			case kbSpM4hands:
				seg14.w <- "SPLIT PT"
				k, ok := getPnoKey()
				if ok {
					issueCmd(h4Set, h4SpP, 0, byte(k+20))
					notify <- (fmt.Sprintf("KEY %d", k))
				} else {
					writeError("cancelled")
				}
			}
		}

	case blkKey[13]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			id := "tuning"
			lowerBound := -26
			upperBound := 26
			seg14.w <- settingTopics[id]
			time.Sleep(1500 * time.Millisecond)
			seg14.w <- scaledValue(mbStateItem(id), 440, 0.5, "Hz")
			k, ok := getPnoKey()
			kMiD := int(k) - 42 // middle-D = 0
			if ok && kMiD >= lowerBound && kMiD <= upperBound {
				notify <- scaledValue(kMiD, 440, 0.5, "Hz")
				keepMbState(id, int(kMiD))
				issueCmd(mainF, mTung, 0x0, 0x0, kMiD)
				issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode")) // necessary only in a few cases
			} else {
				writeError("cancelled")
			}
		}
	case blkKey[14]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("damperHold", mainF, mDpHl, 0, 1)
		}
	case blkKey[15]:
		inputSettingsValue("toneControl", mainF, mTone, 0, 6)
		// TODO: brilliance, user tone control
	case blkKey[16]:
		inputSettingsValue("speakerVolume", mainF, mSpkV, 0, 1)
	case blkKey[17]:
		id := "lineInLevel"
		lowerBound := -10
		upperBound := 10
		seg14.w <- settingTopics[id]
		time.Sleep(1500 * time.Millisecond)
		seg14.w <- scaledValue(mbStateItem(id), 0, 1, "dB")
		k, ok := getPnoKey()
		kMiD := int(k) - 42 // middle-D = 0
		if ok && kMiD >= lowerBound && kMiD <= upperBound {
			notify <- scaledValue(kMiD, 0, 1, "dB")
			keepMbState(id, int(kMiD))
			issueCmd(mainF, mLinV, 0x0, 0x0, kMiD)
			issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode")) // necessary only in a few cases
		} else {
			writeError("cancelled")
		}
	case blkKey[18]:
		inputSettingsValue("wallEq", mainF, mWall, 0, 1)
	case blkKey[19]:
		inputSettingsValue("shsMode", hPhon, phShs, 0, 3)
	case blkKey[20]:
		inputSettingsValue("phonesType", hPhon, phTyp, 0, 5)
	case blkKey[21]:
		inputSettingsValue("phonesVolume", hPhon, phVol, 0, 1)

	case blkKey[22]:
		inputSettingsValue("bluetoothMidi", bluet, btMid, 0, 1)
	case blkKey[23]:
		inputSettingsValue("bluetoothAudio", bluet, btAud, 0, 1)
	case blkKey[24]:
		id := "bluetoothAudioVolume"
		lowerBound := -15
		upperBound := 15
		seg14.w <- settingTopics[id]
		time.Sleep(1500 * time.Millisecond)
		seg14.w <- scaledValue(mbStateItem(id), 0, 1, "dB")
		k, ok := getPnoKey()
		kMiD := int(k) - 42 // middle-D = 0
		if ok && kMiD >= lowerBound && kMiD <= upperBound {
			notify <- scaledValue(kMiD, 0, 1, "dB")
			keepMbState(id, int(kMiD))
			issueCmd(bluet, btAud, 0x0, 0x0, kMiD)
			issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode")) // necessary only in a few cases
		} else {
			writeError("cancelled")
		}
	case blkKey[25]:
		inputSettingsValue("midiChannel", midiI, miCha, 0, 15)
	case blkKey[26]:
		inputSettingsValue("localControl", midiI, miLoc, 0, 1)
	case blkKey[27]:
		inputSettingsValue("transmitPgmNumberOnOff", midiI, miTrP, 0, 1)
	case blkKey[28]:
		inputSettingsValue("multiTimbralMode", midiI, miMul, 0, 2)
	case blkKey[29]:
		inputSettingsValue("channelMute", midiI, miMut, 0, 15) // TODO: individual channels
	// case blkKey[25]:
	// 	inputSettingsValue("lcdContrast")
	// case blkKey[26]:
	// 	inputSettingsValue("autoDisplayOff")
	case blkKey[30]:
		inputSettingsValue("autoPowerOff", mainF, mAOff, 0, 3)
	case blkKey[31]:
		inputSettingsValue("metronomeVolume", metro, mVolu, 0, 9)
	case blkKey[32]:
		inputSettingsValue("recorderGainLevel", auRec, auGai, 0, 15)
	case blkKey[33]:
		inputSettingsValue("recorderFileType", auRec, auTyp, 0, 1)
	case blkKey[34]:
		// usbPlayerVolume", playr, plVol, 0, 100
		id := "usbPlayerVolume"
		cmd := playr
		subCmd := plVol
		notify <- settingTopics[id]
		time.Sleep(1500 * time.Millisecond)
		seg14.w <- fmt.Sprint(mbStateItem(id))
		k, ok := getPnoKey()
		kMiD := int(k) - 42 // middle-D = 0
		if ok && kMiD >= 0 && kMiD <= 88 {
			x := scaleVal(0, 100, 45, kMiD)
			notify <- fmt.Sprint(x)
			keepMbState(id, int(x))
			issueCmd(byte(cmd), byte(subCmd), 0x0, 0x0, byte(x))
			issueCmd(tgMod, tgMod, 0x0, mbStateItem("toneGeneratorMode"))
			time.Sleep(time.Second)
		} else {
			writeError("cancelled")
		}
	case blkKey[35]:
		inputSettingsValue("usbPlayerTranspose", auRec, auTrn, -12, 12)
	default:
		writeError("cancelled")
	}
}

func virtualTechnician() {
	seg14.w <- "V TECHN"
	k, ok := getPnoKey()
	if !ok {
		writeError("cancelled")
		return
	}
	switch k {

	case blkKey[0]:
		inputSettingsValue("smartModeVt", vTech, smart, 0, 10)
	case blkKey[1]:
		inputSettingsValue("touchCurve", vTech, tCurv, 0, 6)
	case blkKey[3]:
		inputSettingsValue("voicing", vTech, voicg, 0, 6)
	case blkKey[4]:
		inputKeySpecificSettingsValue("userVoicing", vTech, uVoic, -5, 5)
	case blkKey[5]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("damperResonance", vTech, dmpRs, 0, 10)
		}
	case blkKey[6]:
		inputSettingsValue("damperNoise", vTech, dmpNs, 0, 10)
	case blkKey[7]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("stringResonance", vTech, strRs, 0, 10)
		}
	case blkKey[8]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("undampedStringResonance", vTech, uStRs, 0, 10)
		}
	case blkKey[9]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("cabinetResonance", vTech, cabRs, 0, 9)
		}
	case blkKey[10]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("keyOffEffect", vTech, koEff, 0, 10)
		}
	case blkKey[11]:
		inputSettingsValue("fallBackNoise", vTech, fBkNs, 0, 10)
	case blkKey[12]:
		inputSettingsValue("hammerDelay", vTech, hmDly, 0, 10)
	case blkKey[13]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("topboard", vTech, topBd, 0, 3)
		}
	case blkKey[14]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("decayTime", vTech, dcayT, 1, 10)
		}
	case blkKey[15]:
		inputSettingsValue("minimumTouch", vTech, miTch, 1, 20)
	case blkKey[16]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("stretchTuning", vTech, streT, 0, 3)
		}
	case blkKey[17]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputKeySpecificSettingsValue("userStretchTuning", vTech, uStrT, -50, 50) // TODO: range too big
		}
	case blkKey[18]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("temperament", vTech, tmpmt, 1, 7)
		}
	case blkKey[19]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputKeySpecificSettingsValue("userTemperament", vTech, uTmpm, -50, 50) // TODO: range too big
		}
	case blkKey[20]:
		if mbStateItem("toneGeneratorMode") == tgPia {
			writeError("notInPianistMode")
		} else {
			inputSettingsValue("temperamentKey", vTech, tmKey, 0, 11)
		}
	case blkKey[21]:
		inputSettingsValue("keyVolume", vTech, keyVo, 0, 5)
	case blkKey[22]:
		inputKeySpecificSettingsValue("userKeyVolume", vTech, uKeyV, -50, 50) // TODO: range too big
	case blkKey[23]:
		inputSettingsValue("halfPedalAdjust", vTech, hfPdl, 1, 10)
	case blkKey[24]:
		inputSettingsValue("softPedalDepth", vTech, sfPdl, 1, 10)
	default:
		writeError("cancelled")
	}
}

func hi() {
	issueCmdAc(commu, commu, 0x0, 0x0)
}

func setLocalDefaults() {
	// issueCmd(regst, rgLoa, 0x0, 0)
	issueDtaRq(request{regst, rgMod, 0, 0x1, 0x0})
	// issueCmd(metro, mVolu, 0x0, uint8(0))
	issueCmd(tgMod, tgMod, 0x0, byte(mbStateItem("toneGeneratorMode")))
	keepMbState("currentPianistSong", 0)
	keepMbState("currentSoundSong", 0)
	keepMbState("currentUsbSong", 0)
	keepMbState("currentUsbSongType", 0)
	// storeCurrentRecorderState <- idle
}
