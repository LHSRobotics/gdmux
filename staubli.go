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
	f, err := os.OpenFile(*armPort, os.O_RDWR, 0)
	if err != nil {
		log.Fatal("couldn't open Arm file", err)
	}
	buf := make([]byte, 100)
	
	for {
		msg := <-armc
		fmt.Printf("Sending: %f %f %f\n", msg.X, msg.Y, msg.Z)
		_, err = fmt.Fprintf(f, "%f %f %f\n", msg.X, msg.Y, msg.Z)
		if err != nil {
			log.Println("error sending coordinates to arm: ", err)
		}
		_, err =  f.Read(buf)
		if err != nil {
			log.Println("error reading ack from arm: ", err)
		}
		fmt.Printf("Got: %v\n", string(buf))
	}
}
