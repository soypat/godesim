package godesim

import (
	"fmt"
	"io"
	"strings"
)

type Logger struct {
	w    io.Writer
	buff strings.Builder
}

func (log *Logger) Logf(format string, a ...interface{}) {
	log.buff.WriteString(fmt.Sprintf(format, a...))
}
func (log *Logger) Flush() {
	log.w.Write([]byte(log.buff.String()))
}

func newLogger(w io.Writer) Logger {
	return Logger{w: w, buff: strings.Builder{}}
}
