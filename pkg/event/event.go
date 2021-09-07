package event

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var maxWorker = 10

//Event 事件
type Event struct {
	Data  interface{} `json:"data"`
	Topic string      `json:"topic"`
}

//Bus 事件
type Bus struct {
	workers    int
	hub        map[string]chan *Event
	subscriber map[string][]Listener
	resultChan chan *Result
	signal     chan bool
	channels   []chan *Event
	closed     bool
	dataFile   string
	mu         sync.RWMutex
	subMu      sync.RWMutex
}

//Result listenner处理结果
type Result struct {
	Topic   string
	Success bool
	Err     error
}

//Listener 事件处理者
type Listener func(*Event) error

//NewBus  returns a new Bus
func NewBus(worker int, dataFile string) *Bus {
	if worker > maxWorker {
		worker = maxWorker
	}

	return &Bus{
		workers:    worker,
		channels:   []chan *Event{},
		hub:        map[string]chan *Event{},
		signal:     make(chan bool, 1),
		subscriber: map[string][]Listener{},
		closed:     false,
		resultChan: make(chan *Result, 4096),
		dataFile:   dataFile,
	}
}

//Publish a new event
func (b *Bus) Publish(topic string, data interface{}) {
	if b.closed {
		return
	}

	b.mu.RLock()
	ch, ok := b.hub[topic]
	b.mu.RUnlock()

	if !ok {
		b.mu.Lock()
		ch = make(chan *Event, 1024)
		b.hub[topic] = ch
		b.mu.Unlock()
	}

	if len(ch) == cap(ch) {
		go func() {
			ch <- &Event{
				Topic: topic,
				Data:  data,
			}
		}()
		return
	}
	ch <- &Event{
		Topic: topic,
		Data:  data,
	}
}

//Subscribe a event
func (b *Bus) Subscribe(topic string, listener ...Listener) {
	b.subMu.RLock()
	subscribers, exist := b.subscriber[topic]
	channel, hubExist := b.hub[topic]
	b.subMu.RUnlock()

	if exist {
		subscribers = append(subscribers, listener...)
	} else {
		subscribers = listener
	}

	b.subMu.Lock()
	if !hubExist {
		channel = make(chan *Event, 1024)
		b.hub[topic] = channel
		b.channels = append(b.channels, channel)
	}
	b.subscriber[topic] = subscribers
	b.subMu.Unlock()

	for _, executor := range listener {
		b.bootstrapListener(executor, channel)
	}
}

//Len len
func (b *Bus) Len() int {
	sum := 0
	for _, ch := range b.channels {
		sum += len(ch)
	}
	return sum
}

//IsEmpty 是否为空
func (b *Bus) IsEmpty() bool {
	return b.Len() == 0
}

//Result listener handle result
func (b *Bus) Result() chan *Result {
	return b.resultChan
}

//Stop stop event
func (b *Bus) Stop() {
	b.closed = true
	b.dump()

	quit := make(chan bool, 1)
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
		b.signal <- true
	}
}

//dump channel数据到文件
func (b *Bus) dump() {
	data := map[string][]*Event{}
	for topic, channel := range b.hub {
		num := len(channel)
		if num < 1 {
			continue
		}
		rows := []*Event{}
		for {
			select {
			case e := <-channel:
				rows = append(rows, e)
			case <-time.After(time.Microsecond):
				goto read
			}
		}
	read:
		data[topic] = rows
	}

	body, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal fail. error:%s", err.Error())
		return
	}

	err = ioutil.WriteFile(b.dataFile, body, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write file fail. error:%s", err.Error())
	}
}

//load 加载数据到channel
func (b *Bus) load() error {
	body, err := ioutil.ReadFile(b.dataFile)
	if err != nil {
		return err
	}

	var data map[string][]*Event
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	for topic, events := range data {
		channel, ok := b.hub[topic]
		if !ok {
			size := len(channel)
			if size < 1024 {
				size = 1024
			}
			channel = make(chan *Event, size)
			b.hub[topic] = channel
		}

		for _, event := range events {
			channel <- event
		}
	}

	return nil
}

//bootstrapListener 启动listener
func (b *Bus) bootstrapListener(listener Listener, ch chan *Event) {
	for i := 0; i < b.workers; i++ {
		go b.listener(listener, ch)
	}
}

//listener event listener
func (b *Bus) listener(listener Listener, ch chan *Event) {
	for {
		if b.closed {
			fmt.Println("event bus exiting.....")
			return
		}

		e, ok := <-ch
		if !ok {
			fmt.Println("channel already close!!!!")
			continue
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

		//re check
		if b.closed {
			return
		}

		if cap(b.resultChan) == len(b.resultChan) {
			go func(res *Result) {
				b.resultChan <- res
			}(res)
		} else {
			b.resultChan <- res
		}
	}
}

//Bootstrap 启动事件监听
func (b *Bus) Bootstrap() {
	b.load()
	t := time.NewTicker(5 * time.Second)
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGHUP)
	signal.Notify(ch, syscall.SIGINT)
	signal.Notify(ch, syscall.SIGTERM)

loop:
	for {
		select {
		case t := <-t.C:
			fmt.Printf("%s-topic length:%d\n", t.Format(time.RFC3339), b.Len())
		case <-ch:
			b.Stop()
			fmt.Println("see you again!!!")
			break loop
		}
	}
}
