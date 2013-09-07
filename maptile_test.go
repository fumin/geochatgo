package geochat

import "testing"

func TestLatLngToTile(t *testing.T) {
	p := LatLng{Latitude: 41.850033, Longitude: -87.65005229999997}
	x, y := LatLngToTile(7, p)
	if !(x == 32 && y == 47) {
		t.Errorf("32 != %d, 47 != %d", x, y)
	}

	x, y = LatLngToTile(14, p)
	if !(x == 4202 && y == 6091) {
		t.Errorf("4202 != %d, 6091 != %d", x, y)
	}
}

func TestTileToLatLng(t *testing.T) {
	p := TileToLatLng(14, 4202, 6091)
	if !(41.837 < p.Latitude && p.Latitude < 41.854) {
		t.Errorf("latitude: %v", p.Latitude)
	}
	if !(-87.671 < p.Longitude && p.Longitude < -87.649) {
		t.Errorf("longitude: %v", p.Longitude)
	}
}
