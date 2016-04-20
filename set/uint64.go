package set

import (
	"fmt"
)

type Uint64 map[uint64]struct{}

func Uint64SetOf(members ...uint64) Uint64 {
	set := make(Uint64)
	set.Insert(members...)
	return set
}

func (s Uint64) Has(member uint64) bool {
	_, exists := s[member]
	return exists
}

func (s Uint64) HasAll(first uint64, rest ...uint64) bool {
	test := append([]uint64{first}, rest...)

	hasAll := true
	for i := 0; i < len(test) && hasAll; i++ {
		hasAll = s.Has(test[i])
	}

	return hasAll
}

func (s Uint64) Insert(members ...uint64) {
	for _, member := range members {
		s[member] = struct{}{}
	}
}

func (s Uint64) Remove(member uint64) {
	delete(s, member)
}

func (s Uint64) Members() []uint64 {
	members := make([]uint64, len(s))

	i := 0
	for member, _ := range s {
		members[i] = member
		i += 1
	}

	return members
}

func (s Uint64) SubsetOf(universe Uint64) bool {
	result := true
	for member, _ := range s {
		if _, exists := universe[member]; !exists {
			result = false
			break
		}
	}

	return result
}

func (s Uint64) String() string {
	return fmt.Sprintf("%v", s.Members())
}
