// Command vplus automates dealing with the V+ console.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var terminal = flag.String("terminal", "/dev/staubli-terminal", "the device file for the Staubli's termnial")

func main() {
	flag.Parse()

	term, err := os.OpenFile(*terminal, os.O_APPEND|os.O_RDWR, os.ModeDevice)
	if err != nil {
		log.Println("error opening device file:", err)
		os.Exit(1)
	}

file:
	for _, fname := range flag.Args() {
		content, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Println("error opening file:", err)
			continue
		}

		_, err = term.WriteString(fmt.Sprintf("edit %s\r\n", fname))
		if err != nil {
			log.Printf("error sending file (%v): %v", fname, err)
			continue file
		}

		buf := make([]byte, 1)
		for _, b := range content {
			if b == '\n' {
				_, err = term.WriteString("\r\n")
				if err != nil {
					log.Printf("error sending file (%v): %v", fname, err)
					continue file
				}
			} else {
				buf[0] = b
				_, err = term.Write(buf)
			}
		}

		_, err = term.WriteString("e\r\n")
		if err != nil {
			log.Printf("error sending file (%v): %v", fname, err)
			continue file
		}
	}
	term.Close()
}
