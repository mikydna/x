package concurrent

import (
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	semaphore := NewSemaphore(10)

	err := semaphore.Release(3)
	if err != nil {
		t.FailNow()
	}

	semaphore.Acquire(3)
	if semaphore.Available() != 0 {
		t.FailNow()
	}

	request := make(chan bool)
	go func() {
		semaphore.Acquire(1)
		request <- true
	}()

	select {
	case <-request:
		t.FailNow()
	case <-time.After(1 * time.Millisecond):
		// expected
	}

	semaphore.Release(1)
	select {
	case <-request:
		// expected
	case <-time.After(1 * time.Millisecond):
		t.FailNow()
	}

}
