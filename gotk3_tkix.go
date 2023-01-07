package main

/*
gotk3_tkix.go: a simple game where you try to catch the bouncing
"TKix" by grabbing both ends with a movable black box.

ISC License

Copyright (c) 2023 Glenn Horton-Smith

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
PERFORMANCE OF THIS SOFTWARE.
*/

import (
	"fmt"
	"math"
	"math/rand"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
)

const (
	KEY_LEFT  uint = 65361
	KEY_UP    uint = 65362
	KEY_RIGHT uint = 65363
	KEY_DOWN  uint = 65364
)

func HueToRgb(h float64) (float64, float64, float64) {
	h *= 3.0
	r, g, b := 0.0, 0.0, 0.0
	if h < 1.0 { 
		r = 1.0 - h 
		g = h
	} else if h < 2.0 {
		g = 2.0 - h
		b = h - 1.0
	} else if h <= 3.0 {
		b = 3.0 - h
		r = h - 2.0
	}
	return r, g, b
}

func main() {
	gtk.Init(nil)

	// gui boilerplate
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	width := 400.0
	height := 200.0
	win.SetDefaultSize(int(width), int(height))
	da, _ := gtk.DrawingAreaNew()
	win.Add(da)
	win.SetTitle("TKix catch!")
	win.Connect("destroy", gtk.MainQuit)
	win.ShowAll()

	// Data
	x0 := 10.0
	y0 := 10.0 
	x1 := 10.0
	y1 := 25.0
	vx0 := 2.0 
	vy0 := 1.0
	vx1 := 1.0
	vy1 := 2.0
	hue := 0
	const Ntrail = 120
	var trail [Ntrail][5]float64
	itrail := 0
	step_and_bounce := func(u, v, limit float64)  (float64, float64) {
		u += v
		if (u >= limit) {
			u = limit-1
			if (v > 0) {
				v = -v
			}
		} else if (u < 0) {
		  u = 0;
		  if (v < 0) {
				v = -v
			}
		}
		return u, v
	}
	unitSize := 20.0
	x := width - unitSize
	y := height - unitSize
	caughtEnds := 0
	wasCaught := false
	level := 0.0

	// motion
	moveStep := func() {
		trail[itrail][0] = x0
		trail[itrail][1] = y0
		trail[itrail][2] = x1
		trail[itrail][3] = y1
		trail[itrail][4] = float64(hue)/360.0
		hue = (hue + 1) % 360
		itrail = (itrail + 1) % Ntrail
		caughtEnds = 0
		if (x0 < x || y0 < y || x0 >= x+unitSize || y0 >= y+unitSize) {
			x0, vx0 = step_and_bounce(x0, vx0, width)
			y0, vy0 = step_and_bounce(y0, vy0, height)
		} else {
			caughtEnds += 1
		}
		if (x1 < x || y1 < y || x1 >= x+unitSize || y1 >= y+unitSize) {
			x1, vx1 = step_and_bounce(x1, vx1, width)
			y1, vy1 = step_and_bounce(y1, vy1, height)
		} else {
			caughtEnds += 1
		}
	}
	
	randomizeStick := func() {
		len := 15.0
		x0 = rand.Float64()*(width-2*len)+len
		y0 = rand.Float64()*(width-2*len)+len
		phi_r := 2.0*math.Pi*rand.Float64()
		x1 = x0 + math.Cos(phi_r)*len
		y1 = y0 + math.Sin(phi_r)*len
		v := 2.0 + level
		phi_v0 := 2.0*math.Pi*rand.Float64()
		vx0 = v * math.Cos(phi_v0)
		vy0 = v * math.Sin(phi_v0)
		phi_v1 := phi_v0 + math.Pi*(0.25+0.5*rand.Float64())
		vx1 = v * math.Cos(phi_v1)
		vy1 = v * math.Sin(phi_v1)
	}
	
	// Event handlers
	keyMap := map[uint]func(){
		KEY_LEFT:  func() { x-=unitSize; if (x<0) { x=0.0 } },
		KEY_UP:    func() { y-=unitSize; if (y<0) { y=0.0 } },
		KEY_RIGHT: func() { x+=unitSize; if (x+unitSize > width) { x = width-unitSize} },
		KEY_DOWN:  func() { y+=unitSize; if (y+unitSize > height) { y = height-unitSize}  },
		' ': func() { moveStep() },
	}
	
	draw := func (da *gtk.DrawingArea, cr *cairo.Context) {
		width = float64(da.GetAllocatedWidth())
		height = float64(da.GetAllocatedHeight())
		moveStep()
		if caughtEnds == 2 {
			moveStep() // double speed after caught
			if !wasCaught {
				win.SetTitle("Caught TKix!")
				wasCaught = true
			}
		} else {
			if wasCaught {
				wasCaught = false
				level += 1.0
				win.SetTitle(fmt.Sprint("TKix released! ", level))
				randomizeStick()
				for j := 0; j < Ntrail; j++ {
					trail[j][0] = x0
					trail[j][1] = y0
					trail[j][2] = x1
					trail[j][3] = y1
				}
			}
		}
		cr.SetSourceRGB(0, 0, 0)
		cr.Rectangle(x, y, unitSize, unitSize)
		cr.Fill()
		for j := 0; j < Ntrail; j++ {
			t := trail[ (j+itrail)%Ntrail ]
			r, g, b := HueToRgb(t[4])
			cr.SetSourceRGB(r, g, b)
			cr.MoveTo(t[0], t[1])
			cr.LineTo(t[2], t[3])
			cr.Stroke()
		}
	}
	da.Connect("draw", draw)
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}
		if move, found := keyMap[keyEvent.KeyVal()]; found {
			move()
			win.QueueDraw()
		}
	})
	glib.TimeoutAdd(100, func() bool {
		win.QueueDraw()
		return true
	})

	gtk.Main()
}
