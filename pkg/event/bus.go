package event

import (
	"log"
	"os"
	"time"
)

const (
	OPEND = iota
	CLOSED
)

var (
	defaultLogger            = Logger(log.New(os.Stderr, "", log.LstdFlags))
	defaultBus               = NewBus(1024)
	defaultCleanIntervalTime = time.Minute
)

type Logger interface {
	Printf(format string, args ...interface{})
}

func Publish(topic string, data interface{}) {
	defaultBus.Publish(topic, data)
}

func Release() {
	defaultBus.Release()
}

func IsClosed() bool {
	return defaultBus.IsClosed()
}

func Subscribe(topic string, listeners ...Handle) {
	defaultBus.Subscribe(topic, listeners...)
}

func Cap() int {
	return defaultBus.Cap()
}

func Running() int {
	return defaultBus.Running()
}

func Reboot() {
	defaultBus.Reboot()
}

func Free() {
	defaultBus.Free()
}
