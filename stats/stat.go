package stats

import (
	"errors"
	"math"
	"time"
)

var (
	ErrNoSuchStat = errors.New("No such stat")
)

type Values map[string]float64

type Stat struct {
	RecordedAt time.Time `json:"recorded_at"`
	Values     Values    `json:"values"`
}

func NewStat(recordedAt time.Time) *Stat {
	return &Stat{
		RecordedAt: recordedAt,
		Values:     make(Values),
	}
}

func (s Stat) Get(name string) (float64, error) {
	if val, exists := s.Values[name]; exists {
		return val, nil
	} else {
		return math.SmallestNonzeroFloat64, ErrNoSuchStat
	}
}

func (s Stat) Set(name string, val float64) {
	s.Values[name] = val
}
