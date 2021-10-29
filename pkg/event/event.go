package event

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var workerChnCap = func() int {
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}
	return 1
}()

//Event 事件
type Event struct {
	Data  interface{} `json:"data"`
	Topic string      `json:"topic"`
}

//Bus 事件
type Bus struct {
	topic       map[string]string
	handles     map[string][]Handle
	opts        *Options
	state       int32
	workerCache sync.Pool
	capacity    int32
	cond        *sync.Cond
	workers     *workerStack
	lock        sync.Locker
	rwLock      sync.RWMutex
	running     int32
	blockingNum int
}

// Handle 事件处理者
type Handle func(*Event) error

func loadOptions(options ...Option) (opts *Options) {
	opts = new(Options)
	for _, option := range options {
		option(opts)
	}

	return
}

//NewBus  returns a new Bus
func NewBus(size int, options ...Option) *Bus {
	opts := loadOptions(options...)

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	if expiry := opts.ExpiryDuration; expiry < 1 {
		opts.ExpiryDuration = defaultCleanIntervalTime
	}

	bus := &Bus{
		topic:    map[string]string{},
		handles:  map[string][]Handle{},
		opts:     opts,
		lock:     NewSpinLock(),
		capacity: int32(size),
	}
	bus.workerCache.New = func() interface{} {
		return &worker{
			task: make(chan func(), workerChnCap),
			bus:  bus,
		}
	}
	bus.workers = newWorkerStack(size)
	bus.cond = sync.NewCond(bus.lock)

	go bus.purgePeriodically()

	return bus
}

//Publish a new event
func (b *Bus) Publish(topic string, data interface{}) {
	b.rwLock.RLock()
	handles, exists := b.handles[topic]
	b.rwLock.RUnlock()

	if !exists {
		return
	}

	event := &Event{
		Topic: topic,
		Data:  data,
	}

	wrap := func(event *Event, handle Handle) func() {
		return func() {
			handle(event)
		}
	}

	for _, handle := range handles {
		task := wrap(event, handle)
		worker := b.retriveWorker()
		if worker == nil {
			return
		}
		worker.task <- task
	}
}

func (b *Bus) Release() {
	atomic.StoreInt32(&b.state, CLOSED)
	b.lock.Lock()
	b.workers.reset()
	b.lock.Unlock()
	b.cond.Broadcast()
}

func (b *Bus) IsClosed() bool {
	return atomic.LoadInt32(&b.state) == CLOSED
}

//Subscribe a event
func (b *Bus) Subscribe(topic string, listener ...Handle) {
	b.registerTopic(topic)
	b.registerHandle(topic, listener...)
}

func (b *Bus) Cap() int {
	return int(atomic.LoadInt32(&b.capacity))
}

func (b *Bus) Running() int {
	return int(atomic.LoadInt32(&b.running))
}

func (b *Bus) Reboot() {
	if atomic.CompareAndSwapInt32(&b.state, CLOSED, OPEND) {
		go b.purgePeriodically()
	}
}

func (b *Bus) Free() int {
	c := b.Cap()
	if c < 0 {
		return -1
	}
	return c - b.Running()
}

func (b *Bus) incRunning() {
	atomic.AddInt32(&b.running, 1)
}

func (b *Bus) decRunning() {
	atomic.AddInt32(&b.capacity, -1)
}

func (b *Bus) registerTopic(topic string) {
	b.rwLock.RLock()
	_, exists := b.topic[topic]
	b.rwLock.RUnlock()

	if exists {
		return
	}

	b.lock.Lock()
	b.topic[topic] = topic
	b.lock.Unlock()
}

func (b *Bus) registerHandle(topic string, handle ...Handle) {
	b.rwLock.RLock()
	handles, exists := b.handles[topic]
	b.lock.Unlock()

	register := func(handle ...Handle) {
		b.lock.Lock()
		b.handles[topic] = handle
		b.lock.Unlock()
	}

	if exists {
		handles = append(handles, handle...)
		register(handles...)
		return
	}

	register(handle...)
}

func (b *Bus) purgePeriodically() {
	heartbeat := time.NewTicker(b.opts.ExpiryDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		if b.IsClosed() {
			break
		}

		b.lock.Lock()
		expiredWorkers := b.workers.retrieve(b.opts.ExpiryDuration)
		b.lock.Unlock()

		for i := range expiredWorkers {
			expiredWorkers[i].task <- nil
			expiredWorkers[i] = nil
		}

		if b.Running() == 0 {
			b.cond.Broadcast()
		}

	}
}

func (b *Bus) retriveWorker() (w *worker) {
	spanWorker := func() {
		w := b.workerCache.Get().(*worker)
		w.run()
	}

	b.lock.Lock()

	w = b.workers.detach()
	if w != nil {
		b.lock.Unlock()
	} else if capacity := b.Cap(); capacity == -1 || capacity > b.Running() {
		b.lock.Unlock()
		spanWorker()
	} else {
		if b.opts.NonBlocking {
			b.lock.Unlock()
			return
		}
	retry:
		if b.opts.MaxBlockingTasks != 0 && b.blockingNum >= b.opts.MaxBlockingTasks {
			b.lock.Unlock()
			return
		}
		b.blockingNum++
		b.cond.Wait()
		var nw int

		if nw = b.Running(); nw == 0 {
			b.lock.Unlock()
			if !b.IsClosed() {
				spanWorker()
			}
			return
		}

		if w = b.workers.detach(); w == nil {
			if nw < capacity {
				b.lock.Unlock()
				spanWorker()
				return
			}
			goto retry
		}
		b.lock.Unlock()
	}
	return
}

func (b *Bus) revertWorker(w *worker) bool {
	if capacity := b.Cap(); (capacity > 0 && b.Running() > capacity) || b.IsClosed() {
		return false
	}
	w.recyleTime = time.Now()
	b.lock.Lock()

	if b.IsClosed() {
		b.lock.Unlock()
		return false
	}

	err := b.workers.insert(w)
	if err != nil {
		b.lock.Unlock()
		return false
	}

	b.cond.Signal()
	b.lock.Unlock()

	return true
}
