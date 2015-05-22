package qlose

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

const MAX_PRIORITY uint = 9

type Task struct {
	Priority uint
	Run      func() interface{}
	end_chan chan interface{}
}

func newTask(p uint, f func() interface{}) *Task {
	return &Task{
		Priority: p,
		Run:      f,
		end_chan: make(chan interface{}, 1),
	}
}

const (
	NOT_WORKING = 1
	WORKING     = 0
)

type Qlose struct {
	isWorking int32 // 1: working, 0: not working
	quit      chan struct{}
	notify    chan struct{}
	queues    [MAX_PRIORITY + 1]chan *Task
	allDone   *sync.WaitGroup
}

func New(num_of_workers uint, buf_size uint) *Qlose {
	q := &Qlose{
		isWorking: WORKING,
		quit:      make(chan struct{}),
		notify:    make(chan struct{}, buf_size),
		allDone:   &sync.WaitGroup{},
	}
	var i uint
	for i = 0; i <= MAX_PRIORITY; i++ {
		q.queues[i] = make(chan *Task, buf_size)
	}
	for i := uint(0); i < num_of_workers; i++ {
		go q.work(q.quit, q.allDone)
	}
	runtime.SetFinalizer(q, func(q *Qlose) {
		<-q.Stop()
	})
	return q
}

func (q *Qlose) work(quit chan struct{}, allDone *sync.WaitGroup) {
	allDone.Add(1)
	defer allDone.Done()
LOOP:
	for {
		select {
		case <-quit:
			break LOOP
		case <-q.notify:
			if atomic.LoadInt32(&q.isWorking) == NOT_WORKING {
				break LOOP
			}
			var i uint
			for i = 0; i <= MAX_PRIORITY; i++ {
				select {
				case t := <-q.queues[i]:
					func() { // scope for defer
						defer func() {
							if x := recover(); x != nil {
								t.end_chan <- x
							}
							close(t.end_chan)
						}()
						r := t.Run()
						t.end_chan <- r
					}()
					continue LOOP
				default:
					// Just for avoiding blocking.
				}
			}
		}
	}
}

func (q *Qlose) Stop() <-chan struct{} {
	r := make(chan struct{})
	if !atomic.CompareAndSwapInt32(&q.isWorking, WORKING, NOT_WORKING) {
		// workers are already stopped by someone else.
		close(r)
		return r
	}
	close(q.quit)
	go func() {
		q.allDone.Wait()
		close(r)
	}()
	return r
}

func (q *Qlose) Enqueue(p uint, f func() interface{}) (<-chan interface{}, error) {
	if p > MAX_PRIORITY {
		return nil, fmt.Errorf("`prio (=%d)` is out of bound", p)
	}
	if atomic.LoadInt32(&q.isWorking) != WORKING {
		return nil, fmt.Errorf("queue is already stopped")
	}
	t := newTask(p, f)
	q.queues[p] <- t
	q.notify <- struct{}{}
	return t.end_chan, nil
}
