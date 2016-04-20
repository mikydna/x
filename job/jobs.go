package job

import (
	"time"
)

// for testing purposes; TODO move to test file
type TimeJob struct {
	hash    uint64
	timeout time.Duration
}

func NewTimeJob(hash uint64, timeout time.Duration) *TimeJob {
	job := &TimeJob{
		hash:    hash,
		timeout: timeout,
	}

	return job
}

func (j *TimeJob) Hash() uint64 {
	return j.hash
}

func (j *TimeJob) Run() (*Result, error) {
	time.Sleep(j.timeout)
	return &Result{j.hash, nil, nil}, nil
}
