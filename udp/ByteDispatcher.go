package udp

import (
	"sync"
)

// ByteDispatcher is specialized component, used to deliver bytes to
// channel only if outgoing channel speed is higher that incoming
// channel.
type ByteDispatcher interface {
	Publish([]byte)
	Channel() <-chan []byte
	Stats() (currentCount, currentSize, dropByCount, dropBySize int)
}

// NewByteDispatcher constructs new byte dispatcher with given constrain.
// When one of given limits is exceeded, dispatcher will drop current packet.
// Zero for each parameters means no limit.
func NewByteDispatcher(limitQueueCount, limitQueueSize int) ByteDispatcher {
	return &byteDispatcher{
		in:         make(chan []byte),
		out:        make(chan []byte),
		limitCount: limitQueueCount,
		limitSize:  limitQueueSize,
	}
}

type byteDispatcher struct {
	in, out chan []byte

	m sync.Mutex

	queueCount, queueSize   int
	limitCount, limitSize   int
	dropByCount, dropBySize int
}

func (d *byteDispatcher) Stats() (currentCount, currentSize, dropByCount, dropBySize int) {
	currentCount = d.queueCount
	currentSize = d.queueSize
	dropByCount = d.dropByCount
	dropBySize = d.dropBySize
	return
}

func (d *byteDispatcher) Channel() <-chan []byte {
	return d.out
}

func (d *byteDispatcher) Publish(b []byte) {
	l := len(b)
	if l > 0 {
		deliver := true

		d.m.Lock()
		if d.limitCount > 0 && d.queueCount > d.limitCount {
			deliver = false
			d.dropByCount++
		} else {
			d.queueCount++
		}
		if d.limitSize > 0 && d.queueSize > d.limitSize {
			deliver = false
			d.dropBySize++
		} else {
			d.queueSize += l
		}
		d.m.Unlock()

		if deliver {
			go d.publish(b)
		}
	}
}

func (d *byteDispatcher) publish(b []byte) {
	d.out <- b
	d.m.Lock()
	d.queueCount--
	d.queueSize -= len(b)
	d.m.Unlock()
}
