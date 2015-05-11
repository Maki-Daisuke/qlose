package qlose

import (
	"fmt"
	"sync"
)

const MAX_PRIORITY uint = 9

type Qlose struct {
	quit_chan   chan struct{}
	notify_chan chan struct{}
	queues      [MAX_PRIORITY + 1]chan func()
	allDone     *sync.WaitGroup
}

func New(num_of_workers int, buffer_size uint) *Qlose {
	q := &Qlose{
		quit_chan:   make(chan struct{}),
		notify_chan: make(chan struct{}, buffer_size),
		allDone:     &sync.WaitGroup{},
	}
	var i uint
	for i = 0; i <= MAX_PRIORITY; i++ {
		q.queues[i] = make(chan func(), buffer_size)
	}
	for i := 0; i < num_of_workers; i++ {
		go q.work()
	}
	return q
}

func (q *Qlose) work() {
	q.allDone.Add(1)
LOOP:
	for {
		select {
		case <-q.quit_chan:
			q.allDone.Done()
			break LOOP
		case <-q.notify_chan:
			var i uint
			for i = 0; i <= MAX_PRIORITY; i++ {
				select {
				case t := <-q.queues[i]:
					t()
					continue LOOP
				default:
					// Just for avoiding blocking.
				}
			}
		}
	}
}

func (q *Qlose) Enqueue(prio uint, task func() interface{}) <-chan interface{} {
	if prio > 9 {
		panic(fmt.Sprintf("`prio (=%d)` is out of bound", prio))
	}
	end_chan := make(chan interface{}, 1)
	q.queues[prio] <- func() {
		defer func() {
			if x := recover(); x != nil {
				end_chan <- x
			}
			close(end_chan)
		}()
		r := task()
		end_chan <- r
	}
	q.notify_chan <- struct{}{}
	return end_chan
}

func (q *Qlose) Stop() <-chan struct{} {
	close(q.quit_chan)
	c := make(chan struct{})
	go func() {
		q.allDone.Wait()
		close(c)
	}()
	return c
}
