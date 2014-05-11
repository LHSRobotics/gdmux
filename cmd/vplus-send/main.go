// Command vplus automates dealing with the V+ console.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/LHSRobotics/gdmux/pkg/vplus"
)

var terminal = flag.String("terminal", "/dev/staubli-terminal", "the device file for the Staubli's termnial")
var execute = flag.Bool("execute", false, "execute the first program after sending")

func Cmd(s string) {
	err := console.Cmd(s)
	if err != nil {
		log.Fatal("error sending line: ", err)
	}
}

var console *vplus.Console

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	term, err := os.OpenFile(*terminal, os.O_APPEND|os.O_RDWR, os.ModeDevice)
	defer term.Close()
	if err != nil {
		log.Fatal("error opening device file: ", err)
	}

	console = vplus.NewConsole(term)

	// Quit the currently running V+ program, if one is.
	Cmd("abort")

	// Remove it from the stack so we can modify it.
	// TODO perhaps parse the output of 'status' and only kill the right process?
	// For now a couple of kill commands should do it, since it's unlikely that we're
	// running any more than that.
	Cmd("kill")
	Cmd("kill")

	for _, f := range flag.Args() {
		err = console.UpdateFile(f)
		if err != nil {
			log.Fatal("error sending file: ", f)
		}
	}

	if *execute {
		// Execute the first argument if we're running
		Cmd(fmt.Sprintf("ex %s", path.Base(flag.Args()[0])))
	}
}
