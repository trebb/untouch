package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"os"
	"time"
)

var buildDate string

type msg [256]byte // Big Enough (TM)

var (
	mbPort      = flag.String("mb", "/dev/ttyUSB0", "the serial interface connected to the mainboard")
	showVersion = flag.Bool("v", false, "print version and exit")
	rawBytes    = make(chan byte, 1000)
	notImplMsgs = make(chan string, 100)
	toMb        = make(chan []byte, 100)
	pnoKeys     = make(chan int8, 100)
)

var seg14 display

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s\n", buildDate)
		os.Exit(0)
	}
	s, err := serial.OpenPort(&serial.Config{Name: *mbPort, Baud: 115200})
	if err != nil {
		log.Fatal(err)
	}
	defer seg14.close()
	go parse()
	go rxd(*s)
	go txd(*s)
	go input()
	keepMbState("mainboardSeen", false)
	for !mbStateItem("mainboardSeen").(bool) {
		hi()
		for i := 0; i < 6; i++ {
			seg14.spn <- spinPattern{runningOutline, []int{7}}
			time.Sleep(50 * time.Millisecond)
		}
	}
	for {
		if _, ok := mbStateItemOk("normalPianoMode"); ok {
			notify("       *", 0, 1500*time.Millisecond)
			break
		}
		if _, ok := mbStateItemOk("toneGeneratorMode"); ok {
			notify("      **", 0, 1500*time.Millisecond)
			break
		}
		if _, ok := mbStateItemOk("serviceMode"); ok {
			notify("    ****", 0, 1500*time.Millisecond)
			break
		}
		for i := 0; i < 6; i++ {
			seg14.spn <- spinPattern{runningPointer, []int{7}}
			time.Sleep(50 * time.Millisecond)
		}
		fmt.Print("x ")
	}
	if mode, ok := mbStateItemOk("serviceMode"); ok {
		service(mode.(byte))
	} else { // normal playing mode
		for {
			if _, ok := mbStateItemOk("toneGeneratorMode"); ok {
				notify("     ***", 0, 1500*time.Millisecond)
				break
			}
		}
		setLocalDefaults()
	}
	for {
		x := <-notImplMsgs
		log.Print("not implemented:", x)
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
	cmdAcHead = []byte{hdr0, hdr1, hdr2, cmdAc} // UI command that expects acknowledgement
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
		case int8:
			m2 = append(m2, byte(x))
			l++
		case uint16:
			m2 = append(m2, uint16Msg(x)...)
			l += 2
		case string:
			m2 = append(m2, []byte(x)...)
			l += len(x)
		default:
			log.Printf("unknown cmd parameter (%T)%v in cmd %X %X %X %v\n", p, p, topic, subtopic, item, params)
		}
	}
	m1 = append(m1, byte(l))
	toMb <- append(m1, m2...)
}

func issueTglCmd(name string, topic byte, subtopic byte, item byte) (newState byte) {
	switch mbStateItem(name).(byte) {
	case 0:
		newState = 0x1
		issueCmd(topic, subtopic, item, newState)
	case 1:
		newState = 0x0
		issueCmd(topic, subtopic, item, newState)
	default:
		log.Printf("unknown %s switch state (%T)%v\n", name, mbStateItem(name), mbStateItem(name))
	}
	return
}

