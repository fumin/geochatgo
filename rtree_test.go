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
	tree.insert(key, make(chan recvMsg_t), [2]float64{0, 0}, [2]float64{0.1, 0.1})
	expectedRect, _ := rtreego.NewRect([]float64{0.0, 0.0}, []float64{0.1, 0.1})
	rect := tree.keyMap[key].receiver.where
	if !rect.Equal(expectedRect) {
		t.Errorf("Expected %v == %v", rect, expectedRect)
	}

	// The rectange of the key should be the updated one.
	err := tree.update(key, [2]float64{0.0, 0.0}, [2]float64{1.0, 1.0})
	if err != nil {
		t.Errorf("Expected tree.update to succeed, but go %v", err)
	}
	expectedRect, _ = rtreego.NewRect([]float64{0.0, 0.0}, []float64{1.0, 1.0})
	rect = tree.keyMap[key].receiver.where
	if !rect.Equal(expectedRect) {
		t.Errorf("Expected %v == %v", rect, expectedRect)
	}
}

func TestDel(t *testing.T) {
	tree := NewRtree()
	key := "key"
	tree.insert(key, make(chan recvMsg_t), [2]float64{0, 0}, [2]float64{0.1, 0.1})
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
	channelA := make(chan recvMsg_t, 1)
	channelB := make(chan recvMsg_t, 1)
	tree.insert("a", channelA, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	tree.insert("b", channelB, [2]float64{1, 1}, [2]float64{1.1, 1.1})

	// Nearest neighbors of {0.01, 0.01} should be [a, b]
	clients := tree.nearestNeighbors(10, [2]float64{0.01, 0.01})
	if l := len(clients); 2 != l {
		t.Errorf("len(%v) != 2", clients)
	}
	clients[0].receiver.channel <- recvMsg_t{"kind", []byte("hello")}
	if msg := <-channelA; string(msg.content) != "hello" {
		t.Errorf("%v != \"hello\"", msg)
	}
	clients[1].receiver.channel <- recvMsg_t{"kind", []byte("world")}
	if msg := <-channelB; string(msg.content) != "world" {
		t.Errorf("%v != \"world\"", msg)
	}

	// Nearest neighbors of {1.01, 1.01} should be [b]
	clients = tree.nearestNeighbors(1, [2]float64{1.01, 1.01})
	clients[0].receiver.channel <- recvMsg_t{"kind", []byte("go!")}
	if l := len(channelA); l != 0 {
		t.Errorf("%v != 0", l)
	}
	if msg := <-channelB; string(msg.content) != "go!" {
		t.Errorf("%v != \"go!\"", msg)
	}
}

func TestUpdateDoesNotAlterPopups(t *testing.T) {
	tree := NewRtree()
	channelA := make(chan recvMsg_t, 1)
	tree.insert("a", channelA, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	popupId, _ := tree.insertPopup("a", [2]float64{3, 3}, [2]float64{0.1, 0.1})

	tree.update("a", [2]float64{2.2, 2.2}, [2]float64{0.1, 0.1})

	clientA := tree.nearestNeighbors(1, [2]float64{1.9, 1.9})[0]
	popupA := clientA.popups[0]

	// Check that the key, rectangle, and channel of popupA is as expected.
	if popupId != popupA.key {
		t.Errorf("%v != %v", popupId, popupA.key)
	}

	expectedRect, _ := rtreego.NewRect([]float64{3.0, 3.0}, []float64{0.1, 0.1})
	rect := popupA.where
	if !rect.Equal(expectedRect) {
		t.Errorf("Expected %v == %v", rect, expectedRect)
	}

	popupA.channel <- recvMsg_t{"kind", []byte("hello")}
	if msg := <-channelA; string(msg.content) != "hello" {
		t.Errorf("%v != \"hello\"", msg)
	}
}

func TestDelPopup(t *testing.T) {
	tree := NewRtree()
	channel := make(chan recvMsg_t, 1)
	tree.insert("a", channel, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
	idA, _ := tree.insertPopup("a", [2]float64{1, 1}, [2]float64{0.1, 0.1})
	idB, _ := tree.insertPopup("a", [2]float64{1, 1}, [2]float64{0.2, 0.2})
	idC, _ := tree.insertPopup("a", [2]float64{1, 1}, [2]float64{0.3, 0.3})

	tree.delPopup("a", idB)

	// popups1 and popups2 should be fully identical.
	popups1 := tree.keyMap["a"].popups
	popups2 := tree.searchContaining(10, [2]float64{1.01, 1.01})
	for _, popups := range []([]*receiver_t){popups1, popups2} {
		if l := len(popups); l != 2 {
			t.Errorf("%v != 2", l)
		}
		if idA != popups[0].key {
			t.Errorf("%v != %v", idA, popups[0].key)
		}
		if idC != popups[1].key {
			t.Errorf("%v != %v", idC, popups[1].key)
		}

		if a, b := popups[0].channel, popups1[0].channel; a != b {
			t.Errorf("%v != %v", a, b)
		}
		if a, b := popups[1].channel, popups1[1].channel; a != b {
			t.Errorf("%v != %v", a, b)
		}
	}
}

func TestDelRemovesPopups(t *testing.T) {
	tree := NewRtree()
	for _, key := range []string{"a", "b"} {
		channel := make(chan recvMsg_t, 1)
		tree.insert(key, channel, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
		tree.insertPopup(key, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
		if key == "a" {
			tree.insertPopup(key, [2]float64{0.0, 0.0}, [2]float64{0.1, 0.1})
		}
	}
	// a has 2 popups and b has 1, thus the total of popups in the tree is 3.
	if l := tree.popupRtree.Size(); l != 3 {
		t.Errorf("%d != 3", l)
	}

	tree.del("a")

	// Since a is deleted, we are left the single popup of b.
	if l := tree.popupRtree.Size(); l != 1 {
		t.Errorf("%d != 1", l)
	}
}
