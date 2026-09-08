package protocol

import (
	"encoding/json"
	"testing"
)

func TestDecodeCreateRoom(t *testing.T) {
	raw := []byte(`{"v":1,"type":"create_room","name":"Ada"}`)
	msg, err := DecodeClient(raw)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := msg.(CreateRoom)
	if !ok || m.Name != "Ada" {
		t.Fatalf("got %#v", msg)
	}
}

func TestDecodeUnknown(t *testing.T) {
	_, err := DecodeClient([]byte(`{"v":1,"type":"nope"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEncodeErrorRoundtrip(t *testing.T) {
	b, err := Encode(NewError("not_implemented", "rooms come later"))
	if err != nil {
		t.Fatal(err)
	}
	var m ErrorMsg
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != TypeError || m.Code != "not_implemented" {
		t.Fatalf("got %#v", m)
	}
}
