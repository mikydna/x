package link2

import (
	"net/http"
)

import (
	"github.com/mikydna/x/redis"
)

type RedisExpander struct {
	client *http.Client

	*redis.Redis
}

func NewRedisExpander(conf redis.Conf, client *http.Client) (*RedisExpander, error) {
	rdb, err := redis.New(conf)
	if err != nil {
		return nil, err
	}

	e := &RedisExpander{
		client: client,
		Redis:  rdb,
	}

	return e, nil
}
