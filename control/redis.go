package control

import (
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"mania/logger"
	"mania/model"
	"time"
)

func ConnectRedis(dataSource *model.Redis) (*redis.Pool, error) {
	redisConn := &redis.Pool{
		MaxIdle:     500,
		MaxActive:   10000,
		IdleTimeout: 10 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", dataSource.Path,
				redis.DialConnectTimeout(time.Second*30),
				redis.DialWriteTimeout(time.Second*30),
				redis.DialReadTimeout(time.Second*30),
				redis.DialDatabase(dataSource.DB),
				redis.DialPassword(dataSource.Password))
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	conn := redisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("", "错误信息", errors.Unwrap(err), "调用栈：", err)
		}
	}(conn)

	_, err := conn.Do("PING")

	if err != nil {
		err = errors.Wrap(err, "ping redis failed")
		return nil, err
	}

	return redisConn, nil
}
