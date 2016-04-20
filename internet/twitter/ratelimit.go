package twitter

import (
	"time"
)

type Resource string

type Stat struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

type RateLimit struct {
	CreatedAt time.Time
	Resources map[TwitterResourceKey]Stat
}

func FromTwitterRateLimit(obj TwitterRateLimit) *RateLimit {
	result := &RateLimit{
		CreatedAt: time.Now(),
		Resources: make(map[TwitterResourceKey]Stat),
	}

	for family, resource := range obj.Resources {
		for name, stat := range resource {
			key := TwitterResourceKey{family, name}

			result.Resources[key] = Stat{
				Limit:     stat.Limit,
				Remaining: stat.Remaining,
				Reset:     stat.Reset.Add(0),
			}
		}
	}

	return result
}
