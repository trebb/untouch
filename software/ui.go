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
)

func main() {
	flag.Parse()
	s, err := serial.OpenPort(&serial.Config{Name: *mbPort, Baud: 115200})
	if err != nil {
		log.Fatal(err)
	}
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
			log.Fatal(err)
		}
		rawBytes <- buf[0]
	}
}

func txd(port serial.Port) {
	for {
		message := <-toMb
		_, err := port.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// func txd(port serial.Port) {
// 	for {
// 		message := <-toMb
// 		fmt.Println(message)
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
	cmdVnHead = []byte{0x55, 0xAA, 0x00, 0x60} // vanilla UI command
	cmdAcHead = []byte{0x55, 0xAA, 0x00, 0x61} // UI command that expects acknoledgement
	dtaRqHead = []byte{0x55, 0xAA, 0x00, 0x62} // data request
)

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
			log.Fatal("unknown cmd parameter")
		case int: // byte, actually
			m2 = append(m2, byte(x))
			l++
		}
	}
	m1 = append(m1, byte(l))
	toMb <- append(m1, m2...)
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
			log.Fatal("unknown cmd parameter")
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
	issueCmd(regst, rgLoa, 0x0, 0xF)
	issueDtaRq(
		request{regst, rgMod, 0xF, 0x1, 0x0})
	issueDtaRq(
		request{regst, rgMod, 0xF, 0x1, 0x0},
		request{regst, rgNam, 0xF, 0x0})
}


func msgUint16(b []byte) uint16 {
	var r uint16
	buf := bytes.NewReader(b[:2])
	err := binary.Read(buf, binary.BigEndian, &r)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func uint16Msg(i uint16) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()[:2]
}

