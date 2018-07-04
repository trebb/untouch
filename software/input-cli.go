package main

import (
	"fmt"
	"log"
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
		case "m": // aggregate mode: pianist, all sound mode splits
			if mbStateItem("toneGeneratorMode") == tgSnd {
				fmt.Println("current keyboard split mode is", kbMode[mbStateItem("keyboardMode")])
			} else {
				fmt.Println("current keyboard split mode is Pianist mode")
			}
			k := getPnoKey()
			if k == 1 { // pianist mode
				issueCmd(tgMod, tgMod, 0x0, tgPia)
			} else if k <= 5 { // sound mode
				issueCmd(tgMod, tgMod, 0x0, tgSnd)
				issueCmd(kbSpl, kbSpM, 0x0, k-2) // KB split mode 0..3
			} else {
				fmt.Println(k, "is not a mode")
			}
		case "mode": // sound(0), pianist(1)
			issueCmd(tgMod, tgMod, 0x0, byte(numarg))
		case "metro":
			issueCmd(metro, mOnOf, 0x0, byte(numarg))
		case "metrovol":
			issueCmd(metro, mVolu, 0x0, byte(numarg))
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
		case "s": // aggregate sound: rendering of pianist mode or first/only sound of sound mode
			switch mbStateItem("toneGeneratorMode") {
			case tgPia:
				fmt.Println("current sound (pianist mode) is",
					renderingCharacter[mbStateItem("renderingCharacter")])
				k := getPnoKey()
				if k <= 10 {
					issueCmd(pmSet, pmRen, 0x0, k-1)
				} else {
					fmt.Println(k, "is not a pianist mode sound")
				}
			case tgSnd:
				switch mbStateItem("keyboardMode") {
				case 0: // single
					fmt.Println("current single sound (sound mode) is",
						instrumentSound[mbStateItem("single")])
					k := getPnoKey()
					issueCmd(instr, iSing, 0x0, k-1)
				case 1: // dual
					fmt.Println("current first dual sound (sound mode) is",
						instrumentSound[mbStateItem("dual1")])
					k := getPnoKey()
					issueCmd(instr, iDua1, 0x0, k-1)
				case 2: // split
					fmt.Println("current first split sound (sound mode) is",
						instrumentSound[mbStateItem("split1")])
					k := getPnoKey()
					issueCmd(instr, iSpl1, 0x0, k-1)
				case 3: // 4hands
					fmt.Println("current first 4hands sound (sound mode) is",
						instrumentSound[mbStateItem("4hands1")])
					k := getPnoKey()
					issueCmd(instr, i4Hd1, 0x0, k-1)
				default:
					log.Print("bad keyboardMode")
				}
			default:
				log.Print("bad toneGeneratorMode")
			}
		case "s2": // aggregate sound: second sound of sound mode
			switch mbState["toneGeneratorMode"] {
			case tgPia:
				fmt.Println("pianist mode has no second sound")
			case tgSnd:
				switch mbStateItem("keyboardMode") {
				case 0: // single
					fmt.Println("single has no second sound")
				case 1: // dual
					fmt.Println("current second dual sound (sound mode) is",
						instrumentSound[mbStateItem("dual2")])
					k := getPnoKey()
					issueCmd(instr, iDua2, 0x0, k-1)
				case 2: // split
					fmt.Println("current second split sound (sound mode) is",
						instrumentSound[mbStateItem("split2")])
					k := getPnoKey()
					issueCmd(instr, iSpl2, 0x0, k-1)
				case 3: // 4hands
					fmt.Println("current second 4hands sound (sound mode) is",
						instrumentSound[mbStateItem("4hands2")])
					k := getPnoKey()
					issueCmd(instr, i4Hd2, 0x0, k-1)
				default:
					log.Print("bad keyboardMode")
				}
			default:
				log.Print("bad toneGeneratorMode")
			}
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
		case "kbmode": // single, dual, split, 4hands
			issueCmd(kbSpl, kbSpM, 0x0, byte(numarg))
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
				request{pmRec, pmEmp, byte(numarg), 0x1, 0x0})
			issueCmd(pmRec, pmSel, 0x0, byte(numarg))
		case "sels": // 0..9
			issueCmd(smRec, smSel, 0x0, byte(numarg))
		case "playpart": // 0..2
			issueCmd(smRec, smPlP, 0x0, byte(numarg))
		case "recpart":
			issueCmd(smRec, smRcP, 0x0, byte(numarg))
		case "selusb":
			issueCmd(auRec, auSel, 0x0, byte(numarg))
		case "rec":
			issueCmd(playr, plSby, 0x0, 0x1)
		case "recusb":
			issueCmd(rpFce, rpUsb, 0x0, 0x0)
			issueCmd(playr, plSby, 0x0, 0x1)
			issueCmd(auRec, 0x20, 0x0, 0x0)
		case "rec2":
			issueCmd(playr, plRec, 0x0, 0x0)
		case "stop":
			issueCmd(playr, plSto, 0x0, 0x0)
		case "play":
			issueCmdAc(playr, plPla, 0x0, 0x0)
		case "save": //  0,1 MP3,WAV; name
			issueCmd(auRec, auTyp, 0x0, byte(numarg))
			issueCmdAc(auRec, auNam, 0xFF, textarg2)
		case "savekso":
			issueCmdAc(files, fSvKs, byte(numarg), textarg2)
		case "savesmf":
			issueCmdAc(files, fSvSm, byte(numarg), textarg2)
		case "erase": // 0..2
			issueCmdAc(pmRec, pmEra, byte(numarg))
		case "erases": // 0..9; 0..2  (internal song; parts set)
			issueCmdAc(smRec, smEra, byte(numarg), byte(numarg2))
		case "eraseall":
			issueCmdAc(pmRec, pmEra, 0xFF)
		case "erasealls":
			issueCmdAc(smRec, smEra, 0xFF, 0x2)
		case "ls":
			for i := 0; i < 0x3; i++ {
				issueDtaRq(
					request{pmRec, pmEmp, byte(i), 0x0})
			}
			for i := 0; i < 0xA; i++ {
				issueDtaRq(
					request{smRec, smEmp, byte(i), 0x0})
			}
		case "loadfromusb1":
			issueDtaRq(
				request{files, fUsNm, 0xFF, 0x0})
		case "loadfromusb2": // sound song, usb song
			// doesn't seem to work for empty sound songs
			issueCmd(files, fUsNu, byte(numarg), byte(numarg2))
			issueCmdAc(files, fUsCf, 0x0)
		case "usbmempl":
			issueDtaRq(
				request{auRec, fUsNm, 0xFF, 0x0})

		case "builtinsong": // song list (0, 2, 3, 5, 7, 9), song number
			issueCmd(biSng, 0x40, byte(numarg), byte(numarg2))
		case "soundsong":
			issueDtaRq(request{smRec, smEmp, byte(numarg), 0x1, 0x0})
			issueCmd(smRec, smSel, 0x0, byte(numarg))
		case "pianistsong":
			issueDtaRq(request{pmRec, pmEmp, byte(numarg), 0x1, 0x0})
			issueCmd(pmRec, pmSel, 0x0, byte(numarg))
		case "audiorecname":
			issueCmdAc(auRec, auNam, 0xFF, textarg)
		case "playbackmode":
			issueCmd(playr, plPbM, 0x0, byte(numarg))
		case "playbackvol":
			issueCmd(playr, plVol, 0x0, byte(numarg))
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
			go func() {
				fmt.Println("piano key", getPnoKey(), "pressed")
			}()
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
