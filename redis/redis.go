package redis

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

type Conf struct {
	Host     string `json:"host"`
	Pool     int    `json:"pool"`
	Database int    `json:"database"`
}

type Redis struct {
	pool *pool.Pool
	db   int
}

func New(conf Conf) (*Redis, error) {
	pool, err := pool.New("tcp", conf.Host, conf.Pool)
	if err != nil {
		return nil, err
	}

	redis := &Redis{
		pool: pool,
		db:   conf.Database,
	}

	return redis, nil
}

func (r *Redis) Conn() (*redis.Client, error) {
	client, err := r.pool.Get()
	if err != nil {
		return client, err
	}

	// workaround :/
	client.Cmd("select", r.db)

	return client, nil
}

func (r *Redis) Release(conn *redis.Client) {
	r.pool.Put(conn)
}

// for testing
func (r *Redis) FlushAll() error {
	conn, err := r.Conn()
	if err != nil {
		return err
	}
	defer r.Release(conn)

	return conn.Cmd("flushdb").Err
}
