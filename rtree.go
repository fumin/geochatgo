package geochat

import (
	"errors"
	"fmt"
	"github.com/dhconnelly/rtreego"
	"sync"
)

type recvMsg_t struct {
	kind    string
	content []byte
}

type receiver_t struct {
	where   *rtreego.Rect
	key     string
	channel chan recvMsg_t
}

func (r *receiver_t) Bounds() *rtreego.Rect {
	return r.where
}

type client_t struct {
	receiver receiver_t
	popups   []*receiver_t
}

func (c *client_t) Bounds() *rtreego.Rect {
	return c.receiver.where
}

type rtree_t struct {
	sync.RWMutex
	rt     *rtreego.Rtree
	keyMap map[string]*client_t

	popupRtree *rtreego.Rtree
}

func NewRtree() *rtree_t {
	rt := rtreego.NewTree(2, 25, 50)
	keyMap := make(map[string]*client_t)
	popupRtree := rtreego.NewTree(2, 25, 50)
	return &rtree_t{rt: rt, keyMap: keyMap, popupRtree: popupRtree}
}

func (t *rtree_t) insert(key string, channel chan recvMsg_t, point, lengths [2]float64) error {
	rect := newRtreegoRect(point, lengths)
	newClient := &client_t{receiver_t{rect, key, channel}, []*receiver_t{}}
	t.Lock()
	defer t.Unlock()

	_, ok := t.keyMap[key]
	if ok {
		return errors.New(fmt.Sprintf("Object already exists for key: %v", key))
	}

	t.rt.Insert(newClient)
	t.keyMap[key] = newClient
	return nil
}

func (t *rtree_t) update(key string, point, lengths [2]float64) error {
	rect, err := rtreego.NewRect(point[:], lengths[:])
	if err != nil {
		panic(err)
	}
	t.Lock()
	defer t.Unlock()
	client, ok := t.keyMap[key]
	if !ok {
		return errors.New(fmt.Sprintf("No object for key %v in keyMap", key))
	}
	assertClientHasKey(client, key)

	ok = t.rt.Delete(client)
	if !ok {
		panic(fmt.Sprintf("Key exists in keyMap but not in rtree: %v", key))
	}
	newReceiver := receiver_t{rect, key, client.receiver.channel}
	newClient := &client_t{newReceiver, client.popups}
	t.rt.Insert(newClient)
	t.keyMap[key] = newClient
	return nil
}

func (t *rtree_t) del(key string) {
	t.Lock()
	defer t.Unlock()
	client, ok := t.keyMap[key]
	if ok {
		assertClientHasKey(client, key)

		// Cleanup popups in popupRtree belonging to this key.
		for _, popup := range client.popups {
			ok = t.popupRtree.Delete(popup)
			if !ok {
				panic(fmt.Sprintf("Popup not found in popupRtree: %v", client))
			}
		}

		ok = t.rt.Delete(client)
		if !ok {
			panic(fmt.Sprintf("Client found in keyMap but not in rtree: %v", client))
		}
		delete(t.keyMap, key)
	}
}

func (t *rtree_t) insertPopup(key string, popupPoint, popupLengths [2]float64) (string, error) {
	t.Lock()
	defer t.Unlock()
	client, ok := t.keyMap[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("No object for key %v in keyMap", key))
	}

	popupId := generateUniquePopupId(client.popups)
	rect := newRtreegoRect(popupPoint, popupLengths)
	popup := &receiver_t{rect, popupId, client.receiver.channel}

	client.popups = append(client.popups, popup)
	t.popupRtree.Insert(popup)
	return popupId, nil
}

func (t *rtree_t) delPopup(key, popupId string) {
	t.Lock()
	defer t.Unlock()
	client, ok := t.keyMap[key]
	if !ok {
		return
	}

	popupIndex := indexOfPopupWithId(client.popups, popupId)
	if popupIndex == -1 { // No popup with id popupId found.
		return
	}

	t.popupRtree.Delete(client.popups[popupIndex])
	// Delete the item at popupIndex from the slice of popups.
	l := len(client.popups)
	client.popups[popupIndex] = client.popups[l-1]
	client.popups = client.popups[0:(l - 1)]
}

func (t *rtree_t) nearestNeighbors(k int, point [2]float64) []*client_t {
	t.RLock()
	defer t.RUnlock()
	spacials := t.rt.NearestNeighbors(k, point[:])
	return removeNilsAndCastToClients(spacials)
}

const margin = 0.000001

func (t *rtree_t) searchContaining(k int, point [2]float64) []*receiver_t {
	rect, _ := rtreego.NewRect(point[:], []float64{margin, margin})
	t.RLock()
	defer t.RUnlock()
	spacials := t.popupRtree.SearchIntersectWithLimit(k, rect)
	return removeNilsAndCastToReceivers(spacials)
}

func indexOfPopupWithId(popups []*receiver_t, id string) int {
	popupIndex := -1
	for i, popup := range popups {
		if popup.key == id {
			popupIndex = i
			break
		}
	}
	return popupIndex
}

func assertClientHasKey(client *client_t, key string) {
	if client.receiver.key != key {
		panic(fmt.Sprintf("Inconsistent keys: %v, %v", client, key))
	}
}

func newRtreegoRect(point, lengths [2]float64) *rtreego.Rect {
	rect, err := rtreego.NewRect(point[:], lengths[:])
	if err != nil {
		panic(err)
	}
	return rect
}

func generateUniquePopupId(popups []*receiver_t) string {
	for {
		id := string(randByteSlice())
		index := indexOfPopupWithId(popups, id)
		if index == -1 {
			return id
		}
	}
}

func removeNilsAndCastToReceivers(spacials []rtreego.Spatial) []*receiver_t {
	receivers := make([]*receiver_t, len(spacials))
	i := 0
	for ; i != len(spacials); i++ {
		spacial := spacials[i]
		if spacial == nil {
			break
		}
		receiver, ok := spacial.(*receiver_t)
		if !ok {
			panic(fmt.Sprint("Non *receiver_t object %v stored in Rtree", spacial))
		}
		receivers[i] = receiver
	}

	return receivers[0:i]
}

func removeNilsAndCastToClients(spacials []rtreego.Spatial) []*client_t {
	clients := make([]*client_t, len(spacials))
	i := 0
	for ; i != len(spacials); i++ {
		spacial := spacials[i]
		if spacial == nil {
			break
		}
		client, ok := spacial.(*client_t)
		if !ok {
			panic(fmt.Sprint("Non *client_t object %v stored in Rtree", spacial))
		}
		clients[i] = client
	}

	return clients[0:i]
}
