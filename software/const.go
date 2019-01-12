package main

const ( // message headers
	hdr0 = 0x55 // first byte of any message
	hdr1 = 0xAA // second byte of any message
	hdr2 = 0x00 // third byte of any message
	// messages sent to the mainboard
	cmdVn = 0x60 // fourth byte of vanilla command
	cmdAc = 0x61 // fourth byte of command that expects acknoledgement (sent back in a mbCAc message)
	dtaRq = 0x62 // fourth byte of a data request (sent back in a dtaRq message)
	// messages sent by the mainboard
	mbMsg = 0x6E // message sent without request
	mbCAc = 0x71 // confirmation of a cmdAc command
	mbDRq = 0x72 // response to a dtaRq data request
)

const ( // principal topics of mainboard operation
	tgMod = 0x00 // tone generator mode (sound/pianist)
	kbSpl = 0x01 // keyboard splitting
	instr = 0x02 // sound-mode instrument
	pmSet = 0x04 // pianist mode settings
	dlSet = 0x05 // dual-mode settings (sound mode)
	spSet = 0x06 // split-mode settings (sound mode)
	h4Set = 0x07 // 4hands-mode settings (sound mode)
	revrb = 0x08 // reverb
	effct = 0x09 // effects
	metro = 0x0A // metronome/drums
	regst = 0x0F // registrations aka favourites
	mainF = 0x10 // transpose, factory reset
	vTech = 0x11 // virtual technician
	hPhon = 0x12 // headphones
	midiI = 0x13 // midi
	files = 0x14 // files on USB storage
	bluet = 0x16 // bluetooth
	lcdCn = 0x17 // LCD contrast
	smRec = 0x20 // recorder, sound mode
	pmRec = 0x21 // recorder, pianist mode
	auRec = 0x22 // USB audio recorder
	msg32 = 0x32 // lesson song volume balance
	biSng = 0x3F // built-in songs
	servi = 0x60 // service screen
	romId = 0x61 // ROM identification
	updat = 0x64 // update
	msg65 = 0x65 // ???
	hardw = 0x70 // user action on hardware: connecting devices, pressing keys
	playr = 0x71 // player
	rpFce = 0x7E // recorder/player face
	commu = 0x7F // communication setup
)

const ( // (tgMod, 0x00) tone generator modes
	// tgMod = 0x00  happens to be equal to the principal topic
	// submodes
	tgSnd = 0x00 // sound mode
	tgPia = 0x01 // pianist mode
)

const ( // (kbSpl, 0x01) keyboard splitting
	kbSpM = 0x00 // single, dual, split, 4hands
)
const ( // the four keyboard spliting modes
	kbSpMSingle = 0x00
	kbSpMDual   = 0x01
	kbSpMSplit  = 0x02
	kbSpM4hands = 0x03
)

const ( // (instr, 0x02) sound-mode instruments
	iSing = 0x00 // single instrument
	iDua1 = 0x01 // 1st dual instrument
	iDua2 = 0x02 // 2nd dual instrument
	iSpl1 = 0x03 // 1st split instrument
	iSpl2 = 0x04 // 2nd split instrument
	i4Hd1 = 0x05 // 1st 4hands instrument
	i4Hd2 = 0x06 // 2nd 4hands instrument
)

const ( // (pmSet, 0x04) pianist mode settings
	pmRen = 0x00 // rendering character
	pmRes = 0x01 // resonance depth
	pmAmb = 0x02 // ambience type
	pmAmD = 0x03 // ambience depth
)

const ( // (dlSet, 0x05) dual-mode settings (sound mode)
	dlBal = 0x00 // balance
	dlOcS = 0x01 // layer octave shift
	dlDyn = 0x02 // dynamics
)

const ( // (spSet, 0x06) split-mode settings (sound mode)
	spBal = 0x00 // balance
	spOcS = 0x01 // lower octave shift
	spPed = 0x02 // lower pedal
	spSpP = 0x03 // split point
)
const ( // (h4Set, 0x07) 4hands-mode settings (sound mode)
	h4Bal = 0x00 // balance
	h4LOS = 0x01 // left octave shift
	h4ROS = 0x02 // right octave shift
	h4SpP = 0x03 // split point
)

