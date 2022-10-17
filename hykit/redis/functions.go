package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// ------------------------------------------------------------ //
// --------------------------- STRING ------------------------- //
// ----------------------------------------------------------- //
// GET
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.String(redisConn.Do(ctx, "GET", key))
}

func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.Bytes(redisConn.Do(ctx, "GET", key))
}

// SET, expiration 单位s, 0永久
func (c *Client) Set(ctx context.Context, key string, val interface{}, expiration int64) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	if expiration == 0 {
		_, err := redisConn.Do(ctx, "SET", key, val)
		return err
	}

	_, err := redisConn.Do(ctx, "SET", key, val, "EX", expiration)
	return err
}

// INCRBY
func (c *Client) IncrBy(ctx context.Context, key string, step int) (int, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.Int(redisConn.Do(ctx, "INCR", key, step))
}

// ------------------------------------------------------------ //
// --------------------------- HASH ------------------------- //
// ----------------------------------------------------------- //
// SET
func (c *Client) HSet(ctx context.Context, key string, field string, val interface{}) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	value, err := c.encode(val)
	if err != nil {
		return err
	}
	_, err = redisConn.Do(ctx, "HSET", key, field, value)

	return err
}

func (c *Client) HMSet(ctx context.Context, key string, val interface{}) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	_, err := redisConn.Do(ctx, "HMSET", redis.Args{}.Add(key).AddFlat(val)...)
	return err
}

func (c *Client) HGet(ctx context.Context, key, field string) ([]byte, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.Bytes(redisConn.Do(ctx, "HGET", key, field))
}

func (c *Client) HMGet(ctx context.Context, key string, field ...interface{}) ([]string, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.Strings(redisConn.Do(ctx, "HMGET", redis.Args{}.Add(key).AddFlat(field)...))
}

func (c *Client) HGetAll(ctx context.Context, key string, val interface{}) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()
	v, err := redis.Values(redisConn.Do(ctx, "HGETALL", key))
	if err != nil {
		return err
	}

	return redis.ScanStruct(v, val)
}

// ------------------------------------------------------------ //
// --------------------------- LIST ------------------------- //
// ----------------------------------------------------------- //
// timeout 单位s， 0表示无限期阻塞
func (c *Client) BLPop(ctx context.Context, key string, timeout int) (interface{}, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	values, err := redis.Values(redisConn.Do(ctx, "BLPOP", key, timeout))
	if err != nil {
		return nil, err
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("redisgo: unexpected number of values, got %d", len(values))
	}

	return values[1], err
}

func (c *Client) BRPop(ctx context.Context, key string, timeout int) (interface{}, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	values, err := redis.Values(redisConn.Do(ctx, "BRPOP", key, timeout))
	if err != nil {
		return nil, err
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("redisgo: unexpected number of values, got %d", len(values))
	}

	return values[1], err
}

func (c *Client) LPop(ctx context.Context, key string) (interface{}, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redisConn.Do(ctx, "LPOP", key)
}

func (c *Client) RPop(ctx context.Context, key string) (interface{}, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redisConn.Do(ctx, "RPOP", key)
}

// LPush 将一个值插入到列表头部
func (c *Client) LPush(ctx context.Context, key string, val interface{}) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	value, err := c.encode(val)
	if err != nil {
		return err
	}

	_, err = redisConn.Do(ctx, "LPUSH", key, value)

	return err
}

// RPush 将一个值插入到列表尾部
func (c *Client) RPush(ctx context.Context, key string, val interface{}) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	value, err := c.encode(val)
	if err != nil {
		return err
	}

	_, err = redisConn.Do(ctx, "RPUSH", key, value)
	return err
}

// 区间以偏移量 start 和 end
// 0 表示列表的第一个元素， 1 表示列表的第二个元素，以此类推。
// -1 表示列表的最后一个元素， -2 表示列表的倒数第二个元素，以此类推。
// end (闭区间)
func (c *Client) LRange(ctx context.Context, key string, start, end int) (interface{}, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redisConn.Do(ctx, "LRANGE", key, start, end)
}

// ------------------------------------------------------------ //
// --------------------------- KEYS ------------------------- //
// ----------------------------------------------------------- //

// key 设置过期时间  单位s
func (c *Client) Expire(ctx context.Context, key string, expiration int64) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	_, err := redis.Bool(redisConn.Do(ctx, "EXPIRE", key, expiration))

	return err
}

// 判断key是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.Bool(redisConn.Do(ctx, "EXISTS", key))
}

// 删除key
func (c *Client) DeleteKey(ctx context.Context, key string) (bool, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	return redis.Bool(redisConn.Do(ctx, "DEL", key))
}

func marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// encode 序列化要保存的值
func (c *Client) encode(val interface{}) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	default:
		b, err := marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}
	return value, nil
}

func (c *Client) SelectDB(ctx context.Context, db int) error {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()
	_, err := redisConn.Do(ctx, "SELECT", db)
	return err
}
