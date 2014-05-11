// Package vplus offers utility functions to script and control the Stäubli's V+ console.
package vplus

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

type Console struct {
	buf []byte
	l   int
	w   io.Writer
}

func NewConsole(w io.Writer) *Console {
	return &Console{
		buf: make([]byte, 40),
		w:   w,
	}
}

func (c *Console) writeLine(b []byte) error {
	// TODO read the output of these commands instead of sleeping arbitrary times...
	i := 0
	for i < len(b) {
		n := copy(c.buf[c.l:], b[i:])
		c.l += n
		i += n
		if c.l == len(c.buf) {
			_, err := c.w.Write(c.buf)
			if err != nil {
				return err
			}
			c.l = 0
			// This function is full of random sleeps because the serial line keeps crapping out on us.
			// We need to sort this out.
			time.Sleep(20 * time.Millisecond)
		}
	}

	if _, err := c.w.Write(c.buf[:c.l]); err != nil {
		return err
	}
	c.buf[0] = '\r'
	c.buf[1] = '\n'
	if _, err := c.w.Write(c.buf[:2]); err != nil {
		return err
	}
	c.l = 0

	time.Sleep(50 * time.Millisecond)
	return nil
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
	err = c.Cmd("y")
	if err != nil {
		return
	}
	err = c.Cmd(fmt.Sprintf("edit %s", path.Base(name)))
	if err != nil {
		return
	}

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
	}
	if err = scanner.Err(); err != nil {
		return
	}

	err = c.Cmd("e")
	return
}

// TODO Cmd/UpdateFile without an object. should probably panic on error.
