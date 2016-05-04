package link2

import (
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"net/http"
	"sync"
	"time"
)

import (
	"github.com/mikydna/x/redis"
	"golang.org/x/net/context"
)

const (
	DefaultCacheExpire      = 24 * time.Hour
	DefaultCacheErrorExpire = 1 * time.Hour
)

// fix later: merge with x/stats
// - use rwmutex
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

	expander := &RedisExpander{
		hashf:    fnv.New64a(),
		stats:    &stats{make(map[string]float64), &sync.Mutex{}},
		Redis:    r,
		Expander: NewExpander(client, processor),
	}

	return expander, nil
}

func (e *RedisExpander) Expand(ctx context.Context, url string) (result *Result, err error) {
	conn, connErr := e.Conn()
	if err != nil {
		err = connErr
		return
	}
	defer e.Release(conn)

	// compute cache key
	e.hashf.Reset()
	e.hashf.Write([]byte(url))
	cacheKey := e.hashf.Sum64()

	// check cache
	rKey := fmt.Sprintf("link:%d", cacheKey)
	rMap, cmdErr := conn.Cmd("hgetall", rKey).Map()
	if err != nil {
		err = cmdErr
		return
	}

	if len(rMap) == 0 { // miss
		var expandErr error

		result, expandErr = e.Expander.Expand(ctx, url)
		if expandErr != nil {
			err = expandErr
		}

		pipeLen := 0

		if result != nil {
			vals, _ := result.MarshalStringMap()
			conn.PipeAppend("hmset", rKey, vals)
			pipeLen += 1

			conn.PipeAppend("expire", rKey, DefaultCacheExpire.Seconds())
			pipeLen += 1
		}

		if expandErr != nil {
			conn.PipeAppend("hset", rKey, "ERR", err)
			pipeLen += 1

			conn.PipeAppend("expire", rKey, DefaultCacheErrorExpire.Seconds())
			pipeLen += 1
		}

		for i := 0; i < pipeLen; i++ {
			if pipeErr := conn.PipeResp().Err; pipeErr != nil {
				err = pipeErr
				return
			}
		}

		e.stats.incr("miss", 1)

	} else { // hit

		if str, exists := rMap["ERR"]; exists {
			cachedErr := errors.New(str)
			err = cachedErr

		} else {
			result = &Result{}
			if decodeErr := result.UnmarshalStringMap(rMap); decodeErr != nil {
				err = decodeErr
				return
			}

		}

		e.stats.incr("hit", 1)

	}

	return
}

func (e *RedisExpander) Stats() map[string]float64 {
	copy := make(map[string]float64)

	for key, value := range e.stats.values {
		copy[key] = value
	}

	return copy
}
