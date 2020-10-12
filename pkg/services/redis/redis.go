package redis

import (
	"errors"
	"time"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/cfg"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

var redisConnections = make(map[string]*redis.Client)

func newRedis(addr string, password string, db int) (*redis.Client, error) {
	log.Info("connect redis: " + addr)
	redisPool := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
		//DialTimeout:  10 * time.Second,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
		//PoolSize:     10,
		//PoolTimeout:  30 * time.Second,
	})

	if err := redisPool.Ping().Err(); err != nil {
		return nil, errors.New("redis connect failed: " + err.Error())
	}

	time.Now().Format("20")
	return redisPool, nil
}

func InitConnections() {
	conn := "default"
	config := cfg.AppConfig.Redis
	redisPool, err := newRedis(config.Addr, config.Password, config.Db)
	if err != nil {
		log.Panic("conn: " + conn + " newRedis redis err: " + err.Error())
	}

	redisConnections[conn] = redisPool
}

func Connection(conn string) *redis.Client {
	if conn == "" {
		conn = "default"
	}

	return redisConnections[conn]
}

func Publish(conn string, channel string, message interface{}) error {
	if err := Connection(conn).Publish(channel, message).Err(); err != nil {
		log.Error("redis publish error, channel: "+channel, zap.Error(err), zap.Any("message", message))
		return err
	}

	log.Debug("redis publish channel:"+channel, zap.Any("message", message))
	return nil
}
