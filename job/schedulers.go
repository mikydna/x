package job

import (
	"errors"
	"log"
	"time"
)

var (
	ErrAlreadyStarted = errors.New("Already started")
	ErrAlreadyStopped = errors.New("Already stopped")
)

type ClockScheduler struct {
	interval time.Duration
	running  bool
	done     chan bool

	*Scheduler
}

func NewClockScheduler(interval time.Duration) *ClockScheduler {
	scheduler := &ClockScheduler{
		interval:  interval,
		running:   false,
		done:      make(chan bool),
		Scheduler: NewScheduler(),
	}

	return scheduler
}

func (s *ClockScheduler) Start() error {
	if s.running {
		return ErrAlreadyStarted
	}

	log.Printf("Starting Clock Scheduler (%p)", s)

	s.running = true
	go func() {
		ticker := time.NewTicker(s.interval)
		alive := true
		for alive {
			select {
			case <-ticker.C:
				s.Update()

			case <-s.done:
				alive = false
			}
		}
	}()

	return nil
}

func (s *ClockScheduler) Stop() error {
	if !s.running {
		return ErrAlreadyStopped
	}

	log.Printf("Stopping Clock Scheduler (%p)", s)

	s.done <- true
	s.running = false

	return nil
}

func (s *ClockScheduler) Running() bool {
	return s.running
}
