package set

import (
	"fmt"
	"sync"
)

type Uint64 struct {
	data map[uint64]struct{}
	*sync.Mutex
}

func NewUint64Set() *Uint64 {
	set := &Uint64{
		data:  make(map[uint64]struct{}),
		Mutex: &sync.Mutex{},
	}

	return set
}

func Uint64SetOf(members ...uint64) *Uint64 {
	set := NewUint64Set()

	set.Insert(members...)

	return set
}

func (s *Uint64) Has(member uint64) bool {
	s.Lock()
	defer s.Unlock()

	_, exists := s.data[member]
	return exists
}

func (s *Uint64) HasAll(first uint64, rest ...uint64) bool {
	test := append([]uint64{first}, rest...)

	hasAll := true
	for i := 0; i < len(test) && hasAll; i++ {
		hasAll = s.Has(test[i])
	}

	return hasAll
}

func (s *Uint64) Insert(members ...uint64) {
	s.Lock()
	defer s.Unlock()

	for _, member := range members {
		s.data[member] = struct{}{}
	}
}

func (s *Uint64) Remove(member uint64) {
	s.Lock()
	defer s.Unlock()

	delete(s.data, member)
}

func (s *Uint64) Members() []uint64 {
	s.Lock()
	defer s.Unlock()

	members := make([]uint64, len(s.data))

	i := 0
	for member, _ := range s.data {
		members[i] = member
		i += 1
	}

	return members
}

func (s *Uint64) SubsetOf(universe *Uint64) bool {
	result := true
	for member, _ := range s.data {
		if exists := universe.Has(member); !exists {
			result = false
			break
		}
	}

	return result
}

func (s *Uint64) String() string {
	return fmt.Sprintf("%v", s.Members())
}
