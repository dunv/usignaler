package signaler

import (
	"context"
	"sync"
)

// A container which supports
// - waiting for a condition to be true
// - with respecting context
// - without busy-checking the condition
type Signaler struct {
	lock  *sync.Mutex
	cond  *sync.Cond
	ready bool
}

func NewSignaler(ready bool) *Signaler {
	lock := &sync.Mutex{}
	return &Signaler{
		lock:  lock,
		cond:  sync.NewCond(lock),
		ready: ready,
	}
}

// Wait for condition to be true OR context to run out
func (s *Signaler) WaitForReady(ctx context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	useCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-useCtx.Done()
		s.cond.Signal()
	}()

	if s.ready {
		return nil
	}

	for !s.ready || ctx.Err() != nil {
		if err := ctx.Err(); err != nil {
			return ctx.Err()
		}
		s.cond.Wait()
	}

	return nil
}

// Set condition to be true
func (s *Signaler) Ready() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ready = true
	s.cond.Signal()
}

// Set condition to be false
func (s *Signaler) NotReady() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ready = false
	s.cond.Signal()
}

