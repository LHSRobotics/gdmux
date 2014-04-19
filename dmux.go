package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
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
