package main

import (
	"image/color"
	"machine"
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
	//colorId := uint16(rand.Uint32() >> 8)
	//rgba := currentPalette.ColorAt(colorId)
	l.Color = ledsgo.White
	if l.SinceBlink == currentIteration {
		l.Color = ledsgo.Black
		WriteStringSerial("Blinking ", l.Identifier, " iterCount: ", strconv.Itoa(l.SinceBlink), " currentIter: ", strconv.Itoa(currentIteration))
	}

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
		arr = SetLeds(arr, let.Offset, let.Count, let.Color, let.Identifier)
	}
	return arr
}

func IsBlack(ledColor color.RGBA) bool {
	return ledColor.B == 0 && ledColor.G == 0 && ledColor.R == 0
}

func SetLeds(arr []color.RGBA, offset, count int, ledColor color.RGBA, letName string) []color.RGBA {
	if offset > len(arr) {
		for i := len(arr); i < offset; i++ {
			arr = append(arr, color.RGBA{R: 0, G: 0, B: 0, A: 0})
		}
	}
	c := ledColor

	if !IsBlack(c) {
		// Get a noise value!
		hue := ledsgo.Noise2(uint32(time.Now().UnixNano()>>22), uint32(byte(letName[0]))*getRandomUint32())
		c = ledsgo.ApplyAlpha(ledsgo.Color{hue, 0xff, 0xff}.Spectrum(), ALPHA)
	}
	// WriteStringSerial(colorString(c))

	for i := 0; i < count; i++ {
		arr = append(arr, c)
	}
	return arr
}

func getRandomUint32() uint32 {
	rand, err := machine.GetRNG()
	if err != nil {
		WriteStringSerial(err.Error())
	}
	return rand
}

func randIntBetween(min,max int) int {
	rand := int(getRandomUint32())
	if rand < min || rand > max {
		return randIntBetween(min, max)
	}
	return rand
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
	min := 250 * int(time.Millisecond)
	max := 1000 * int(time.Millisecond)
	rand := randIntBetween(min,max)
	ticker := time.NewTicker(time.Duration(rand))
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

func main() {
	machine.InitSerial()
	InitLetters(allTheLetters)
	min := 250 * int(time.Millisecond)
	max := 1000 * int(time.Millisecond)
	rand := randIntBetween(min,max)
	WriteStringSerial(strconv.Itoa(int(rand)))

	iterTicker := time.NewTicker(time.Duration(rand))
	defer iterTicker.Stop()

	for {
		select {
		case <-iterTicker.C:
			incrementCurrentIteration()
		}
	}
}

var allTheLetters = map[machine.Pin][]Letter{
	machine.GP0: {
		{
			Identifier: "G",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      15,
			SinceBlink: 1,
		},
		{
			Identifier: "C",
			Color:      ledsgo.Black,
			Offset:     15,
			Count:      11,
			SinceBlink: 5,
		},
	},
	machine.GP2: {
		{
			Identifier: "L",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      8,
			SinceBlink: 2,
		},
		{
			Identifier: "H",
			Color:      ledsgo.Black,
			Offset:     8,
			Count:      13,
			SinceBlink: 6,
		},
	},
	machine.GP4: {
		{
			Identifier: "O",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      14,
			SinceBlink: 3,
		},
		{
			Identifier: "I",
			Color:      ledsgo.Black,
			Offset:     14,
			Count:      10,
			SinceBlink: 7,
		},
	},
	machine.GP6: {
		{
			Identifier: "W",
			Color:      ledsgo.Black,
			Offset:     0,
			Count:      14,
			SinceBlink: 4,
		},
		{
			Identifier: "C2",
			Color:      ledsgo.Black,
			Offset:     14,
			Count:      11,
			SinceBlink: 8,
		},
	},
}
