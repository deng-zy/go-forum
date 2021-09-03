package snowflake

import (
	"fmt"
	"sync"
	"time"
)

var (
	DataCenterBits uint8 = 5
	NodeBits       uint8 = 5
	StepBits       uint8 = 12

	mu              sync.Mutex
	nodeMax         int64 = -1 ^ (-1 << NodeBits)
	dataCenterMax   int64 = -1 ^ (-1 << DataCenterBits)
	stepMask        int64 = -1 ^ (-1 << StepBits)
	timeShift             = DataCenterBits + NodeBits + StepBits
	nodeShift             = StepBits
	dataCenterShift       = StepBits + NodeBits
)

type Node struct {
	mu         sync.Mutex
	epoch      time.Time
	time       int64
	dataCenter int64
	node       int64
	step       int64

	nodeMax         int64
	stepMask        int64
	timeShift       uint8
	nodeShift       uint8
	dataCenterShift uint8
}

type ID int64

func NewNode(node, dataCenter int64, epoch time.Time) (*Node, error) {
	mu.Lock()
	nodeMax = -1 ^ (-1 << NodeBits)
	timeShift = NodeBits + DataCenterBits + StepBits
	nodeShift = StepBits
	stepMask = -1 ^ (-1 << StepBits)
	dataCenterShift = StepBits + NodeBits

	mu.Unlock()

	if node < 0 || node > nodeMax {
		return nil, fmt.Errorf("Node number must be between 0 and %d", nodeMax)
	}

	if dataCenter < 0 || dataCenter > dataCenterMax {
		return nil, fmt.Errorf("dataCenter number must be between 0 and %d", dataCenterMax)
	}

	n := &Node{
		node:            node,
		stepMask:        stepMask,
		dataCenter:      dataCenter,
		nodeMax:         nodeMax,
		timeShift:       timeShift,
		nodeShift:       nodeShift,
		dataCenterShift: dataCenterShift,
		epoch:           epoch,
	}

	return n, nil
}

func (n *Node) Generate() ID {
	n.mu.Lock()
	now := time.Since(n.epoch).Nanoseconds() / 1000000

	if now == n.time {
		n.step = (n.step + 1) & n.stepMask
		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		n.step = 0
	}
	n.time = now

	r := ID((now << n.timeShift) |
		(n.dataCenter << n.dataCenterShift) |
		(n.node << n.nodeShift) |
		(n.step),
	)
	n.mu.Unlock()
	return r
}
