package main

import (
	"os"
	"log"
	"fmt"
)

type armMsg struct {
	X, Y, Z float64
}

// armCtl handles communication with the Staubli arm for us. For each move, we output the coordinates
// separated by spaces. This is easy to parse in V+ using READ.
func armCtl() {
	f, err := os.Open(*armPort)
	if err != nil {
		log.Fatal("couldn't open Arm file", err)
	}
	
	for {
		msg := <-armc
		fmt.Fprintf(f, "%f %f %f\r\n", msg.X, msg.Y, msg.Z)
	}
}
