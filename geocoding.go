package geochat

import "fmt"

type LatLng struct {
  Latitude float64
  Longitude float64
}

// Accuracy of different Geohash precision levels
// http://en.wikipedia.org/wiki/Geohash
//
// | geohash length | lat error  | lng error | km error
// |       1        | +-23       | +-23      | +-2500
// |       2        | +-2.8      | +-5.6     | +-630
// |       3        | +-0.7      | +-0.7     | +-78
// |       4        | +-0.087    | +-0.18    | +-20
// |       5        | +-0.022    | +-0.022   | +-2.4
// |       6        | +-0.0027   | +-0.0055  | +-0.61
// |       7        | +-0.00068  | +-0.00068 | +-0.076
// |       8        | +-0.000085 | +-0.00017 | +-0.019
//
const (
  GeohashBitsInChar = 5 // Geohash uses base32 encoding, thus 5 bits in a char
  MaxGeohashPrecision = 8 // Max precision is 8 for memory and speed reasons
)

func LatLngToInt(p LatLng, precision int) uint64 {
  numberOfZeroPaddingBits := GeohashBitsInChar*MaxGeohashPrecision - precision
  if numberOfZeroPaddingBits < 0 {
    panic(fmt.Sprintf("Max precision exceeded: %d", numberOfZeroPaddingBits))
  }

  latrange := [2]float64{-90, 90}
  lngrange := [2]float64{-180, 180}
  var mid float64 = 0
  var n uint64 = 0

  for i := 0; i != precision; i++ {
    if i % 2 != 0 {
      mid = (latrange[0] + latrange[1]) / 2
      if p.Latitude > mid {
        latrange[0] = mid
        n |= 1
      } else {
        latrange[1] = mid
      }
    } else {
      mid = (lngrange[0] + lngrange[1]) / 2
      if p.Longitude > mid {
        lngrange[0] = mid
        n |= 1
      } else {
        lngrange[1] = mid
      }
    }

    if i != precision - 1 {
      n <<= 1
    }
  }

  return n << uint8(numberOfZeroPaddingBits)
}
