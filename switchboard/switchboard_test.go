package switchboard

import (
	"gitlab.com/xx_network/primitives/id"
	"strings"
	"testing"
	"time"
)

//testing message structure
type Message struct {
	Contents    []byte
	Sender      *id.ID
	MessageType int32
}

func (m *Message) GetSender() *id.ID {
	return m.Sender
}

func (m *Message) GetMessageType() int32 {
	return m.MessageType
}

func TestNew(t *testing.T) {
	sw := New()

	if sw.id == nil {
		t.Errorf("did not create an id map")
	}

	if sw.messageType == nil {
		t.Errorf("did not create a messageType map")
	}
}

func TestSwitchboard_RegisterListener_Errors(t *testing.T) {
	sw := New()
	_, err := sw.RegisterListener(nil, 0, &funcListener{})

	if err == nil {
		t.Errorf("A nil userID should have caused an error")
	}

	if err != nil && !strings.Contains(err.Error(), "cannot register listener to nil user") {
		t.Errorf("A nil userID caused the wrong error")
	}

	_, err = sw.RegisterListener(id.NewIdFromUInt(42, id.User, t), 0, nil)

	if err == nil {
		t.Errorf("A nil listener should have caused an error")
	}

	if err != nil && !strings.Contains(err.Error(), "cannot register nil listener") {
		t.Errorf("A nil listener caused the wrong error")
	}
}

func TestSwitchboard_RegisterListener(t *testing.T) {
	sw := New()

	l := &funcListener{}

	uid := id.NewIdFromUInt(42, id.User, t)

	mt := int32(69)

	lid, err := sw.RegisterListener(uid, mt, l)

	//check the returns
	if err != nil {
		t.Errorf("Register Listener should not have errored: %s", err)
	}

	if lid.messageType != mt {
		t.Errorf("ListenerID message type is wrong")
	}

	if !lid.userID.Cmp(uid) {
		t.Errorf("ListenerID userID is wrong")
	}

	if lid.listener != l {
		t.Errorf("ListenerID listener is wrong")
	}

	//check that the listener is registered in the appropriate location
	setID := sw.id.Get(uid)

	if !setID.Has(l) {
		t.Errorf("Listener is not registered by ID")
	}

	setType := sw.messageType.Get(mt)

	if !setType.Has(l) {
		t.Errorf("Listener is not registered by Message Type")
	}

}

func TestSwitchboard_RegisterFunc_Errors(t *testing.T) {
	sw := New()
	_, err := sw.RegisterFunc("test", nil, 0, func(Item) {})

	if err == nil {
		t.Errorf("A nil userID should have caused an error")
	}

	if err != nil && !strings.Contains(err.Error(), "cannot register listener to nil user") {
		t.Errorf("A nil userID caused the wrong error")
	}

	_, err = sw.RegisterFunc("test", id.NewIdFromUInt(42, id.User, t), 0, nil)

	if err == nil {
		t.Errorf("A nil listener func should have caused an error")
	}

	if err != nil && !strings.Contains(err.Error(), "cannot register listener with nil func") {
		t.Errorf("A nil listener caused func the wrong error")
	}
}

func TestSwitchboard_RegisterFunc(t *testing.T) {
	sw := New()

	heard := false

	l := func(Item) { heard = true }

	uid := id.NewIdFromUInt(42, id.User, t)

	mt := int32(69)

	lid, err := sw.RegisterFunc("test", uid, mt, l)

	//check the returns
	if err != nil {
		t.Errorf("Register Listener should not have errored: %s", err)
	}

	if lid.messageType != mt {
		t.Errorf("ListenerID message type is wrong")
	}

	if !lid.userID.Cmp(uid) {
		t.Errorf("ListenerID userID is wrong")
	}

	//check that the listener is registered in the appropriate location
	setID := sw.id.Get(uid)

	if !setID.Has(lid.listener) {
		t.Errorf("Listener is not registered by ID")
	}

	setType := sw.messageType.Get(mt)

	if !setType.Has(lid.listener) {
		t.Errorf("Listener is not registered by Message Type")
	}

	lid.listener.Hear(nil)
	if !heard {
		t.Errorf("Func listener not registered correctly")
	}
}

