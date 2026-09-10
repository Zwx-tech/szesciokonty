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

func TestRematchAndLeaveForfeit(t *testing.T) {
	s := NewStore()
	hostC, guestC := &memClient{}, &memClient{}
	r, host, _ := s.Create("Ada", hostC)
	_, guest, _ := s.Join(r.Code, "Bob", guestC)
	_, _ = s.SetArmy(r.Code, host.ID, protocol.ArmyRed)
	_, _ = s.SetArmy(r.Code, guest.ID, protocol.ArmyBlue)
	_, _ = s.SetReady(r.Code, host.ID, true)
	_, _ = s.SetReady(r.Code, guest.ID, true)
	r, err := s.Start(r.Code, host.ID)
	if err != nil {
		t.Fatal(err)
	}

	r.Match.Forfeit(guest.ID)
	r, err = s.Rematch(r.Code, host.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Phase != protocol.PhaseLobby || r.Match != nil || r.Host.Ready || r.Guest.Ready {
		t.Fatalf("rematch lobby: %+v ready h/g=%v/%v", r.Phase, r.Host.Ready, r.Guest.Ready)
	}

	_, _ = s.SetReady(r.Code, host.ID, true)
	_, _ = s.SetReady(r.Code, guest.ID, true)
	r, _ = s.Start(r.Code, host.ID)

	closed, rem, err := s.Leave(r.Code, guest.ID)
	if err != nil || closed || rem == nil {
		t.Fatalf("leave: %v %v %v", closed, rem, err)
	}
	if rem.Match == nil || rem.Match.Phase != protocol.MatchEnded || rem.Match.Result.WinnerID != host.ID {
		t.Fatalf("want forfeit win for host, got %+v", rem.Match)
	}
}
