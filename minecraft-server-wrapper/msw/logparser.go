package msw

import (
	"log"
	"regexp"
)

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

type LogParser func(string) Event

var eventToRegexp = map[Event]*regexp.Regexp{
	StartedEvent: regexp.MustCompile(`Done (?s)(.*)! For help, type "help"`),
	StartEvent:   regexp.MustCompile(`Starting minecraft server version (.*)`),
	StopEvent:    regexp.MustCompile(`Stopping (.*) server`),
}

func LogParserFunc(line string) Event {
	ll := ParseToLogLine(line)
	ev := ll.Event()
	log.Println(ll.output, ev)
	return ev
}
