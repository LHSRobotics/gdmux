package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	addr  = flag.String("addr", "127.0.0.1:5000", "tcp address on which to listen")
	stdin = flag.Bool("stdin", false, "read a gcode file from stdin")
)

func dmux(read io.Reader) {
	r := NewParser(read)
	for {
		line, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("parse error: %v", err)
		}

		if len(line.Words) == 0 {
			continue
		}

		switch line.Words[0].Content {
		case "G21":
			fmt.Printf(" mm ")
		case "G1":
			fmt.Printf("g1, ")
		case "M107":
			fmt.Printf(" Msomething ")
		default:
			log.Printf("unknown gcode: %v (%v)",line.Words[0].Content, line)
		}
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
	if *stdin {
		log.Println("reading from stdin")
		dmux(os.Stdin)
		return
	}

	listen()
}
