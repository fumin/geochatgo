package geochat

import (
  "testing"
  "strconv"
  "strings"
)

func TestLatLngToInt(t *testing.T) {
  totalBits := GeohashBitsInChar * MaxGeohashPrecision
  str := "1110011000101101011100111"
  strPadded := str + strings.Repeat("0", totalBits - len(str))
  i, _ := strconv.ParseUint(strPadded, 2, 64)

  x := LatLngToInt(LatLng{25.19, 121.44}, len(str));

  if i != x {
    t.Errorf("%v != %v", strPadded, strconv.FormatUint(x, 2))
  }
}
