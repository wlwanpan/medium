package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"time"

	"github.com/looplab/fsm"
)

// STEP 1: Exec
func JavaExecCmd(serverPath string, iniHeapSize, maxHeapSize int) *exec.Cmd {
	iniHeapFlag := fmt.Sprintf("-Xms%dM", iniHeapSize)
	maxHeapFlag := fmt.Sprintf("-Xmx%dM", maxHeapSize)
	return exec.Command("java", iniHeapFlag, maxHeapFlag, "-jar", serverPath, "nogui")
}

// STEP 2: Console
type Console struct {
	cmd    *exec.Cmd
	stdout *bufio.Reader
	stdin  *bufio.Writer
}

func NewConsole(cmd *exec.Cmd) *Console {
	c := &Console{
		cmd: cmd,
	}

	stdout, _ := cmd.StdoutPipe()
	c.stdout = bufio.NewReader(stdout)

	stdin, _ := c.cmd.StdinPipe()
	c.stdin = bufio.NewWriter(stdin)

	return c
}

func (c *Console) Start() error {
	return c.cmd.Start()
}

func (c *Console) Kill() error {
	return c.cmd.Process.Kill()
}

func (c *Console) WriteCmd(cmd string) error {
	wrappedCmd := fmt.Sprintf("%s\r\n", cmd)
	_, err := c.stdin.WriteString(wrappedCmd)
	if err != nil {
		return err
	}
	return c.stdin.Flush()
}

func (c *Console) ReadLog() (string, error) {
	return c.stdout.ReadString('\n')
}

// STEP 3: LogParser
var logRegex = regexp.MustCompile(`(\[[0-9:]*\]) \[([A-z(-| )#0-9]*)\/([A-z #]*)\]: (.*)`)

type LogLine struct {
	timestamp  string
	threadName string
	level      string
	output     string
}

func ParseToLogLine(line string) *LogLine {
	matches := logRegex.FindAllStringSubmatch(line, 4)
	return &LogLine{
		timestamp:  matches[0][1],
		threadName: matches[0][2],
		level:      matches[0][3],
		output:     matches[0][4],
	}
}

func (ll *LogLine) Match(r *regexp.Regexp) bool {
	return r.MatchString(ll.output)
}

func (ll *LogLine) Event() Event {
	for e, reg := range eventToRegexp {
		if ll.Match(reg) {
			return e
		}
	}
	return EmptyEvent
}

type Event string

const (
	EmptyEvent   Event = "empty"
	StartedEvent       = "started"
	StoppedEvent       = "stopped"
	StartEvent         = "start"
	StopEvent          = "stop"
)

var eventToRegexp = map[Event]*regexp.Regexp{
	StartedEvent: regexp.MustCompile(`Done (?s)(.*)! For help, type "help"`),
	StartEvent:   regexp.MustCompile(`Starting minecraft server version (.*)`),
	StopEvent:    regexp.MustCompile(`Stopping (.*) server`),
}

func LogParser(line string) Event {
	ll := ParseToLogLine(line)
	ev := ll.Event()
	return ev
}

// STEP 4: Wrapper
const (
	ServerOffline  = "offline"
	ServerOnline   = "online"
	ServerStarting = "starting"
	ServerStopping = "stopping"
)

type Wrapper struct {
	console *Console
	machine *fsm.FSM
}

func NewWrapper(c *Console) *Wrapper {
	return &Wrapper{
		console: c,
		machine: fsm.NewFSM(
			ServerOffline,
			fsm.Events{
				fsm.EventDesc{
					Name: StopEvent,
					Src:  []string{ServerOnline},
					Dst:  ServerStopping,
				},
				fsm.EventDesc{
					Name: StoppedEvent,
					Src:  []string{ServerStopping},
					Dst:  ServerOffline,
				},
				fsm.EventDesc{
					Name: StartEvent,
					Src:  []string{ServerOffline},
					Dst:  ServerStarting,
				},
				fsm.EventDesc{
					Name: StartedEvent,
					Src:  []string{ServerStarting},
					Dst:  ServerOnline,
				},
			},
			nil,
		),
	}
}

func (w *Wrapper) processLogEvents() {
	for {
		line, err := w.console.ReadLog()
		if err == io.EOF {
			w.updateState(StoppedEvent)
			return
		}

		event := LogParser(line)
		w.updateState(event)
	}
}

func (w *Wrapper) updateState(ev Event) error {
	if ev == EmptyEvent {
		return nil
	}
	return w.machine.Event(string(ev))
}

func (w *Wrapper) State() string {
	return w.machine.Current()
}

func (w *Wrapper) Start() error {
	go w.processLogEvents()
	return w.console.Start()
}

func (w *Wrapper) Stop() error {
	return w.console.WriteCmd("stop")
}

func main() {
	cmd := JavaExecCmd("server.jar", 1024, 1024)
	console := NewConsole(cmd)
	wrapper := NewWrapper(console)
	wrapper.Start()

	go func() {
		time.Sleep(15 * time.Second)
		wrapper.Stop()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	console.Kill()
}
