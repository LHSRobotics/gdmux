package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"code.google.com/p/go.net/websocket"
	"github.com/tarm/goserial"

	"github.com/LHSRobotics/gdmux/pkg/staubli"
	"github.com/LHSRobotics/gdmux/pkg/vplus"
)

var (
	// armPort is the serial file connected to the arm controller's data line. For the Staubli
	// its baudrate 19200, we assume that's already set for the device file. (I.e. with stty.)
	armFile     = flag.String("arm", "/dev/staubli-data", "serial file to the Staubli console data line")
	consoleFile = flag.String("console", "/dev/staubli-terminal", "serial file to the Staubli console prompt")
	baudrate    = flag.Int("baudrate", 19200, "baud rate for the staubli's console")
	dummy       = flag.Bool("dummy", false, "don't actually send commands to the arm")
	addr        = flag.String("addr", "0.0.0.0:5000", "tcp address on which to listen")
	stdin       = flag.Bool("stdin", false, "read a gcode file from stdin")
	nosendvplus = flag.Bool("nosendvplus", false, "don't send over the V+ code on startup")
	verbose     = flag.Bool("verbose", false, "print lots output")
	dataRoot    = flag.String("data",
		strings.Split(os.Getenv("GOPATH"), ":")[0]+"/src/github.com/LHSRobotics/gdmux",
		"repository root to find static files")

	arm staubli.Arm

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
		go dmux(conn)
	}
}

var sessionLock = sync.Mutex{}

func handleRun(w http.ResponseWriter, r *http.Request) {
	// TODO: communicate the running state to js, so the right buttons get enabled/disabled.
	if running {
		weblog(fmt.Sprintf("Got run request from %s, but the arm is already running.\n", r.RemoteAddr))
		return
	}
	weblog(fmt.Sprintf("Got run request from %s\n", r.RemoteAddr))
	sessionLock.Lock()
	running = true
	fmt.Println("comingata")
	weblog("RUNNING GCODE!\n")
	dmux(r.Body)
	running = false
	sessionLock.Unlock()
	weblog("Done.\n")
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	if running {
		weblog(fmt.Sprintf("Got stop request from %s\n", r.RemoteAddr))
		running = false
		weblog("Stopped sending Gcode\n")
	} else {
		weblog(fmt.Sprintf("Got stop request from %s, but the arm isn't running.\n", r.RemoteAddr))
	}
}

var clients struct {
	sync.Mutex
	m map[chan string]bool
}

var logc = make(chan string, 100)

func weblog(msg string) {
	log.Printf("%s", msg)
	logc <- msg
}

func logger() {
	for {
		msg := <-logc
		for c, _ := range clients.m {
			select {
			case c <- msg:
			default:
			}
		}
	}
}

func handleLog(ws *websocket.Conn) {
	var msgc = make(chan string, 200)

	// TODO: Move this to weblog.Register()/Unregister() methods?
	clients.Lock()
	clients.m[msgc] = true
	clients.Unlock()
	defer func() {
		clients.Lock()
		delete(clients.m, msgc)
		clients.Unlock()
	}()

	enc := json.NewEncoder(ws)
	for {
		err := enc.Encode(<-msgc)
		if err != nil {
			break
		}
	}
}

func main() {
	flag.Parse()

	clients.m = make(map[chan string]bool)
	go logger()

	if *dummy {
		arm = staubli.Dummy
	} else {
		log.Println("Opening ", *armFile)
		s, err := serial.OpenPort(&serial.Config{Name: *armFile, Baud: *baudrate})
		if err != nil {
			log.Fatal(err)
		}

		arm = staubli.NewStaubli(s)
	}

	if !*nosendvplus {
		log.Println("Sending over V+ code")
		term, err := os.OpenFile(*consoleFile, os.O_APPEND|os.O_RDWR, os.ModeDevice)
		if err != nil {
			log.Fatal("error opening device file: ", err)
		}
		console := vplus.NewConsole(term)

		f := *dataRoot + "/pkg/staubli/V+/gcode.pg"

		err = console.Cmd("abort")
		if err != nil {
			log.Fatal("error sending file: ", f)
		}
		err = console.Cmd("kill")
		if err != nil {
			log.Fatal("error sending file: ", f)
		}
		err = console.UpdateFile(f)
		if err != nil {
			log.Fatal("error sending file: ", f)
		}
		err = console.Cmd(fmt.Sprintf("ex %s", path.Base(f)))
		if err != nil {
			log.Fatal("error sending file: ", f)
		}
		term.Close()
	}

	if *stdin {
		log.Println("reading from stdin")
		running = true
		dmux(os.Stdin)
		os.Exit(0)
	}

	log.Println("listening on ", *addr)
	http.HandleFunc("/run", handleRun)
	http.HandleFunc("/stop", handleStop)
	http.Handle("/log", websocket.Handler(handleLog))
	http.Handle("/", http.FileServer(http.Dir(*dataRoot+"/cmd/gdmux/ui")))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
