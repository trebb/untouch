package main

const ( // message headers
	hdr0 byte = 0x55 // first byte of any message
	hdr1 byte = 0xAA // second byte of any message
	hdr2 byte = 0x00 // third byte of any message
	// messages sent to the mainboard
	cmdVn byte = 0x60 // fourth byte of vanilla command
	cmdAc byte = 0x61 // fourth byte of command that expects acknoledgement (sent back in a mbCAc message)
	dtaRq byte = 0x62 // fourth byte of a data request (sent back in a dtaRq message)
	// messages sent by the mainboard
	mbMsg byte = 0x6E // message sent without request
	mbCAc byte = 0x71 // confirmation of a cmdAc command
	mbDRq byte = 0x72 // response to a dtaRq data request
)

const ( // principal topics of mainboard operation
	tgMod byte = 0x00 // tone generator mode (sound/pianist)
	kbSpl byte = 0x01 // keyboard splitting
	instr byte = 0x02 // sound-mode instrument
	pmSet byte = 0x04 // pianist mode settings
	dlSet byte = 0x05 // dual-mode settings (sound mode)
	spSet byte = 0x06 // split-mode settings (sound mode)
	h4Set byte = 0x07 // 4hands-mode settings (sound mode)
	revrb byte = 0x08 // reverb
	effct byte = 0x09 // effects
	metro byte = 0x0A // metronome/drums
	regst byte = 0x0F // registrations aka favourites
	mainF byte = 0x10 // transpose, factory reset
	vTech byte = 0x11 // virtual technician
	hPhon byte = 0x12 // headphones
	midiI byte = 0x13 // midi
	files byte = 0x14 // files on USB storage
	bluet byte = 0x16 // bluetooth
	lcdCn byte = 0x17 // LCD contrast
	smRec byte = 0x20 // recorder, sound mode
	pmRec byte = 0x21 // recorder, pianist mode
	auRec byte = 0x22 // USB audio recorder
	msg32 byte = 0x32 // lesson song volume balance
	biSng byte = 0x3F // built-in songs
	servi byte = 0x60 // service screen
	romId byte = 0x61 // ROM identification
	mbUpd byte = 0x63 // MB update
	uiUpd byte = 0x64 // UI update
	mrket byte = 0x65 // model, market destination
	hardw byte = 0x70 // user action on hardware: connecting devices, pressing keys
	playr byte = 0x71 // player
	pFace byte = 0x7E // recorder/player face
	commu byte = 0x7F // communication setup
)

const ( // (tgMod, 0x00) tone generator modes
	// tgMod = 0x00  happens to be equal to the principal topic
	// submodes
	tgSnd byte = 0x00 // sound mode
	tgPia byte = 0x01 // pianist mode
)

const ( // (kbSpl, 0x01) keyboard splitting
	kbSpM byte = 0x00 // single, dual, split, 4hands
)
const ( // the four keyboard spliting modes
	kbSpMSingle byte = 0x00
	kbSpMDual   byte = 0x01
	kbSpMSplit  byte = 0x02
	kbSpM4hands byte = 0x03
)

const ( // (instr, 0x02) sound-mode instruments
	iSing byte = 0x00 // single instrument
	iDua1 byte = 0x01 // 1st dual instrument
	iDua2 byte = 0x02 // 2nd dual instrument
	iSpl1 byte = 0x03 // 1st split instrument
	iSpl2 byte = 0x04 // 2nd split instrument
	i4Hd1 byte = 0x05 // 1st 4hands instrument
	i4Hd2 byte = 0x06 // 2nd 4hands instrument
)

const ( // (pmSet, 0x04) pianist mode settings
	pmRen byte = 0x00 // rendering character
	pmRes byte = 0x01 // resonance depth
	pmAmb byte = 0x02 // ambience type
	pmAmD byte = 0x03 // ambience depth
)

const ( // (dlSet, 0x05) dual-mode settings (sound mode)
	dlBal byte = 0x00 // balance
	dlOcS byte = 0x01 // layer octave shift
	dlDyn byte = 0x02 // dynamics
)

const ( // (spSet, 0x06) split-mode settings (sound mode)
	spBal byte = 0x00 // balance
	spOcS byte = 0x01 // lower octave shift
	spPed byte = 0x02 // lower pedal
	spSpP byte = 0x03 // split point
)
const ( // (h4Set, 0x07) 4hands-mode settings (sound mode)
	h4Bal byte = 0x00 // balance
	h4LOS byte = 0x01 // left octave shift
	h4ROS byte = 0x02 // right octave shift
	h4SpP byte = 0x03 // split point
)

const ( // (revrb, 0x08) reverb settings
	rOnOf byte = 0x00 // reverb on/off
	rType byte = 0x01 // reverb type
	rDpth byte = 0x02 // reverb depth
	rTime byte = 0x03 // reverb time
)

