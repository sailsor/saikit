package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"

	"code.jshyjdtech.com/godev/hykit/config"
	"github.com/stretchr/testify/assert"
)

func TestRedisClient_Funcs(t *testing.T) {
	var (
		it     = assert.New(t)
		ctx    = context.Background()
		client *Client
	)

	options := config.ViperConfOptions{}
	conf := config.NewViperConfig(options.WithConfigType("yaml"),
		options.WithConfFile([]string{"../config/a.yaml"}))
	redisClientOptions := ClientOptions{}
	conf.Set("debug", false)

	it.NotPanics(func() {
		client = NewClient(
			redisClientOptions.WithConf(conf),
			redisClientOptions.WithStateTicker(10*time.Microsecond),
		)
	})

	time.Sleep(100 * time.Millisecond)

	// string
	{
		key := "test_string"
		val := `{"name": "HY", "sex": "man"}`
		err := client.Set(ctx, key, val, 3)
		it.Nil(err)

		time.Sleep(3 * time.Second)

		result, err := client.Get(ctx, key)
		it.Nil(err)
		it.Equal(val, result)

		buf, err := client.GetBytes(ctx, key)
		it.Nil(err)
		fmt.Println(string(buf))
	}

	// hash
	{
		key := "test_hash"
		val := `{"name": "HY", "sex": "man"}`
		err := client.HSet(ctx, key, "h01", val)
		it.Nil(err)

		buf, err := client.HGet(ctx, key, "h01")
		it.Nil(err)
		it.Equal(val, string(buf))

		m := make(map[string]interface{})
		m["name"] = "huang"
		m["age"] = 23
		err = client.HMSet(ctx, key, m)
		it.Nil(err)

		keys := []interface{}{"name", "age"}
		aa, err := client.HMGet(ctx, key, keys...)
		it.Nil(err)
		fmt.Println(aa)

		type msTest struct {
			Key string
			Val string
		}
		type msTest2 struct {
			Elm []msTest
		}
		var ms = new(msTest2)
		err = client.HGetAll(ctx, key, ms)
		fmt.Printf("%+v\n", ms)
		it.Nil(err)
	}

	// list
	{
		key := "list_test"
		val := "1"
		err := client.LPush(ctx, key, val)
		it.Nil(err)

		v, err := redis.String(client.LPop(ctx, key))
		it.Nil(err)
		it.Equal("1", v)
	}

	// keys
	{
		key := "list_test"
		err := client.Expire(ctx, key, 5)
		it.Nil(err)

		time.Sleep(time.Second)
		b, err := client.Exists(ctx, key)
		it.Nil(err)
		it.Equal(false, b)
	}
}
