package link2

import (
	"fmt"
	"hash"
	"hash/fnv"
	"net/http"
	"sync"
)

import (
	"github.com/mikydna/x/redis"
)

func ToStrMap(r *Result) map[string]string {
	strMap := make(map[string]string)
	return strMap
}

func FromStrMap(map[string]string) *Result {
	return nil
}

// fix later: merge with x/stats
type stats struct {
	values map[string]float64
	*sync.Mutex
}

func (s *stats) incr(key string, incr float64) {
	s.Lock()
	defer s.Unlock()
	s.values[key] += incr
}

func (s *stats) get(key string) float64 {
	s.Lock()
	defer s.Unlock()
	return s.values[key]
}

type RedisExpander struct {
	hashf hash.Hash64
	stats *stats
	*Expander
	*redis.Redis
}

func NewRedisExpander(conf redis.Conf, client *http.Client, processor ContentFunc) (*RedisExpander, error) {
	r, err := redis.New(conf)
	if err != nil {
		return nil, err
	}

	e, err := NewExpander(client, processor)
	if err != nil {
		return nil, err
	}

	expander := &RedisExpander{
		hashf:    fnv.New64a(),
		stats:    &stats{make(map[string]float64), &sync.Mutex{}},
		Redis:    r,
		Expander: e,
	}

	return expander, nil
}

func (e *RedisExpander) Expand(ctx context.Context, url string) (*Result, error) {
	conn, rErr := e.Conn()
	if err != nil {
		return nil, rErr
	}
	defer e.Release(conn)

	// compute cache key
	e.hashf.Reset()
	e.hashf.Write([]byte(url))
	cacheKey := e.hashf.Sum64()

	// check cache
	rKey := fmt.Sprintf("link:%d", cacheKey)
	rMap, rErr := conn.Cmd("hgetall", rKey).Map()
	if rErr != nil {
		// db-level err, return immediately
		return nil, rErr
	}

	var result Result
	if len(rMap) == 0 {
		// miss
		e.stats.incr("miss", 1)

		// - expand link
		result, err = e.Expander.Expand(ctx, url)
		if err != nil {
			return nil, err
		}

		// - cache result
		if err := conn.Cmd("hmset", rKey, ToStrMap(result)); err != nil {
			return nil, err
		}

	} else {
		// hit
		e.stats.incr("hit", 1)

		result = FromStrMap(rMap)
	}

	return result, err
}

func (e *RedisExpander) Stats() map[string]float64 {
	copy := make(map[string]float64)

	for key, value := range e.stats.values {
		copy[key] = value
	}

	return copy
}