const ( // (effct, 0x09) effects settings
	eOnOf byte = 0x00 // effects on/off
	eType byte = 0x01 // effects type
	ePar1 byte = 0x02 // effects parameter 1
	ePar2 byte = 0x03 // effects parameter 2
)

const ( // (metro, 0x0A) metronome settings
	mOnOf byte = 0x00 // metronome on/off
	mTmpo byte = 0x01 // tempo
	mSign byte = 0x02 // time signature/rhythm pattern
	mVolu byte = 0x03 // metronome volume
	mBeat byte = 0x08 // metronome beat count
)

const ( // (regst, 0x0F) registrations
	rgOpn byte = 0x00 // open registrations
	rgLoa byte = 0x01 // registration to load
	rgSto byte = 0x02 // registration to store to
	rgNam byte = 0x40 // name of registration
	rgMod byte = 0x50 // mode of registration (sound/pianist)
)

const ( // (mainF, 0x10) mainF
	mTran byte = 0x00 // transpose
	mTone byte = 0x02 // tone control
	mSpkV byte = 0x03 // speaker volume
	mLinV byte = 0x04 // line-in volume
	mWall byte = 0x05 // wall EQ
	mTung byte = 0x06 // tuning
	mDpHl byte = 0x07 // damper hold
	m__08 byte = 0x08 //
	mFact byte = 0x09 // factory reset
	mAOff byte = 0x0A // auto power-off
	m__0B byte = 0x0B //
	m__0C byte = 0x0C //
	mUTon byte = 0x41 // user tone control
)

const ( // (vTech, 0x11) virtual technician settings
	tCurv byte = 0x00 // touch curve
	voicg byte = 0x01 // voicing
	dmpRs byte = 0x02 // damper resonance
	dmpNs byte = 0x03 // damper noise
	strRs byte = 0x04 // string resonance
	uStRs byte = 0x05 // undamped-string resonance
	cabRs byte = 0x06 // cabinet resonance
	koEff byte = 0x07 // key-off effect
	fBkNs byte = 0x08 // fall-back noise
	hmDly byte = 0x09 // hammer delay
	topBd byte = 0x0A // topboard
	dcayT byte = 0x0B // decay time
	miTch byte = 0x0C // minimum touch
	streT byte = 0x0D // stretch tuning
	tmpmt byte = 0x0E // temperament
	tmKey byte = 0x0F // temperament key
	keyVo byte = 0x10 // key volume
	hfPdl byte = 0x11 // half-pedal adjust
	sfPdl byte = 0x12 // soft-pedal depth
	smart byte = 0x20 // virtual technician smart mode
	toSnd byte = 0x2F // store to sound
	uVoic byte = 0x41 // user voicing
	uStrT byte = 0x42 // user stretch tuning
	uTmpm byte = 0x43 // user temperament
	uKeyV byte = 0x44 // user key volume
)

const ( // (hPhon, 0x12) headphones properties
	phShs byte = 0x00 // SHS mode
	phTyp byte = 0x01 // phones type
	phVol byte = 0x02 // phones volume
)

const ( // (midiI, 0x13) MIDI
	miCha byte = 0x00 // midi channel
	miPgC byte = 0x01 // send pgm change number
	miLoc byte = 0x02 // local control
	miTrP byte = 0x03 // transmit pgm change numbers
	miMul byte = 0x04 // multi-timbral mode
	miMut byte = 0x40 // channel mute
)

const ( // (files, 0x14) file operations
	fUsNu byte = 0x00 // file number to load from USB
	fMvNu byte = 0x09 // file number to rename
	fRmNu byte = 0x0A // file number to delete
	fPgrs byte = 0x2D // progress during USB formatting
	fUsbE byte = 0x2F // USB error
	fUsNm byte = 0x40 // directory entry (name part) for file to load from USB
	fMvNm byte = 0x49 // directory entry (name part) for file to rename
	fRmNm byte = 0x4A // directory entry (name part) for file to delete
	fUsEx byte = 0x50 // directory entry (extension part) for file to load from USB
	fMvEx byte = 0x59 // directory entry (extension part) for file to rename
	fRmEx byte = 0x5A // directory entry (extension part) for file to delete
	fUsCf byte = 0x60 // load from USB
	fSvKs byte = 0x64 // save as .KSO file
	fSvSm byte = 0x65 // save as .SMF file
	fName byte = 0x69 // new filename
	fRmCf byte = 0x6A // delete file
	fFmat byte = 0x6B // format USB
)

const ( // (bluet, 0x16) bluetooth
	btAud byte = 0x00 // bluetooth audio
	btAuV byte = 0x01 // bluetooth audio volume
	btMid byte = 0x02 // bluetooth MIDI
)

