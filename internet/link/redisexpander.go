package link

import (
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"log"
	"net/url"
	"strconv"
	"time"
)

import (
	"github.com/mikydna/x/redis"
	"github.com/mikydna/x/stats"
)

var (
	DefaultExpire          = (24 * 7 * time.Hour).Seconds()
	DefaultExpireForErrors = (1 * time.Hour).Seconds()
)

type RedisExpander struct {
	hash     hash.Hash64
	expander Expander
	stats    map[string]float64

	*redis.Redis
}

func NewRedisExpander(conf redis.Conf, expander Expander) (*RedisExpander, error) {
	redis, err := redis.New(conf)
	if err != nil {
		return nil, err
	}

	redisExpander := &RedisExpander{
		hash:     fnv.New64(),
		expander: expander,
		stats:    make(map[string]float64),
		Redis:    redis,
	}

	return redisExpander, nil
}

func (e *RedisExpander) Expand(rawurl string) *Result {
	parsed, err := url.Parse(rawurl)
	if err != nil {
		return &Result{Err: err}
	}

	norm := Normalize(parsed)
	norm.RawQuery = ""

	e.hash.Reset()
	e.hash.Write([]byte(norm.String()))

	conn, err := e.Conn()
	defer e.Release(conn)

	if err != nil {
		return &Result{Err: err}
	}

	hash := e.hash.Sum64()
	key := fmt.Sprintf("link:%d:expand", hash)

	conn.PipeAppend("hgetall", key)
	cacheEntry, err := conn.PipeResp().Map()
	if err != nil {
		return &Result{Err: err}
	}

	hit := len(cacheEntry) > 0 && err == nil

	var result *Result
	if hit {
		// e.stats["hit"] += 1

		// create result from redis response
		var target *url.URL
		var storedErr error
		var responseTime time.Duration
		var statusCode int64

		if cachedUrl := cacheEntry["url"]; cachedUrl != "" {
			target, _ = url.Parse(cachedUrl)
		}

		if cachedErrStr := cacheEntry["error"]; cachedErrStr != "" {
			storedErr = errors.New(cachedErrStr)
		}

		if cachedResponseTime := cacheEntry["responseTime"]; cachedResponseTime != "" {
			responseTimeNs, _ := strconv.ParseInt(cachedResponseTime, 10, 64)
			responseTime = time.Duration(responseTimeNs) * time.Nanosecond
		}

		if cachedStatusCode := cacheEntry["statusCode"]; cachedStatusCode != "" {
			statusCode, _ = strconv.ParseInt(cachedStatusCode, 10, 64)
		}

		result = &Result{
			URL:          target,
			Domain:       cacheEntry["domain"],
			Err:          storedErr,
			StatusCode:   int(statusCode),
			ResponseTime: responseTime,
		}

	} else /* miss */ {
		// e.stats["miss"] += 1

		result = e.expander.Expand(rawurl)
		expire := DefaultExpire
		entry := make(map[string]string)

		if result.URL != nil {
			entry["url"] = result.URL.String()
			entry["domain"] = result.Domain
		}

		if result.ResponseTime != 0 {
			entry["responseTime"] = fmt.Sprintf("%d", result.ResponseTime.Nanoseconds())
		}

		if result.StatusCode != 0 {
			entry["statusCode"] = fmt.Sprintf("%d", result.StatusCode)
		}

		if err := result.Err; err != nil {
			entry["error"] = err.Error()
			expire = DefaultExpireForErrors
		}

		pipeLen := 0

		conn.PipeAppend("hmset", key, entry)
		pipeLen += 1

		conn.PipeAppend("expire", key, expire)
		pipeLen += 1

		for i := 0; i < pipeLen; i++ {
			if err := conn.PipeResp().Err; err != nil {
				log.Println(err)
			}
		}
	}

	return result
}

func (e *RedisExpander) Stats() *stats.Stat {
	stat := stats.NewStat(time.Now())
	for key, val := range e.stats {
		stat.Set(key, val)
	}

	return stat
}
