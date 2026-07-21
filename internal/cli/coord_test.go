package cli

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/danielriddell21/crucible/hub"

	"github.com/danielriddell21/hegemony/internal/gui"
)

func TestMsgJSONRoundTrip(t *testing.T) {
	orig := gui.Msg{
		Type: "state", Tick: 12, Over: true, Winner: 3,
		Factions: []gui.FactionStat{{ID: 1, Name: "Greedy", Share: 0.4}},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(orig); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("encoding is not line-delimited")
	}
	var got gui.Msg
	if err := json.NewDecoder(&buf).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("round trip changed message:\n%+v\n%+v", got, orig)
	}
}

func TestRoutePolicy(t *testing.T) {
	if got := route(gui.Msg{Type: "state"}); got != hub.RouteState {
		t.Errorf("state route = %v", got)
	}
	for _, typ := range []string{"quit", "", "other"} {
		if got := route(gui.Msg{Type: typ}); got != hub.RouteNone {
			t.Errorf("route(%q) = %v, want RouteNone", typ, got)
		}
	}
}

func TestHubConfigChildArgs(t *testing.T) {
	cfg := hubConfig("hegemony")
	if cfg.Self != "hegemony" || cfg.Quit.Type != "quit" {
		t.Errorf("cfg self=%q quit=%q", cfg.Self, cfg.Quit.Type)
	}
	if args := cfg.ChildArgs(1); len(args) != 2 || args[0] != "run" || args[1] != "--child=1" {
		t.Errorf("ChildArgs(1) = %v", args)
	}
}
