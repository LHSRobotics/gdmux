// Package staubli provides an interface to control the London Hackspace's Staübli arm.
package staubli

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

type Arm interface {
	Move(x, y, z float64) error
	MoveStraight(x, y, z float64) error
	ArcCenter(x, y, z, i, j, k, direction float64) error
}

type Staubli struct {
	rw io.ReadWriter
	sync.Mutex
	buf []byte
}

// Move the arm to the point (x,y,z), without guaranteeing a staight line.
func (s *Staubli) Move(x, y, z float64) error {
	// we probably need a lock here...
	_, err := fmt.Fprintf(s.rw, "0 %.3f %.3f %.3f\r\n", x, y, z)
	if err != nil {
		return fmt.Errorf("error sending coordinates to arm: %s", err)
	}

	if r := s.readReply(); r != "OK" {
		return fmt.Errorf("error from arm: %s", r)
	}
	return nil
}

// Move the arm to the point (x,y,z) in a straight line.
func (s *Staubli) MoveStraight(x, y, z float64) error {
	_, err := fmt.Fprintf(s.rw, "1 %.3f %.3f %.3f\r\n", x, y, z)
	if err != nil {
		return fmt.Errorf("error sending coordinates to arm: %s", err)
	}

	if r := s.readReply(); r != "OK" {
		return fmt.Errorf("error from arm: %s", r)
	}
	return nil
}

const (
	Clockwise     = 1
	Anticlockwise = -1
)

// Move the arm to the point (x,y,z) following the path of an arc whose centre is at (i,j,k).
//
// The distance between the current position and (i,j,k) must equal that between (x,y,z) and (i,j,k).
// This method is likely to be removed.
func (s *Staubli) ArcCenter(x, y, z, i, j, k, direction float64) error {
	_, err := fmt.Fprintf(s.rw, "2 %.3f %.3f %.3f %.3f %.3f %.3f %.3f\r\n", x, y, z, i, j, k, direction)
	if err != nil {
		return fmt.Errorf("error sending coordinates to arm: %s", err)
	}

	if r := s.readReply(); r != "OK" {
		return fmt.Errorf("error from arm: %s", r)
	}
	return nil
}

func (s *Staubli) readReply() string {
	n, err := s.rw.Read(s.buf)
	if err != nil {
		log.Println("error reading ack from arm: ", err)
	}
	return strings.TrimSpace(string(s.buf[:n]))
}

func NewStaubli(rw io.ReadWriter) *Staubli {
	a := &Staubli{
		rw:  rw,
		buf: make([]byte, 255),
	}

	return a
}
