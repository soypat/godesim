package godesim

import (
	"fmt"
	"io"
	"strings"
)

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
