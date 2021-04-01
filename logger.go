package godesim

import (
	"fmt"
	"io"
	"strings"
)

type LoggerOptions struct {
	Results struct {
		AllStates bool
		FormatLen int    `yaml:"format_len"`
		Separator string `yaml:"separator"`
		// Defines X in formatter %a.Xg for floating point values.
		// A value of -1 specifies default precision
		Precision int `yaml:"prec"`
		// EventPadding int    `yaml:"event_padding"`
		// EventPrefix  string `yaml:"event_prefix"`
	} `yaml:"results"`
}

// Logger accumulates messages during simulation
// run and writes them to Output once simulation finishes.
type Logger struct {
	Output io.Writer
	buff   strings.Builder
}

// Logf formats message to simulation logger. Messages are printed
// when simulation finishes. This is a rudimentary implementation of a logger.
func (log *Logger) Logf(format string, a ...interface{}) {
	log.buff.WriteString(fmt.Sprintf(format, a...))
}

func (log *Logger) flush() {
	log.Output.Write([]byte(log.buff.String()))
	log.buff.Reset()
}

func newLogger(w io.Writer) Logger {
	return Logger{Output: w, buff: strings.Builder{}}
}
