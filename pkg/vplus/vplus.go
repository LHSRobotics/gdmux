// Package vplus offers utility functions to script and control the Stäubli's V+ console.
package vplus

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
)

type Console struct {
	buf []byte
	l   int
	w   io.ReadWriter
}

func NewConsole(w io.ReadWriter) *Console {
	return &Console{
		buf: make([]byte, 80),
		w:   w,
	}
}

// BUG: long lines fail.
func (c *Console) writeLine(b []byte) error {
	if _, err := c.w.Write(b); err != nil {
		return err
	}
	crlf := []byte{'\r', '\n'}
	if _, err := c.w.Write(crlf); err != nil {
		return err
	}
	return nil
}

func (c *Console) Expect() {
	var b [100]byte
	buf := b[:]
	for {
		fmt.Printf(".")
		n, err := c.w.Read(buf)
		// This is really shitty. We should probably do a more fancy expect sort of thing.
		if buf[n-1] == '.' || buf[n-1] == '?' || (buf[n-1] == ' ' && buf[n-2] == '?') {
			return
		}
		if err != nil {
			return
		}
	}
}

// Cmd sends a single line to the Stäubli console.
func (c *Console) Cmd(cmd string) error {
	return c.writeLine([]byte(cmd))
}

// UpdateFile sends a V+ file to the Stäubli console.
func (c *Console) UpdateFile(name string) (err error) {
	err = c.Cmd(fmt.Sprintf("delete %s", path.Base(name)))
	if err != nil {
		return
	}
	c.Expect()
	err = c.Cmd("y")
	if err != nil {
		return
	}
	c.Expect()
	err = c.Cmd(fmt.Sprintf("edit %s", path.Base(name)))
	if err != nil {
		return
	}
	c.Expect()

	f, err := os.Open(name)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		err = c.Cmd(scanner.Text())
		if err != nil {
			return
		}
	c.Expect()
	}
	if err = scanner.Err(); err != nil {
		return
	}

	err = c.Cmd("e")
	c.Expect()
	return
}

// TODO Cmd/UpdateFile without an object. Should probably panic on error.
