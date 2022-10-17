package main

import (
	"image/color"
	"machine"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aykevl/ledsgo"
	"tinygo.org/x/drivers/ws2812"
)

var ALPHA = uint8(64)

func microcontrollerTimeString() string {
	return time.Now().Format("15:04:05.00000")

}

func WriteStringSerial(args ...string) (int, error) {
	leaderCount, err := machine.Serial.Write([]byte(microcontrollerTimeString() + ": "))
	total := leaderCount
	if err != nil {
		return total, err
	}
	for _, arg := range args {
		count, err := machine.Serial.Write([]byte(arg))
		total += count
		if err != nil {
			return total, err
		}
	}
	footerCount, err := machine.Serial.Write([]byte("\n"))
	total += footerCount
	return total, err
}

var currentIteration = 0
var maxIteration = 24

func (l *Letter) UpdateColors() {
	colorId := uint16(rand.Uint32()  >> 8)
	rgba := currentPalette.ColorAt(colorId)
	rgba = ledsgo.ApplyAlpha(rgba, ALPHA)
	if l.SinceBlink == currentIteration {
		rgba = ledsgo.Black
		WriteStringSerial("Blinking ", l.Identifier, " iterCount: ", strconv.Itoa(l.SinceBlink), " currentIter: ", strconv.Itoa(currentIteration))
	}

	l.Color = rgba

	err := l.Column.Device.WriteColors(LettersColors(l.Column.Letters)[:])
	if err != nil {
		WriteStringSerial("failed to write colors: ", err.Error())
	}
}

func incrementCurrentIteration() {
	currentIteration++
	if currentIteration > maxIteration {
		currentIteration = 0
	}
}

func printableLeds(sliceC []color.RGBA) string {
	output := "["
	for _, c := range sliceC {
		output += colorString(c) + " "
	}
	output += "]"
	return output
}

func colorString(rgba color.RGBA) string {
	return strings.Join(
		[]string{
			partialColorString("R", int(rgba.R)),
			partialColorString("G", int(rgba.G)),
			partialColorString("B", int(rgba.B)),
		}, " ")
}

func partialColorString(name string, val int) string {
	return strings.Join([]string{name, strconv.Itoa(val)}, ": ")
}

func LettersColors(lets []*Letter) []color.RGBA {
	var arr []color.RGBA
	for _, let := range lets {
		arr = SetLeds(arr, let.Offset, let.Count, let.Color)
	}
	return arr
}

func SetLeds(arr []color.RGBA, offset, count int, ledColor color.RGBA) []color.RGBA {
	if offset > len(arr) {
		for i := len(arr); i < offset; i++ {
			arr = append(arr, color.RGBA{R: 0, G: 0, B: 0, A: 0})
		}
	}
	for i := 0; i < count; i++ {
		c := ledColor
		if USE_NOISE{
			hue := ledsgo.Noise2(uint32(time.Now().UnixNano() >> 22) , uint32(i))
			c = ledsgo.ApplyAlpha(ledsgo.Color{hue, 0xff, 0xff}.Spectrum(), ALPHA)
		}
		arr = append(arr, c)
	}
	return arr
}

// InitLetters will create one column for each pin supplied, and set them to output
func InitLetters(columnsMap map[machine.Pin][]Letter) {
	for pin, letters := range columnsMap {
		thePin := pin
		thePin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		theWS := ws2812.New(thePin)
		col := &Column{Device: theWS}
		for l := range letters {
			let := letters[l]
			let.Column = col
			col.Letters = append(col.Letters, &let)
			RunTicker(&let)
		}
	}
}

func RunTicker(l *Letter) {
	ticker := time.NewTicker(l.WaitTime)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.UpdateColors()
			}
		}
	}()
}

type Column struct {
	Letters []*Letter
	Device  ws2812.Device
}

type Letter struct {
	Identifier string
	Color      color.RGBA
	Offset     int
	Count      int
	SinceBlink int
	Column     *Column
	WaitTime   time.Duration
}

func getRand() uint8 {
	n := rand.Int()
	if n > 255 {
		return uint8(n % 255)
	}
	return uint8(n)

}

func toggleUseNoise() {
	USE_NOISE = !USE_NOISE
}

func main() {
	machine.InitSerial()
	InitLetters(allTheLetters)

	currentPalette = palettes[0]
	next := 1
	iterTicker := time.NewTicker(time.Millisecond * 250)
	defer iterTicker.Stop()
	paletteTicker := time.NewTicker(time.Second * 60)
	defer paletteTicker.Stop()

	noiseTicker := time.NewTicker(time.Second * 120)
	defer noiseTicker.Stop()

	for {
		select {
		case <-iterTicker.C:
			incrementCurrentIteration()
		case <-paletteTicker.C:
			currentPalette = palettes[next]
			next++
			if next >= len(palettes) {
				next = 0
			}

		case <-noiseTicker.C:
			toggleUseNoise()

		}
	}
}

var USE_NOISE = false

var currentPalette ledsgo.Palette16

var colors = ledsgo.Palette16{
	ledsgo.AliceBlue,
	ledsgo.Blue,  // blue
	ledsgo.Red,   // red
	ledsgo.Azure, // purple
	ledsgo.Chartreuse,
	ledsgo.Black,
	ledsgo.AntiqueWhite, // Light white?
	ledsgo.Amethyst,
	ledsgo.AliceBlue,
	ledsgo.Blue,  // blue
	ledsgo.Red,   // red
	ledsgo.Azure, // purple
	ledsgo.Chartreuse,
	ledsgo.Black,
	ledsgo.AntiqueWhite, // Light white?
	ledsgo.Amethyst,     // purple?
}

var bluePalette = ledsgo.Palette16{
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.Black,
	ledsgo.Black,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.DarkBlue,
	ledsgo.Black,
}

var palettes = []ledsgo.Palette16{
	colors,
	bluePalette,
	ledsgo.LavaColors,
	ledsgo.CloudColors,
	ledsgo.ForestColors,
	ledsgo.HeatColors,
}

var allTheLetters = map[machine.Pin][]Letter{
	machine.GP0: {
		{
			Identifier: "G",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      14,
			SinceBlink: 1,
			WaitTime:   time.Millisecond * 250,
		},
		{
			Identifier: "C",
			Color:      ledsgo.Black,
			Offset:     14,
			Count:      14,
			SinceBlink: 5,
			WaitTime:   time.Millisecond * 255,
		},
	},
	machine.GP2: {
		{
			Identifier: "L",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      8,
			SinceBlink: 2,
			WaitTime:   time.Millisecond * 240,
		},
		{
			Identifier: "H",
			Color:      ledsgo.Black,
			Offset:     8,
			Count:      13,
			SinceBlink: 6,
			WaitTime:   time.Millisecond * 245,
		},
	},
	machine.GP4: {
		{
			Identifier: "O",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      14,
			SinceBlink: 3,
			WaitTime:   time.Millisecond * 230,
		},
		{
			Identifier: "I",
			Color:      ledsgo.Black,
			Offset:     14,
			Count:      10,
			SinceBlink: 7,
			WaitTime:   time.Millisecond * 235,
		},
	},
	machine.GP6: {
		{
			Identifier: "W",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      14,
			SinceBlink: 4,
			WaitTime:   time.Millisecond * 260,
		},
		{
			Identifier: "C2",
			Color:      ledsgo.Black,
			Offset:     14,
			Count:      11,
			SinceBlink: 8,
			WaitTime:   time.Millisecond * 265,
		},
	},
}
