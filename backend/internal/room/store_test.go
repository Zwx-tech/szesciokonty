package room

import (
	"testing"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
)

type memClient struct {
	msgs []any
}

func (c *memClient) Send(msg any) { c.msgs = append(c.msgs, msg) }

func TestCreateJoinLeave(t *testing.T) {
	s := NewStore()
	hostC, guestC := &memClient{}, &memClient{}

	r, host, err := s.Create("Ada", hostC)
	if err != nil {
		t.Fatal(err)
	}
	if r.Code == "" || host.Token == "" {
		t.Fatal("missing code/token")
	}

	_, guest, err := s.Join(r.Code, "Bob", guestC)
	if err != nil {
		t.Fatal(err)
	}
	if guest.Host {
		t.Fatal("guest should not be host")
	}

	if _, _, err := s.Join(r.Code, "Eve", &memClient{}); err != ErrFull {
		t.Fatalf("want full, got %v", err)
	}

	closed, rem, err := s.Leave(r.Code, guest.ID)
	if err != nil || closed || rem == nil || rem.Guest != nil {
		t.Fatalf("guest leave: closed=%v rem=%v err=%v", closed, rem, err)
	}

	closed, _, err = s.Leave(r.Code, host.ID)
	if err != nil || !closed {
		t.Fatalf("host leave: closed=%v err=%v", closed, err)
	}
	if _, _, err := s.Join(r.Code, "Bob", &memClient{}); err != ErrNotFound {
		t.Fatalf("want not found after host leave, got %v", err)
	}
}

func TestArmyReadyStartReconnect(t *testing.T) {
	s := NewStore()
	hostC, guestC := &memClient{}, &memClient{}
	r, host, _ := s.Create("Ada", hostC)
	_, guest, _ := s.Join(r.Code, "Bob", guestC)

	if _, err := s.SetArmy(r.Code, host.ID, protocol.ArmyRed); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetArmy(r.Code, guest.ID, protocol.ArmyRed); err != ErrBadArmy {
		t.Fatalf("want bad army, got %v", err)
	}
	if _, err := s.SetArmy(r.Code, guest.ID, protocol.ArmyBlue); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetReady(r.Code, host.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetReady(r.Code, guest.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Start(r.Code, guest.ID); err != ErrNotHost {
		t.Fatalf("want not host, got %v", err)
	}
	r, err := s.Start(r.Code, host.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Phase != protocol.PhaseMatch {
		t.Fatalf("phase %s", r.Phase)
	}

	s.Disconnect(r.Code, guest.ID)
	newC := &memClient{}
	_, g2, err := s.Reconnect(r.Code, guest.Token, newC)
	if err != nil {
		t.Fatal(err)
	}
	if !g2.Connected || g2.ID != guest.ID {
		t.Fatal("reconnect failed")
	}
}

func TestStateFor(t *testing.T) {
	s := NewStore()
	c := &memClient{}
	r, host, _ := s.Create("Ada", c)
	st := r.stateFor(host)
	if st.Type != protocol.TypeRoomState || st.Token != host.Token || !st.You.Host {
		t.Fatalf("%+v", st)
	}
}
