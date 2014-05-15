package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
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
	ttyData     = flag.String("datatty", "/dev/ttyS0", "serial tty to the Staubli data line")
	baudData    = flag.Int("datarate", 19200, "baud rate for the staubli's data line")
	ttyConsole  = flag.String("consoletty", "/dev/ttyUSB0", "serial tty to the Staubli console prompt")
	baudConsole = flag.Int("consolerate", 38400, "baud rate for the staubli's console")

	originx = flag.Float64("x", 500, "x coordinates for the origin")
	originy = flag.Float64("y", 0, "y coordinates for the origin")
	originz = flag.Float64("z", -100, "z coordinates for the origin")

	dummy       = flag.Bool("dummy", false, "don't actually send commands to the arm")
	httpAddr    = flag.String("http", "", "tcp address on which to listen")
	nosendvplus = flag.Bool("skipv", false, "don't send over the V+ code on startup")
	verbose     = flag.Bool("verbose", false, "print lots output")
	dataRoot    = flag.String("root",
		strings.Split(os.Getenv("GOPATH"), ":")[0]+"/src/github.com/LHSRobotics/gdmux",
		"repository root to find static files")

	arm     staubli.Arm
	running = false
)

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

func sendPg() {
	log.Println("Sending over V+ code")
	s, err := serial.OpenPort(&serial.Config{Name: *ttyConsole, Baud: *baudConsole})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	console := vplus.NewConsole(s)

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
}

func initArm() {
	origin.x, origin.y, origin.z = *originx, *originy, *originz

	if *dummy {
		arm = staubli.Dummy
	} else {
		log.Println("Opening ", *ttyData)
		s, err := serial.OpenPort(&serial.Config{Name: *ttyData, Baud: *baudData})
		if err != nil {
			log.Fatal(err)
		}
		arm = staubli.NewStaubli(s)
	}

	if !*nosendvplus {
		sendPg()
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [gcode.nc]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s -http :5000\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	clients.m = make(map[chan string]bool)
	go logger()

	if *httpAddr != "" {
		initArm()
		log.Println("Listening on ", *httpAddr)
		http.HandleFunc("/run", handleRun)
		http.HandleFunc("/stop", handleStop)
		http.Handle("/log", websocket.Handler(handleLog))
		http.Handle("/", http.FileServer(http.Dir(*dataRoot+"/cmd/gdmux/ui")))
		log.Fatal(http.ListenAndServe(*httpAddr, nil))
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	initArm()
	running = true
	for _, fn := range flag.Args() {
		f, err := os.Open(fn)
		if err != nil {
			log.Fatal(err)
		}
		dmux(f)
	}
}
