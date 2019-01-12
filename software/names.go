package main

// We assume an 8-character 14-segment alphanumeric display
// The lower-case characters
//            gjkpqxy
// are ugly and should be avoided.

var names = map[string][]string{
	"toneGeneratorMode": {
		"Sound mode",
		"Pianist mode",
	},
	"keyboardMode": {
		"SINGLE", // Single
		"DUAL",   // Dual
		"SPLIT",  // Split
		"4HANDS", // 4Hands
	},
	"renderingCharacter": {
		"CLASSIC",  // Classic
		"ROMANTIC", // Romantic
		"FULL",     // Full
		"JAZZ",     // Jazz
		"BRILLNT",  // Brilliant
		"RICH",     // Rich
		"BALLAD",   // Ballad
		"POP",      // Pop
		"VINTAGE",  // Vintage
		"BOOGIE",   // Boogie
	},
	"instrumentSound": {
		// Piano 1
		"SK CN GD", // SK ConcertGrand
		"EX CN GD", // EX ConcertGrand
		"SK-5 GD",  // SK-5 Grand
		"JZZ CLN",  // Jazz Clean
		"JZZ O SC", // Jazz Old School
		"WARM GD",  // Warm Grand
		"WARM GD2", // Warm Grand 2
		"STD GD",   // Standard Grand
		// Piano 2
		"POP GD",   // Pop Grand
		"POP GD2",  // Pop Grand 2
		"POP PNO",  // Pop Piano
		"NW AG GD", // New Age Grand
		"UPRT PNO", // Upright Piano
		"MODN PNO", // Modern Piano
		"BOOG PNO", // Boogie Piano
		"HONKY TK", // Honky Tonk
		// Electric Piano
		"CL EPNO",  // Classic Electric Piano
		"60 EPNO",  // 60's Electric Piano
		"MD EPNO",  // Modern Electric Piano
		"Cl EPNO2", // Classic Electric Piano 2
		"CL EPNO3", // Classic Electric Piano 3
		"CRS EPNO", // Crystal Electric Piano
		"MD EPNO2", // Modern Electric Piano 2
		"MD EPNO3", // Modern Electric Piano 3
		// Organ
		"JZZ ORGN", // Jazz Organ
		"BLS ORGN", // Blues Organ
		"BLD ORGN", // Ballad Organ
		"GSP ORGN", // Gospel Organ
		"DRB ORGN", // Drawbar Organ
		"DRB ORG2", // Drawbar Organ 2
		"DRB ORG3", // Drawbar Organ 3
		"DRB ORG4", // Drawbar Organ 4
		"CHR ORGN", // Church Organ
		"DIAPASON", // Diapason
		"FUL ENSB", // Full Ensemble
		"DIAP OCT", // Diapason Oct.
		"CHFY TIB", // Chiffy Tibia
		"PCPL OCT", // Principal Oct.
		"BAROQUE",  // Baroque
		"SFT DIAP", // Soft Diapasn
		"SFT STRI", // Soft Strings
		"MEL FLUT", // Mellow Flutes
		"MED ENSB", // Medium Ensemble
		"LD ENSB",  // Loud Ensembe
		"BRT ENSB", // Bright Ensemble
		"FUL ORGN", // Full Organ
		"REED ENS", // Reed Ensemble
		// Harpsi & Mallets
		"HARPSICH", // Harpsichord
		"HRPSI OT", // Harpsichord Oct
		"VIBRAPHN", // Vibraphone
		"CLAVI",    // Clavi
		"MARIMBA",  // Marimba
		"CELESTA",  // Celesta
		"HARPSI2",  // Harpsichord 2
		"BELL SPL", // Bell Split
		// Strings
		"SLO STRG", // Slow Strings
		"STRG PAD", // String Pad
		"WRM STRG", // Warm Strings
		"STRG ENS", // String Ensemble
		"SFT ORCH", // Soft Orchestra
		"CHB STRG", // Chamber Strings
		"HARP",     // Harp
		"PIZZ STR", // Pizzicato Str.
		// Vocal & Pad
		"CHOIR",    // Choir
		"POP OOH",  // Pop Ooh
		"POP AAH",  // Pop Aah
		"CHOIR 2",  // Choir 2
		"JZZ ENSB", // Jazz Ensemble
		"POP ENSB", // Pop Ensemble
		"SLO CHR",  // Slow Choir
		"BTHY CHR", // Breathy Choir
		"NW AG PD", // New Age Pad
		"ATMOSPH",  // Atmosphere
		"ITOPIA",   // Itopia
		"BRIGHTNS", // Brightness
		"NW AG P2", // New Age Pad 2
		"BRSS PAD", // Brass Pad
		"BOWD PAD", // Bowed Pad
		"BT WM PD", // Bright Warm Pad
		// Bass & Guitar
		"WOOD BS",  // Wood Bass
		"FINGR BS", // Finger Bass
		"FRTLS BS", // Fretless Bass
		"PR CHOIR", // Principal Choir
		"W BS+RDE", // W.Bass & Ride
		"E BS+RDE", // E.Bass & Ride
		"BLD GTAR", // Ballad Guitar
		"PC NY GT", // Pick Nylon Gt.
		"FG NY GT", // Finger Nylon Gt
	},
	"metronomeOnOff": {
		"Off",
		"On",
	},
	"rhythmGroup": {
		"TIME SGN", // time signature
		"8 BEAT",   // 8 Beat
		"8 B ROCK", // 8 Beat Rock
		"16 BEAT",  // 16 Beat
		"8 B BLD",  // 8 Beat Ballad
		"16 B BLD", // 16 Beat Ballad
		"16 B DNC", // 16 Beat Dance
		"16 B SWI", // 16 Beat Swing
		"8 B SWI",  // 8 Beat Swing
		"TRIPLET",  // Triplet
		"JAZZ",     // Jazz
		"LATIN",    // Latin/Traditional
	},
	"rhythmPattern": {
		// time signature
		0:      "1/4", // 1/4
		"2/4",  // 2/4
		"3/4",  // 3/4
		"4/4",  // 4/4
		"5/4",  // 5/4
		"3/8",  // 3/8
		"6/8",  // 6/8
		"7/8",  // 7/8
		"9/8",  // 9/8
		"12/8", // 12/8
		// 8 Beat
		10:         "8 BEAT 1", // 8 Beat 1
		"8 BEAT 2", // 8 Beat 2
		"8 BEAT 3", // 8 Beat 3
		"POP 1",    // Pop 1
		"POP 2",    // Pop 2
		"POP 3",    // Pop 3
		"POP 4",    // Pop 4
		"POP 5",    // Pop 5
		"POP 6",    // Pop 6
		"RDE BT 1", // Ride Beat 1
		"RDE BT 2", // Ride Beat 2
		"DNC POP1", // Dance Pop 1
		"CTRY POP", // Country Pop
		"SMTH BT",  // Smooth Beat
		"RIM BEAT", // Rim Beat
		// 8 Beat Rock
		25:         "MDN RCK1", // Modern Rock 1
		"MDN RCK2", // Modern Rock 2
		"MDN RCK3", // Modern Rock 3
		"MDN RCK4", // Modern Rock 4
		"POP ROCK", // Pop Rock
		"RDE ROCK", // Ride Rock
		"JZZ ROCK", // Jazz Rock
		"SRF ROCK", // Surf Rock
		// 16 Beat
		33:         "16 BEAT", // 16 Beat
		"INDIE P1", // Indie Pop 1
		"ACD JZZ1", // Acid Jazz 1
		"RIDE BT3", // Ride Beat 3
		"DANCE P2", // Dance Pop 2
		"DANCE P3", // Dance Pop 3
		"DANCE P4", // Dance Pop 4
		"DANCE P5", // Dance Pop 5
		"DANCE P6", // Dance Pop 6
		"DANCE P7", // Dance Pop 7
		"DANCE P8", // Dance Pop 8
		"INDIE P2", // Indie Pop 2
		"CAJUN RK", // Cajun Rock
		// 8 Beat Ballad
		46:         "POP BLD1", // Pop Ballad 1
		"POP BLD2", // Pop Ballad 2
		"POP BLD3", // Pop Ballad 3
		"RCK BLD1", // Rock Ballad 1
		"RCK BLD2", // Rock Ballad 2
		"SLOW JAM", // Slow Jam
		"6/8 RB B", // 6/8 R&B Ballad
		"TPL BLD1", // Triplet Ballad 1
		"TPL BLD2", // Triplet Ballad 2
		// 16 Beat Ballad
		55:         "16 BLD 1", // 16 Ballad 1
		"DNC BLD1", // Dance Ballad 1
		"DNC BLD2", // Dance Ballad 2
		"DNC BLD3", // Dance Ballad 3
		"E POP",    // Electro Pop
		"16 BLD 2", // 16 Ballad 2
		"MD POP B", // Mod Pop Ballad
		// 16 Beat Dance
		62:         "DANCE 1", // Dance 1
		"DANCE 2",  // Dance 2
		"DANCE 3",  // Dance 3
		"DISCO",    // Disco
		"TECHNO 1", // Techno 1
		"TECHNO 2", // Techno 2
		// 16 Beat Swing
		68:         "16 SHFL1", // 16 Shuffle 1
		"16 SHFL2", // 16 Shuffle 2
		"16 SHFL3", // 16 Shuffle 3
		"ACD JZZ2", // Acid Jazz 2
		"ACD JZZ3", // Acid Jazz 3
		"NW JC SW", // New Jack Swing
		"MD DANCE", // Modern Dance
		"INDIE P3", // Indie Pop 3
		// 8 Beat Swing
		76:         "SWING BT", // Swing Beat
		"MOTOWN",   // Motown
		"CTRY 2BT", // Country 2 Beat
		"BOOGIE",   // Boogie
		// Triplet
		80:         "8 Shffl1", // 8 Shuffle 1
		"8 Shffl2", // 8 Shuffle 2
		"8 Shffl3", // 8 Shuffle 3
		"Dnce Shf", // Dance Shuffle
		"TRIPLET1", // Triplet 1
		"TRIPLET2", // Triplet 2
		"TRIPL RC", // Triplet Rock
		"REGGAE",   // Reggae
		// Jazz
		88:         "HH SWING", // H.H. Swing
		"RD SWING", // Ride Swing
		"FAST4BT",  // Fast 4 Beat
		"AFROCUBA", // Afro Cuban
		"JZ BOSSA", // Jazz Bossa
		"JZ WALTZ", // Jazz Waltz
		"5/4 Swng", // 5/4 Swing
		// Latin/Traditional
		95:         "HH BOSSA", // H.H. Bossa Nova
		"RD BOSSA", // Ride Bossa Nova
		"BEGUINE",  // Beguine
		"RHUMBA",   // Rhumba
		"CHA CHA",  // Cha Cha
		"MAMBO",    // Mambo
		"SAMBA",    // Samba
		"SALA",     // Sala
		"MERENGE",  // Merenge
		"TANGO",    // Tango
		"HABANERA", // Habanera
		"WALTZ",    // Waltz
		"RAGTIME",  // Ragtime
		"MARCH",    // March
		"6/8 MRCH", // 6/8 March
	},
	"ambienceType": {
		"NATURAL",  // Natural
		"SML ROOM", // Small Room
		"MED ROOM", // Medium Room
		"LGE ROOM", // Large Room
		"STUDIO",   // Studio
		"WOOD STD", // Wood Studio
		"MELLOW L", // Mellow Lounge
		"BRIGHT L", // Bright Lounge
		"LIVE STG", // Live Stage
		"ECHO",     // Echo
	},
	"reverbOnOff": {
		"revrbOff", // Off
		"revrbOn",  // On
	},
	"effectsOnOff": {
		"effctOff", // Off
		"effctOn",  // On
	},
	"reverbType": {
		"ROOM",     // Room
		"LOUNGE",   // Lounge
		"SMLL HLL", // Small Hall
		"CCRT HLL", // Concert Hall
		"LIVE HLL", // Live Hall
		"CATHEDRL", // Cathedral
	},
	"effectsType": {
		"MONO DLY", // Mono Delay
		"PING DLY", // Ping Delay
		"TRPL DLY", // Triple Delay
		"CHORUS",   // Chorus
		"CLSS CHS", // Classic Chorus
		"ENSEMBLE", // Ensemble
		"TREMOLO",  // Tremolo
		"CLSS TRL", // Classic Tremolo
		"VBRT TRL", // Vibrato Tremolo
		"AUTO PAN", // Auto Pan
		"CLSS AtP", // Classic Auto Pan
		"PHASER",   // Phaser
		"CLSS PHS", // Classic Phaser
		"ROTARY 1", // Rotary 1
		"ROTARY 2", // Rotary 2
		"ROTARY 3", // Rotary 3
		"ROTARY 4", // Rotary 4
		"ROTARY 5", // Rotary 5
	},
	"damperHold": {
		"DMPR OfF", // Off
		"DMPR ON",  // On
	},
	"smartModeVt": {
		"OFF",      // Off
		"NOISELSS", // Noiseless
		"DEEP RES", // Deep Resonance
		"LGHT RES", // Light Resonance
		"SOFT",     // Soft
		"BRILLNT",  // Brilliant
		"CLEAN",    // Clean
		"FULL",     // Full
		"DARK",     // Dark
		"RICH",     // Rich
		"CLASSICL", // Classical
	},
	"touchCurve": {
		"LIGHT+", // Light+
		"LIGHT",  // Light
		"NORMAL", // Normal
		"HEAVY",  // Heavy
		"HEAVY+", // Heavy+
		"OFF",    // Off
		"User",   // User
	},
	"voicing": {
		"NORMAL",   // Normal
		"MELLOW 1", // Mellow 1
		"MELLOW 2", // Mellow 2
		"DYNAMIC",  // Dynamic
		"BRIGHT 1", // Bright 1
		"BRIGHT 2", // Bright 2
		"User",     // User
	},
	"topboard": {
		"OPEN 3", // Open 3
		"OPEN 2", // Open 2
		"OPEN 1", // Open 1
		"CLOSED", // Closed
	},
	"stretchTuning": {
		"OFF",    // Off
		"NORMAL", // Normal
		"WIDE",   // Wide
		"User",   // User
	},
	"temperament": {
		"EQUAL",    // Equal
		"PR MJ/MN", // Pure Major/Pure Minor
		"PYTHAGRN", // Pythagorean
		"MEANTONE", // Meantone
		"WRCKMSTR", // Werckmeister
		"KIRNBRGR", // Kirnberger
		"User",     // User
	},
	"temperamentKey": {
		"C",
		"Db",
		"D",
		"Eb",
		"E",
		"F",
		"Gb",
		"G",
		"Ab",
		"A",
		"Bb",
		"B",
	},
	"keyVolume": {
		"NORMAL",   // Normal
		"HI DAMPG", // High Damping
		"LO DAMPG", // Low Damping
		"HI+LO DP", // High & Low Damping
		"CENTR DP", // Center Damping
		"User",     // User
	},
	"shsMode": {
		"OFF",     // Off
		"FORWARD", // Forward
		"NORMAL",  // Normal
		"WIDE",    // Wide
	},
	"phonesType": {
		"NORMAL",   // Normal
		"OPEN",     // Open
		"SEMIOPEN", // Semi-open
		"CLOSED",   // Closed
		"INNR-EAR", // Inner-ear
		"CANAL",    // Canal
	},
	"usbThumbDrivePresence": {
		"NO USB", // unplugged
		"USB IN", // plugged
	},
	"phonesPresence": {
		"NO HdPhs", // unplugged
		"HdPhones", // plugged
	},
	"toneControl": {
		"OFF",      // Off
		"BRILLNCE", // Brilliance
		"LOUDNESS", // Loudness
		"BS BOOST", // Bass Boost
		"TB BOOST", // Treble Boost
		"MID CUT",  // Mid Cut
		"User",     // User
	},
	"userToneControl": {
		"LO dB",    // low dB
		"MIDLO f",  // mid-low freqency
		"MIDLO dB", // mid-low dB
		"MIDHI f",  // mid-high frequency
		"MIDHI dB", // mid-high dB
		"HI f",     // high frequency
	},
	"speakerVolume": {
		"LOW",    // Low
		"NORMAL", // Normal
	},
	"phonesVolume": {
		"NORMAL", // Normal
		"HIGH",   // High
	},
	"wallEq": {
		"Wall Off", // Off
		"Wall On",  // On
	},
	"autoPowerOff": {
		"OFF",     // Off
		"15 min",  // 15 min
		"60 min",  // 60 min
		"120 min", // 120 min
	},
	"fileExt": {
		2: "*dir", //  (dir)
		6: ".KSO", // .KSO
		7: ".MID", // .MID
		8: ".MP3", // .MP3
		9: ".WAV", // .WAV
	},
	"recorderFileType": {
		"MP3", // .MP3
		"WAV", // .WAV
	},
	"emptiness": { // of various songs
		"EMPTY", // empty
		"USED",  // used
	},
	"mutedness": { // of various midi channels
		"MUTED",   // muted
		"UNMUTED", // unmuted
	},
	"bluetoothMidi": {
		"BtMi OFF", // Off
		"BtMi ON",  // On
	},
	"bluetoothAudio": {
		"BtAu OFF", // Off
		"BtAu ON",  // On
	},
	"midiLocalControl:": {
		"LOCL OFF", // Off
		"LOCL ON",  // On
	},
	"transmitPgmNumberOnOff": {
		"PRGM OFF", // Off
		"PRGM ON",  // On
	},
	"multiTimbralMode": {
		"MULT OFF", // Off
		"MULT ON1", // On1
		"MULT ON2", // On2
	},
}

