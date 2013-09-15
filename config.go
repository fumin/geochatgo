package geochat

import (
	"github.com/garyburd/redigo/redis"
	"os"
	"time"
)

var rTree *rtree_t
var redisServer string
var redisPassword string
var redisPool *redis.Pool

func initConfig() {
	rTree = NewRtree()

	if os.Getenv("OPENSHIFT_APP_NAME") != "" {
		redisServer = os.Getenv("OPENSHIFT_REDIS_HOST") +
			":" + os.Getenv("OPENSHIFT_REDIS_PORT")
		redisPassword = os.Getenv("REDIS_PASSWORD")
	} else {
		redisServer = ":6379"
		redisPassword = ""
	}

	redisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        NewRedisConn,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	if err != nil {
		panic(err)
	}
}

func NewRedisConn() (redis.Conn, error) {
	c, err := redis.Dial("tcp", redisServer)
	if err != nil {
		return nil, err
	}
	if redisPassword != "" {
		if _, err := c.Do("AUTH", redisPassword); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, nil
}

func NewRedisSubscriber(c redis.Conn, channel string) chan interface{} {
	sc := redis.PubSubConn{c}
	sc.Subscribe(channel)
	ch := make(chan interface{}, 32)
	go func() {
		for {
			v := sc.Receive()
			select {
			case ch <- v:
			default:
			}

			_, is_error := v.(error)
			if is_error {
				return
			}
		}
	}()
	return ch
}

// Redis keys/prefixes
const rediskeyGeo = "geo"
const rediskeyTileChatlog = "tile_chatlog"