const ( // (revrb, 0x08) reverb settings
	rOnOf = 0x00 // reverb on/off
	rType = 0x01 // reverb type
	rDpth = 0x02 // reverb depth
	rTime = 0x03 // reverb time
)

const ( // (effct, 0x09) effects settings
	eOnOf = 0x00 // effects on/off
	eType = 0x01 // effects type
	ePar1 = 0x02 // effects parameter 1
	ePar2 = 0x03 // effects parameter 2
)

const ( // (metro, 0x0A) metronome settings
	mOnOf = 0x00 // metronome on/off
	mTmpo = 0x01 // tempo
	mSign = 0x02 // time signature/rhythm pattern
	mVolu = 0x03 // metronome volume
	mBeat = 0x08 // metronome beat count
)

const ( // (regst, 0x0F) registrations
	rgOpn = 0x00 // open registrations
	rgLoa = 0x01 // registration to load
	rgSto = 0x02 // registration to store to
	rgNam = 0x40 // name of registration
	rgMod = 0x50 // mode of registration (sound/pianist)
)

const ( // (mainF, 0x10) mainF
	mTran = 0x00 // transpose
	mTone = 0x02 // tone control
	mSpkV = 0x03 // speaker volume
	mLinV = 0x04 // line-in volume
	mWall = 0x05 // wall EQ
	mTung = 0x06 // tuning
	mDpHl = 0x07 // damper hold
	m__08 = 0x08 //
	mFact = 0x09 // factory reset
	mAOff = 0x0A // auto power-off
	m__0B = 0x0B //
	m__0C = 0x0C //
	mUTon = 0x41 // user tone control
)

const ( // (vTech, 0x11) virtual technician settings
	tCurv = 0x00 // touch curve
	voicg = 0x01 // voicing
	dmpRs = 0x02 // damper resonance
	dmpNs = 0x03 // damper noise
	strRs = 0x04 // string resonance
	uStRs = 0x05 // undamped-string resonance
	cabRs = 0x06 // cabinet resonance
	koEff = 0x07 // key-off effect
	fBkNs = 0x08 // fall-back noise
	hmDly = 0x09 // hammer delay
	topBd = 0x0A // topboard
	dcayT = 0x0B // decay time
	miTch = 0x0C // minimum touch
	streT = 0x0D // stretch tuning
	tmpmt = 0x0E // temperament
	tmKey = 0x0F // temperament key
	keyVo = 0x10 // key volume
	hfPdl = 0x11 // half-pedal adjust
	sfPdl = 0x12 // soft-pedal depth
	smart = 0x20 // virtual technician smart mode
	toSnd = 0x2F // store to sound
	uVoic = 0x41 // user voicing
	uStrT = 0x42 // user stretch tuning
	uTmpm = 0x43 // user temperament
	uKeyV = 0x44 // user key volume
)

const ( // (hPhon, 0x12) headphones properties
	phShs = 0x00 // SHS mode
	phTyp = 0x01 // phones type
	phVol = 0x02 // phones volume
)

const ( // (midiI, 0x13) MIDI
	miCha = 0x00 // midi channel
	miPgC = 0x01 // pgm change number
	miLoc = 0x02 // local control
	miTrP = 0x03 // transmit pgm change numbers
	miMul = 0x04 // multi-timbral mode
	miMut = 0x40 // channel mute
)

