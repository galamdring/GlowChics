package main

import (
	// "github.com/aykevl/ledsgo"
	"image/color"
	"image/color/palette"
	"machine"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"tinygo.org/x/drivers/ws2812"
)

var brightnessAdjustPercentage uint32 = 20

func getRGBA(rgb color.Color) color.RGBA {
	R, G, B, _ := rgb.RGBA()
	adR := uint8((R * brightnessAdjustPercentage) / 100)
	adG := uint8((G * brightnessAdjustPercentage) / 100)
	adB := uint8((B * brightnessAdjustPercentage) / 100)
	return color.RGBA{
		R: adR,
		G: adG,
		B: adB,
		A: 0,
	}
}

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
	colorId := rand.Int() % len(currentPalette)
	ourColor := currentPalette[colorId]
	rgba := getRGBA(ourColor)
	if l.SinceBlink == currentIteration {
		rgba = getRGBA(ledsgo.Black)
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
		arr = SetLeds(arr, let.Offset, let.Count, getRGBA(let.Color))
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
		arr = append(arr, ledColor)
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
		for l, _ := range letters {
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
	Color      color.Color
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

func main() {
	machine.InitSerial()
	InitLetters(allTheLetters)

	currentPalette = palettes[0]
	next := 1
	iterTicker := time.NewTicker(time.Millisecond * 250)
	defer iterTicker.Stop()
	paletteTicker := time.NewTicker(time.Second * 60)
	defer paletteTicker.Stop()

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
		}
	}
}

func getRandomColors() color.Palette {
	var lotsOfColors color.Palette
	_ = palette.WebSafe
	for range make([]struct{}, 254) {
		R := getRand()
		G := getRand()
		B := getRand()
		lotsOfColors = append(lotsOfColors,
			color.RGBA{
				R: R / 2,
				G: G / 2,
				B: B / 2,
			})
	}
	return lotsOfColors
}

var currentPalette []color.Color
var notSoDamnBright = uint8(0x6f)
var blinkingYellowPalette = color.Palette{
	color.RGBA{R: notSoDamnBright, G: notSoDamnBright, B: 0, A: 0},
	// color.RGBA{R: 0, G: 0, B: 0, A: 0},
	color.RGBA{R: notSoDamnBright, G: notSoDamnBright, B: 0, A: 0},
	// color.RGBA{R: 0, G: 0, B: 0, A: 0},
	color.RGBA{R: notSoDamnBright, G: notSoDamnBright, B: 0, A: 0},
	// color.RGBA{R: 0, G: 0, B: 0, A: 0},
	color.RGBA{R: notSoDamnBright, G: notSoDamnBright, B: 0, A: 0},
	color.RGBA{R: 0, G: 0, B: 0, A: 0},
}

var colors = []color.Color{
	color.White,
	color.RGBA{R: 0, G: 0, B: 0xaa},    // blue
	color.RGBA{R: 0xaa, G: 0, B: 0},    // red
	color.RGBA{R: 0xaa, G: 0xaa, B: 0}, // purple
	color.RGBA{R: 0xaa, G: 0, B: 0xaa}, // yellow
	color.Black,
	color.RGBA{R: 0x66, G: 0x66, B: 0x66}, // Light white?
	color.RGBA{R: 0x55, G: 0, B: 0x75},    // purple?
}

var bluePalette = color.Palette{
	color.RGBA{R: 36, G: 2, B: 230, A: 0},
	color.RGBA{R: 36, G: 2, B: 230, A: 0},
	color.RGBA{R: 36, G: 2, B: 230, A: 0},
	color.RGBA{R: 36, G: 2, B: 230, A: 0},
	color.RGBA{R: 36, G: 2, B: 230, A: 0},
	color.Black,
	color.Black,
}

var dimPalette = color.Palette{
	color.RGBA{R: 1, G: 2, B: 3, A: 255},
	color.RGBA{R: 1, G: 2, B: 3, A: 150},
	color.RGBA{R: 1, G: 2, B: 3, A: 50},
	color.RGBA{R: 1, G: 2, B: 3, A: 1},
	color.RGBA{R: 1, G: 2, B: 3, A: 1},
	color.Black,
	color.Black,
}

var palettes = []color.Palette{
	// colors,
	// palette.Plan9,
	// blinkingYellowPalette,
	// palette.WebSafe,
	// bluePalette,
	dimPalette,
}

var allTheLetters = map[machine.Pin][]Letter{
	machine.GP0: {
		{
			Identifier: "G",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     0,
			Count:      14,
			SinceBlink: 1,
			WaitTime:   time.Millisecond * 250,
		},
		{
			Identifier: "C",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     14,
			Count:      14,
			SinceBlink: 5,
			WaitTime:   time.Millisecond * 255,
		},
	},
	machine.GP2: {
		{
			Identifier: "L",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     0,
			Count:      8,
			SinceBlink: 2,
			WaitTime:   time.Millisecond * 240,
		},
		{
			Identifier: "H",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     8,
			Count:      13,
			SinceBlink: 6,
			WaitTime:   time.Millisecond * 245,
		},
	},
	machine.GP4: {
		{
			Identifier: "O",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     0,
			Count:      14,
			SinceBlink: 3,
			WaitTime:   time.Millisecond * 230,
		},
		{
			Identifier: "I",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     14,
			Count:      10,
			SinceBlink: 7,
			WaitTime:   time.Millisecond * 235,
		},
	},
	machine.GP6: {
		{
			Identifier: "W",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     0,
			Count:      14,
			SinceBlink: 4,
			WaitTime:   time.Millisecond * 260,
		},
		{
			Identifier: "C2",
			Color:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
			Offset:     14,
			Count:      11,
			SinceBlink: 8,
			WaitTime:   time.Millisecond * 265,
		},
	},
}
