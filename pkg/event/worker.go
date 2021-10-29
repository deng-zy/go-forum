package event

import (
	"time"

	"github.com/pkg/errors"
)

type worker struct {
	task       chan func()
	recyleTime time.Time
	bus        *Bus
}

func (w *worker) run() {
	w.bus.incRunning()
	go func() {
		defer func() {
			w.bus.decRunning()
			w.bus.workerCache.Put(w)

			if p := recover(); p != nil {
				if ph := w.bus.opts.PanicHandler; ph != nil {
					ph(p)
				} else {
					err := errors.Errorf("worker exits from panic: %v\n", p)
					w.bus.opts.Logger.Printf("worker exits from a panic: %v\n", err)
				}
			}
			w.bus.cond.Signal()
		}()

		for f := range w.task {
			if f == nil {
				return
			}

			f()

			if ok := w.bus.revertWorker(w); !ok {
				return
			}
		}
	}()
}
