// Utilities to transform between Latitude, Longitude and
// map tiles (x, y) on a particular zoom level.
//
// http://cloudmade.com/documentation/map-tiles

package geochat

import "math"

func sec(x float64) float64 {
  return 1.0 / math.Cos(x)
}

func LatLngToTile(z uint8, p LatLng) (int, int) {
  n := math.Pow(2, float64(z))
  x := ((float64(p.Longitude) + 180) / 360) * n
  lat := float64(p.Latitude / 180 * math.Pi)
  y := (1 - (math.Log(math.Tan(lat) + sec(lat)) / math.Pi)) / 2 * n
  return int(x), int(y)
}

func TileToLatLng(z uint8, x int, y int) LatLng {
  n := math.Pow(2, float64(z))
  longitude := float64(x) / n * 360 - 180
  lat_radius := math.Atan(math.Sinh(math.Pi*(1 - 2 * float64(y) / n)))
  return LatLng{Latitude: float32(lat_radius * 180 / math.Pi),
                Longitude: float32(longitude)}
}
