////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package switchboard_test

import (
	"gitlab.com/elixxir/primitives/cmixproto"
	"gitlab.com/elixxir/primitives/id"
	"gitlab.com/elixxir/primitives/switchboard"
	"sync"
	"testing"
	"time"
)

// Demonstrates that all the messages can be heard when multiple threads are
// producing items
func TestListeningQueue_Hear(t *testing.T) {
	numItems := 2000
	numThreads := 8
	var wg sync.WaitGroup
	wg.Add(numThreads * numItems)

	s := switchboard.NewSwitchboard()
	_, queue := s.ListenChannel(cmixproto.OuterType_NONE,
		cmixproto.InnerType_NO_TYPE, id.ZeroID, 12)

	var items []switchboard.Item

	user := new(id.User).SetUints(&[4]uint64{0, 0, 0, 3})
	// Hopefully this would be enough to cause a race condition
	for j := 0; j < numThreads; j++ {
		go func() {
			for i := 0; i < numItems; i++ {
				s.Speak(&switchboard.Message{
					Contents:  []byte{},
					Sender:    user,
					InnerType: cmixproto.InnerType_TEXT_MESSAGE,
					OuterType: cmixproto.OuterType_NODE,
				})
				wg.Done()
				time.Sleep(time.Millisecond)
			}
		}()
	}
	// Listen to the heard messages
	// If there aren't enough items, this will block forever instead of failing
	// the test
	for len(items) < numThreads * numItems {
		items = append(items, <-queue)
	}
	// Check that all items are represented
	wg.Wait()
	time.Sleep(50*time.Millisecond)
	if len(items) != numThreads * numItems {
		t.Error("Didn't get the expected number of items on the channel")
	}
	// Make sure there isn't anything else available on the channel: there
	// should be exactly the right number of items available
	select {
	case <-queue:
		t.Error("There was another item on the channel that shouldn't have" +
			" been there")
	default:
	}
}
