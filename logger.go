package a2amux

import "log"

type Logger interface {
	Printf(format string, v ...interface{})
}

type DefaultLogger struct{}

func (d DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
