package main

import (
	"fmt"
	"golang.org/x/exp/io/i2c"
	"log"
)

type display struct {
	blinkState        byte
	spinState         int
	breatheState      int
	defaultBrightness int // Ox0..0xF
	waxing            bool
	buf               [16]byte
	d0                *i2c.Device
	d1                *i2c.Device
	w                 chan string
	brth              chan struct{}
	spn               chan spinPattern
}

func openDisplay(addr0 int, addr1 int) (d display, err error) {
	d0, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, addr0)
	if err != nil {
		return
	}
	d1, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, addr1)
	if err != nil {
		return
	}
	d.w = make(chan string)
	d.brth = make(chan struct{})
	d.spn = make(chan spinPattern)
	d.d0 = d0
	d.d1 = d1
	d.d0.Write([]byte{0x21})
	d.d1.Write([]byte{0x21})
	d.d0.Write([]byte{0x81})
	d.d1.Write([]byte{0x81})
	return
}

func (d *display) close() {
	d.d0.Close()
	d.d1.Close()
}

func (d *display) displayMonitor() {
	for {
		select {
		case s := <-d.w:
			d.write(s)
		case <-d.brth:
			d.breathe()
		case p := <-d.spn:
			d.spin(p)
		}
	}
}

func init() {
	var err error
	seg14, err = openDisplay(0x70, 0x71)
	if err != nil {
		log.Print(err)
	}
	seg14.defaultBrightness = 0xC
	go seg14.displayMonitor()
}

func (d *display) write(txt string) {
	d.dim(d.defaultBrightness)
	var dotlessTxt [8]byte
	var dots [8]bool
	i := 0
	for _, c := range []byte(txt) {
		if c == '.' && i > 0 {
			dots[i-1] = true
		} else {
			dotlessTxt[i] = c
			i++
		}
		if i > 8 {
			break
		}
	}
	fmt.Println(dotlessTxt, dots)
	for i, c := range fmt.Sprintf("%-8s", dotlessTxt)[:8] {
		shape := ascii14Segment[c]
		if dots[i] {
			shape |= ascii14Segment['.']
		}
		d.buf[2*i] = byte(shape & 0x00FF)
		d.buf[2*i+1] = byte(shape >> 8)
	}
	d.d0.WriteReg(0, d.buf[:8])
	d.d1.WriteReg(0, d.buf[8:])
}

func (d *display) setDots(pat [8]bool) {
	d.dim(d.defaultBrightness)
	for i, dot := range pat {
		if dot {
			d.buf[2*i+1] |= 0x40
		} else {
			d.buf[2*i+1] &= ^byte(0x40)
		}
	}
	d.d0.WriteReg(0, d.buf[:8])
	d.d1.WriteReg(0, d.buf[8:])
}

type spinPattern struct {
	spinMap   []uint16
	positions []int
}

func (d *display) spin(s spinPattern) {
	d.dim(d.defaultBrightness)
	if d.spinState > len(s.spinMap)-1 {
		d.spinState = 0
	}
	shape := s.spinMap[d.spinState]
	for _, pos := range s.positions {
		d.buf[2*pos] = byte(shape & 0x00FF)
		d.buf[2*pos+1] = byte(shape >> 8)
	}
	d.d0.WriteReg(0, d.buf[:8])
	d.d1.WriteReg(0, d.buf[8:])
	d.spinState += 1
}

func (d *display) breathe() {
	if d.waxing {
		d.breatheState++
	} else {
		d.breatheState--
	}
	d.dim(d.breatheState)
	switch d.breatheState {
	case 0:
		d.waxing = true
	case d.defaultBrightness:
		d.waxing = false
	}
}

func (d *display) dim(b int) {
	brightness := byte(b & ^0xF0)
	d.d0.Write([]byte{0xE0 | brightness})
	d.d1.Write([]byte{0xE0 | brightness})
	d.breatheState = int(brightness)
}

