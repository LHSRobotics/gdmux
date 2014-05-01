// Command vplus automates dealing with the V+ console.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type SlowWriter struct {
	buf []byte
	l   int
	w   io.Writer
}

func (sw *SlowWriter) Write(b []byte) error {
	i := 0
	for i < len(b) {
		n := copy(sw.buf[sw.l:], b[i:])
		sw.l += n
		i += n
		if sw.l == len(sw.buf) {
			_, err := sw.w.Write(sw.buf)
			if err != nil {
				return err
			}
			sw.l = 0
			time.Sleep(20 * time.Millisecond)
		}
	}
	return nil
}

func (sw *SlowWriter) Flush() {
	sw.w.Write(sw.buf[:sw.l])
	time.Sleep(40 * time.Millisecond)
	sw.l = 0
}

func NewSlowWriter(w io.Writer) *SlowWriter {
	return &SlowWriter{
		buf: make([]byte, 40),
		w:   w,
	}
}

var terminal = flag.String("terminal", "/dev/staubli-terminal", "the device file for the Staubli's termnial")

func sendString(s string) error {
	err := slow.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

var slow *SlowWriter

func main() {
	flag.Parse()

	var err error
	term, err := os.OpenFile(*terminal, os.O_APPEND|os.O_RDWR, os.ModeDevice)
	if err != nil {
		log.Println("error opening device file:", err)
		os.Exit(1)
	}

	slow = NewSlowWriter(term)

file:
	for _, fname := range flag.Args() {
		content, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Println("error opening file:", err)
			continue
		}

		err = sendString(fmt.Sprintf("delete %s\r\n", fname))
		if err != nil {
			log.Printf("error sending file (%v): %v", fname, err)
			continue file
		}
		slow.Flush()
		err = sendString("y\r\n")
		if err != nil {
			log.Printf("error sending file (%v): %v", fname, err)
			continue file
		}
		slow.Flush()
		err = sendString(fmt.Sprintf("edit %s\r\n", fname))
		if err != nil {
			log.Printf("error sending file (%v): %v", fname, err)
			continue file
		}
		slow.Flush()

		for _, b := range content {
			if b == '\n' {
				err = sendString("\r\n")
				slow.Flush()
			} else {
				err = sendString(string(b))
			}

			if err != nil {
				log.Printf("error sending file (%v): %v", fname, err)
				continue file
			}
		}

		err = sendString("e\r\n")
		if err != nil {
			log.Printf("error sending file (%v): %v", fname, err)
			continue file
		}
	}
	slow.Flush()
	term.Close()
}
