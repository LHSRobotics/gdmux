package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	// armPort is the serial file connected to the arm controller's data line. For the Staubli
	// its baudrate 38400, we assume that's already set for the device file. (I.e. with stty.)
	armFile      = flag.String("arm", "/dev/ttyUSB0", "serial file to talk to the staubli's console")
	extruderFile = flag.String("extruder", "/dev/ttyS1", "serial file to talk to the extruder's firmware")
	addr         = flag.String("addr", "127.0.0.1:5000", "tcp address on which to listen")
	stdin        = flag.Bool("stdin", false, "read a gcode file from stdin")
	tcp          = flag.Bool("tcp", false, "listen on tcp for gcode")
	verbose      = flag.Bool("verbose", false, "print lots output")
	dataRoot     = flag.String("data",
		strings.Split(os.Getenv("GOPATH"), ":")[0]+"/src/github.com/LHSRobotics/gdmux/ui",
		"html directory")

	armc  = make(chan armMsg)
	stopc = make(chan bool)

	running = false
)

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
		go dmux(conn, make(chan bool))
	}
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	log.Println("Running some gcode!")
	running = true
	dmux(r.Body, stopc)
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	log.Println("Got stop request")
	if running {
		log.Println("stopping")
		stopc <- true
	}
	log.Println("stopped")
	running = false
}

func main() {
	flag.Parse()

	go armCtl() // Launch the arm controlling goroutine. We talk to this using armc.

	if *stdin {
		log.Println("reading from stdin")
		dmux(os.Stdin, make(chan bool))
		os.Exit(0)
	}

	if *tcp {
		listen()
		os.Exit(0)
	}

	http.HandleFunc("/run", handleRun)
	http.HandleFunc("/stop", handleStop)
	http.Handle("/", http.FileServer(http.Dir(*dataRoot)))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
