package dataStructures

import (
	"github.com/pkg/errors"
	"sync"
)

type idFunc func(interface{}) int
type compFunc func(interface{}, interface{}) bool

type RingBuff struct {
	buff             []interface{}
	len, first, last int
	id               idFunc
	lock             sync.Mutex
}

// next is a helper function for ringbuff
// it handles incrementing the first & last markers
func (rb *RingBuff) next() {
	rb.last = (rb.last + 1) % rb.len
	if rb.last == rb.first {
		rb.first = (rb.first + 1) % rb.len
	}
	if rb.first == -1 {
		rb.first = 0
	}
}

// getIndex is a helper function for ringbuff
// it returns an index relative to the first/last position of the buffer
func (rb *RingBuff) getIndex(i int) int {
	var index int
	if i < 0 {
		index = (rb.last + rb.len + i) % rb.len
	} else {
		index = (rb.first + i) % rb.len
	}
	return index
}

// Initialize a new ring buffer with length n
func New(n int, id idFunc) *RingBuff {
	rb := &RingBuff{
		buff:  make([]interface{}, 0),
		len:   n,
		first: -1,
		last:  0,
		id:    id,
	}
	return rb
}

// Push a round to the buffer
func (rb *RingBuff) Push(val interface{}) {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	rb.buff[rb.last] = val
	rb.next()
}

// push a round to a relative index in the buffer
func (rb *RingBuff) UpsertById(val interface{}, id idFunc, comp compFunc) error {
	rb.lock.Lock()
	defer rb.lock.Unlock()
	newId := id(val)

	if id(rb.buff[rb.first]) > newId {
		return errors.Errorf("Did not upsert value %+v; id is older than first tracked", val)
	}

	lastId := id(rb.Get())
	if lastId+1 == newId {
		rb.Push(val)
	} else if (lastId + 1) < newId {
		for i := lastId + 1; i <= newId; i++ {
			rb.Push(nil)
		}
		rb.Push(val)
	} else if lastId+1 > newId {
		i := rb.getIndex(newId - (lastId + 1))
		if comp(rb.buff[i], val) {
			rb.buff[i] = val
		} else {
			return errors.Errorf("Did not upsert value %+v; comp function returned false", val)
		}
	}
	return nil
}

func (rb *RingBuff) Get() interface{} {
	mostRecentIndex := (rb.last + rb.len - 1) % rb.len
	return rb.buff[mostRecentIndex]
}

func (rb *RingBuff) GetById(i int) (interface{}, error) {
	firstId := rb.id(rb.buff[rb.first])
	if i < firstId {
		return nil, errors.Errorf("requested ID %d is lower than oldest id %d", i, firstId)
	}

	lastId := rb.id(rb.Get())
	if i > lastId {
		return nil, errors.Errorf("requested id %d is higher than most recent id %d", i, lastId)
	}

	index := rb.getIndex(firstId - i)
	return rb.buff[index], nil
}

func (rb *RingBuff) Len() int {
	return rb.len
}
