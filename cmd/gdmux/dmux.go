package main

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/LHSRobotics/gdmux/pkg/gcode"
	"github.com/LHSRobotics/gdmux/pkg/staubli"
)

type point struct {
	x, y, z, a, b, c float64
}

type Cmd struct {
	env    map[byte]float64
	zero   point
	ops    []func(c *Cmd)
	inches bool
	line   *gcode.Line
}

func (c *Cmd) Exec() {
	if *verbose {
		log.Printf("executing line %v", c.line)
	}

	for _, op := range c.ops {
		op(c)
	}
}

// SetVar parses a variable-setting code, such as X, Y, or E.
func (c *Cmd) SetVar(code gcode.Code) {
	value, err := strconv.ParseFloat(string(code[1:]), 32)
	if err != nil {
		// TODO return this error instead of panicing
		log.Fatal("couldn't parse float value")
	}

	c.env[code[0]] = value
}

// AddOp parses and adds an G- or M-code to the operation queue.
func (c *Cmd) AddOp(code gcode.Code) {
	switch code {
	case "G0":
		// TODO(s): I don't like how this is done, need to rethink this package...
		c.ops = append(c.ops, func(c *Cmd) {
			weblog(fmt.Sprintf("Move %8.2f %8.2f %8.2f", c.env['X'], c.env['Y'], c.env['Z']))
			err := arm.Move(c.env['X']+c.zero.x, c.env['Y']+c.zero.y, c.env['Z']+c.zero.z)
			if err != nil {
				weblog(fmt.Sprintf(" → %s\n", err))
				return
			}
			weblog(" → OK\n")
		})
	case "G1":
		c.ops = append(c.ops, func(c *Cmd) {
			weblog(fmt.Sprintf("Line %8.2f %8.2f %8.2f", c.env['X'], c.env['Y'], c.env['Z']))
			err := arm.MoveStraight(c.env['X']+c.zero.x, c.env['Y']+c.zero.y, c.env['Z']+c.zero.z)
			if err != nil {
				weblog(fmt.Sprintf(" → %s\n", err))
				return
			}
			weblog(" → OK\n")
		})
	case "G2":
		// Follow a clockwise arc.
		//
		// For now we only support the 'centre format arc'. This format gives us target coordinates
		// and the coordinates of the centre of the circle whose arc we're following.
		// It's not great but it's what all the slicers spit out.
		//
		// The other format is 'radius format arc' and that gives us target coordinates and a radius.
		// It's probably worth supporting that at some point.
		c.ops = append(c.ops, func(c *Cmd) {
			weblog(fmt.Sprintf("Clockwise Arc to %8.2f %8.2f %8.2f, around %8.2f %8.2f %8.2f", c.env['X'], c.env['Y'], c.env['Z'], c.env['I'], c.env['J'], c.env['K']))
			// TODO add a step argument here and use negative to go anti-clockwise.
			err := arm.ArcCenter(c.env['X']+c.zero.x, c.env['Y']+c.zero.y, c.env['Z']+c.zero.z,
				c.env['I']+c.zero.x, c.env['J']+c.zero.y, c.env['K']+c.zero.z, staubli.Clockwise)
			if err != nil {
				weblog(fmt.Sprintf(" → %s\n", err))
				return
			}
			weblog(" → OK\n")
		})
	case "G3":
		// Follow an anti-clockwise arc.
		c.ops = append(c.ops, func(c *Cmd) {
			weblog(fmt.Sprintf("Anti-clockwise Arc to %8.2f %8.2f %8.2f, around %8.2f %8.2f %8.2f", c.env['X'], c.env['Y'], c.env['Z'], c.env['I'], c.env['J'], c.env['K']))
			// TODO add a step argument here and use negative to go anti-clockwise.
			err := arm.ArcCenter(c.env['X']+c.zero.x, c.env['Y']+c.zero.y, c.env['Z']+c.zero.z,
				c.env['I']+c.zero.x, c.env['J']+c.zero.y, c.env['K']+c.zero.z, staubli.Anticlockwise)
			if err != nil {
				weblog(fmt.Sprintf(" → %s\n", err))
				return
			}
			weblog(" → OK\n")
		})
	case "G21":
		c.inches = false
	case "M107":
		log.Printf("ignoring: fanoff M107.")
	case "M103":
	case "M101":
	default:
		log.Printf("unknown code: %v", code) // should return an error here instead
	}
}

func dmux(read io.Reader) {
	r := gcode.NewParser(read)
	cmd := Cmd{
		env:  make(map[byte]float64),
		zero: point{x: 500, y: 0, z: -100},
	}
	n := 1
	for {
		var err error
		cmd.line, err = r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			// TODO probably better to pause on errors
			log.Println("parse error: %v", err)
			continue
		}

		for _, c := range cmd.line.Codes {
			switch c[0] {
			case 'G', 'M':
				cmd.AddOp(c)
			case 'X', 'Y', 'Z', 'E', 'F', 'I', 'J', 'K':
				cmd.SetVar(c)
			default:
				log.Printf("unknown code class: %v (%v)", c, cmd.line)
			}
		}
		// TODO handle pausing as well
		if !running {
			return
		}
		cmd.Exec()
		cmd.ops = cmd.ops[:0]
		n++
	}
}
