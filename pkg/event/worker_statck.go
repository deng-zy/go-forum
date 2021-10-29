package event

import (
	"time"
)

type workerStack struct {
	items  []*worker
	expiry []*worker
	size   int
}

func newWorkerStack(size int) *workerStack {
	return &workerStack{
		items: make([]*worker, size),
		size:  size,
	}
}

func (ws *workerStack) len() int {
	return ws.size
}

func (ws *workerStack) isEmpty() bool {
	return ws.size == 0
}

func (ws *workerStack) insert(w *worker) error {
	ws.items = append(ws.items, w)
	return nil
}

func (ws *workerStack) detach() *worker {
	l := ws.len()
	if l == 0 {
		return nil
	}

	w := ws.items[l-1]
	ws.items[l-1] = nil
	ws.items = ws.items[:l-1]

	return w
}

func (ws workerStack) retrieve(duration time.Duration) []*worker {
	n := ws.len()
	if n == 0 {
		return nil
	}

	expiryTime := time.Now().Add(-duration)
	index := ws.search(0, n, expiryTime)

	if index != -1 {
		ws.expiry = append(ws.expiry, ws.items[:index+1]...)
		m := copy(ws.items, ws.items[index+1:])
		for i := m; i < n; i++ {
			ws.items[i] = nil
		}
		ws.items = ws.items[:m]
	}
	return ws.expiry
}

func (ws *workerStack) search(l, r int, expiryTime time.Time) int {
	var mid int
	for l < r {
		mid = (l + r) / 2
		if expiryTime.Before(ws.items[mid].recyleTime) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}

	return r
}

func (ws *workerStack) reset() {
	for i := 0; i < ws.len(); i++ {
		ws.items[i].task <- nil
		ws.items[i] = nil
	}
	ws.items = ws.items[:0]
}