func issueCmdAc(topic byte, subtopic byte, item byte, params ...interface{}) {
	m1 := append(cmdAcHead, 0x1, topic, subtopic, item)
	var m2 []byte
	l := 0
	for _, p := range params {
		switch x := p.(type) {
		case byte:
			m2 = append(m2, x)
			l++
		case int8:
			m2 = append(m2, byte(x))
			l++
		case uint16:
			m2 = append(m2, uint16Msg(x)...)
			l += 2
		case string:
			m2 = append(m2, []byte(x)...)
			l += len(x)
		default:
			log.Printf("unknown cmdAc parameter (%T)%v in cmd %X %X %X %v\n", p, p, topic, subtopic, item, params)
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

func name(tableKey string, i interface{}) string {
	switch i := i.(type) {
	case int:
		if i >= 0 && i < len(names[tableKey]) {
			return names[tableKey][i]
		}
		return fmt.Sprint(i)
	case int8:
		if int(i) >= 0 && int(i) < len(names[tableKey]) {
			return names[tableKey][i]
		}
		return fmt.Sprint(i)
	case byte:
		if int(i) >= 0 && int(i) < len(names[tableKey]) {
			return names[tableKey][i]
		}
		return fmt.Sprint(i)
	case string:
		return i
	default:
		return "FALLTHROUGH ERROR"
	}
}

var notifyC = make(chan notification)
var notifyLockC = make(chan string)
var notifyUnlockC = make(chan notification)

type notification struct {
	s          string
	precedence int
	expiry     time.Time
}

func notify(s string, precedence int, ttl time.Duration) {
	notifyC <- notification{s, precedence, time.Now().Add(ttl)}
}

func notifyLock(s string) {
	notifyLockC <- s
}

func notifyUnlock(s string, precedence int, ttl time.Duration) {
	notifyUnlockC <- notification{s, precedence, time.Now().Add(ttl)}
}

func notifyCMonitor() {
	var nStack []notification
	var nActive notification
	locked := false

	for {
		select {
		case s := <-notifyLockC:
			locked = true
			seg14.w <- s
		case n := <-notifyUnlockC:
			locked = false
			nStack = append(nStack, n) // push
		case n := <-notifyC:
			if !locked {
				nStack = append(nStack, n) // push
			}
		default:
			if !locked {
				if len(nStack) >= 1 {
					nTop := nStack[len(nStack)-1]
					if nTop.expiry.Before(time.Now()) {
						nStack = nStack[:len(nStack)-1] // pop & discard nTop
					} else if nActive.precedence <= nTop.precedence {
						nActive.precedence = nActive.precedence - 1
						nStack[len(nStack)-1] = nActive // pop nTop, push nActive
						nActive = nTop
						seg14.w <- nActive.s
					}
				}
				if nActive.expiry.Before(time.Now()) {
					seg14.w <- ""
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func init() {
	go notifyCMonitor()
}

type mbStateUpdateItem struct {
	key string
	val interface{}
}

type mbStateQuery struct {
	key    string
	result chan mbStateQueryResult
}

type mbStateQueryResult struct {
	val interface{}
	ok  bool
}

var (
	mbStateUpdates = make(chan mbStateUpdateItem)
	mbStateQueries = make(chan mbStateQuery)
)

func mbStateMonitor() {
	var mbState = make(map[string]interface{})
	for {
		select {
		case in := <-mbStateUpdates:
			mbState[in.key] = in.val
		case out := <-mbStateQueries:
			val, ok := mbState[out.key]
			out.result <- mbStateQueryResult{val, ok}
		}
	}
}

func init() {
	go mbStateMonitor()
}

func mbStateItemOk(key string) (x interface{}, ok bool) {
	var r = make(chan mbStateQueryResult)
	mbStateQueries <- mbStateQuery{key, r}
	result := <-r
	return result.val, result.ok
}

func mbStateItem(key string) interface{} {
	s, ok := mbStateItemOk(key)
	if ok {
		return s
	} else {
		log.Print("missing item ", key, " in mbState")
		return byte(0)
	}
}

func keepMbState(key string, payload interface{}) {
	switch x := payload.(type) {
	case byte:
		mbStateUpdates <- mbStateUpdateItem{key, x}
		fmt.Println("NOTICED:", key, name(key, x))
	case int8:
		mbStateUpdates <- mbStateUpdateItem{key, x}
		fmt.Println("NOTICED:", key, name(key, x))
	case int16:
		mbStateUpdates <- mbStateUpdateItem{key, x}
		fmt.Println("NOTICED:", key, "=", x)
	case string:
		mbStateUpdates <- mbStateUpdateItem{key, x}
		fmt.Println("NOTICED:", key, "=", x)
	case bool:
		mbStateUpdates <- mbStateUpdateItem{key, x}
		fmt.Println("NOTICED:", key, "=", x)
	default:
		log.Printf("can't keep (%T)%v as an mbStateItem for %s\n", payload, payload, key)
	}
}

func pianistSongsItem(i int) pianistSong {
	c := make(chan pianistSong)
	pianistSongQueries <- pianistSongQuery{i, c}
	s := <-c
	return s
}

func keepPianistSongsData(n int, d bool) {
	pianistSongData <- songsBool{n, d}
}

func keepPianistSongsSeen(n int, d bool) {
	pianistSongSeen <- songsBool{n, d}
}

type (
	songsBool struct {
		n int
		d bool
	}
	pianistSongQuery struct {
		n      int
		result chan pianistSong
	}
	pianistSong struct {
		seen bool
		data bool
	}
	soundSongQuery struct {
		n      int
		result chan soundSong
	}
	soundSong struct {
		seen  bool
		data  bool
		part1 bool
		part2 bool
	}
)

var (
	pianistSongSeen    = make(chan songsBool)
	pianistSongData    = make(chan songsBool)
	pianistSongQueries = make(chan pianistSongQuery)
	soundSongSeen      = make(chan songsBool)
	soundSongData      = make(chan songsBool)
	soundSongPart1     = make(chan songsBool)
	soundSongPart2     = make(chan songsBool)
	soundSongQueries   = make(chan soundSongQuery)
)

func soundSongsItem(i int) soundSong {
	c := make(chan soundSong)
	soundSongQueries <- soundSongQuery{i, c}
	s := <-c
	return s
}

func keepSoundSongsData(n int, d bool) {
	soundSongData <- songsBool{n, d}
}

func keepSoundSongsSeen(n int, d bool) {
	soundSongSeen <- songsBool{n, d}
}

func keepSoundSongsPart1(n int, d bool) {
	soundSongPart1 <- songsBool{n, d}
}

func keepSoundSongsPart2(n int, d bool) {
	soundSongPart2 <- songsBool{n, d}
}

func pianistSongsMonitor() {
	var songs [10]pianistSong
	for {
		select {
		case s := <-pianistSongSeen:
			songs[s.n].seen = s.d
		case s := <-pianistSongData:
			songs[s.n].data = s.d
		case s := <-pianistSongQueries:
			s.result <- songs[s.n]
		}
	}
}

func soundSongsMonitor() {
	var songs [10]soundSong
	for {
		select {
		case s := <-soundSongSeen:
			songs[s.n].seen = s.d
		case s := <-soundSongData:
			songs[s.n].data = s.d
		case s := <-soundSongPart1:
			songs[s.n].part1 = s.d
		case s := <-soundSongPart2:
			songs[s.n].part2 = s.d
		case s := <-soundSongQueries:
			s.result <- songs[s.n]
		}
	}
}

var (
	userKeySetting      = make(chan int)
	userKeySettingSeen  = make(chan bool)
	storeUserKeySetting = make(chan int)
	clearUserKeySetting = make(chan struct{})
)

func userKeySettingMonitor() {
	var n int
	var seen bool
	for {
		select {
		case <-clearUserKeySetting:
			seen = false
		case s := <-storeUserKeySetting:
			n = s
			seen = true
		case userKeySetting <- n:
		case userKeySettingSeen <- seen:
		}
	}
}

var (
	storePlayerMsg = make(chan string)
	getPlayerMsg   = make(chan string)
)

func playerMsgMonitor() {
	var msg string
	for {
		select {
		case s := <-storePlayerMsg:
			msg = s
		case getPlayerMsg <- msg:
		}
	}
}

var (
	storeCurrentRecorderState = make(chan int)
	getCurrentRecorderState   = make(chan int)
)

func currentRecorderStateMonitor() {
	var state int
	for {
		select {
		case n := <-storeCurrentRecorderState:
			state = n
		case getCurrentRecorderState <- state:
		}
	}
}

var (
	storeConfirmedUsbSong = make(chan string)
	getConfirmedUsbSong   = make(chan string)
)

func confirmedUsbSongMonitor() {
	var song string
	for {
		select {
		case s := <-storeConfirmedUsbSong:
			song = s
			fmt.Println("stored", s)
		case getConfirmedUsbSong <- song:
		}
	}
}

func init() {
	go pianistSongsMonitor()
	go soundSongsMonitor()
	go userKeySettingMonitor()
	go playerMsgMonitor()
	go currentRecorderStateMonitor()
	go confirmedUsbSongMonitor()
}

var (
	tapTempoTap = make(chan struct{})
	tapDuration = make(chan time.Duration)
)

func tapTempo() {
	tapTempoTap <- struct{}{}
}

func tapTimeMonitor() {
	var lastTime = time.Now()
	for {
		<-tapTempoTap
		tapDuration <- time.Since(lastTime)
		lastTime = time.Now()
	}
}

func tapDurationMonitor() {
	var lastDuration time.Duration
	for {
		d := <-tapDuration
		if d-lastDuration > -200*time.Millisecond && d-lastDuration < 200*time.Millisecond {
			tempo := 60 * time.Second / d
			notify(fmt.Sprint(tempo, "/min"), 0, 1500*time.Millisecond)
			issueCmd(metro, mTmpo, 0x0, uint16(tempo))
			issueCmd(tgMod, tgMod, 0x0, (mbStateItem("toneGeneratorMode")))
		}
		lastDuration = d
	}
}

func init() {
	go tapDurationMonitor()
	go tapTimeMonitor()
}

var metronomeBeatTotal int

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
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmRes}: func(m msg) { keepMbState("resonanceDepth", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmAmb}: func(m msg) { keepMbState("ambienceType", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmSet, pmAmD}: func(m msg) { keepMbState("ambienceDepth", m[9]) },
	// 55    AA    00    6E    01    08
	{hdr0, hdr1, hdr2, mbMsg, 0x01, revrb, rOnOf}: func(m msg) { keepMbState("reverbOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, revrb, rDpth}: func(m msg) { keepMbState("reverbDepth", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, revrb, rTime}: func(m msg) { keepMbState("reverbTime", m[9]) },
	// 55    AA    00    6E    01    09
	{hdr0, hdr1, hdr2, mbMsg, 0x01, effct, eOnOf}: func(m msg) { keepMbState("effectsOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, effct, ePar1}: func(m msg) { keepMbState("effectsParam1", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, effct, ePar2}: func(m msg) { keepMbState("effectsParam2", m[9]) },
	// 55    AA    00    6E    01    0A
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mOnOf}: func(m msg) { keepMbState("metronomeOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mTmpo}: func(m msg) { keepMbState("metronomeTempo", msgInt16(m[9:11])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mSign}: func(m msg) { keepMbState("rhythmPattern", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mVolu}: func(m msg) { keepMbState("metronomeVolume", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, metro, mBeat}: func(m msg) {
		metronomeBeatTotal += 1
		p := 8
		if m[9] < 8 {
			p = int(m[9]) + 1
		}
		if mbStateItem("rhythmPattern") == 0 { // 1/1
			notify(fmt.Sprintf("%*d", p+metronomeBeatTotal%2, m[9]+1), 0, 1500*time.Millisecond)
		} else {
			notify(fmt.Sprintf("%*d", p, m[9]+1), 0, 1500*time.Millisecond)
		}
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
			default:
				notImpl(m, "unknown registration screen state")
			}
		default:
			notImpl(m, "unknown registration screen stuff")
		}
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, regst, rgLoa}: func(m msg) { keepMbState("currentRegistration", m[9]) },
	// 55    AA    00    6E    01    10
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mainF, mTran}: func(m msg) { keepMbState("transpose", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mainF, m__0B}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    14
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fPgrs}: func(m msg) { notify(fmt.Sprintf("FMT %3d", m[9]), 0, 1500*time.Millisecond) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, files, fUsbE}: func(m msg) { notify(errors["usbError"], 0, 1500*time.Millisecond) },
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
		switch m[7] {
		case 0:
			keepMbState("soundSongPart1", m[9])
			keepMbState("soundSongPart1Seen", true)
		case 1:
			keepMbState("soundSongPart2", m[9])
			keepMbState("soundSongPart2Seen", true)
		}
	},
	// 55    AA    00    6E    01    21
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pmRec, pmSel}: func(m msg) {
		fmt.Println("Pianist mode song", m[9])
	},
	// 55    AA    00    6E    01    22
	{hdr0, hdr1, hdr2, mbMsg, 0x01, auRec, 0x30}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, auRec, auPNm}: func(m msg) {
		fmt.Printf("USB playback filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, auRec, auPEx}: func(m msg) {
		fmt.Printf("USB playback filename extension(%d)=%s\n", m[7], name("fileExt", int(m[9])))
	},
	// 55    AA    00    6E    01    32
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg32, 0x00}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg32, 0x04}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, msg32, 0x05}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    3F
	{hdr0, hdr1, hdr2, mbMsg, 0x01, biSng, 0x40}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    60
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srPdV}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srTgA}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srUBt}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srTCk}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srKRw}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srWCk}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srAlK}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srKAd}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, servi, srTcS}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    61
	{hdr0, hdr1, hdr2, mbMsg, 0x01, romId, roNam}: func(m msg) { keepMbState("romName", string(m[9:9+m[8]])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, romId, roVer}: func(m msg) { keepMbState("romVersion", string(m[9:9+m[8]])) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, romId, roCkS}: func(m msg) {
		keepMbState("romChecksum", fmt.Sprintf("%X%X", m[9], m[10]))
	},
	// 55    AA    00    6E    01    63
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mbUpd, muUOk}: func(m msg) {
		notify(serviceNames["updateOk"], 0, 5*time.Hour)
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mbUpd, muNam}: func(m msg) { notify(string(m[9:9+m[8]]), 1, 3*time.Second) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mbUpd, muCnt}: func(m msg) {
		notify(fmt.Sprintf("%X", m[9:12]), -1, 5*time.Second)
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mbUpd, muDne}: func(m msg) {
		notify(serviceNames["updateDone"], 0, 5*time.Second)
	},
	// 55    AA    00    6E    01    64
	{hdr0, hdr1, hdr2, mbMsg, 0x01, uiUpd, upErr}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    65
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mrket, mkMdl}: func(m msg) { keepMbState("pianoModel", m[9]) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, mrket, mkDst}: func(m msg) { keepMbState("marketDestination", m[9]) },
	// 55    AA    00    6E    01    70
	{hdr0, hdr1, hdr2, mbMsg, 0x01, hardw, hwUsb}: func(m msg) {
		notify(name("usbThumbDrivePresence", int(m[9])), 0, 1500*time.Millisecond)
		keepMbState("usbThumbDrivePresence", m[9])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, hardw, hwHPh}: func(m msg) {
		notify(name("phonesPresence", int(m[9])), 0, 1500*time.Millisecond)
		keepMbState("phonesPresence", m[9])
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, hardw, hw_03}: func(m msg) { notImpl(m) },
	// 55    AA    00    6E    01    71
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plDur}: func(m msg) { // really duration?
		fmt.Println("duration1", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, pl_01}: func(m msg) {
		playerMsg := <-getPlayerMsg
		notify(fmt.Sprintf("%-4s%4d", playerMsg, msgInt16(m[9:11])), 0, 1500*time.Millisecond)
		fmt.Println("bar/second count", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, pl_02}: func(m msg) { // really duration?
		fmt.Println("duration2", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plBrC}: func(m msg) {
		currentRecorderState := <-getCurrentRecorderState
		if currentRecorderState == recording || currentRecorderState == playing {
			playerMsg := <-getPlayerMsg
			notify(fmt.Sprintf("%-4s%4d", playerMsg, msgInt16(m[9:11])), 0, 1500*time.Millisecond)
		}
		fmt.Println("bar count", msgInt16(m[9:11]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plBea}: func(m msg) {
		fmt.Println("beat", m[9], msgInt16(m[10:12]))
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, pl_08}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, pl_09}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plRec}: func(m msg) {
		noticeRecording()
		fmt.Println("PLAYR:PLREC")
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plSto}: func(m msg) { storeCurrentRecorderState <- idle },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, playr, plA_B}: func(m msg) {
		fmt.Println("A-B repeat mode", m[9])
	},
	// 55    AA    00    6E    01    7E
	{hdr0, hdr1, hdr2, mbMsg, 0x01, pFace, 0x00}: func(m msg) {
		switch m[7] {
		case pFPno:
			keepMbState("normalPianoMode", byte(1))
			fmt.Println("normal piano mode")
		case pFInt:
			fmt.Println("open internal recorder/player")
		case pFUsb:
			fmt.Println("open USB recorder/player")
		case pFDem:
			fmt.Println("open demo song player")
		case pFLes:
			fmt.Println("open lesson song player")
		case pFCon:
			fmt.Println("open concert magic player")
		case pFPMu:
			fmt.Println("open piano music player")
		default:
			notImpl(m, "unknown recorder/player face request")
		}
	},
	// 55    AA    00    6E    01    7F
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, coSvc}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, coVer}: func(m msg) {
		keepMbState("serviceMode", coVer)
		fmt.Println("Version screen")
	},
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, coMUd}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbMsg, 0x01, commu, coUUd}: func(m msg) { notImpl(m) },
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
		storeConfirmedUsbSong <- string(m[9:])
		fmt.Printf("audio recorder filename=%s\n", m[9:])
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
		keepMbState("mainboardSeen", true)
		requestInitialMbData()
	},
	// 55    AA    00    72    01
	// 55    AA    00    72    01    04
	{hdr0, hdr1, hdr2, mbDRq, 0x01, pmSet, pmAmb}: func(m msg) { keepMbState("ambienceType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, pmSet, pmAmD}: func(m msg) { keepMbState("ambienceDepth", m[9]) },
	// 55    AA    00    72    01    05
	{hdr0, hdr1, hdr2, mbDRq, 0x01, dlSet, dlBal}: func(m msg) { keepMbState("dualBalance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, dlSet, dlOcS}: func(m msg) { keepMbState("dualLayerOctaveShift", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, dlSet, dlDyn}: func(m msg) { keepMbState("dualDynamics", m[9]) },
	// 55    AA    00    72    01    06
	{hdr0, hdr1, hdr2, mbDRq, 0x01, spSet, spBal}: func(m msg) { keepMbState("splitBalance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, spSet, spOcS}: func(m msg) { keepMbState("splitLowerOctaveShift", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, spSet, spPed}: func(m msg) { keepMbState("splitLowerPedal", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, spSet, spSpP}: func(m msg) { keepMbState("splitSplitPoint", m[9]) },
	// 55    AA    00    72    01    07
	{hdr0, hdr1, hdr2, mbDRq, 0x01, h4Set, h4Bal}: func(m msg) { keepMbState("4handsBalance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, h4Set, h4LOS}: func(m msg) { keepMbState("4handsLeftOctaveShift", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, h4Set, h4ROS}: func(m msg) { keepMbState("4handsRightOctaveShift", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, h4Set, h4SpP}: func(m msg) { keepMbState("4handsSplitPoint", m[9]) },
	// 55    AA    00    72    01    08
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rOnOf}: func(m msg) { keepMbState("reverbOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rType}: func(m msg) { keepMbState("reverbType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rDpth}: func(m msg) { keepMbState("reverbDepth", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, revrb, rTime}: func(m msg) { keepMbState("reverbTime", m[9]) },
	// 55    AA    00    72    01    09
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, eOnOf}: func(m msg) { keepMbState("effectsOnOff", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, eType}: func(m msg) { keepMbState("effectsType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, ePar1}: func(m msg) { keepMbState("effectsParam1", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, effct, ePar2}: func(m msg) { keepMbState("effectsParam2", m[9]) },
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
			fmt.Println("registration", reg, "is for", name("toneGeneratorMode", int(m[9])))
		} else {
			fmt.Println("unknown registration (mode)", reg)
		}
	},
	// 55    AA    00    72    01    10
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mTran}: func(m msg) { keepMbState("transpose", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mTone}: func(m msg) { keepMbState("toneControl", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mSpkV}: func(m msg) { keepMbState("speakerVolume", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mLinV}: func(m msg) { keepMbState("lineInLevel", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mWall}: func(m msg) { keepMbState("wallEq", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mTung}: func(m msg) { keepMbState("tuning", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mDpHl}: func(m msg) { keepMbState("damperHold", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mAOff}: func(m msg) { keepMbState("autoPowerOff", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, m__0B}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, m__0C}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, mainF, mUTon}: func(m msg) {
		if m[7] < 6 {
			fmt.Println("user tone control,", name("userToneControl", int(m[7])), (m[9]))
		} else {
			notImpl(m, "unknown user tone control msg")
		}
	},
	// 55    AA    00    72    01    11
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, tCurv}: func(m msg) { keepMbState("touchCurve", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, voicg}: func(m msg) { keepMbState("voicing", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, dmpRs}: func(m msg) { keepMbState("damperResonance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, dmpNs}: func(m msg) { keepMbState("damperNoise", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, strRs}: func(m msg) { keepMbState("stringResonance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uStRs}: func(m msg) { keepMbState("undampedStringResonance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, cabRs}: func(m msg) { keepMbState("cabinetResonance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, koEff}: func(m msg) { keepMbState("keyOffResonance", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, fBkNs}: func(m msg) { keepMbState("fallBackNoise", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, hmDly}: func(m msg) { keepMbState("hammerDelay", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, topBd}: func(m msg) { keepMbState("topboard", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, dcayT}: func(m msg) { keepMbState("decayTime", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, miTch}: func(m msg) { keepMbState("minimumTouch", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, streT}: func(m msg) { keepMbState("stretchTuning", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, tmpmt}: func(m msg) { keepMbState("temperament", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, tmKey}: func(m msg) { keepMbState("temperamentKey", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, keyVo}: func(m msg) { keepMbState("keyVolume", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, hfPdl}: func(m msg) { keepMbState("halfPedalAdjust", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, sfPdl}: func(m msg) { keepMbState("softPedalDepth", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, smart}: func(m msg) { keepMbState("smartModeVt", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uVoic}: func(m msg) {
		storeUserKeySetting <- int(m[9])
		fmt.Println("key", m[7], "has user voicing", m[9])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uStrT}: func(m msg) {
		storeUserKeySetting <- int(m[9])
		fmt.Println("key", m[7], "has user stretch", m[9])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uTmpm}: func(m msg) {
		storeUserKeySetting <- int(m[9])
		fmt.Println("key", m[7], "has user temperament", m[9])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, vTech, uKeyV}: func(m msg) {
		storeUserKeySetting <- int(m[9])
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
		fmt.Println("MIDI channel", m[7], name("mutedness", int(m[9])))
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
		fmt.Printf("load from USB filename extension(%d)=%s\n", m[7], name("fileExt", int(m[9])))
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fMvEx}: func(m msg) {
		fmt.Printf("rename: USB filename extension(%d)=%s\n", m[7], name("fileExt", int(m[9])))
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, files, fRmEx}: func(m msg) {
		fmt.Printf("delete: USB filename extension(%d)=%s\n", m[7], name("fileExt", int(m[9])))
	},
	// 55    AA    00    72    01    16
	{hdr0, hdr1, hdr2, mbDRq, 0x01, bluet, btAud}: func(m msg) { keepMbState("bluetoothAudio", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, bluet, btAuV}: func(m msg) { keepMbState("bluetoothAudioVolume", int8(m[9])) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, bluet, btMid}: func(m msg) { keepMbState("bluetoothMidi", m[9]) },
	// 55    AA    00    72    01    17
	{hdr0, hdr1, hdr2, mbDRq, 0x01, lcdCn, 0x00}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, lcdCn, 0x02}: func(m msg) { notImpl(m) },
	// 55    AA    00    72    01    20
	{hdr0, hdr1, hdr2, mbDRq, 0x01, smRec, smPlP}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, smRec, smPEm}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, smRec, smEmp}: func(m msg) {
		keepSoundSongsData(int(m[7]), m[9] != 0)
		keepSoundSongsSeen(int(m[7]), true)
		fmt.Println("sound mode song", m[7], name("emptiness", int(m[9])))
	},
	// 55    AA    00    72    01    21
	{hdr0, hdr1, hdr2, mbDRq, 0x01, pmRec, pmEmp}: func(m msg) {
		keepPianistSongsData(int(m[7]), m[9] != 0)
		keepPianistSongsSeen(int(m[7]), true)
		fmt.Println("pianist mode song", m[7], name("emptiness", int(m[9])))
	},
	// 55    AA    00    72    01    22
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auTrn}: func(m msg) { keepMbState("usbPlayerTranspose", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auTyp}: func(m msg) { keepMbState("usbPlayerFileType", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auGai}: func(m msg) { keepMbState("usbPlayerGainLevel", m[9]) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, au_30}: func(m msg) { notImpl(m) },
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auPNm}: func(m msg) {
		fmt.Printf("USB playback filename(%d)=%s\n", m[7], m[9:])
	},
	{hdr0, hdr1, hdr2, mbDRq, 0x01, auRec, auPEx}: func(m msg) {
		fmt.Printf("USB playback filename extension(%d)=%s\n", m[7], name("fileExt", int(m[9])))
	},
	// 55    AA    00    72    01    32
	{hdr0, hdr1, hdr2, mbDRq, 0x01, msg32, 0x02}: func(m msg) { notImpl(m) },
	// 55    AA    00    72    01    70
	{hdr0, hdr1, hdr2, mbDRq, 0x01, hardw, hwKey}: func(m msg) {
		pnoKeys <- int8(m[9])
	},
	// 55    AA    00    72    01    71
	{hdr0, hdr1, hdr2, mbDRq, 0x01, playr, 0x07}: func(m msg) { notImpl(m) },
}