func init() {
	names["single"] = names["instrumentSound"]
	names["dual1"] = names["instrumentSound"]
	names["dual2"] = names["instrumentSound"]
	names["split1"] = names["instrumentSound"]
	names["split2"] = names["instrumentSound"]
	names["4hands1"] = names["instrumentSound"]
	names["4hands2"] = names["instrumentSound"]
}

var rhythmGroupIndex = [...]int{
	0,  // time signature
	10, // 8 Beat
	25, // 8 Beat Rock
	33, // 16 Beat
	46, // 8 Beat Ballad
	55, // 16 Beat Ballad
	62, // 16 Beat Dance
	68, // 16 Beat Swing
	76, // 8 Beat Swing
	80, // Triplet
	88, // Jazz
	95, // Latin/Traditional
}

var settingTopics = map[string]string{
	// virtual technician
	"smartModeVt":             "SMART VT",
	"touchCurve":              "TCH CURV",
	"voicing":                 "VOICING",
	"userVoicing":             "UVOICING",
	"damperResonance":         "DMPR RES",
	"damperNoise":             "DAMPR NS",
	"stringResonance":         "STRI RES",
	"undampedStringResonance": "USTR RES",
	"cabinetResonance":        "CABI RES",
	"keyOffEffect":            "KOFF EFF",
	"fallBackNoise":           "FALLB NS",
	"hammerDelay":             "HMR DLAY",
	"topboard":                "TOPBOARD",
	"decayTime":               "DCAY TME",
	"minumumTouch":            "MINTOUCH",
	"stretchTuning":           "STRTCH T",
	"userStretchTuning":       "U STRT T",
	"temperament":             "TEMPERAM",
	"temperamentKey":          "TMPR KEY",
	"userTemperament":         "UTEMPERM",
	"keyVolume":               "KEY VOL",
	"userKeyVolume":           "UKEY VOL",
	"halfPedalAdjust":         "HALF PDL",
	"softPedalDepth":          "SOFT PDL",
	// other settings
	"ambienceType":  "AMB TYPE",
	"ambienceDepth": "AMB DPTH",
	"reverbType":    "RVB TYPE",
	"reverbDepth":   "RVB DPTH",
	"effectsType":   "EFF TYPE",
	"effectsParam1": "EFF PRM1",
	"effectsParam2": "EFF PRM2",
	"transpose":     "TRNSPOSE",
	// TODO: split-keyboard mode parameters still missing
	"tuning":                 "TUNING",
	"damperHold":             "DMP HOLD",
	"toneControl":            "TONE CRT",
	"speakerVolume":          "SPKR VOL",
	"lineInLevel":            "L IN VOL",
	"wallEq":                 "WALL EQ",
	"shsMode":                "SHS MODE",
	"phonesType":             "HP TYPE",
	"phonesVolume":           "HP VOL",
	"bluetoothMidi":          "BT MIDI",
	"bluetoothAudio":         "BT AUDIO",
	"bluetoothAudioVolume":   "BT A VOL",
	"midiChannel":            "MIDI CH",
	"localControl":           "LOC CTRL",
	"transmitPgmNumberOnOff": "TRM PGM",
	"multiTimbralMode":       "MUL TIMB",
	"channelMute":            "CH MUTE",
	"lcdContrast":            "LCD CONT",
	"autoDisplayOff":         "ADSP OFF",
	"autoPowerOff":           "AUTO OFF",
	"metronomeVolume":        "METR VOL",
	"recorderGainLevel":      "REC GAIN",
	"recorderFileType":       "MP3orWAV",
	"usbPlayerVolume":        "USB VOL",
	"usbPlayerTranspose":     "USB TRNS",
}

var immediateActionNames = map[string]string{
	"factoryReset": "FACTORY",
	"usbFormat":    "FRMT USB",
}

var errors = map[string]string{
	"cancelled":         "CANCELLD",
	"notInPianistMode":  "UNAVL PM",
	"notInSoundMode":    "UNAVL SM",
	"onlyIn2SoundModes": "2SD ONLY",
	"onlyInSplitModes":  "SPL ONLY",
	"usbError":          "USB ERR",
}
