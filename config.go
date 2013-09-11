package geochat

import (
	"github.com/dhconnelly/rtreego"
	"github.com/fumin/rtree"
	"github.com/garyburd/redigo/redis"
	"os"
	"sync"
	"time"
)

var rTree *rtree_t
var redisServer string
var redisPassword string
var redisPool *redis.Pool

func initConfig() {
	rTree = &rtree_t{t: rtree.NewTree(2)}

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

type rtree_t struct {
	sync.RWMutex
	t *rtree.Rtree
}

func (rt *rtree_t) insert(key string, point, lengths []float64) {
	rect, err := rtreego.NewRect(point, lengths)
	if err != nil {
		panic(err)
	}
	rt.Lock()
	defer rt.Unlock()
	rt.t.Insert(key, rect)
}

func (rt *rtree_t) del(key string) {
	rt.Lock()
	defer rt.Unlock()
	rt.t.Delete(key)
}

func (rt *rtree_t) update(key string, point, lengths []float64) {
	rect, err := rtreego.NewRect(point, lengths)
	if err != nil {
		panic(err)
	}
	rt.Lock()
	defer rt.Unlock()
	rt.t.Update(key, rect)
}

func (rt *rtree_t) nearestNeighbors(k int, point []float64) []string {
	rt.RLock()
	defer rt.RUnlock()
	return rt.t.NearestNeighbors(k, point)
}
