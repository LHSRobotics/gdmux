// Package staubli provides an interface to control the London Hackspace's Staübli arm.
package staubli

import (
	"fmt"
	"log"
)

type dummy struct{}

var Dummy = &dummy{}

func dummyMove(x, y, z float64) error {
	// We just make up some bounding box to return some errors
	if x > 200 || x < -200 ||
		y > 200 || y < -200 ||
		z > 200 || z < -200 {
		return fmt.Errorf("out of range")
	}
	log.Printf("dummy move!")
	return nil
}

func (s *dummy) Move(x, y, z float64) error {
	return dummyMove(x, y, z)
}

func (s *dummy) MoveStraight(x, y, z float64) error {
	return dummyMove(x, y, z)
}

// Move the arm to the point (x,y,z) following the path of an arc whose centre is at (i,j,k).
func (s *dummy) ArcCenter(x, y, z, i,j,k float64) error {
	return dummyMove(x, y, z)
}
