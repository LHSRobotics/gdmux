package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/tarm/goserial"
)

type armMsg struct {
	X, Y, Z float64
}

func armReader(c chan string, r io.Reader) {
	buf := make([]byte, 255)
	for {
		n, err := r.Read(buf)
		if err != nil {
			log.Println("error reading ack from arm: ", err)
		}
		c <- strings.TrimSpace(string(buf[:n]))
	}
}

// armCtl handles communication with the Staubli arm for us. For each move, we output the coordinates
// separated by spaces. This is easy to parse in V+ using READ.
func armCtl() {
	c := &serial.Config{Name: *armPort, Baud: 38400}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	feedback := make(chan string)
	go armReader(feedback, s)
	//fmt.Printf("Staubli says: %s", <-feedback)

	for {
		msg := <-armc
		fmt.Printf("%8.2f %8.2f %8.2f", msg.X, msg.Y, msg.Z)
		_, err = fmt.Fprintf(s, "%f %f %f\r\n", msg.X, msg.Y, msg.Z)
		if err != nil {
			log.Println("error sending coordinates to arm: ", err)
		}
		fmt.Printf("    → %s\n", <-feedback)
	}
}
