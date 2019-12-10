////////////////////////////////////////////////////////////////////////////////
// Copyright © 2019 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package switchboard

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/primitives/id"
	"reflect"
	"strconv"
	"sync"
)

type Item interface {
	// To reviewer: Is this the correct name for this method? It's always the
	// sender ID in the client, but that might not be the case on the nodes
	GetSender() *id.User
	GetMessageType() int32
}

// This is an interface so you can receive callbacks through the Gomobile
// boundary
type Listener interface {
	Hear(item Item, isHeardElsewhere bool, i ...interface{})
}

type listenerRecord struct {
	l  Listener
	id string
	i  []interface{}
}

type Switchboard struct {
	// By matching with the keys for each level of the map,
	// you can find the listeners that meet each criterion
	listeners map[id.User]map[int32][]*listenerRecord
	lastID    int
	mux       sync.RWMutex
}

var Listeners = NewSwitchboard()

func NewSwitchboard() *Switchboard {
	return &Switchboard{
		listeners: make(map[id.User]map[int32][]*listenerRecord),
		lastID:    0,
	}
}

// Add a new listener to the map
// Returns ID of the new listener. Keep this around if you want to be able to
// delete the listener later.
//
// user: 0 for all,
// or any user ID to listen for messages from a particular user.
// messageType: 0 for all, or any message type to listen for messages of that
// type.
// newListener: something implementing the Listener callback interface.
// Don't pass nil to this.
//
// If a message matches multiple listeners, all of them will hear the message.
func (lm *Switchboard) Register(user *id.User,
	messageType int32, newListener Listener, i ...interface{}) string {
	lm.mux.Lock()
	defer lm.mux.Unlock()

	lm.lastID++
	if lm.listeners[*user] == nil {
		lm.listeners[*user] = make(map[int32][]*listenerRecord)
	}

	newListenerRecord := &listenerRecord{
		l:  newListener,
		id: strconv.Itoa(lm.lastID),
		i:  i,
	}
	lm.listeners[*user][messageType] = append(
		lm.listeners[*user][messageType],
		newListenerRecord)

	return newListenerRecord.id
}

func (lm *Switchboard) Unregister(listenerID string) {
	lm.mux.Lock()
	defer lm.mux.Unlock()

	// Iterate over all listeners in the map.
	for u, perUser := range lm.listeners {
		for messageType, perMessageType := range perUser {
			for i, listener := range perMessageType {
				if listener.id == listenerID {
					// this matches, so remove listener from data structure
					lm.listeners[u][messageType] = append(
						perMessageType[:i], perMessageType[i+1:]...)
					// since the id is unique per listener,
					// we can terminate the loop early.
					return
				}
			}
		}
	}
}

func (lm *Switchboard) matchListeners(item Item) []*listenerRecord {

	matches := make([]*listenerRecord, 0)

	// 8 cases total, for matching both specific and general listeners
	// This seems inefficient
	for _, listener := range lm.listeners[*item.GetSender()][item.GetMessageType()] {
		matches = appendIfUnique(matches, listener)
	}
	for _, listener := range lm.listeners[*id.ZeroID][item.GetMessageType()] {
		matches = appendIfUnique(matches, listener)
	}
	for _, listener := range lm.listeners[*item.GetSender()][0] {
		matches = appendIfUnique(matches, listener)
	}
	for _, listener := range lm.listeners[*id.ZeroID][0] {
		matches = appendIfUnique(matches, listener)
	}
	for _, listener := range lm.listeners[*item.GetSender()][0] {
		matches = appendIfUnique(matches, listener)
	}
	for _, listener := range lm.listeners[*id.ZeroID][0] {
		matches = appendIfUnique(matches, listener)
	}
	// Match all, but with generic outer type
	for _, listener := range lm.listeners[*item.GetSender()][item.GetMessageType()] {
		matches = appendIfUnique(matches, listener)
	}
	for _, listener := range lm.listeners[*id.ZeroID][item.GetMessageType()] {
		matches = appendIfUnique(matches, listener)
	}

	return matches
}

func appendIfUnique(matches []*listenerRecord, newListener *listenerRecord) []*listenerRecord {
	// Search for the listener ID
	found := false
	for _, l := range matches {
		found = found || (l.id == newListener.id)
	}
	if !found {
		// If we didn't find it, it's OK to append it to the slice
		return append(matches, newListener)
	} else {
		// We already matched this listener, and shouldn't append it
		return matches
	}
}

// Broadcast a message to the appropriate listeners
func (lm *Switchboard) Speak(item Item) {
	lm.mux.RLock()
	defer lm.mux.RUnlock()

	// Matching listeners include those that match all criteria perfectly,
	// as well as those that don't care about certain criteria.
	matches := lm.matchListeners(item)

	if len(matches) > 0 {
		// notify all normal listeners
		for _, listener := range matches {
			jww.INFO.Printf("Hearing on listener %v of type %v",
				listener.id, reflect.TypeOf(listener.l))
			// If you want to be able to hear things on the switchboard on
			// multiple goroutines, you should call Speak() on the switchboard
			// from multiple goroutines
			go listener.l.Hear(item, len(matches) > 1)
		}
	} else {
		jww.ERROR.Printf(
			"Message of type %v from user %q didn't match any listeners in"+
				" the map", item.GetMessageType(), item.GetSender())
		// dump representation of the map
		for u, perUser := range lm.listeners {
			for messageType, perMessageType := range perUser {
				for i, listener := range perMessageType {

					jww.ERROR.Printf("Listener %v: %v, user %v, "+
						" type %v, ",
						i, listener.id, u, messageType)
				}
			}
		}
	}
}
