package geochat

import (
  "fmt"
  "os"
  "time"
  "github.com/garyburd/redigo/redis"
  "github.com/fumin/rtree"
)

var rtreeClient rtree.Client
var redisPool redis.Pool

func initConfig() {
  var redisServer string
  var redisPassword string
  var rtreeAddr string
  if os.Getenv("OPENSHIFT_APP_NAME") != "" {
    redisServer = os.Getenv("OPENSHIFT_REDIS_HOST") +
                    ":" + os.Getenv("OPENSHIFT_REDIS_PORT")
    redisPassword = os.Getenv("REDIS_PASSWORD")
    rtreeAddr = ":6389"
  } else {
    redisServer = ":6379"
    redisPassword = ""
    rtreeAddr = ":6389"
  }

  // Setup the global redisPool
  redisPool := &redis.Pool{
    MaxIdle: 3,
    IdleTimeout: 240 * time.Second,
    Dial: func () (redis.Conn, error) {
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
      return c, err
    },
    TestOnBorrow: func(c redis.Conn, t time.Time) error {
      _, err := c.Do("PING")
      return err
    },
  }
  conn := redisPool.Get()
  defer conn.Close()
  _, err := conn.Do("PING")
  if err != nil { panic(err) }

  // Setup the global rtreeClient
  rtreeClient, err := rtree.NewClient("tcp", rtreeAddr)
  if err != nil { panic(err) }
  _ , err = rtreeClient.RtreeNearestNeighbors(rtreekeyUser, 1, []float64{0, 0})
  if err != nil { panic(err) }
}

// Redis keys
const rediskeyGeo = "geo"
func rediskeyTileChatlog(z int, x int, y int) string {
  return fmt.Sprintf("tile_chatlog:%d:%d:%d", x, y, z)
}

// Rtree keys
const rtreekeyUser = "user"
