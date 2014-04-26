package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/tarm/goserial"
)

type Arm interface {
	Move(x, y, z float64)
	MoveStraight(x, y, z float64)
}

type Staubli struct {
	rw io.ReadWriter
	sync.Mutex
	buf []byte
}

func (s *Staubli) Move(x, y, z float64) {
	if *verbose {
		log.Printf("moving arm to %f, %f, %f", x, y, z)
	}

	weblog(fmt.Sprintf("%8.2f %8.2f %8.2f", x, y, z))

	// we probably need a lock here...
	_, err := fmt.Fprintf(s.rw, "%f %f %f\r\n", x, y, z)
	if err != nil {
		log.Println("error sending coordinates to arm: ", err)
	}
	weblog(fmt.Sprintf(" → %s\n", s.readReply()))
}

func (s *Staubli) MoveStraight(x, y, z float64) {
	if *verbose {
		log.Printf("straight moving arm to %f, %f, %f", x, y, z)
	}

	weblog(fmt.Sprintf("%8.2f %8.2f %8.2f", x, y, z))

	_, err := fmt.Fprintf(s.rw, "%f %f %f\r\n", x, y, z)
	if err != nil {
		log.Println("error sending coordinates to arm: ", err)
	}
	weblog(fmt.Sprintf(" → %s\n", s.readReply()))
}

func (s *Staubli) readReply() string {
	n, err := s.rw.Read(s.buf)
	if err != nil {
		log.Println("error reading ack from arm: ", err)
	}
	return strings.TrimSpace(string(s.buf[:n]))
}

func NewStaubli(serialPort string) *Staubli {
	s, err := serial.OpenPort(&serial.Config{Name: serialPort, Baud: 38400})
	if err != nil {
		log.Fatal(err)
	}

	a := &Staubli{
		rw:  s,
		buf: make([]byte, 255),
	}
	
	log.Println("first", a.readReply())
	return a
}
