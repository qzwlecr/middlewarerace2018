package mempool

import (
	"container/list"
	"fmt"
	"log"
	"sync"
)

type BufferCond struct {
	buf []byte
	ref uint32
}

// ZhwkMemoryPool : mmp is the short for memory pool, dont mistake it!
type ZhwkMemoryPool struct {
	allocList *list.List
	mu        sync.Mutex
	buffers   []BufferCond
	clens     uint32
}

const INITIAL_CNT = 10

// NewZhwkMemoryPool returns a zhwkmmp with a presettled set of buffers.
func NewZhwkMemoryPool(theLength uint32) *ZhwkMemoryPool {
	var mmp ZhwkMemoryPool
	mmp.clens = theLength
	mmp.buffers = make([]BufferCond, INITIAL_CNT)
	mmp.allocList = list.New()
	for i := 0; i < INITIAL_CNT; i++ {
		mmp.buffers[i] = BufferCond{make([]byte, mmp.clens), 0}
		mmp.allocList.PushBack(uint32(i))
	}
	return &mmp
}

// Acquire returns a currently free memory.
func (mmp *ZhwkMemoryPool) Acquire() (uint32, []byte, error) {
	mmp.mu.Lock()
	defer mmp.mu.Unlock()
	if mmp.allocList.Len() == 0 {
		// nothing left. ALLOC ONE!
		mmp.buffers = append(mmp.buffers, BufferCond{make([]byte, mmp.clens), 1})
		log.Println("MemoryPool SIZE changed. Now ", len(mmp.buffers))
		// no need to enqueue
		return uint32(len(mmp.buffers) - 1), mmp.buffers[len(mmp.buffers)-1].buf, nil
	}
	// deque one, then
	theone := mmp.allocList.Front()
	theid := theone.Value.(uint32)
	mmp.allocList.Remove(theone)
	err := assert(mmp.buffers[theid].ref == 0, func() error { return fmt.Errorf("attempt to acquire an already referenced memory") })
	if err != nil {
		return 0, nil, err
	}
	mmp.buffers[theid].ref = 1
	return theid, mmp.buffers[theid].buf, nil
}

// Release will release an acquired or shared memory.
func (mmp *ZhwkMemoryPool) Release(id uint32) (err error) {
	// let's get through the assertions
	err = assert(uint32(len(mmp.buffers)) > id, func() error { return fmt.Errorf("releasing non-exist memory") })
	if err != nil {
		return err
	}
	// the references
	err = assert(mmp.buffers[id].ref > 0, func() error { return fmt.Errorf("double free or corruption") })
	if err != nil {
		return err
	}
	// then all checks passed.
	mmp.mu.Lock()
	defer mmp.mu.Unlock()
	mmp.buffers[id].ref = mmp.buffers[id].ref - 1
	if mmp.buffers[id].ref == 0 {
		mmp.allocList.PushBack(uint32(id))
	}
	return nil
}

// Attempt will try to acquire a memory with the id, if already acquired or shared, return error.
func (mmp *ZhwkMemoryPool) Attempt(id uint32) (arr []byte, err error) {
	err = assert(uint32(len(mmp.buffers)) > id, func() error { return fmt.Errorf("releasing non-exist memory") })
	if err != nil {
		return nil, err
	}
	mmp.mu.Lock()
	defer mmp.mu.Unlock()
	if mmp.buffers[id].ref != 0 {
		return nil, fmt.Errorf("already acquired or shared")
	}
	// first, find and remove the fucking stuff from the list.
	for e := mmp.allocList.Front(); e != nil; e = e.Next() {
		theid := e.Value.(uint32)
		if theid == id {
			mmp.allocList.Remove(e)
			break
		}
	}
	// second, increase the reference
	mmp.buffers[id].ref = 1
	return mmp.buffers[id].buf, nil
}

// Share will surely return a byte array on almost every condition. references counter increased.
func (mmp *ZhwkMemoryPool) Share(id uint32) (arr []byte, err error) {
	// first, let's try attempt..
	arr, err = mmp.Attempt(id)
	if err == nil {
		return arr, err
	}
	// that didn't work..
	err = assert(uint32(len(mmp.buffers)) > id, func() error { return fmt.Errorf("releasing non-exist memory") })
	if err != nil {
		return nil, err
	}
	mmp.mu.Lock()
	defer mmp.mu.Unlock()
	mmp.buffers[id].ref = mmp.buffers[id].ref + 1
	return mmp.buffers[id].buf, nil
}
