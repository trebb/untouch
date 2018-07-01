package main

import (
	"bytes"
	"fmt"
)

// func input(){
// 	for {
// 		fmt.Print(getChar(), " ")
// 	}
// }
func input() {
	for {
		fmt.Print("CMD-> ")
		var cmd, arg, arg2, textarg, textarg2 string
		var numarg, numarg2 uint
		var signumarg int
		fmt.Scanln(&cmd, &arg, &arg2)
		fmt.Sscanf(arg, "%d", &numarg)
		fmt.Sscanf(arg, "%d", &signumarg)
		fmt.Sscanf(arg, "%s", &textarg)
		fmt.Sscanf(arg2, "%s", &textarg2)
		fmt.Sscanf(arg2, "%d", &numarg2)
		switch cmd {
		case "hi":
			issueCmdAc(commu, 0x7F, 0x0, 0x0)
		case "loadreg":
			loadRegistration(byte(numarg))
		case "storereg":
			issueCmd(regst, rgNam, byte(numarg), textarg2)
			issueCmd(regst, rgSto, 0x0, byte(numarg))
		case "openreg": // close(0), open(1)
			issueCmd(regst, rgOpn, 0x0, byte(numarg))
		case "mode": // sound(0), pianist(1)
			toMb <- append(cmdVnHead, 0x1, 0x0, 0x0, 0x0, 0x1, byte(numarg))
		case "kbmode": // single, dual, split, 4hands
			issueCmd(kbSpl, 0x1, 0x0, byte(numarg))
		case "metro":
			issueCmd(metro, mOnOf, 0x0, byte(numarg))
		case "metrovol":
			issueCmd(metro, mVolu, byte(numarg))
		case "tempo":
			issueCmd(metro, mTmpo, 0x0, uint16(numarg))
		case "timesig":
			issueCmd(metro, mSign, 0x0, byte(numarg))
		case "rendering":
			issueCmd(pmSet, pmRen, 0x0, byte(numarg))
		case "resodepth":
			issueCmd(pmSet, pmRes, 0x0, byte(numarg))
		case "ambience":
			issueCmd(pmSet, pmAmb, 0x0, byte(numarg))
		case "ambiencedepth":
			issueCmd(pmSet, pmAmD, 0x0, byte(numarg))
		case "sound":
			issueCmd(instr, iSing, 0x0, byte(numarg))
		case "sounddual1":
			issueCmd(instr, iDua1, 0x0, byte(numarg))
		case "sounddual2":
			issueCmd(instr, iDua2, 0x0, byte(numarg))
		case "soundsplit1":
			issueCmd(instr, iSpl1, 0x0, byte(numarg))
		case "soundsplit2":
			issueCmd(instr, iSpl2, 0x0, byte(numarg))
		case "sound4hd1":
			issueCmd(instr, i4Hd1, 0x0, byte(numarg))
		case "sound4hd2":
			issueCmd(instr, i4Hd2, 0x0, byte(numarg))
		case "splitting":
			issueCmd(kbSpl, 0x0, 0x0, byte(numarg))
		case "reverb":
			issueCmd(revrb, rOnOf, 0x0, byte(numarg))
		case "reverbtype":
			issueCmd(revrb, rType, 0x0, byte(numarg))
		case "reverbdepth":
			issueCmd(revrb, rDpth, 0x0, byte(numarg))
		case "reverbtime":
			issueCmd(revrb, rTime, 0x0, byte(numarg))
		case "effects":
			issueCmd(effct, eOnOf, 0x0, byte(numarg))
		case "effecttype":
			issueCmd(effct, eType, 0x0, byte(numarg))
		case "effectp1":
			issueCmd(effct, ePar1, 0x0, byte(numarg))
		case "effectp2":
			issueCmd(effct, ePar2, 0x0, byte(numarg))

		case "sel": // 0..3
			// issueCmd(0x7E, 0x2, 0x0, 0x0)
			issueDtaRq(
				request{0x21, 0x61, byte(numarg), 0x1, 0x0})
			issueCmd(0x21, 0x0, 0x0, byte(numarg))
		case "sels": // 0..9
			issueCmd(0x20, 0x0, 0x0, byte(numarg))
		case "playpart": // 0..2
			issueCmd(0x20, 0x1, 0x0, byte(numarg))
		case "recpart":
			issueCmd(0x20, 0x2, 0x0, byte(numarg))
		case "selusb":
			issueCmd(0x22, 0x0, 0x0, byte(numarg))
		case "rec":
			issueCmd(0x71, 0x14, 0x0, 0x1)
		case "recusb":
			issueCmd(0x7E, 0x3, 0x0, 0x0)
			issueCmd(0x71, 0x14, 0x0, 0x1)
			issueCmd(0x22, 0x20, 0x0, 0x0)
		case "rec2":
			issueCmd(0x71, 0x11, 0x0, 0x0)
		case "stop":
			issueCmd(0x71, 0x12, 0x0, 0x0)
		case "play":
			issueCmdAc(0x71, 0x10, 0x0, 0x0)
		case "save": //  0,1 MP3,WAV; name
			issueCmd(0x22, 0x22, 0x0, byte(numarg))
			issueCmdAc(0x22, 0x50, 0xFF, textarg2)
		case "savekso":
			issueCmdAc(0x14, 0x64, byte(numarg), textarg2)
		case "savesmf":
			issueCmdAc(0x14, 0x65, byte(numarg), textarg2)
		case "erase": // 0..2
			issueCmdAc(0x21, 0x40, byte(numarg))
		case "erases": // 0..9; 0..2  (internal song; parts set)
			issueCmdAc(0x20, 0x40, byte(numarg), byte(numarg2))
		case "eraseall":
			issueCmdAc(0x21, 0x40, 0xFF)
		case "erasealls":
			issueCmdAc(0x20, 0x40, 0xFF, 0x2)
		case "ls":
			for i := 0; i < 0x3; i++ {
				issueDtaRq(
					request{0x21, 0x61, byte(i), 0x0})
			}
			for i := 0; i < 0xA; i++ {
				issueDtaRq(
					request{0x20, 0x61, byte(i), 0x0})
			}
		case "loadfromusb1":
			issueDtaRq(
				request{0x14, 0x40, 0xFF, 0x0})
		case "loadfromusb2": // sound song, usb song
			// doesn't seem to work for empty sound songs
			issueCmd(files, 0x0, byte(numarg), byte(numarg2))
			issueCmdAc(files, 0x60, 0x0)
		case "usbmempl":
			issueDtaRq(
				request{0x22, 0x40, 0xFF, 0x0})

		case "builtinsong": // song list (0, 2, 3, 5, 7, 9), song number
			issueCmd(biSng, 0x40, byte(numarg), byte(numarg2))
		case "soundsong":
			issueDtaRq(request{smRec, 0x61, byte(numarg), 0x1, 0x0})
			issueCmd(smRec, 0x0, 0x0, byte(numarg))
		case "pianistsong":
			issueDtaRq(request{pmRec, 0x61, byte(numarg), 0x1, 0x0})
			issueCmd(pmRec, 0x0, 0x0, byte(numarg))
		case "audiorecname":
			issueCmdAc(auRec, 0x50, 0xFF, textarg)
		case "playbackmode":
			issueCmd(playr, 0x18, 0x0, byte(numarg))
		case "playbackvol":
			issueCmd(playr, 0x7, 0x0, byte(numarg))
		case "transpose":
			issueCmd(mainF, mTran, 0x0, byte(signumarg))
		case "btmidi":
			issueCmd(bluet, btMid, 0x0, byte(numarg))
		case "btaudio":
			issueCmd(bluet, btAud, 0x0, byte(numarg))
		case "btaudiovol":
			issueCmd(bluet, btAuV, 0x0, byte(signumarg))
		case "walleq":
			issueCmd(mainF, mWall, 0x0, byte(numarg))
		case "autopoweroff":
			issueCmd(mainF, mAOff, 0x0, byte(numarg))
		case "autopoweroff?":
			issueDtaRq(
				request{mainF, mAOff, 0x0, 0x1, 0x0})
		case "factory":
			issueCmdAc(mainF, mFact, 0x0, 0x0)
		case "key":
			issueDtaRq(request{hardw, hwKey, 0x0, 0x1, 0x0})
		case "instrparams":
			// pianist mode and sound mode parameters
			issueDtaRq(
				request{pmSet, pmAmb, 0x0, 0x1, 0x0},
				request{pmSet, pmAmD, 0x0, 0x1, 0x0})
			// sound mode parameters
			issueDtaRq(
				request{revrb, rOnOf, 0x0, 0x1, 0x0},
				request{effct, eOnOf, 0x0, 0x1, 0x0},
				request{mainF, mTran, 0x0, 0x1, 0x0})
		case "vtparams":
			// pianist mode parameters
			issueDtaRq(
				request{vTech, tCurv, 0x0, 0x1, 0x0},
				request{vTech, voicg, 0x0, 0x1, 0x0},
				request{vTech, dmpNs, 0x0, 0x1, 0x0},
				request{vTech, fBkNs, 0x0, 0x1, 0x0},
				request{vTech, hmDly, 0x0, 0x1, 0x0},
				request{vTech, miTch, 0x0, 0x1, 0x0},
				request{vTech, keyVo, 0x0, 0x1, 0x0},
				request{vTech, hfPdl, 0x0, 0x1, 0x0},
				request{vTech, sfPdl, 0x0, 0x1, 0x0})
			// sound mode parameters
			issueDtaRq(
				request{vTech, tCurv, 0x0, 0x1, 0x0},
				request{vTech, voicg, 0x0, 0x1, 0x0},
				request{vTech, dmpRs, 0x0, 0x1, 0x0},
				request{vTech, dmpNs, 0x0, 0x1, 0x0},
				request{vTech, strRs, 0x0, 0x1, 0x0},
				request{vTech, uStRs, 0x0, 0x1, 0x0},
				request{vTech, cabRs, 0x0, 0x1, 0x0},
				request{vTech, koEff, 0x0, 0x1, 0x0},
				request{vTech, fBkNs, 0x0, 0x1, 0x0},
				request{vTech, hmDly, 0x0, 0x1, 0x0},
				request{vTech, topBd, 0x0, 0x1, 0x0},
				request{vTech, dcayT, 0x0, 0x1, 0x0},
				request{vTech, miTch, 0x0, 0x1, 0x0},
				request{vTech, streT, 0x0, 0x1, 0x0},
				request{vTech, tmpmt, 0x0, 0x1, 0x0},
				request{vTech, vt_0F, 0x0, 0x1, 0x0},
				request{vTech, keyVo, 0x0, 0x1, 0x0},
				request{vTech, hfPdl, 0x0, 0x1, 0x0},
				request{vTech, sfPdl, 0x0, 0x1, 0x0})
			// smart setting state
			issueDtaRq(
				request{vTech, smart, 0x0, 0x1, 0x0})
		case "reverbparams":
			issueDtaRq(
				request{revrb, rType, 0x0, 0x1, 0x0},
				request{revrb, rDpth, 0x0, 0x1, 0x0},
				request{revrb, rTime, 0x0, 0x1, 0x0})
		case "effectparams":
			issueDtaRq(
				request{effct, eType, 0x0, 0x1, 0x0},
				request{effct, ePar1, 0x0, 0x1, 0x0},
				request{effct, ePar2, 0x0, 0x1, 0x0})
		case "soundsettings":
			issueDtaRq(
				request{mainF, mTone, 0x0, 0x1, 0x0},
				request{mainF, mSpkV, 0x0, 0x1, 0x0},
				request{mainF, mLinV, 0x0, 0x1, 0x0},
				request{mainF, mWall, 0x0, 0x1, 0x0},
				request{hPhon, phShs, 0x0, 0x1, 0x0},
				request{hPhon, phVol, 0x0, 0x1, 0x0})
		case "uvoicing":
			issueDtaRq(
				request{vTech, uVoic, byte(numarg), 4, 0x1, 0x0})
		case "ustretch":
			issueDtaRq(
				request{vTech, uStrT, byte(numarg), 0x1, 0x0})
		case "utemperament":
			issueDtaRq(
				request{vTech, uTmpm, byte(numarg), 0x1, 0x0})
		case "ukeyvolume":
			issueDtaRq(
				request{vTech, uKeyV, byte(numarg), 0x1, 0x0})
		case "utone":
			issueDtaRq(
				request{mainF, mUTon, 0x0, 0x1, 0x0},
				request{mainF, mUTon, 0x1, 0x1, 0x0},
				request{mainF, mUTon, 0x2, 0x1, 0x0},
				request{mainF, mUTon, 0x3, 0x1, 0x0},
				request{mainF, mUTon, 0x4, 0x1, 0x0},
				request{mainF, mUTon, 0x5, 0x1, 0x0})
		case "usbrename1":
			issueDtaRq(
				request{files, fMvNm, 0xFF, 0x0})
		case "usbrename2":
			issueCmd(files, fMvNu, 0x0, byte(numarg))
			issueCmdAc(files, fName, 0xFF, textarg2)
		case "usbdelete1":
			issueDtaRq(
				request{files, fRmNm, 0xFF, 0x0})
		case "usbdelete2":
			issueCmd(files, fRmNu, 0x0, byte(numarg))
			issueCmdAc(files, fRmCf, 0xFF, 0x0)
		case "usbformat":
			issueDtaRq(
				request{files, fFmat, 0xFF, 0x1, 0x0})
		case "midimute":
			issueDtaRq(
				request{midiI, miMut, byte(numarg), 0x1, 0x0})

		default:
			fmt.Println("???", cmd, arg)
		}
	}
}

func parse() {
	for {
		hdr := []byte{}
		msg := msg{}
		var msgIndex int
		for msgIndex = -1; msgIndex < 0; msgIndex = bytes.Index(hdr, []byte{0x55, 0xAA, 0x0}) {
			b := <-rawBytes
			hdr = append(hdr, b)
		}
		if msgIndex > 0 {
			notImpl(hdr[:msgIndex], "Headless rubbish")
		}
		for i, b := range hdr[msgIndex:] {
			msg[i] = b
		}
		for i := 3; i < 7; i++ {
			msg[i] = <-rawBytes
		}
		if a, ok := actions[msg]; ok {
			var i int
			for i = 7; i < 9; i++ {
				msg[i] = <-rawBytes
			}
			for k := 0; k < int(msg[8]); k++ {
				msg[i+k] = <-rawBytes
			}
			a(msg)
		} else {
			notImpl(msg[:7])
		}
	}
}
