package msw

import (
	"io"
	"log"

	"github.com/looplab/fsm"
)

type Event string

const (
	EmptyEvent   Event = "empty"
	StartedEvent       = "started"
	StoppedEvent       = "stopped"
	StartEvent         = "start"
	StopEvent          = "stop"
)

const (
	ServerOffline  = "offline"
	ServerOnline   = "online"
	ServerStarting = "starting"
	ServerStopping = "stopping"
)

type Wrapper struct {
	console Console
	parser  LogParser
	machine *fsm.FSM
}

func NewDefaultWrapper(server string, initial, max int) *Wrapper {
	cmd := JavaExecCmd(server, initial, max)
	console := NewConsole(cmd)
	return NewWrapper(console, LogParserFunc)
}

func NewWrapper(c Console, p LogParser) *Wrapper {
	return &Wrapper{
		console: c,
		parser:  p,
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
			fsm.Callbacks{
				"enter_state": func(ev *fsm.Event) {
					log.Printf("State transitioned: %s -> %s", ev.Src, ev.Dst)
				},
			},
		),
	}
}

func (w *Wrapper) processLogEvents() {
	for {
		line, err := w.console.ReadLine()
		if err == io.EOF {
			w.updateState(StoppedEvent)
			return
		}

		event := w.parseLineToEvent(line)
		w.updateState(event)
	}
}

func (w *Wrapper) parseLineToEvent(line string) Event {
	return w.parser(line)
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

func (w *Wrapper) Kill() error {
	return w.console.Kill()
}
