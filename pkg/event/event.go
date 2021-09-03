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
	workers    int
	hub        map[string]chan *Event
	subscriber map[string][]Listener
	resultChan chan *Result
	signal     chan bool
	channels   []chan *Event

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
		workers:    worker,
		channels:   []chan *Event{},
		hub:        map[string]chan *Event{},
		signal:     make(chan bool),
		subscriber: map[string][]Listener{},
		resultChan: make(chan *Result, 4096),
		timeout:    timeout,
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
	channel := make(chan *Event, 1024)
	b.subscriber[topic] = subscribers
	b.hub[topic] = channel
	b.channels = append(b.channels, channel)
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

func (b *Bus) Fails() chan *Result {
	return b.resultChan
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

		res := &Result{
			Topic:   e.Topic,
			Success: success,
			Err:     err,
		}

		if cap(b.resultChan) == 0 {
			go func(res *Result) {
				b.resultChan <- res
			}(res)
		}
		b.resultChan <- res
	}
}

func (b *Bus) Bootstrap() {
	b.dispatch()

	timer := time.NewTimer(1 * time.Second)

	for {
		t := <-timer.C
		fmt.Printf("%s-topic length:%d\n", t.Format(time.RFC3339), b.Len())
	}
}