const ( // (files, 0x14) file operations
	fUsNu = 0x00 // file number to load from USB
	fMvNu = 0x09 // file number to rename
	fRmNu = 0x0A // file number to delete
	fPgrs = 0x2D // progress during USB formatting
	fUsbE = 0x2F // USB error
	fUsNm = 0x40 // directory entry (name part) for file to load from USB
	fMvNm = 0x49 // directory entry (name part) for file to rename
	fRmNm = 0x4A // directory entry (name part) for file to delete
	fUsEx = 0x50 // directory entry (extension part) for file to load from USB
	fMvEx = 0x59 // directory entry (extension part) for file to rename
	fRmEx = 0x5A // directory entry (extension part) for file to delete
	fUsCf = 0x60 // load from USB
	fSvKs = 0x64 // save as .KSO file
	fSvSm = 0x65 // save as .SMF file
	fName = 0x69 // new filename
	fRmCf = 0x6A // delete file
	fFmat = 0x6B // format USB
)

const ( // (bluet, 0x16) bluetooth
	btAud = 0x00 // bluetooth audio
	btAuV = 0x01 // bluetooth audio volume
	btMid = 0x02 // bluetooth MIDI
)

const ( // (smRec, 0x20) sound mode recorder
	smSel = 0x00 // select sound mode song
	smPlP = 0x01 // part(s) to play
	smRcP = 0x02 // part to record to
	sm_04 = 0x04 //
	smEra = 0x40 // sound mode song or song parts to erase
	smPEm = 0x60 // whether part of sound mode song contains a recording
	smEmp = 0x61 // whether sound mode song contains a recording
)

const ( // (pmRec, 0x21) pianist mode recorder
	pmSel = 0x00 // select pianist mode song
	pmEra = 0x40 // pianist mode song number to erase
	pmEmp = 0x61 // whether pianist mode song contains a recording
)

const ( // (auRec, 0x22) USB audio recorder
	auSel = 0x00 // select USB song
	auTrn = 0x13 // song transpose
	au_20 = 0x20 //
	auTyp = 0x22 // file type to write
	auGai = 0x23 // recorder gain level
	au_30 = 0x30 //
	auPNm = 0x40 // file name to play
	auPEx = 0x41 // extension of file name to play
	auNam = 0x50 // file name to write
)

const ( // (servi, 0x60) service
	srM00 = 0x00 // service mode 00
	srM01 = 0x01 // service mode 01
	srM02 = 0x02 // service mode 02
	srM03 = 0x03 // service mode 03
	srM04 = 0x04 // service mode 04
	srM05 = 0x05 // service mode 05
	srM06 = 0x06 // service mode 06
	srM07 = 0x07 // service mode 07
	srM08 = 0x08 // service mode 08
	srM09 = 0x09 // service mode 09
	srM10 = 0x0A // service mode 10
	srM11 = 0x0B // service mode 11
	srM12 = 0x0C // service mode 12
	srM13 = 0x0D // service mode 13
	srM14 = 0x0E // service mode 14
)

const ( // (updat, 0x64) update
	upErr = 0x00 // error: no USB
	upNow = 0x02 // update now
	upLtr = 0x03 // update later
)

const ( // (hardw, 0x70) user actions on hardware
	hwKey = 0x00 // piano key
	hwUsb = 0x01 // USB stick
	hwHPh = 0x02 // headphones
	hw_03 = 0x03 // ???
)

const ( // (playr, 0x71) player
	plDur = 0x00 // duration
	pl_01 = 0x01 // bar/second count
	pl_02 = 0x02 //
	plBrC = 0x03 // bar count
	plBea = 0x04 // beat
	plVol = 0x07 // playback volume
	pl_08 = 0x08 //
	pl_09 = 0x09 //
	plPla = 0x10 // start playing
	plRec = 0x11 // start recording
	plSto = 0x12 // stop recorder/player
	plA_B = 0x13 // A-B repeat mode
	plSby = 0x14 // put recorder into standby
	plPbM = 0x18 // playback mode
)

const ( // (rpFce, 0x7E) recorder/player face
	rpClo = 0x00 // close player/recorder
	rpInt = 0x02 // internal recorder/player
	rpUsb = 0x03 // USB recorder/player
	rpDem = 0x05 // demo songs
	rpLes = 0x07 // lesson songs
	rpCon = 0x08 // concert magic
	rpPno = 0x09 // piano music
)

const ( // (commu, 0x7E) communication setup
//      commu = 0x7F // init
)
