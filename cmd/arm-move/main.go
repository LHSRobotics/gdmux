// Command arm-move moves the arm relative to its current position
package main

import (
	"flag"
	"log"

	"github.com/tarm/goserial"

	"github.com/LHSRobotics/gdmux/pkg/staubli"
)

var (
	// armPort is the serial file connected to the arm controller's data line. For the Staubli
	// its baudrate 19200, we assume that's already set for the device file. (I.e. with stty.)
	ttyData  = flag.String("datatty", "/dev/ttyS0", "serial tty to the Staubli data line")
	baudData = flag.Int("datarate", 19200, "baud rate for the staubli's data line")

	x = flag.Float64("x", 0, "x coordinate")
	y = flag.Float64("y", 0, "y coordinate")
	z = flag.Float64("z", 0, "z coordinate")
)

func main() {
	flag.Parse()

	log.Println("Opening ", *ttyData)
	s, err := serial.OpenPort(&serial.Config{Name: *ttyData, Baud: *baudData})
	if err != nil {
		log.Fatal(err)
	}
	arm := staubli.NewStaubli(s)

	if err := arm.MoveRel(*x, *y, *z); err != nil {
		log.Fatal(" %s", err)
	}
}
