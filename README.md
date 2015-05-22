Qlose - A Prioritized Task Queue
================================

_**NOTE: This is experimental and still heavily under development. Most of the
features documented here is not implemeted yet.**_

Description
-----------

Qlose (pronounced kloːzə) is a very simple and prioritized task queue library
for Go, which executes multiple tasks in order in background.


Synopsis
--------

```go
import (
  "os/exec"
  "runtime"
  "github.com/Maki-Daisuke/qlose"
)

q := qlose.New(runtime.NumCPU, 100)  // Specify number of workers and buffer size

// A task is just a function with signature "func()interface{}".
p, err := q.Enqueue(5, fun()error{
  cmd := exec.Command("a long running command")
  err := cmd.Run()  // Take long time
  retutn err
})
if err !=nil {
  fmt.Println("can't enqueue somehow: %v", err)
}


// This task is prioritized over the former.
q.Enqueue(0, fun()error{
  ...
})
// You can just ignore the return value, if you are not interested in completion
// of the task.


// You can process other jobs, here.
// If you need the result of the task, receive from the channel
r := <-p  // Block until the task is done.
if err := r.(error); err != nil {
  // error handling
}
```

How to Install
--------------

```
go get github.com/Maki-Daisuke/qlose
```


API
---

### `func New(num_of_workers int, buffer_size uint) *Qlose`

Create new Qlose and start speficied number of workers.

### `func (q *Qlose) Enqueue(prio uint, task func()interface{}) (promise <-chan interface{}, err error)`

Enqueue `task` with priority `prio`, where 0 is the highest priority and 9 is
the lowest. You can wait and retrieve the result of task by receiving from the
returned channel `primise`. `primise` will be sent just a single value and closed.
Return value `err` will be non-nil value in the following situations, otherwise
it will be nil.

- `prio` is is out of bound
- `q` is already stopped

### `func (q *Qlose) Stop() <-chan struct{}`

Stop all workers of this Qlose. Once Qlose is stopped, it can never be restarted.
The returned channel is closed when all workers end.
