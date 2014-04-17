package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

var (
	// armPort is the serial file connected to the arm controller's data line. For the Staubli
	// its baudrate 38400, we assume that's already set for the device file. (I.e. with stty.)
	armPort      = flag.String("arm", "/dev/ttyUSB0", "serial file to talk to the staubli's console")
	extruderPort = flag.String("extruder", "/dev/ttyS1", "serial file to talk to the extruder's firmware")
	addr         = flag.String("addr", "127.0.0.1:5000", "tcp address on which to listen")
	stdin        = flag.Bool("stdin", false, "read a gcode file from stdin")
	verbose      = flag.Bool("verbose", false, "print lots output")

	armc = make(chan armMsg)
)

type Cmd struct {
	// TODO these variables are stateful. Maybe we should add a way to tell if they've
	// been set in this line or inhereted from the previous line?

	x, y, z, e, f float64 // TODO perhaps this should be a map? Also, perhaps keep it as string?
	ops           []func()
	inches        bool
	line          *Line
}

func (c *Cmd) move() {
	if *verbose {
		log.Println("moving arm")
	}

	armc <- armMsg{c.x, c.y, c.z}
}

func (c *Cmd) Exec() {
	if *verbose {
		log.Printf("executing line %v", c.line)
	}

	for _, op := range c.ops {
		op()
	}
}

// SetVar parses a variable-setting code, such as X, Y, or E.
func (c *Cmd) SetVar(code Code) {
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
func (c *Cmd) AddOp(code Code) {
	switch code {
	case "G1":
		c.ops = append(c.ops, c.move)
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
	r := NewParser(read)
	cmd := Cmd{}
	for {
		var err error
		cmd.line, err = r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("parse error: %v", err)
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
		cmd.Exec()
		cmd.ops = cmd.ops[:0]
	}
}

func listen() {
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("couldn't listen on socket:", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("couldn't accept connection:", err)
			continue
		}
		log.Println("accepted connection:", err)
		go dmux(conn)
	}
}

func main() {
	flag.Parse()

	go armCtl() // Launch the arm controlling goroutine. We talk to this using armc.

	if !*stdin {
		listen()
	}

	log.Println("reading from stdin")
	dmux(os.Stdin)
}
