package qlose

import "testing"

func TestSingleWorker(t *testing.T) {
	q := New(1, 100)
	defer q.Stop()

	out := make(chan int, 10)
	task := func(n int) func() interface{} {
		return func() interface{} {
			out <- n
			return n
		}
	}

	stopper := make(chan struct{})
	q.Enqueue(0, func() interface{} {
		<-stopper
		return 0
	})

	var p [7]<-chan interface{}
	p[0], _ = q.Enqueue(0, task(0))
	p[1], _ = q.Enqueue(5, task(1))
	p[2], _ = q.Enqueue(5, task(2))
	p[3], _ = q.Enqueue(5, task(3))
	p[4], _ = q.Enqueue(1, task(4))
	p[5], _ = q.Enqueue(1, task(5))
	p[6], _ = q.Enqueue(1, task(6))

	close(stopper)

	if n := <-out; n != 0 {
		t.Errorf("expected=0, but actual=%d", n)
	}
	if n := <-out; n != 4 {
		t.Errorf("expected=4, but actual=%d", n)
	}
	if n := <-out; n != 5 {
		t.Errorf("expected=5, but actual=%d", n)
	}
	if n := <-out; n != 6 {
		t.Errorf("expected=6, but actual=%d", n)
	}
	if n := <-out; n != 1 {
		t.Errorf("expected=1, but actual=%d", n)
	}
	if n := <-out; n != 2 {
		t.Errorf("expected=2, but actual=%d", n)
	}
	if n := <-out; n != 3 {
		t.Errorf("expected=3, but actual=%d", n)
	}
}
