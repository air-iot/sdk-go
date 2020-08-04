package redis

import (
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Client *redis.Client
var PoolSize int

func Init() {
	var (
		host     = viper.GetString("redis.host")
		port     = viper.GetInt("redis.port")
		password = viper.GetString("redis.password")
		db       = viper.GetInt("redis.db")
	)
	PoolSize = viper.GetInt("redis.poolSize")
	Client = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(host, strconv.Itoa(port)),
		Password: password, // no password set
		DB:       db,       // use default DB
		PoolSize: PoolSize,
		//MinIdleConns: 10,
		//PoolTimeout:10*time.Second
		//IdleTimeout:  2 * time.Second,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	p := Client.Ping()
	if p.Err() != nil {
		logrus.Panic(p.Err())
	}
}

func Close() {
	if Client != nil {
		if err := Client.Close(); err != nil {
			logrus.Errorf("Redis关闭失败:%s", err.Error())
		}
	}
}

// Set 保存数据
func Set(c *redis.Client, data map[string]interface{}, expire time.Duration) error {
	p := c.TxPipeline()
	for k, v := range data {
		p.Set(k, v, expire)
	}
	if _, err := p.Exec(); err != nil {
		return err
	}
	return nil
}

// Get 获取数据
func Get(c *redis.Client, keys ...string) (map[string]string, error) {
	p := c.Pipeline()

	query := make(map[string]*redis.StringCmd)
	for _, key := range keys {
		query[key] = p.Get(key)
	}
	if _, err := p.Exec(); err != nil {
		return nil, err
	}
	results := make(map[string]string)
	for k, cmd := range query {
		if cmd.Err() != nil {
			return nil, cmd.Err()
		}
		results[k] = cmd.Val()
	}

	return results, nil
}

// SetMap 保存map数据
func SetMap(c *redis.Client, key string, value map[string]interface{}) error {
	if cmd := c.HMSet(key, value); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

// SetMaps 保存map数组数据到map
func SetMaps(c *redis.Client, data map[string]map[string]interface{}) error {
	p := c.TxPipeline()
	for k, v := range data {
		p.HMSet(k, v)
	}
	if _, err := p.Exec(); err != nil {
		return err
	}
	return nil
}

// GetMaps 从map批量查询数据
func GetMaps(c *redis.Client, keys ...string) (map[string]map[string]string, error) {
	p := c.Pipeline()

	query := make(map[string]*redis.StringStringMapCmd)
	for _, key := range keys {
		query[key] = p.HGetAll(key)
	}
	if _, err := p.Exec(); err != nil {
		return nil, err
	}
	results := make(map[string]map[string]string)
	for k, cmd := range query {
		val, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		results[k] = val
	}

	return results, nil
}
