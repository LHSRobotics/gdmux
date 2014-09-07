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

func (s *dummy) Move6DOF(x, y, z, yaw, pitch, roll float64) error {
	return dummyMove(x, y, z)
}

func (s *dummy) Move(x, y, z float64) error {
	return dummyMove(x, y, z)
}

func (s *dummy) MoveStraight(x, y, z float64) error {
	return dummyMove(x, y, z)
}

func (s *dummy) ArcCenter(x, y, z, i, j, k, direction float64) error {
	return dummyMove(x, y, z)
}

func (s *dummy) Break() error {
	return nil
}
