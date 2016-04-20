package concurrent

import (
	"errors"
	"sync"
)

var (
	ErrExceedsLimit = errors.New("")
)

type Permit struct{}

type Semaphore struct {
	Max int

	permits chan Permit

	*sync.Mutex
}

func NewSemaphore(max int) *Semaphore {
	return &Semaphore{
		max,
		make(chan Permit, max),
		&sync.Mutex{},
	}
}

func (s *Semaphore) Release(n int) error {
	l := s.Available()
	if l+n > s.Max {
		// ?
		close(s.permits)
		return ErrExceedsLimit
	}

	s.Lock()
	defer s.Unlock()

	permit := Permit{}
	for i := 0; i < n; i++ {
		s.permits <- permit
	}

	return nil
}

func (s *Semaphore) Acquire(n int) <-chan Permit {
	callback := make(chan Permit, n)
	defer close(callback)

	for i := 0; i < n; i++ {
		<-s.permits
		callback <- Permit{}
	}

	return callback
}

func (s *Semaphore) Drain() {
	s.Lock()
	defer s.Unlock()

	s.Acquire(len(s.permits))
}

func (s *Semaphore) Available() int {
	return len(s.permits)
}