const ( // (smRec, 0x20) sound mode recorder
	smSel byte = 0x00 // select sound mode song
	smPlP byte = 0x01 // part(s) to play
	smRcP byte = 0x02 // part to record to
	sm_04 byte = 0x04 //
	smEra byte = 0x40 // sound mode song or song parts to erase
	smPEm byte = 0x60 // whether part of sound mode song contains a recording
	smEmp byte = 0x61 // whether sound mode song contains a recording
)

const ( // (pmRec, 0x21) pianist mode recorder
	pmSel byte = 0x00 // select pianist mode song
	pmEra byte = 0x40 // pianist mode song number to erase
	pmEmp byte = 0x61 // whether pianist mode song contains a recording
)

const ( // (auRec, 0x22) USB audio recorder
	auSel byte = 0x00 // select USB song
	auTrn byte = 0x13 // song transpose
	au_20 byte = 0x20 //
	auTyp byte = 0x22 // file type to write
	auGai byte = 0x23 // recorder gain level
	au_30 byte = 0x30 //
	auPNm byte = 0x40 // file name to play
	auPEx byte = 0x41 // extension of file name to play
	auNam byte = 0x50 // file name to write
)

const ( // (servi, 0x60) service
	sMSel byte = 0x00 // select service mode
	// selectable service modes
	sMLcd byte = 0x00 // service mode 00, LCD
	sMPdV byte = 0x01 // service mode 01, pedal, volume, keyboard, midi, USB midi
	sMEfR byte = 0x02 // service mode 02, effect, reverb
	sMTgA byte = 0x03 // service mode 03, TG all channel
	sML_R byte = 0x04 // service mode 04, L/R
	sMEqL byte = 0x05 // service mode 05, EQ level
	sMUBt byte = 0x06 // service mode 06, USB device, bluetooth audio
	sMMTc byte = 0x07 // service mode 07, max touch
	sMTCk byte = 0x08 // service mode 08, tone check
	sMKRw byte = 0x09 // service mode 09, keyboard S1, S2, S3; AD raw value
	sMWCk byte = 0x0A // service mode 10, wave checksum
	sMAlK byte = 0x0B // service mode 11, all key on
	sMKAd byte = 0x0C // service mode 12, key adjust
	sMTcS byte = 0x0D // service mode 13, touch select
	sMVer byte = 0x0E // service mode 14, version (of UI)
	// service mode in use
	srPdV byte = 0x41 // pedal, volume, keyboard, midi, USB midi
	srEfR byte = 0x42 // effect, reverb
	srTgA byte = 0x43 // TG all channel
	srL_R byte = 0x44 // L/R
	srEqL byte = 0x45 // EQ level
	srUBt byte = 0x46 // USB device, bluetooth audio
	srMTc byte = 0x47 // max touch
	srTCk byte = 0x48 // tone check
	srKRw byte = 0x49 // keyboard S1, S2, S3; AD raw value
	srWCk byte = 0x4A // wave checksum
	srAlK byte = 0x4B // all key on
	srKAd byte = 0x4C // key adjust
	srTcS byte = 0x4D // touch select
)

const ( // (romId, 0x61) ROM identification
	roNam byte = 0x00 // ROM name
	roVer byte = 0x01 // ROM version string
	roCkS byte = 0x02 // checksum
)

const ( // (mbUpd, 0x63) MB update
	muUOk byte = 0x01 // update ok
	muNam byte = 0x40 // filename
	muCnt byte = 0x41 // byte count
	muDne byte = 0x42 // update done
)

const ( // (uiUpd, 0x64) UI update
	upErr byte = 0x00 // error: no USB
	upNow byte = 0x02 // update later
	upLtr byte = 0x03 // update now
)

const ( // (mrket = 0x65) model, market destination
	mkMdl byte = 0x00 // model (8=NV10)
	mkDst byte = 0x01 // market destination (1=EU)
)

const ( // (hardw, 0x70) user actions on hardware
	hwKey byte = 0x00 // piano key
	hwUsb byte = 0x01 // USB stick
	hwHPh byte = 0x02 // headphones
	hw_03 byte = 0x03 // ???
)

const ( // (playr, 0x71) player
	plDur byte = 0x00 // duration
	pl_01 byte = 0x01 // bar/second count
	pl_02 byte = 0x02 //
	plBrC byte = 0x03 // bar count
	plBea byte = 0x04 // beat
	plVol byte = 0x07 // playback volume
	pl_08 byte = 0x08 //
	pl_09 byte = 0x09 //
	plPla byte = 0x10 // start playing
	plRec byte = 0x11 // start recording
	plSto byte = 0x12 // stop recorder/player
	plA_B byte = 0x13 // A-B repeat mode
	plSby byte = 0x14 // put recorder into standby
	plPbM byte = 0x18 // playback mode
)

