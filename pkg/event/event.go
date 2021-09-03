package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var maxWorker = 10

type Event struct {
	Data  interface{}
	Topic string
}

type Bus struct {
	workers      int
	hub          map[string]chan *Event
	subscriber   map[string][]Listener
	resultChan   chan *Result
	signal       chan bool
	workerSignal chan bool
	channels     []chan *Event

	timeout time.Duration
	mu      sync.RWMutex
}

type Result struct {
	Topic   string
	Success bool
	Err     error
}

type Listener func(*Event) error

func NewBus(worker int, timeout time.Duration) *Bus {
	if worker > maxWorker {
		worker = maxWorker
	}

	return &Bus{
		workers:      worker,
		channels:     []chan *Event{},
		hub:          map[string]chan *Event{},
		resultChan:   make(chan *Result, 4096),
		signal:       make(chan bool),
		workerSignal: make(chan bool, worker),
		subscriber:   map[string][]Listener{},
		timeout:      timeout,
	}
}

func NewEvent(topic string, data interface{}) *Event {
	return &Event{
		Topic: topic,
		Data:  data,
	}
}

func (b *Bus) Publish(e *Event) {
	b.mu.RLock()
	ch, ok := b.hub[e.Topic]
	b.mu.RUnlock()

	if !ok {
		return
	}

	ch <- e
}

func (b *Bus) Subscribe(topic string, listener Listener) {
	b.mu.RLock()
	subscribers, exist := b.subscriber[topic]
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if exist {
		subscribers = append(subscribers, listener)
		b.subscriber[topic] = subscribers
		return
	}

	subscribers = []Listener{listener}
	ch := make(chan *Event, 1024)
	b.subscriber[topic] = subscribers
	b.hub[topic] = ch
	b.channels = append(b.channels, ch)
}

func (b *Bus) Len() int {
	sum := 0
	for _, ch := range b.channels {
		sum += len(ch)
	}
	return sum
}

func (b *Bus) IsEmpty() bool {
	return b.Len() == 0
}

func (b *Bus) Stop() {
	for _, ch := range b.channels {
		close(ch)
	}
	quit := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		for {
			if b.IsEmpty() {
				break
			}
		}
		quit <- true
	}()

	select {
	case <-ctx.Done():
	case <-quit:
	}
	b.signal <- true
}

func (b *Bus) getListener(topic string) []Listener {
	b.mu.RLock()
	defer b.mu.RUnlock()

	listeners, exists := b.subscriber[topic]
	if exists {
		return listeners
	}
	return nil
}

func (b *Bus) dispatch() {
	for topic, ch := range b.hub {
		listeners := b.getListener(topic)
		for _, listener := range listeners {
			b.bootstrapListener(listener, ch)
		}
	}
}

func (b *Bus) bootstrapListener(listener Listener, ch chan *Event) {
	for i := 0; i < b.workers; i++ {
		go b.listener(listener, ch)
	}
}

func (b *Bus) listener(listener Listener, ch chan *Event) {
	for {
		e, ok := <-ch
		if !ok {
			fmt.Println("listener say bye byebye!!!!")
			break
		}

		err := listener(e)
		success := true
		if err != nil {
			success = false
		}

		b.resultChan <- &Result{
			Topic:   e.Topic,
			Success: success,
			Err:     err,
		}
	}
}

func (b *Bus) Bootstrap() {
	b.dispatch()

	for {
		select {
		case <-b.signal:
			fmt.Println("see you again!!!!!")
		case <-time.After(time.Second):
			fmt.Printf("topic length:%d\n", b.Len())
		}
	}
}
