package geochat

import (
	"errors"
	"fmt"
	"github.com/dhconnelly/rtreego"
	"sync"
)

// Our custom implementation of rtreego's Spatial interface.
// This is the type to be stored in the rtree.
type receiver_t struct {
	where   *rtreego.Rect
	key     string
	channel chan []byte
}

func (r *receiver_t) Bounds() *rtreego.Rect {
	return r.where
}

type rtree_t struct {
	sync.RWMutex
	rt     *rtreego.Rtree
	keyMap map[string]*receiver_t
}

func NewRtree() *rtree_t {
	rt := rtreego.NewTree(2, 25, 50)
	return &rtree_t{rt: rt, keyMap: make(map[string]*receiver_t)}
}

func (t *rtree_t) insert(key string, channel chan []byte, point, lengths [2]float64) error {
	rect, err := rtreego.NewRect(point[:], lengths[:])
	if err != nil {
		return err
	}
	newReceiver := &receiver_t{rect, key, channel}
	t.Lock()
	defer t.Unlock()
	receiver, ok := t.keyMap[key]
	if ok {
		ok = t.rt.Delete(receiver)
		if !ok {
			panic(fmt.Sprintf("Object found in keyMap but not in rtree: %v", receiver))
		}
	}
	t.rt.Insert(newReceiver)
	t.keyMap[key] = newReceiver
	return nil
}

// Similar to insert, but returns an error if the key does not exist yet.
func (t *rtree_t) update(key string, point, lengths [2]float64) error {
	rect, err := rtreego.NewRect(point[:], lengths[:])
	if err != nil {
		return err
	}
	t.Lock()
	defer t.Unlock()
	receiver, ok := t.keyMap[key]
	if !ok {
		return errors.New(fmt.Sprintf("No object for key %v in keyMap", key))
	}
	if receiver.key != key {
		panic(fmt.Sprint("Inconsistent receiver keys: %v, %v", key, receiver))
	}
	ok = t.rt.Delete(receiver)
	newReceiver := &receiver_t{rect, key, receiver.channel}
	t.rt.Insert(newReceiver)
	t.keyMap[key] = newReceiver
	return nil
}

func (t *rtree_t) del(key string) {
	t.Lock()
	defer t.Unlock()
	receiver, ok := t.keyMap[key]
	if ok {
		ok = t.rt.Delete(receiver)
		if !ok {
			panic(fmt.Sprintf("Object found in keyMap but not in rtree: %v", receiver))
		}
		delete(t.keyMap, key)
	}
}

func (t *rtree_t) nearestNeighbors(k int, point [2]float64) []*receiver_t {
	t.RLock()
	defer t.RUnlock()
	neighbors := t.rt.NearestNeighbors(k, point[:])

	// Drop nil elements and cast rtreego.Spatial to *receiver_t
	receivers := make([]*receiver_t, len(neighbors))
	i := 0
	for ; i != len(neighbors); i++ {
		neighbor := neighbors[i]
		if neighbor == nil {
			break
		}
		receiver, ok := neighbor.(*receiver_t)
		if !ok {
			panic(fmt.Sprint("Non *receiver_t object %v stored in Rtree", neighbor))
		}
		receivers[i] = receiver
	}

	return receivers[0:i]
}