const ( // (pFace, 0x7E) piano/recorder/player face
	pFPno byte = 0x00 // normal piano mode
	pFInt byte = 0x02 // internal recorder/player
	pFUsb byte = 0x03 // USB recorder/player
	pFDem byte = 0x05 // demo songs
	pFLes byte = 0x07 // lesson songs
	pFCon byte = 0x08 // concert magic
	pFPMu byte = 0x09 // piano music
)

const ( // (commu, 0x7F) communication setup
	// commu = 0x7F // init
	coSvc byte = 0x00 // service screen (power on + pedals 1, 2)
	coVer byte = 0x01 // version screen (power on + pedals 2, 3)
	coMUd byte = 0x03 // mainboard firmware update screen (power on 10 sec)
	coUUd byte = 0x04 // UI update screen (power on + pedals 1, 2, 3)
)

var rhythmPatternMax = [...]int{ // number of beats
	// time signature
	0:  1, // 1/4
	2,  // 2/4
	3,  // 3/4
	4,  // 4/4
	5,  // 5/4
	3,  // 3/8
	6,  // 6/8
	7,  // 7/8
	9,  // 9/8
	12, // 12/8
	// 8 Beat
	10: 4, // 8 Beat 1
	4,  // 8 Beat 2
	4,  // 8 Beat 3
	4,  // Pop 1
	4,  // Pop 2
	4,  // Pop 3
	4,  // Pop 4
	4,  // Pop 5
	4,  // Pop 6
	4,  // Ride Beat 1
	4,  // Ride Beat 2
	4,  // Dance Pop 1
	4,  // Country Pop
	4,  // Smooth Beat
	4,  // Rim Beat
	// 8 Beat Rock
	25: 4, // Modern Rock 1
	4,  // Modern Rock 2
	4,  // Modern Rock 3
	4,  // Modern Rock 4
	4,  // Pop Rock
	4,  // Ride Rock
	4,  // Jazz Rock
	4,  // Surf Rock
	// 16 Beat
	33: 4, // 16 Beat
	4,  // Indie Pop 1
	4,  // Acid Jazz 1
	4,  // Ride Beat 3
	4,  // Dance Pop 2
	4,  // Dance Pop 3
	4,  // Dance Pop 4
	4,  // Dance Pop 5
	4,  // Dance Pop 6
	4,  // Dance Pop 7
	4,  // Dance Pop 8
	4,  // Indie Pop 2
	4,  // Cajun Rock
	// 8 Beat Ballad
	46: 4, // Pop Ballad 1
	4,  // Pop Ballad 2
	4,  // Pop Ballad 3
	4,  // Rock Ballad 1
	4,  // Rock Ballad 2
	4,  // Slow Jam
	4,  // 6/8 R&B Ballad
	4,  // Triplet Ballad 1
	4,  // Triplet Ballad 2
	// 16 Beat Ballad
	55: 4, // 16 Ballad 1
	4,  // Dance Ballad 1
	4,  // Dance Ballad 2
	4,  // Dance Ballad 3
	4,  // Electro Pop
	4,  // 16 Ballad 2
	4,  // Mod Pop Ballad
	// 16 Beat Dance
	62: 4, // Dance 1
	4,  // Dance 2
	4,  // Dance 3
	4,  // Disco
	4,  // Techno 1
	4,  // Techno 2
	// 16 Beat Swing
	68: 4, // 16 Shuffle 1
	4,  // 16 Shuffle 2
	4,  // 16 Shuffle 3
	4,  // Acid Jazz 2
	4,  // Acid Jazz 3
	4,  // New Jack Swing
	4,  // Modern Dance
	4,  // Indie Pop 3
	// 8 Beat Swing
	76: 4, // Swing Beat
	4,  // Motown
	4,  // Country 2 Beat
	4,  // Boogie
	// Triplet
	80: 4, // 8 Shuffle 1
	4,  // 8 Shuffle 2
	4,  // 8 Shuffle 3
	4,  // Dance Shuffle
	4,  // Triplet 1
	4,  // Triplet 2
	4,  // Triplet Rock
	4,  // Reggae
	// Jazz
	88: 4, // H.H. Swing
	4,  // Ride Swing
	4,  // Fast 4 Beat
	4,  // Afro Cuban
	4,  // Jazz Bossa
	3,  // Jazz Waltz
	5,  // 5/4 Swing
	// Latin/Traditional
	95:  4, // H.H. Bossa Nova
	4,   // Ride Bossa Nova
	4,   // Beguine
	4,   // Rhumba
	4,   // Cha Cha
	4,   // Mambo
	4,   // Samba
	4,   // Sala
	4,   // Merenge
	4,   // Tango
	4,   // Habanera
	3,   // Waltz
	4,   // Ragtime
	4,   // March
	109: 6, // 6/8 March
}