// func (d *display) hide(y bool) {
// 	if y {
// 		d.blinkState = (d.blinkState | 0x80) & ^byte(0x01)
// 	} else {
// 		d.blinkState = d.blinkState | 0x81
// 	}
// 	d.d0.Write([]byte{d.blinkState})
// 	d.d1.Write([]byte{d.blinkState})
// }

var runningNeedle []uint16 = []uint16{
	// the individual arms of "*"
	0x00C0,
	0x2100,
	0x1200,
	0x0C00,
}

var runningPointer []uint16 = []uint16{
	// the individual arms of "*"
	0x0040,
	0x0100,
	0x0200,
	0x0400,
	0x0080,
	0x2000,
	0x1000,
	0x0800,
}

var runningDoublePointer []uint16 = []uint16{
	// one or two of the individual arms of "*"
	0x0040,
	0x0140,
	0x0100,
	0x0300,
	0x0200,
	0x0600,
	0x0400,
	0x0480,
	0x0080,
	0x2080,
	0x2000,
	0x3000,
	0x1000,
	0x1800,
	0x0800,
	0x0840,
}

var runningOutline []uint16 = []uint16{
	0x0003,
	0x0006,
	0x000C,
	0x0018,
	0x0030,
	0x0021,
}

var ascii14Segment []uint16 = []uint16{
	// github.com/dmadison/Segmented-LED-Display-ASCII
	// Copyright (c) 2017 David Madison

	// Permission is hereby granted, free of charge, to any person obtaining a copy
	// of this software and associated documentation files (the "Software"), to deal
	// in the Software without restriction, including without limitation the rights
	// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	// copies of the Software, and to permit persons to whom the Software is
	// furnished to do so, subject to the following conditions:

	// The above copyright notice and this permission notice shall be included in
	// all copies or substantial portions of the Software.

	// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	// THE SOFTWARE.
	0x20:   0x0000, // (space)
	0x4006, // !
	0x0202, // "
	0x12CE, // #
	0x12ED, // $
	0x3FE4, // %
	0x2359, // &
	0x0200, // '
	0x2400, // (
	0x0900, // )
	0x3FC0, // *
	0x12C0, // +
	0x0800, // ,
	0x00C0, // -
	0x4000, // .
	0x0C00, // /
	0x0C3F, // 0
	0x0406, // 1
	0x00DB, // 2
	0x008F, // 3
	0x00E6, // 4
	0x2069, // 5
	0x00FD, // 6
	0x0007, // 7
	0x00FF, // 8
	0x00EF, // 9
	0x1200, // :
	0x0A00, // ;
	0x2440, // <
	0x00C8, // =
	0x0980, // >
	0x5083, // ?
	0x02BB, // @
	0x00F7, // A
	0x128F, // B
	0x0039, // C
	0x120F, // D
	0x0079, // E
	0x0071, // F
	0x00BD, // G
	0x00F6, // H
	0x1209, // I
	0x001E, // J
	0x2470, // K
	0x0038, // L
	0x0536, // M
	0x2136, // N
	0x003F, // O
	0x00F3, // P
	0x203F, // Q
	0x20F3, // R
	0x00ED, // S
	0x1201, // T
	0x003E, // U
	0x0C30, // V
	0x2836, // W
	0x2D00, // X
	0x00EE, // Y
	0x0C09, // Z
	0x0039, // [
	0x2100, // \
	0x000F, // ]
	0x2800, // ^
	0x0008, // _
	0x0100, // `
	0x1058, // a
	0x2078, // b
	0x00D8, // c
	0x088E, // d
	0x0858, // e
	0x14C0, // f
	0x048E, // g
	0x1070, // h
	0x1000, // i
	0x0A10, // j
	0x3600, // k
	0x0030, // l
	0x10D4, // m
	0x1050, // n
	0x00DC, // o
	0x0170, // p
	0x0486, // q
	0x0050, // r
	0x2088, // s
	0x0078, // t
	0x001C, // u
	0x0810, // v
	0x2814, // w
	0x2D00, // x
	0x028E, // y
	0x0848, // z
	0x0949, // {
	0x1200, // |
	0x2489, // }
	0x0CC0, // ~
	0x0000, // (del)
}
