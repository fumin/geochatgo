package geochat

import (
	"github.com/dhconnelly/rtreego"
	"testing"
)

func TestUpdateError(t *testing.T) {
	tree := NewRtree()
	key := "key"
	err := tree.update(key, [2]float64{0.0, 0.0}, [2]float64{1.0, 1.0})
	if err == nil {
		t.Errorf("Expected tree.update to return error since key doesn't exists")
	}
}

func TestUpdate(t *testing.T) {
	// The rectangle of the key should be the same as inserted.
	tree := NewRtree()
	key := "key"
	tree.insert(key, make(chan []byte), [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	expectedRect, _ := rtreego.NewRect([]float64{0.0, 0.0}, []float64{0.1, 0.1})
	rect := tree.keyMap[key].where
	if !rect.Equal(expectedRect) {
		t.Errorf("Expected %v == %v", rect, expectedRect)
	}

	// The rectange of the key should be the updated one.
	err := tree.update(key, [2]float64{0.0, 0.0}, [2]float64{1.0, 1.0})
	if err != nil {
		t.Errorf("Expected tree.update to succeed, but go %v", err)
	}
	expectedRect, _ = rtreego.NewRect([]float64{0.0, 0.0}, []float64{1.0, 1.0})
	rect = tree.keyMap[key].where
	if !rect.Equal(expectedRect) {
		t.Errorf("Expected %v == %v", rect, expectedRect)
	}
}

func TestDel(t *testing.T) {
	tree := NewRtree()
	key := "key"
	tree.insert(key, make(chan []byte), [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	_, ok := tree.keyMap[key]
	if !ok {
		t.Errorf("Expected %v in %v, but no found.", key, tree.keyMap)
	}
	if size := tree.rt.Size(); size != 1 {
		t.Errorf("%v != 1", size)
	}

	tree.del(key)
	_, ok = tree.keyMap[key]
	if ok {
		t.Errorf("Expected %v not in %v, but found.", key, tree.keyMap)
	}
	if size := tree.rt.Size(); size != 0 {
		t.Errorf("%v != 0", size)
	}
}

func TestNearestNeighbors(t *testing.T) {
	tree := NewRtree()
	channelA := make(chan []byte, 1)
	channelB := make(chan []byte, 1)
	tree.insert("a", channelA, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	tree.insert("b", channelB, [2]float64{1, 1}, [2]float64{1.1, 1.1})

	// Nearest neighbors of {0.01, 0.01} should be [a, b]
	receivers := tree.nearestNeighbors(10, [2]float64{0.01, 0.01})
	if l := len(receivers); 2 != l {
		t.Errorf("len(%v) != 2", receivers)
	}
	receivers[0].channel <- []byte("hello")
	if msg := string(<-channelA); msg != "hello" {
		t.Errorf("%v != \"hello\"", msg)
	}
	receivers[1].channel <- []byte("world")
	if msg := string(<-channelB); msg != "world" {
		t.Errorf("%v != \"world\"", msg)
	}

	// Nearest neighbors of {1.01, 1.01} should be [b]
	receivers = tree.nearestNeighbors(1, [2]float64{1.01, 1.01})
	receivers[0].channel <- []byte("go!")
	if l := len(channelA); l != 0 {
		t.Errorf("%v != 0", l)
	}
	if msg := string(<-channelB); msg != "go!" {
		t.Errorf("%v != \"go!\"", msg)
	}
}
