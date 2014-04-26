package main

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/LHSRobotics/gdmux/gcode"
)

type Cmd struct {
	x, y, z, e, f float64 // TODO perhaps this should be a map? Also, perhaps keep it as string?
	ops           []func(c *Cmd)
	inches        bool
	line          *gcode.Line
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
	f, err := strconv.ParseFloat(string(code[1:]), 32)
	if err != nil {
		// TODO return this error instead of panicing
		log.Fatal("couldn't parse float value")
	}

	switch code[0] {
	case 'X':
		c.x = f
	case 'Y':
		c.y = f
	case 'Z':
		c.z = f
	case 'E':
		c.e = f
	case 'F':
		c.f = f
	default:
		log.Printf("unknown class: %v", c) // should return an error here instead
	}
}

// AddOp parses and adds an G- or M-code to the operation queue.
func (c *Cmd) AddOp(code gcode.Code) {
	switch code {
	case "G0":
		c.ops = append(c.ops, func(c *Cmd) { arm.Move(c.x, c.y, c.z) })
	case "G1":
		c.ops = append(c.ops, func(c *Cmd) { arm.MoveStraight(c.x, c.y, c.z) })
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

func dmux(read io.Reader, stop chan bool) {
	r := gcode.NewParser(read)
	cmd := Cmd{}
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
			case 'X', 'Y', 'Z', 'E', 'F':
				cmd.SetVar(c)
			default:
				log.Printf("unknown code class: %v (%v)", c, cmd.line)
			}
		}
		// TODO handle pausing as well
		if !running {
			return
		}
		weblog(fmt.Sprintf("Executing line %d: %s\n", n, cmd.line.Text))
		cmd.Exec()
		cmd.ops = cmd.ops[:0]
		n++
	}
}