func TestSwitchboard_RegisterChan_Errors(t *testing.T) {
	sw := New()
	_, err := sw.RegisterChannel("test", nil, 0, make(chan Item))

	if err == nil {
		t.Errorf("A nil userID should have caused an error")
	}

	if err != nil && !strings.Contains(err.Error(), "cannot register listener to nil user") {
		t.Errorf("A nil userID caused the wrong error")
	}

	_, err = sw.RegisterChannel("test", id.NewIdFromUInt(42, id.User, t), 0, nil)

	if err == nil {
		t.Errorf("A nil channel func should have caused an error")
	}

	if err != nil && !strings.Contains(err.Error(), "cannot register listener with nil channel") {
		t.Errorf("A nil listener caused func the wrong error")
	}
}

func TestSwitchboard_RegisterChan(t *testing.T) {
	sw := New()

	ch := make(chan Item, 1)

	uid := id.NewIdFromUInt(42, id.User, t)

	mt := int32(69)

	lid, err := sw.RegisterChannel("test", uid, mt, ch)

	//check the returns
	if err != nil {
		t.Errorf("Register Listener should not have errored: %s", err)
	}

	if lid.messageType != mt {
		t.Errorf("ListenerID message type is wrong")
	}

	if !lid.userID.Cmp(uid) {
		t.Errorf("ListenerID userID is wrong")
	}

	//check that the listener is registered in the appropriate location
	setID := sw.id.Get(uid)

	if !setID.Has(lid.listener) {
		t.Errorf("Listener is not registered by ID")
	}

	setType := sw.messageType.Get(mt)

	if !setType.Has(lid.listener) {
		t.Errorf("Listener is not registered by Message Type")
	}

	lid.listener.Hear(nil)
	select {
	case <-ch:
	case <-time.After(5 * time.Millisecond):
		t.Errorf("Chan listener not registered correctly")
	}
}

//tests all combinations of hits and misses for speak
func TestSwitchboard_Speak(t *testing.T) {

	uids := []*id.ID{{}, id.NewIdFromUInt(42, id.User, t), id.NewIdFromUInt(69, id.User, t)}
	mts := []int32{AnyType, 42, 69}

	for _, uidReg := range uids {
		for _, mtReg := range mts {

			//create the registrations
			sw := New()
			ch1 := make(chan Item, 1)
			ch2 := make(chan Item, 1)

			_, err := sw.RegisterChannel("test", uidReg, mtReg, ch1)

			if err != nil {
				t.Errorf("Register Listener should not have errored: %s", err)
			}

			_, err = sw.RegisterChannel("test", uidReg, mtReg, ch2)

			if err != nil {
				t.Errorf("Register Listener should not have errored: %s", err)
			}

			//send every possible message
			for _, uid := range uids {
				for _, mt := range mts {
					if uid.Cmp(&id.ID{}) || mt == AnyType {
						continue
					}

					m := &Message{
						Contents:    []byte{0, 1, 2, 3},
						Sender:      uid,
						MessageType: mt,
					}

					sw.Speak(m)

					shouldHear := (m.Sender.Cmp(uidReg) || uidReg.Cmp(&id.ID{})) && (m.MessageType == mtReg || mtReg == AnyType)

					var heard1 bool

					select {
					case <-ch1:
						heard1 = true
					case <-time.After(5 * time.Millisecond):
						heard1 = false
					}

					if shouldHear != heard1 {
						t.Errorf("Correct operation not recorded "+
							"for listener 1: Expected: %v, Occured: %v",
							shouldHear, heard1)
					}

					var heard2 bool

					select {
					case <-ch2:
						heard2 = true
					case <-time.After(5 * time.Millisecond):
						heard2 = false
					}

					if shouldHear != heard2 {
						t.Errorf("Correct operation not recorded "+
							"for listener 2: Expected: %v, Occured: %v",
							shouldHear, heard2)
					}
				}
			}
		}
	}
}
