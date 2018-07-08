package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"time"
)

type msg [256]byte // Big Enough (TM)

var (
	mbPort = flag.String("mb", "/dev/ttyUSB0", "the serial interface connected to the mainboard")

	rawBytes    = make(chan byte, 1000)
	notImplMsgs = make(chan string, 100)
	toMb        = make(chan []byte, 100)
	pnoKeys     = make(chan uint8, 100)
)

func main() {
	flag.Parse()
	s, err := serial.OpenPort(&serial.Config{Name: *mbPort, Baud: 115200})
	if err != nil {
		log.Print(err)
	}
	addDuplicateNames()
	go parse()
	go rxd(*s)
	go txd(*s)
	go input()
	for {
		select {
		case x := <-notImplMsgs:
			fmt.Println(x)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func rxd(port serial.Port) {
	buf := make([]byte, 1)
	for {
		_, err := port.Read(buf)
		if err != nil {
			log.Print(err)
		}
		rawBytes <- buf[0]
	}
}

func txd(port serial.Port) {
	for {
		message := <-toMb
		_, err := port.Write(message)
		if err != nil {
			log.Print(err)
		}
	}
}

// func txd(port serial.Port) { // debugging version
// 	for {
// 		message := <-toMb
// 		notImpl(message, "TX")
// 	}
// }

// func rxd(port string) {
// 	payload := []byte{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0,
// 		0x55, 0xAA, 0x0, 0x6E, 0x1, 0x71, 0x1, 0x0, 0x2, 0xFF, 0x2,
// 		0x55, 0xAA, 0x0, 0x6E, 0x1, 0x71, 0x1, 0x0, 0x2, 0xFF, 0xFE,
// 	}
// 	for _, i := range payload {
// 		rawBytes <- i
// 	}
// }

var ( // message headers
	cmdVnHead = []byte{hdr0, hdr1, hdr2, cmdVn} // vanilla UI command
	cmdAcHead = []byte{hdr0, hdr1, hdr2, cmdAc} // UI command that expects acknoledgement
	dtaRqHead = []byte{hdr0, hdr1, hdr2, dtaRq} // data request
)

func parse() {
	for {
		hdr := []byte{}
		msg := msg{}
		var msgIndex int
		for msgIndex = -1; msgIndex < 0; msgIndex = bytes.Index(hdr, []byte{hdr0, hdr1, hdr2}) {
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

func issueCmd(topic byte, subtopic byte, item byte, params ...interface{}) {
	m1 := append(cmdVnHead, 0x1, topic, subtopic, item)
	var m2 []byte
	l := 0
	for _, p := range params {
		switch x := p.(type) {
		case byte:
			m2 = append(m2, x)
			l++
		case uint16:
			m2 = append(m2, uint16Msg(x)...)
			l += 2
		case string:
			m2 = append(m2, []byte(x)...)
			l += len(x)
		default:
			log.Print("unknown cmd parameter")
		case int: // byte, actually
			m2 = append(m2, byte(x))
			l++
		}
	}
	m1 = append(m1, byte(l))
	toMb <- append(m1, m2...)
}

func issueTglCmd(name string, topic byte, subtopic byte, item byte) (newState byte) {
	switch mbStateItem(name) {
	case 0:
		newState = 0x1
		issueCmd(topic, subtopic, item, newState)
	case 1:
		newState = 0x0
		issueCmd(topic, subtopic, item, newState)
	default:
		log.Print("unknown ", name, " switch state")
	}
	return
}

func issueCmdAc(topic byte, subtopic byte, item byte, params ...interface{}) {
	m1 := append(cmdAcHead, 0x1, topic, subtopic, item)
	var m2 []byte
	l := 0
	for _, p := range params {
		switch x := p.(type) {
		case uint16:
			m2 = append(m2, uint16Msg(x)...)
			l += 2
		case string:
			m2 = append(m2, []byte(x)...)
			l += len(x)
		case byte:
			m2 = append(m2, x)
			l++
		case int: // byte, actually
			m2 = append(m2, byte(x))
			l++
		default:
			log.Print("unknown cmd parameter")
		}
	}
	m1 = append(m1, byte(l))
	toMb <- append(m1, m2...)
}

type request []byte

func issueDtaRq(requests ...request) {
	var m2 []byte
	var i int
	var r request
	for i, r = range requests {
		m2 = append(m2, r...)
	}
	i++
	m1 := append(dtaRqHead, byte(i))
	toMb <- append(m1, m2...)
}

func loadRegistration(reg byte) {
	issueCmd(regst, rgLoa, 0x0, reg)
	issueDtaRq(
		request{regst, rgMod, reg, 0x1, 0x0})
	issueDtaRq(
		request{regst, rgMod, reg, 0x1, 0x0},
		request{regst, rgNam, reg, 0x0})
}

func msgInt16(b []byte) int16 {
	var r int16
	buf := bytes.NewReader(b[:2])
	err := binary.Read(buf, binary.BigEndian, &r)
	if err != nil {
		log.Print(err)
	}
	return r
}

func uint16Msg(i uint16) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		log.Print(err)
	}
	return buf.Bytes()[:2]
}

func notImpl(m interface{}, what ...string) {
	var line string
	for _, w := range what {
		line += w
	}
	if line == "" {
		line = "Not implemented"
	}
	line += ":"
	switch message := m.(type) {
	case []byte:
		for _, c := range message {
			line += fmt.Sprintf(" %2X", c)
		}
	case msg:
		for _, c := range message[:9+message[8]] {
			line += fmt.Sprintf(" %2X", c)
		}
	}
	notImplMsgs <- line
}

func name(tableKey string, i uint8) string {
	if i >= 0 && i < uint8(len(names[tableKey])) {
		return names[tableKey][i]
	}
	return fmt.Sprint("undef:", i)
}

var mbState = make(map[string]int)

func mbStateItem(key string) int {
	s, ok := mbState[key]
	if ok {
		return s
	} else {
		log.Print("missing item ", key, " in mbState")
		return 0
	}
}

func keepMbState(key string, payload interface{}) {
	switch x := payload.(type) {
	case byte:
		mbState[key] = int(x)
		fmt.Println("NOTICED:", key, name(key, x))
	case int:
		mbState[key] = x
		fmt.Println("NOTICED:", key, "=", x)
	default:
		log.Print(payload, "has an unknown type")
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
		request{vTech, smart, 0x0, 0x1, 0x0},
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
		request{vTech, keyVo, 0x0, 0x1, 0x0},
		request{vTech, hfPdl, 0x0, 0x1, 0x0},
		request{vTech, sfPdl, 0x0, 0x1, 0x0},
		// request{vTech, vt_0F, 0x0, 0x1, 0x0},
	)
	mbState["metronomeOnOff"] = mbState["metronomeOnOff"] // create or leave unchanged
}

var actions = map[msg]func(msg){
	// key contains the bytes 0..6 of the raw message in an otherwise pristine msg

	// 55    AA    00    6E    01
	// 55    AA    00    6E    01    00
	{hdr0, hdr1, hdr2, mbMsg, 0x01, tgMod, 0x00}: func(m msg) { keepMbState("toneGeneratorMode", m[9]) },
	// 55    AA    00    6E    01    01
	{hdr0, hdr1, hdr2, mbMsg, 0x01, kbSpl, 0x00}: func(m msg) { keepMbState("keyboardMode", m[9]) },
	// 55    AA    00    6E    01    02
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, iSing}: func(m msg) { keepMbState("single", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, iDua1}: func(m msg) { keepMbState("dual1", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, iDua2}: func(m msg) { keepMbState("dual2", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, iSpl1}: func(m msg) { keepMbState("split1", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, iSpl2}: func(m msg) { keepMbState("split2", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, i4Hd1}: func(m msg) { keepMbState("4hands1", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, instr, i4Hd2}: func(m msg) { keepMbState("4hands2", m[9]) },
	// 55    AA    00    6E    01    04
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmRen}: func(m msg) { keepMbState("renderingCharacter", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmRes}: func(m msg) { keepMbState("resonanceDepth", int(m[9])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmAmb}: func(m msg) { keepMbState("ambienceType", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmAmD}: func(m msg) { keepMbState("ambienceDepth", int(m[9])) },
	// 55    AA    00    6E    01    08
	{hdr0, hdr1, hdr2, mbMsg, 0x01, revrb, rOnOf}: func(m msg) { keepMbState("reverbOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, revrb, rDpth}: func(m msg) { keepMbState("reverbDepth", int(m[9])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, revrb, rTime}: func(m msg) { keepMbState("reverbTime", int(m[9])) },
	// 55    AA    00    6E    01    09
	{hdr0, hdr1, hdr2, mbMsg, 0x01, effct, eOnOf}: func(m msg) { keepMbState("effectsOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, effct, ePar1}: func(m msg) { keepMbState("effectsParam1", int(m[9])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, effct, ePar2}: func(m msg) { keepMbState("effectsParam2", int(m[9])) },
	// 55    AA    00    6E    01    0A
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mOnOf}: func(m msg) { keepMbState("metronomeOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mTmpo}: func(m msg) { keepMbState("metronomeTempo", int(msgInt16(m[9:11]))) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mSign}: func(m msg) { keepMbState("rhythmPattern", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mVolu}: func(m msg) { keepMbState("metronomeVolume", int(m[9])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mBeat}: func(m msg) {
		fmt.Println("metronome beat", m[9])
	},
	// 55    AA    00    6E    01    0F
	{hdr0, hdr1, hdr2, mbMsg, 0x01, regst, rgOpn}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("close registration screen")
			case 1:
				fmt.Println("open registration screen")
				loadRegistration(0x0)
			default:
				notImpl(m, "unknown registration screen state")
			}
		default:
			notImpl(m, "unknown registration screen stuff")
		}
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, regst, rgLoa}: func(m msg) { keepMbState("currentRegistration", int(m[9])) },
	// 55    AA    00    6E    01    10
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mainF, mTran}: func(m msg) { keepMbState("transpose", int(int8(m[9]))) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mainF, m__0B}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    14
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fPgrs}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fMvNm}: func(m msg) {
		fmt.Printf("rename done: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fRmNm}: func(m msg) {
		fmt.Printf("delete: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fMvEx}: func(m msg) {
		fmt.Printf("rename: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fRmEx}: func(m msg) {
		fmt.Printf("delete: USB filename extension(%d)=%d\n", m[7], m[9])
	},
	// 55    AA    00    6E    01    16
	{hdr0, hdr1, hdr2, mbMsg, 0x01, bluet, btAud}: func(m msg) { keepMbState("bluetoothAudio", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, bluet, btMid}: func(m msg) { keepMbState("bluetoothMidi", m[9]) },
	// 55    AA    00    6E    01    20
	{hdr0, hdr1, hdr2, mbMsg, 0x01, smRec, smSel}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, smRec, smPlP}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, smRec, smRcP}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, smRec, smPEm}: func(m msg) {
		switch m[9] {
		case 0:
			fmt.Println("part", m[7], "empty")
		case 1:
			fmt.Println("part", m[7], "contains a recording")
		default:
			notImpl(m, fmt.Sprint("unknown state of part", m[7]))
		}
	},
	// 55    AA    00    6E    01    21
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmRec, pmSel}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    22
	{hdr0, hdr1, hdr2, mbMsg, 0x01, auRec, 0x30}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, auRec, auPNm}: func(m msg) {
		fmt.Printf("USB playback filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, auRec, auPEx}: func(m msg) {
		fmt.Printf("USB playback filename extension(%d)=%s\n", m[7], name("fileExt", m[9]))
	},
	// 55    AA    00    6E    01    32
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg32, 0x00}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg32, 0x04}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg32, 0x05}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    3F
	{hdr0, hdr1, hdr2, mbMsg, 0x01, biSng, 0x40}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    60
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x41}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x43}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x46}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x48}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x49}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x4A}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, 0x4D}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    61
	{hdr0, hdr1, hdr2, mbMsg, 0x01, romId, 0x00}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, romId, 0x01}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, romId, 0x02}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    65
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg65, 0x00}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg65, 0x01}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    70
	{hdr0, hdr1, hdr2, mbMsg, 0x01, hardw, hwUsb}: func(m msg) { keepMbState("UsbThumbDrivePresence", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, hardw, hwHPh}: func(m msg) { keepMbState("phonesPresence", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, hardw, hw_03}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    71
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plDur}: func(m msg) { // really duration?
		fmt.Println("duration", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, 0x01}: func(m msg) {
		fmt.Println("bar/second count", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, 0x02}: func(m msg) { // really duration?
		fmt.Println("duration", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plBrC}: func(m msg) {
		fmt.Println("bar count", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plBea}: func(m msg) {
		fmt.Println("beat", m[9], msgInt16(m[10:12]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, 0x08}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, 0x09}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, 0x11}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, 0x12}: func(m msg) {
		switch m[9] {
		case 0x0:
			fmt.Println("stopped")
		default:
			notImpl(m, "unknown stopped mode")
		}
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plA_B}: func(m msg) {
		fmt.Println("A-B repeat mode", m[9])
	},
	// 55    AA    00    6E    01    7E
	{hdr0, hdr1, hdr2, mbMsg, 0x01, rpFce, 0x00}: func(m msg) {
		switch m[7] {
		case rpClo:
			fmt.Println("close recorder/player")
		case rpInt:
			fmt.Println("open internal recorder/player")
		case rpUsb:
			fmt.Println("open USB recorder/player")
		case rpDem:
			fmt.Println("open demo song player")
		case rpLes:
			fmt.Println("open lesson song player")
		case rpCon:
			fmt.Println("open concert magic player")
		case rpPno:
			fmt.Println("open piano music player")
		default:
			notImpl(m, "unknown recorder/player face request")
		}
	},
	// 55    AA    00    6E    01    7F
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, 0x00}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, 0x01}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, 0x04}: func(m msg) { notImpl(m) },
	// 55    AA    00    71    01
	// 55    AA    00    71    01    10
	{hdr0, hdr1, hdr2, mbCAc, 0x01, mainF, mFact}: func(m msg) {
		fmt.Println("Ok, factory reset")
	},
	// 55    AA    00    71    01    14
	{hdr0, hdr1, hdr2, mbCAc, 0x01, files, fUsCf}: func(m msg) {
		fmt.Println("done: load from USB")
	},
	{hdr0, hdr1, hdr2, mbCAc, 0x01, files, fSvKs}: func(m msg) {
		fmt.Printf("save .KSO to USB: confirming filename=%s\n", m[9:])
	},
	{hdr0, hdr1, hdr2, mbCAc, 0x01, files, fSvSm}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("save .MID to USB: confirming filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown save .MID file stuff")
		}
	},
	{hdr0, hdr1, hdr2, mbCAc, 0x01, files, fName}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("rename: confirming new USB filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown rename USB file stuff")
		}
	},
	{hdr0, hdr1, hdr2, mbCAc, 0x01, files, fRmCf}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("delete: confirming USB file deletion\n")
		default:
			notImpl(m, "unknown delete USB file stuff")
		}
	},
	{hdr0, hdr1, hdr2, mbCAc, 0x01, files, fFmat}: func(m msg) {
		fmt.Println("USB format done")
	},
	// 55    AA    00    71    01    20
	{hdr0, hdr1, hdr2, mbCAc, 0x01, smRec, smEra}: func(m msg) { notImpl(m) },
	// 55    AA    00    71    01    21
	{hdr0, hdr1, hdr2, mbCAc, 0x01, pmRec, pmEra}: func(m msg) { notImpl(m) },
	// 55    AA    00    71    01    22
	{hdr0, hdr1, hdr2, mbCAc, 0x01, auRec, auNam}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("audio recorder filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown audio recorder filename stuff")
		}
	},
	// 55    AA    00    71    01    71
	{hdr0, hdr1, hdr2, mbCAc, 0x01, playr, 0x10}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("duration", msgInt16(m[9:11]))
		default:
			notImpl(m, "unknown duration count stuff")
		}
	},
	// 55    AA    00    71    01    7F
	{hdr0, hdr1, hdr2, mbCAc, 0x01, commu, commu}: func(m msg) {
		keepMbState("mainboardSeen", int(1))
		requestInitialMbData()
	},
	// 55    AA    00    72    01
	// 55    AA    00    72    01    04
	{hdr0, hdr1, hdr2, mbDRq, 0x01, pmSet, pmAmb}: func(m msg) { keepMbState("ambienceType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, pmSet, pmAmD}: func(m msg) { keepMbState("ambienceDepth", int(m[9])) },
	// 55    AA    00    72    01    08
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rOnOf}: func(m msg) { keepMbState("reverbOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rType}: func(m msg) { keepMbState("reverbType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rDpth}: func(m msg) { keepMbState("reverbDepth", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rTime}: func(m msg) { keepMbState("reverbTime", int(m[9])) },
	// 55    AA    00    72    01    09
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, eOnOf}: func(m msg) { keepMbState("effectsOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, eType}: func(m msg) { keepMbState("effectsType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, ePar1}: func(m msg) { keepMbState("effectsParam1", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, ePar2}: func(m msg) { keepMbState("effectsParam2", int(m[9])) },
	// 55    AA    00    72    01    0F
	{hdr0, hdr1, hdr2, mbDRq, 0x01, regst, rgNam}: func(m msg) {
		if reg := m[7]; reg <= 0xF {
			fmt.Printf("name of registration %d = %s\n", reg, m[9:])
		} else {
			fmt.Println("unknown registration (name)", reg)
		}
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, regst, rgMod}: func(m msg) {
		if reg := m[7]; reg <= 0xF {
			fmt.Println("registration", reg, "is for", name("toneGeneratorMode", m[9]))
		} else {
			fmt.Println("unknown registration (mode)", reg)
		}
	},
	// 55    AA    00    72    01    10
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mTran}: func(m msg) { keepMbState("transpose", int(int8(m[9]))) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mTone}: func(m msg) { keepMbState("toneControl", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mSpkV}: func(m msg) { keepMbState("speakerVolume", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mLinV}: func(m msg) { keepMbState("line-in level", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mWall}: func(m msg) { keepMbState("wallEq", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, m__06}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, m__07}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mAOff}: func(m msg) { keepMbState("autoPowerOff", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, m__0B}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, m__0C}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mUTon}: func(m msg) {
		if m[7] < 6 {
			fmt.Println("user tone control,", name("userToneControl", m[7]), int8(m[9]))
		} else {
			notImpl(m, "unknown user tone control msg")
		}
	},
	// 55    AA    00    72    01    11
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, tCurv}: func(m msg) { keepMbState("touchCurve", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, voicg}: func(m msg) { keepMbState("voicing", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, dmpRs}: func(m msg) { keepMbState("damperResonance", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, dmpNs}: func(m msg) { keepMbState("damperNoise", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, strRs}: func(m msg) { keepMbState("stringResonance", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uStRs}: func(m msg) { keepMbState("undampedStringResonance", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, cabRs}: func(m msg) { keepMbState("cabinetResonance", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, koEff}: func(m msg) { keepMbState("keyOffResonance", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, fBkNs}: func(m msg) { keepMbState("fallBackNoise", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, hmDly}: func(m msg) { keepMbState("hammerDelay", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, topBd}: func(m msg) { keepMbState("topboard", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, dcayT}: func(m msg) { keepMbState("decayTime", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, miTch}: func(m msg) { keepMbState("minimumTouch", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, streT}: func(m msg) { keepMbState("stretchTuning", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, tmpmt}: func(m msg) { keepMbState("temperament", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, vt_0F}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, keyVo}: func(m msg) { keepMbState("keyVolume", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, hfPdl}: func(m msg) { keepMbState("halfPedalAdjust", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, sfPdl}: func(m msg) { keepMbState("softPedalDepth", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, smart}: func(m msg) { keepMbState("smartModeVt", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uVoic}: func(m msg) {
		fmt.Println("key", m[7], "has user voicing", m[9])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uStrT}: func(m msg) {
		fmt.Println("key", m[7], "has user stretch", m[9])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uTmpm}: func(m msg) {
		fmt.Println("key", m[7], "has user temperament", m[9])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uKeyV}: func(m msg) {
		fmt.Println("key", m[7], "has user key volume", m[9])
	},
	// 55    AA    00    72    01    12
	{hdr0, hdr1, hdr2, mbDRq, 0x01, hPhon, phShs}: func(m msg) { keepMbState("shsMode", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, hPhon, phTyp}: func(m msg) { keepMbState("phonesType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, hPhon, phVol}: func(m msg) { keepMbState("phonesVolume", m[9]) },
	// 55    AA    00    72    01    13
	{hdr0, hdr1, hdr2, mbDRq, 0x01, midiI, miCha}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, midiI, miPgC}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, midiI, miLoc}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, midiI, miTrP}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, midiI, miMul}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, midiI, miMut}: func(m msg) {
		fmt.Println("MIDI channel", m[7], name("mutedness", m[9]))
	},
	// 55    AA    00    72    01    14
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fUsNm}: func(m msg) {
		fmt.Printf("load from USB filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fMvNm}: func(m msg) {
		fmt.Printf("rename: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fRmNm}: func(m msg) {
		fmt.Printf("delete: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fUsEx}: func(m msg) {
		fmt.Printf("load from USB filename extension(%d)=%s\n", m[7], name("fileExt", m[9]))
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fMvEx}: func(m msg) {
		fmt.Printf("rename: USB filename extension(%d)=%s\n", m[7], name("fileExt", m[9]))
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fRmEx}: func(m msg) {
		fmt.Printf("delete: USB filename extension(%d)=%s\n", m[7], name("fileExt", m[9]))
	},
	// 55    AA    00    72    01    16
	{hdr0, hdr1, hdr2, mbDRq, 0x01, bluet, btAud}: func(m msg) { keepMbState("bluetoothAudio", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, bluet, btAuV}: func(m msg) { keepMbState("bluetoothAudioVolume", int(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, bluet, btMid}: func(m msg) { keepMbState("bluetoothMidi", m[9]) },
	// 55    AA    00    72    01    17
	{hdr0, hdr1, hdr2, mbDRq, 0x01, lcdCn, 0x00}:  func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, lcdCn, 0x02}:  func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, smRec, smPEm}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, smRec, smEmp}: func(m msg) {
		fmt.Println("sound mode song", m[7], name("emptiness", m[9]))
	},
	// 55    AA    00    72    01    21
	{hdr0, hdr1, hdr2, mbDRq, 0x01, pmRec, pmEmp}: func(m msg) {
		fmt.Println("pianist mode song", m[7], name("emptiness", m[9]))
	},
	// 55    AA    00    72    01    22
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, 0x13}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, 0x22}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, 0x23}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, 0x30}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auPNm}: func(m msg) {
		fmt.Printf("USB playback filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auPEx}: func(m msg) {
		fmt.Printf("USB playback filename extension(%d)=%s\n", m[7], name("fileExt", m[9]))
	},
	// 55    AA    00    72    01    32
	{hdr0, hdr1, hdr2, mbDRq, 0x01, msg32, 0x02}: func(m msg) { notImpl(m) },
	// 55    AA    00    72    01    70
	{hdr0, hdr1, hdr2, mbDRq, 0x01, hardw, hwKey}: func(m msg) {
		pnoKeys <- m[9]
	},
	// 55    AA    00    72    01    71
	{hdr0, hdr1, hdr2, mbDRq, 0x01, playr, 0x07}: func(m msg) { notImpl(m) },
}
