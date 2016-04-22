package set

import (
	"sync"
	"testing"
)

func TestUint64_Concurrent(t *testing.T) {
	// blast the set, see if there is a concurrent mod exception

	set := NewUint64Set()
	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go func() {
		for i := 0; i < 1000000; i++ {
			set.Insert(uint64(i))
		}
		waitGroup.Done()
	}()

	waitGroup.Add(1)
	go func() {
		for i := 0; i < 1000000; i++ {
			set.SubsetOf(Uint64SetOf(uint64(i)))
		}
		waitGroup.Done()
	}()

	waitGroup.Wait()
}
