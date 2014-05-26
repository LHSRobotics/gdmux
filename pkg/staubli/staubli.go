// Package staubli provides an interface to control the London Hackspace's Stäubli arm.
package staubli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"sync"
)

type Arm interface {
	Move(x, y, z float64) error
	MoveStraight(x, y, z float64) error
	ArcCenter(x, y, z, i, j, k, direction float64) error
	Break() error
}

type point struct {
	x, y, z, yaw, pitch, roll float64
}

type Staubli struct {
	rw io.ReadWriter
	sync.Mutex
	cur    point
	reader *bufio.Reader
}

// Move the arm to the point (x,y,z), without guaranteeing a staight line.
func (s *Staubli) Move(x, y, z float64) error {
	// we probably need a lock here...
	_, err := fmt.Fprintf(s.rw, "0 %.3f %.3f %.3f\r\n", x, y, z)
	if err != nil {
		return fmt.Errorf("error sending coordinates to arm: %s", err)
	}

	if r := s.readReply(); !strings.HasPrefix(r, "OK") {
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

	if r := s.readReply(); !strings.HasPrefix(r, "OK") {
		return fmt.Errorf("error from arm: %s", r)
	}
	return nil
}

// Wait until the arm reaches its current destination.
//
// Currently, this also updates the local state to the arm's latest coordinates.
func (s *Staubli) Break() error {
	_, err := fmt.Fprintf(s.rw, "2\r\n")
	if err != nil {
		return fmt.Errorf("error sending command to arm: %s", err)
	}

	r := s.readReply()
	if !strings.HasPrefix(r, "OK") {
		return fmt.Errorf("error from arm: %s", r)
	}

	var x, y, z float64
	_, err = fmt.Sscan(r[2:], &x, &y, &z)
	if err != nil {
		return fmt.Errorf("error parsing reply from arm: %s", err)
	}

	s.cur.x, s.cur.y, s.cur.z = x, y, z

	return nil
}

// Move the arm to the point (x,y,z) in a straight line, using its current position as origin.
func (s *Staubli) MoveRel(x, y, z float64) error {
	_, err := fmt.Fprintf(s.rw, "3 %.3f %.3f %.3f\r\n", x, y, z)
	if err != nil {
		return fmt.Errorf("error sending coordinates to arm: %s", err)
	}

	if r := s.readReply(); !strings.HasPrefix(r, "OK") {
		return fmt.Errorf("error from arm: %s", r)
	}
	return nil
}

const (
	Clockwise     = -1
	Anticlockwise = 1
)

// Move the arm to the point (x,y,z) following the path of an arc whose centre is at (i,j,k).
//
// The distance between the current position and (i,j,k) must equal that between (x,y,z) and (i,j,k).
func (s *Staubli) ArcCenter(x, y, z, i, j, k, direction float64) error {
	// TODO rewrite this nicer. This was copypaster'd from the V+ code. It can be a lot nicer.
	i += s.cur.x
	j += s.cur.y
	k += s.cur.z

	startAngle := math.Atan2(s.cur.y-j, s.cur.x-i)
	endAngle := math.Atan2(y-j, x-i)

	rX := (s.cur.x - i)
	rY := (s.cur.y - j)
	radius := math.Sqrt(rX*rX + rY*rY)

	arcLen := math.Abs((endAngle - startAngle) / direction)
	zStep := (z - s.cur.z) / arcLen

	for a := 0.0; a < arcLen; a++ {
		angle := direction*a + startAngle

		x = radius*math.Cos(angle) + i
		y = radius*math.Sin(angle) + j
		z = s.cur.z + a*zStep

		err := s.MoveStraight(x, y, z)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Staubli) readReply() string {
	line, err := s.reader.ReadString('\n')
	if err != nil {
		log.Println("error reading ack from arm: ", err)
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return s.readReply()
	}
	return line
}

func NewStaubli(rw io.ReadWriter) *Staubli {
	a := &Staubli{
		rw:     rw,
		reader: bufio.NewReader(rw),
	}

	return a
}
