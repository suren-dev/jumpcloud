package requestqueue

import (
  "sync"
)

// Queue to store HashDesc
type Queue struct {
  lock *sync.Mutex
  Values []HashDesc
}

func Init() *Queue {
  return &Queue{&sync.Mutex{}, make([]HashDesc, 0)}
}

// add entry to the end of the queue
func (q *Queue) Enqueue(x HashDesc) {
  for {
    q.lock.Lock()
    q.Values = append(q.Values, x)
    q.lock.Unlock()
    return
  }
}

// return the top of the queue and removing it
func (q *Queue) Dequeue() *HashDesc {
  for {
    if (len(q.Values) > 0) {
      q.lock.Lock()
      x := q.Values[0]
      q.Values = q.Values[1:]
      q.lock.Unlock()
      return &x
    }
    return nil
  }
}

// return the top of the queue without removing it
func (q *Queue) Peek() *HashDesc {
  for {
    if (len(q.Values) > 0) {
      q.lock.Lock()
      x := q.Values[0]
      q.lock.Unlock()
      return &x
    }
    return nil
  }
}
