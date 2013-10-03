package geochat

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math"
)

const MaxZoom = 18

type redisConn struct {
	c   redis.Conn
	err error
}

func (c *redisConn) send(cmd string, args ...interface{}) {
	if c.err != nil {
		return
	}
	c.err = c.c.Send(cmd, args...)
}

func (c *redisConn) flush() {
	if c.err != nil {
		return
	}
	c.err = c.c.Flush()
}

func (c *redisConn) recv() interface{} {
	if c.err != nil {
		return nil
	}
	reply, err := c.c.Receive()
	c.err = err
	return reply
}

// Stores a jsonable object according to its latitude and longitude.
// Assuming the object is at latitude: 25.2, longitude: 121.4
// and the map tiles at different zoom levels enclosing this point are:
// | z  | x     | y     |
// | 1  | 1     | 0     |
// | 2  | 3     | 1     |
// ...
// | 15 | 27434 | 14012 |
//
// The object is stored repeatedly MaxZoom times in Redis under the keys
//   tile_chatlog:1:1:0
//   tile_chatlog:2:3:1
//   ...
//   tile_chatlog:15:27434:14012
//
func maptileStore(keyPrefix string, obj map[string]interface{},
	conn redis.Conn) error {
	latitude, ok := obj["latitude"].(float64)
	if !ok {
		panic(fmt.Sprintf("No latitude in %v", obj))
	}
	longitude, ok := obj["longitude"].(float64)
	if !ok {
		panic(fmt.Sprintf("No longitude in %v", obj))
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	p := LatLng{latitude, longitude}
	c := &redisConn{c: conn}

	for z := 1; z <= MaxZoom; z++ {
		x, y := LatLngToTile(z, p)
		redisKey := fmt.Sprintf("%v:%d:%d:%d", keyPrefix, z, x, y)
		c.send("LPUSH", redisKey, data)
		c.send("LTRIM", redisKey, 0, 499)
	}
	c.flush()
	for z := 1; z <= MaxZoom; z++ {
		c.recv() // response of LPUSH
		c.recv() // response of LTRIM
	}
	return c.err
}

func maptileRead(keyPrefix string, z int, x int, y int,
	offset int, limit int, conn redis.Conn) ([]string, error) {
	redisKey := fmt.Sprintf("%v:%d:%d:%d", keyPrefix, z, x, y)
	v, err := redis.Strings(conn.Do("LRANGE", redisKey, offset, limit))
	if err != nil {
		return make([]string, 0), err
	}
	return v, nil
}

// Utilities to transform between Latitude, Longitude and
// map tiles (x, y) on a particular zoom level.
//
// http://cloudmade.com/documentation/map-tiles
func LatLngToTile(z int, p LatLng) (int, int) {
	n := math.Pow(2, float64(z))
	x := ((p.Longitude + 180) / 360) * n
	lat := p.Latitude / 180 * math.Pi
	y := (1 - (math.Log(math.Tan(lat)+sec(lat)) / math.Pi)) / 2 * n
	return int(x), int(y)
}

func TileToLatLng(z int, x int, y int) LatLng {
	n := math.Pow(2, float64(z))
	longitude := float64(x)/n*360 - 180
	lat_radius := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	return LatLng{Latitude: lat_radius * 180 / math.Pi,
		Longitude: longitude}
}

func sec(x float64) float64 {
	return 1.0 / math.Cos(x)
}
