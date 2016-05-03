package link2

import (
	"fmt"
	"hash"
	"hash/fnv"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

import (
	"github.com/mikydna/x/redis"
	"golang.org/x/net/context"
)

const (
	DefaultCacheExpire = 24 * time.Hour
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

// fix: MarshalRedisMap(r) ???
func toStrMap(r *Result) map[string]string {
	strMap := make(map[string]string)
	strMap["statusCode"] = fmt.Sprintf("%d", r.StatusCode)
	strMap["responseTime"] = fmt.Sprintf("%d", r.ResponseTime.Nanoseconds())
	strMap["resolvedURL"] = r.ResolvedURL.String()

	for key, val := range r.Content {
		rhkey := fmt.Sprintf("c_%d", key)
		strMap[rhkey] = val
	}

	return strMap
}

// fix: UnmarshalRedisMap(r) ???
func fromStrMap(rmap map[string]string) (*Result, error) {
	var result Result

	if str, exists := rmap["statusCode"]; exists {
		statusCode, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return nil, err
		}

		result.StatusCode = int(statusCode)
	}

	if str, exists := rmap["responseTime"]; exists {
		responseTime, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}

		result.ResponseTime = time.Duration(responseTime)
	}

	if str, exists := rmap["resolvedURL"]; exists {
		resolvedURL, err := url.Parse(str)
		if err != nil {
			return nil, err
		}

		result.ResolvedURL = resolvedURL
	}

	result.Content = make(Content)
	for _, key := range []ContentType{Title, Description} {
		rhkey := fmt.Sprintf("c_%d", key)
		if str, exists := rmap[rhkey]; exists {
			result.Content[key] = str
		}
	}

	return &result, nil
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

func (e *RedisExpander) Expand(ctx context.Context, url string) (*Result, error) {
	conn, err := e.Conn()
	if err != nil {
		return nil, err
	}
	defer e.Release(conn)

	// compute cache key
	e.hashf.Reset()
	e.hashf.Write([]byte(url))
	cacheKey := e.hashf.Sum64()

	// check cache
	rKey := fmt.Sprintf("link:%d", cacheKey)
	rMap, err := conn.Cmd("hgetall", rKey).Map()
	if err != nil {
		// db-level err, return immediately
		return nil, err
	}

	var result *Result
	if len(rMap) == 0 { // miss
		result, err = e.Expander.Expand(ctx, url)
		if err != nil {
			// fix: should cache errors
			return nil, err
		}

		pipeLen := 0

		conn.PipeAppend("hmset", rKey, toStrMap(result))
		pipeLen += 1

		conn.PipeAppend("expire", rKey, DefaultCacheExpire.Seconds())
		pipeLen += 1

		for i := 0; i < pipeLen; i++ {
			if err := conn.PipeResp().Err; err != nil {
				return nil, err
			}
		}

		e.stats.incr("miss", 1)

	} else { // hit
		result, err = fromStrMap(rMap)
		if err != nil {
			return nil, err
		}

		e.stats.incr("hit", 1)
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
