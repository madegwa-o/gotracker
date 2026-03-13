package queue

import "sync"

// Buffer is a bounded in-memory FIFO queue for failed cloud sends.
type Buffer struct {
	mu    sync.Mutex
	items [][]byte
	max   int
}

func NewBuffer(max int) *Buffer {
	if max <= 0 {
		max = 1000
	}
	return &Buffer{max: max, items: make([][]byte, 0, max)}
}

// Push appends an item; drops the oldest if full.
func (b *Buffer) Push(msg []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cpy := make([]byte, len(msg))
	copy(cpy, msg)

	if len(b.items) >= b.max {
		b.items = b.items[1:]
	}
	b.items = append(b.items, cpy)
}

// Pop returns oldest buffered item.
func (b *Buffer) Pop() ([]byte, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.items) == 0 {
		return nil, false
	}
	item := b.items[0]
	b.items = b.items[1:]
	return item, true
}

func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.items)
}
