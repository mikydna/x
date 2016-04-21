package job

import (
	"errors"
	"log"
	"sync"
)

import (
	"github.com/mikydna/x/set"
)

var (
	ErrJobAlreadySubmitted = errors.New("Job already exists")
	ErrJobAlreadyCompleted = errors.New("Job already completed")
)

type Job interface {
	Hash() uint64
	Run() (*Result, error)
}

type Result struct {
	Of   uint64
	Data interface{}
	Err  error
}

type Scheduler struct {
	waiting   map[uint64]*scheduled
	running   set.Uint64
	completed set.Uint64
	results   chan *Result
	listeners []chan *Result
	*sync.Mutex
}

type scheduled struct {
	job Job
	dep set.Uint64
}

type schedulerInfo struct {
	Waiting   set.Uint64
	Running   set.Uint64
	Completed set.Uint64
}

func NewScheduler() *Scheduler {
	scheduler := &Scheduler{
		waiting:   make(map[uint64]*scheduled),
		running:   make(set.Uint64),
		completed: make(set.Uint64),
		results:   make(chan *Result),
		listeners: []chan *Result{},
		Mutex:     &sync.Mutex{},
	}

	go func() {
		// fix: need a way to stop? move into clock?
		for result := range scheduler.results {
			for _, listener := range scheduler.listeners {
				select {
				case listener <- result:
				default:
					// remove listener?
				}
			}
		}
	}()

	return scheduler
}

func (s *Scheduler) Add(job Job, deps ...uint64) error {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.waiting[job.Hash()]; exists {
		return ErrJobAlreadySubmitted
	}

	if s.completed.Has(job.Hash()) {
		return ErrJobAlreadyCompleted
	}

	s.waiting[job.Hash()] = &scheduled{job, set.Uint64SetOf(deps...)}

	return nil
}

func (s *Scheduler) Listen() <-chan *Result {
	callback := make(chan *Result)
	s.listeners = append(s.listeners, callback)
	return callback
}

func (s *Scheduler) Info() schedulerInfo {
	waiting := make(set.Uint64)
	for hash, _ := range s.waiting {
		waiting.Insert(hash)
	}

	return schedulerInfo{
		Waiting:   waiting,
		Running:   s.running,
		Completed: s.completed,
	}
}

func (s *Scheduler) Update() {
	s.Lock()
	defer s.Unlock()

	for _, scheduled := range s.waiting {
		job := scheduled.job
		ready := scheduled.dep.SubsetOf(s.completed)

		if ready {
			delete(s.waiting, job.Hash())
			s.running.Insert(job.Hash())

			go func() {
				result, err := job.Run()
				if err != nil {
					log.Printf("[CRITICAL] %v", err)
				}

				s.completed.Insert(job.Hash())
				s.running.Remove(job.Hash())

				// fix
				if result != nil {
					s.results <- result
				}

			}()
		}
	}
}