func notImpl(m interface{}, what ...string) {
	var line string
	for _, w := range what {
		line += w
	}
	if line == "" {
		line = "Unexpected"
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

func item(a []string, i uint8) string {
	if i >= 0 && i < uint8(len(a)) {
		return a[i]
	}
	return fmt.Sprint("undef:", i)
}

var actions = map[msg]func(msg){
	// key contains the bytes 0..6 of the raw message in an otherwise pristine msg

	// 55    AA    00    6E    01
	// 55    AA    00    6E    01    00
	{0x55, 0xAA, 0x00, 0x6E, 0x01, tgMod, 0x00}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("sound mode")
			case 1:
				fmt.Println("pianist mode")
			default:
				notImpl(m, "unknown mode")
			}
		default:
			notImpl(m, "unknown mode stuff")
		}
	},
	// 55    AA    00    6E    01    01
	{0x55, 0xAA, 0x00, 0x6E, 0x01, kbSpl, 0x00}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("keyboard mode", m[9])
		default:
			notImpl(m, "unknown keyboard mode msg")
		}
	},
	// 55    AA    00    6E    01    02
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, iSing}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("single sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown single sound msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, iDua1}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("first dual sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown first dual sound msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, iDua2}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("second dual sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown second dual sound msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, iSpl1}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("first split sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown first split sound msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, iSpl2}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("second split sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown second split sound msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, i4Hd1}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("first 4hands sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown first 4hands sound msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, instr, i4Hd2}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("second 4hands sound", item(instrumentSound[:], m[9]))
		default:
			notImpl(m, "unknown second 4hands sound msg")
		}
	},
	// 55    AA    00    6E    01    04
	{0x55, 0xAA, 0x00, 0x6E, 0x01, pmSet, pmRen}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("rendering character", item(renderingCharacter[:], m[9]))
		default:
			notImpl(m, "unknown rendering character msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, pmSet, pmRes}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("resonance depth", m[9])
		default:
			notImpl(m, "unknown resonance depth msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, pmSet, pmAmb}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("ambience type", item(ambienceType[:], m[9]))
		default:
			notImpl(m, "unknown ambience type msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, pmSet, pmAmD}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("ambience depth", m[9])
		default:
			notImpl(m, "unknown ambience depth msg")
		}
	},
	// 55    AA    00    6E    01    08
	{0x55, 0xAA, 0x00, 0x6E, 0x01, revrb, rOnOf}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("reverb off")
			case 1:
				fmt.Println("reverb on")
			default:
				notImpl(m, "unknown reverb state")
			}
		default:
			notImpl(m, "unknown reverb stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, revrb, rDpth}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, revrb, rTime}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    09
	{0x55, 0xAA, 0x00, 0x6E, 0x01, effct, eOnOf}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("effects off")
			case 1:
				fmt.Println("effects on")
			default:
				notImpl(m, "unknown effects state")
			}
		default:
			notImpl(m, "unknown effects stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, effct, ePar1}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, effct, ePar2}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    0A
	{0x55, 0xAA, 0x00, 0x6E, 0x01, metro, mOnOf}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("metronome off")
			case 1:
				fmt.Println("metronome on")
			default:
				notImpl(m, "unknown metronome state")
			}
		default:
			notImpl(m, "unknown metronome stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, metro, mTmpo}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("metronome tempo", msgUint16(m[9:11]))
		default:
			notImpl(m, "unknown metronome tempo msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, metro, mSign}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("metronome time signature", item(rhythmPattern[:], m[9]))
		default:
			notImpl(m, "unknown metronome time signature msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, metro, mVolu}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("metronome volume", m[9])
		default:
			notImpl(m, "unknown metronome volume msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, metro, mBeat}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("metronome beat", m[9])
		default:
			notImpl(m, "unknown metronome beat msg")
		}
	},
	// 55    AA    00    6E    01    0F
	{0x55, 0xAA, 0x00, 0x6E, 0x01, regst, rgOpn}: func(m msg) {
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
	{0x55, 0xAA, 0x00, 0x6E, 0x01, regst, rgLoa}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("current favourite is", int8(m[9]))
		default:
			notImpl(m, "unknown current favourite msg")
		}
	},
	// 55    AA    00    6E    01    10
	{0x55, 0xAA, 0x00, 0x6E, 0x01, mainF, mTran}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("transpose", int8(m[9]))
		default:
			notImpl(m, "unknown transpose msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, mainF, m__0B}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    14
	{0x55, 0xAA, 0x00, 0x6E, 0x01, files, fPgrs}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, files, fMvNm}: func(m msg) {
		fmt.Printf("rename done: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, files, fRmNm}: func(m msg) {
		fmt.Printf("delete: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, files, fMvEx}: func(m msg) {
		fmt.Printf("rename: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, files, fRmEx}: func(m msg) {
		fmt.Printf("delete: USB filename extension(%d)=%d\n", m[7], m[9])
	},
	// 55    AA    00    6E    01    16
	{0x55, 0xAA, 0x00, 0x6E, 0x01, bluet, btAud}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("no BT audio")
			case 1:
				fmt.Println("BT audio")
			default:
				notImpl(m, "unknown BT audio state")
			}
		default:
			notImpl(m, "unknown bt audio stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, bluet, btMid}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("no BT midi")
			case 1:
				fmt.Println("BT midi")
			default:
				notImpl(m, "unknown BT midi state")
			}
		default:
			notImpl(m, "unknown bt midi stuff")
		}
	},
	// 55    AA    00    6E    01    20
	{0x55, 0xAA, 0x00, 0x6E, 0x01, smRec, 0x00}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, smRec, 0x01}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, smRec, 0x02}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, smRec, 0x60}: func(m msg) {
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
	{0x55, 0xAA, 0x00, 0x6E, 0x01, pmRec, 0x00}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    22
	{0x55, 0xAA, 0x00, 0x6E, 0x01, auRec, 0x30}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, auRec, 0x40}: func(m msg) {
		fmt.Printf("USB playback filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, auRec, 0x41}: func(m msg) {
		fmt.Printf("USB playback filename extension(%d)=%s\n", m[7], item(fileExt[:], m[9]))
	},
	// 55    AA    00    6E    01    32
	{0x55, 0xAA, 0x00, 0x6E, 0x01, msg32, 0x00}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, msg32, 0x04}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, msg32, 0x05}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    3F
	{0x55, 0xAA, 0x00, 0x6E, 0x01, biSng, 0x40}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    60
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x41}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x43}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x46}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x48}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x49}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x4A}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, servi, 0x4D}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    61
	{0x55, 0xAA, 0x00, 0x6E, 0x01, romId, 0x00}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, romId, 0x01}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, romId, 0x02}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    65
	{0x55, 0xAA, 0x00, 0x6E, 0x01, msg65, 0x00}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, msg65, 0x01}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    70
	{0x55, 0xAA, 0x00, 0x6E, 0x01, hardw, hwUsb}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("no USB thumb drive")
			case 1:
				fmt.Println("USB thumb drive present")
			default:
				notImpl(m, "unknown USB state")
			}
		default:
			notImpl(m, "unknown USB stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, hardw, hwHPh}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("headphones unlinked")
			case 1:
				fmt.Println("headphones linked")
			default:
				notImpl(m, "unknown phones state")
			}
		default:
			notImpl(m, "unknown phones stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, hardw, hw_03}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    71
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x00}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("duration", msgUint16(m[9:11]))
		default:
			notImpl(m, "unknown duration count stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x01}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("bar/second count", msgUint16(m[9:11]))
		default:
			notImpl(m, "unknown bar/second count stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x02}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("duration", msgUint16(m[9:11]))
		default:
			notImpl(m, "unknown duration count stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x03}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("bar count", msgUint16(m[9:11]))
		default:
			notImpl(m, "unknown bar count stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x04}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("beat", m[9], msgUint16(m[10:12]))
		default:
			notImpl(m, "unknown beat stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x08}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x09}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x11}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x12}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0x0:
				fmt.Println("stopped")
			default:
				notImpl(m, "unknown stopped mode")
			}
		default:
			notImpl(m, "unknown stopped stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x6E, 0x01, playr, 0x13}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("A-B repeat mode", m[9])
		default:
			notImpl(m, "unknown A-B repeat mode stuff")
		}
	},
	// 55    AA    00    6E    01    7E
	{0x55, 0xAA, 0x00, 0x6E, 0x01, rpFce, 0x00}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("close recorder/player")
		case 0x2:
			fmt.Println("open internal recorder/player")
		case 0x3:
			fmt.Println("open USB recorder/player")
		case 0x5:
			fmt.Println("open demo song player")
		case 0x7:
			fmt.Println("open lesson song player")
		case 0x8:
			fmt.Println("open concert magic player")
		case 0x9:
			fmt.Println("open piano music player")
		default:
			notImpl(m, "unknown recorder/player face request")
		}
	},
	// 55    AA    00    6E    01    7F
	{0x55, 0xAA, 0x00, 0x6E, 0x01, commu, 0x00}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, commu, 0x01}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x6E, 0x01, commu, 0x04}: func(m msg) { notImpl(m) },
	// 55    AA    00    71    01
	// 55    AA    00    71    01    10
	{0x55, 0xAA, 0x00, 0x71, 0x01, mainF, mFact}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("Ok, factory reset")
		default:
			notImpl(m, "unknown reset msg")
		}
	},
	// 55    AA    00    71    01    14
	{0x55, 0xAA, 0x00, 0x71, 0x01, files, 0x60}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("done: load from USB")
		default:
			notImpl(m, "unknown load from USB stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x71, 0x01, files, 0x64}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Printf("save .KSO to USB: confirming filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown save .KSO filename stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x71, 0x01, files, 0x65}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("save .MID to USB: confirming filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown save .MID file stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x71, 0x01, files, fName}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("rename: confirming new USB filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown rename USB file stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x71, 0x01, files, fRmCf}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("delete: confirming USB file deletion\n")
		default:
			notImpl(m, "unknown delete USB file stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x71, 0x01, files, fFmat}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Println("USB format done")
		default:
			notImpl(m, "unknown USB format stuff")
		}
	},
	// 55    AA    00    71    01    20
	{0x55, 0xAA, 0x00, 0x71, 0x01, smRec, 0x40}: func(m msg) { notImpl(m) },
	// 55    AA    00    71    01    21
	{0x55, 0xAA, 0x00, 0x71, 0x01, pmRec, 0x40}: func(m msg) { notImpl(m) },
	// 55    AA    00    71    01    22
	{0x55, 0xAA, 0x00, 0x71, 0x01, auRec, 0x50}: func(m msg) {
		switch m[7] {
		case 0xFF:
			fmt.Printf("audio recorder filename=%s\n", m[9:])
		default:
			notImpl(m, "unknown audio recorder filename stuff")
		}
	},
	// 55    AA    00    71    01    71
	{0x55, 0xAA, 0x00, 0x71, 0x01, playr, 0x10}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("duration", msgUint16(m[9:11]))
		default:
			notImpl(m, "unknown duration count stuff")
		}
	},
	// 55    AA    00    71    01    7F
	{0x55, 0xAA, 0x00, 0x71, 0x01, commu, 0x7F}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("Hi, I'm a mainboard")
		default:
			notImpl(m, "unknown greeting")
		}
	},
	// 55    AA    00    72    01
	// 55    AA    00    72    01    04
	{0x55, 0xAA, 0x00, 0x72, 0x01, pmSet, pmAmb}: func(m msg) {
		switch m[7] {
		case 0x0:
			if t := m[9]; t < 0xA {
				fmt.Println("ambience type =", ambienceType[t])
			} else {
				fmt.Println("unknown ambience type", t)
			}
		default:
			notImpl(m, fmt.Sprint("unknown ambience type stuff"))
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, pmSet, pmAmD}: func(m msg) {
		switch m[7] {
		case 0x0:
			if d := m[9]; d < 0xA {
				fmt.Println("ambience depth =", d)
			} else {
				fmt.Println("unknown ambience depth", d)
			}
		default:
			notImpl(m, fmt.Sprint("unknown ambience depth stuff"))
		}
	},
	// 55    AA    00    72    01    08
	{0x55, 0xAA, 0x00, 0x72, 0x01, revrb, rOnOf}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("reverb is off")
			case 1:
				fmt.Println("reverb is on")
			default:
				fmt.Println("unknown reverb state")
			}
		default:
			notImpl(m, "unknown reverb stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, revrb, rType}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("reverb type", m[9])
		default:
			notImpl(m, "unknown reverb type msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, revrb, rDpth}: func(m msg) {
		switch m[7] {
		case 0x0:
			if d := m[9]; d >= 1 && d < 0xB {
				fmt.Println("reverb depth =", d)
			} else {
				fmt.Println("unknown reverb depth", d)
			}
		default:
			notImpl(m, fmt.Sprint("unknown reverb depth stuff"))
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, revrb, rTime}: func(m msg) {
		switch m[7] {
		case 0x0:
			if d := m[9]; d >= 1 && d < 0xB {
				fmt.Println("reverb time =", d)
			} else {
				fmt.Println("unknown reverb time", d)
			}
		default:
			notImpl(m, fmt.Sprint("unknown reverb time stuff"))
		}
	},
	// 55    AA    00    72    01    09
	{0x55, 0xAA, 0x00, 0x72, 0x01, effct, eOnOf}: func(m msg) {
		switch m[7] {
		case 0x0:
			switch m[9] {
			case 0:
				fmt.Println("effects are off")
			case 1:
				fmt.Println("effects are on")
			default:
				fmt.Println("unknown effects state")
			}
		default:
			notImpl(m, "unknown effects stuff")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, effct, eType}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("effects type", item(effectsType[:], m[9]))
		default:
			notImpl(m, "unknown effects type msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, effct, ePar1}: func(m msg) {
		switch m[7] {
		case 0x0:
			if d := m[9]; d >= 1 && d < 0xB {
				fmt.Println("effects parameter1 =", d)
			} else {
				fmt.Println("unknown effects parameter1", d)
			}
		default:
			notImpl(m, fmt.Sprint("unknown effects parameter1 stuff"))
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, effct, ePar2}: func(m msg) {
		switch m[7] {
		case 0x0:
			if d := m[9]; d >= 1 && d < 0xB {
				fmt.Println("effects parameter2 =", d)
			} else {
				fmt.Println("unknown effects parameter2", d)
			}
		default:
			notImpl(m, fmt.Sprint("unknown effects parameter2 stuff"))
		}
	},
	// 55    AA    00    72    01    0F
	{0x55, 0xAA, 0x00, 0x72, 0x01, regst, rgNam}: func(m msg) {
		if reg := m[7]; reg < 0xF {
			fmt.Printf("name of registration %d = %s\n", reg, m[9:])
		} else {
			fmt.Println("unknown registration", reg)
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, regst, rgMod}: func(m msg) {
		if reg := m[7]; reg < 0xF {
			switch m[9] {
			case 0:
				fmt.Println("registration", reg, "is for sound mode")
			case 1:
				fmt.Println("registration", reg, "is for pianist mode")
			default:
				notImpl(m, fmt.Sprint("unknown mode for registration", reg))
			}
		} else {
			fmt.Println("unknown registration", reg)
		}
	},
	// 55    AA    00    72    01    10
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mTran}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("transpose", int8(m[9]))
		default:
			notImpl(m, "unknown transpose msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mTone}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("tone control", item(toneControl[:], m[9]))
		default:
			notImpl(m, "unknown tone control msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mSpkV}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("speaker volume", item(speakerVolume[:], m[9]))
		default:
			notImpl(m, "unknown speaker volume msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mLinV}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("line-in level", int8(m[9]))
		default:
			notImpl(m, "unknown line-in level msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mWall}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("wall EQ", m[9])
		default:
			notImpl(m, "unknown wall EQ msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, m__06}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, m__07}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mAOff}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("auto power off", m[9])
		default:
			notImpl(m, "unknown auto power off msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, m__0B}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, m__0C}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, mainF, mUTon}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("user tone control, low dB=", int8(m[9]))
		case 0x1:
			fmt.Println("user tone control, mid-low freqency", m[9])
		case 0x2:
			fmt.Println("user tone control, mid-low dB", int8(m[9]))
		case 0x3:
			fmt.Println("user tone control, mid-high frequency", m[9])
		case 0x4:
			fmt.Println("user tone control, mid-high dB", int8(m[9]))
		case 0x5:
			fmt.Println("user tone control, high frequency", m[9])
		default:
			notImpl(m, "unknown user tone control msg")
		}
	},
	// 55    AA    00    72    01    11
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, tCurv}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("touch curve", item(touchCurve[:], m[9]))
		default:
			notImpl(m, "unknown touch curve msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, voicg}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("voicing", item(voicing[:], m[9]))
		default:
			notImpl(m, "unknown voicing msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, dmpRs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("damper resonance", m[9])
		default:
			notImpl(m, "unknown damper resonance msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, dmpNs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("damper noise", m[9])
		default:
			notImpl(m, "unknown damper noise msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, strRs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("string resonance", m[9])
		default:
			notImpl(m, "unknown string resonance msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, uStRs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("undamped string resonance", m[9])
		default:
			notImpl(m, "unknown undamped string resonance msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, cabRs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("cabinet resonance", m[9])
		default:
			notImpl(m, "unknown cabinet resonance msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, koEff}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("key-off resonance", m[9])
		default:
			notImpl(m, "unknown key-off resonance msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, fBkNs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("fall-back noise", m[9])
		default:
			notImpl(m, "unknown fall-back noise msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, hmDly}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("hammer delay", m[9])
		default:
			notImpl(m, "unknown hammer delay msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, topBd}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("topboard", item(topboard[:], m[9]))
		default:
			notImpl(m, "unknown topboard msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, dcayT}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("decay time", m[9])
		default:
			notImpl(m, "unknown decay time msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, miTch}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("minimum touch", m[9])
		default:
			notImpl(m, "unknown minimum touch msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, streT}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("stretch tuning", item(stretchTuning[:], m[9]))
		default:
			notImpl(m, "unknown stretch tuning msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, tmpmt}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("temperament", item(temperament[:], m[9]))
		default:
			notImpl(m, "unknown temperament msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, vt_0F}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, keyVo}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("key volume", item(keyVolume[:], m[9]))
		default:
			notImpl(m, "unknown key-volume msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, hfPdl}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("half-pedal adjust", m[9])
		default:
			notImpl(m, "unknown half-pedal adjust msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, sfPdl}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("soft-pedal depth", m[9])
		default:
			notImpl(m, "unknown soft-pedal depth msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, smart}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("virtual-technician smart mode", m[9])
		default:
			notImpl(m, "unknown virtual-technician smart mode msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, uVoic}: func(m msg) {
		fmt.Println("key", m[7], "has user voicing", m[9])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, uStrT}: func(m msg) {
		fmt.Println("key", m[7], "has user stretch", m[9])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, uTmpm}: func(m msg) {
		fmt.Println("key", m[7], "has user temperament", m[9])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, vTech, uKeyV}: func(m msg) {
		fmt.Println("key", m[7], "has user key volume", m[9])
	},
	// 55    AA    00    72    01    12
	{0x55, 0xAA, 0x00, 0x72, 0x01, hPhon, phShs}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("SHS mode", item(shsMode[:], m[9]))
		default:
			notImpl(m, "unknown SHS mode msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, hPhon, phTyp}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Println("phones type", item(phonesType[:], m[9]))
		default:
			notImpl(m, "unknown phones type msg")
		}
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, hPhon, phVol}: func(m msg) {
		switch m[7] {
		case 0x0:
			if x := m[9]; int(x) < len(m) {
				fmt.Println("phones volume", phonesVolume[x])
			}
		default:
			notImpl(m, "unknown volume type msg")
		}
	},
	// 55    AA    00    72    01    13
	{0x55, 0xAA, 0x00, 0x72, 0x01, midiI, miCha}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, midiI, miPgC}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, midiI, miLoc}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, midiI, miTrP}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, midiI, miMul}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, midiI, miMut}: func(m msg) {
		if ch := m[7]; ch < 0x10 {
			switch m[9] {
			case 0:
				fmt.Println("MIDI channel", ch, "muted")
			case 1:
				fmt.Println("MIDI channel", ch, "unmuted")
			default:
				notImpl(m, fmt.Sprint("unknown MIDI channel state", ch))
			}

		} else {
			fmt.Println("unknown MIDI channel number", ch)
		}
	},
	// 55    AA    00    72    01    14
	{0x55, 0xAA, 0x00, 0x72, 0x01, files, 0x40}: func(m msg) {
		fmt.Printf("load from USB filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, files, fMvNm}: func(m msg) {
		fmt.Printf("rename: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, files, fRmNm}: func(m msg) {
		fmt.Printf("delete: USB filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, files, 0x50}: func(m msg) {
		fmt.Printf("load from USB filename extension(%d)=%s\n", m[7], item(fileExt[:], m[9]))
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, files, fMvEx}: func(m msg) {
		fmt.Printf("rename: USB filename extension(%d)=%s\n", m[7], item(fileExt[:], m[9]))
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, files, fRmEx}: func(m msg) {
		fmt.Printf("delete: USB filename extension(%d)=%s\n", m[7], item(fileExt[:], m[9]))
	},
	// 55    AA    00    72    01    16
	{0x55, 0xAA, 0x00, 0x72, 0x01, bluet, btAud}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, bluet, btAuV}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, bluet, btMid}: func(m msg) { notImpl(m) },
	// 55    AA    00    72    01    17
	{0x55, 0xAA, 0x00, 0x72, 0x01, lcdCn, 0x00}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, lcdCn, 0x02}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, smRec, 0x60}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, smRec, 0x61}: func(m msg) {
		if song := m[7]; song < 0xA {
			switch m[9] {
			case 0:
				fmt.Println("sound mode song", song, "empty")
			case 1:
				fmt.Println("sound mode song", song, "contains a recording")
			default:
				notImpl(m, fmt.Sprint("unknown sound mode song state", song))
			}
		} else {
			fmt.Println("unknown sound mode song number", song)
		}
	},
	// 55    AA    00    72    01    21
	{0x55, 0xAA, 0x00, 0x72, 0x01, pmRec, 0x61}: func(m msg) {
		if song := m[7]; song < 0x3 {
			switch m[9] {
			case 0:
				fmt.Println("pianist mode song", song, "empty")
			case 1:
				fmt.Println("pianist mode song", song, "contains a recording")
			default:
				notImpl(m, fmt.Sprint("unknown pianist mode song state", song))
			}
		} else {
			fmt.Println("unknown pianist mode song number", song)
		}
	},
	// 55    AA    00    72    01    22
	{0x55, 0xAA, 0x00, 0x72, 0x01, auRec, 0x13}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, auRec, 0x22}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, auRec, 0x23}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, auRec, 0x30}: func(m msg) { notImpl(m) },
	{0x55, 0xAA, 0x00, 0x72, 0x01, auRec, 0x40}: func(m msg) {
		fmt.Printf("USB playback filename(%d)=%s\n", m[7], m[9:])
	},
	{0x55, 0xAA, 0x00, 0x72, 0x01, auRec, 0x41}: func(m msg) {
		fmt.Printf("USB playback filename extension(%d)=%s\n", m[7], item(fileExt[:], m[9]))
	},
	// 55    AA    00    72    01    32
	{0x55, 0xAA, 0x00, 0x72, 0x01, msg32, 0x02}: func(m msg) { notImpl(m) },
	// 55    AA    00    72    01    70
	{0x55, 0xAA, 0x00, 0x72, 0x01, hardw, hwKey}: func(m msg) {
		switch m[7] {
		case 0x0:
			fmt.Printf("key %d pressed\n", m[9])
		default:
			notImpl(m, "unknown keypress stuff")
		}
	},
	// 55    AA    00    72    01    71
	{0x55, 0xAA, 0x00, 0x72, 0x01, playr, 0x07}: func(m msg) { notImpl(m) },
}
// TODO: make actions funcs accept bool
