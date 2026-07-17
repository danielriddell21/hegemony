package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

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

func TestHubBroadcastsToOthersAndStoresLast(t *testing.T) {
	h := newHub("self")
	a := make(chan gui.Msg, 4)
	b := make(chan gui.Msg, 4)
	ida := h.addParticipant(a, nil)
	h.addParticipant(b, nil)

	h.handle(ida, gui.Msg{Type: "state", Tick: 7})

	select {
	case m := <-b:
		if m.Tick != 7 {
			t.Fatalf("b received tick %d, want 7", m.Tick)
		}
	default:
		t.Fatal("other participant did not receive the state")
	}
	select {
	case <-a:
		t.Fatal("sender should not receive its own message")
	default:
	}
	if h.last.Tick != 7 {
		t.Fatalf("hub.last tick = %d, want 7 stored", h.last.Tick)
	}
}

func TestHubSendsLastStateToNewParticipant(t *testing.T) {
	h := newHub("self")
	ida := h.addParticipant(make(chan gui.Msg, 4), nil)
	h.handle(ida, gui.Msg{Type: "state", Tick: 3})

	late := make(chan gui.Msg, 4)
	h.addParticipant(late, nil)
	select {
	case m := <-late:
		if m.Tick != 3 {
			t.Fatalf("late participant got tick %d, want 3", m.Tick)
		}
	default:
		t.Fatal("late participant did not receive the last state")
	}
}

func TestHubDropClosesChannel(t *testing.T) {
	h := newHub("self")
	a := make(chan gui.Msg, 4)
	id := h.addParticipant(a, nil)
	h.handle(id, gui.Msg{Type: eofType})
	if _, ok := <-a; ok {
		t.Fatal("participant channel not closed after drop")
	}
}

func TestHubRunAndShutdown(t *testing.T) {
	h := newHub("self")
	a := make(chan gui.Msg, 4)
	b := make(chan gui.Msg, 4)
	ida := h.addParticipant(a, nil)
	h.addParticipant(b, nil)

	go h.run()
	h.inbox <- srcMsg{id: ida, m: gui.Msg{Type: "state", Tick: 9}}
	select {
	case m := <-b:
		if m.Tick != 9 {
			t.Fatalf("relayed tick %d, want 9", m.Tick)
		}
	case <-time.After(time.Second):
		t.Fatal("hub.run did not relay the message")
	}
	h.shutdown()
}

func TestHubSpawnChildReapsOnExit(t *testing.T) {
	// `true` ignores the args and exits 0, so its stdout EOFs immediately and the
	// child is dropped — exercising spawnChild's pipes, goroutines, and reaping.
	h := newHub("true")
	// Prime a last state so the new child is handed a message to encode, covering
	// the hub->child writer path.
	h.mu.Lock()
	h.last = gui.Msg{Type: "state", Tick: 4}
	h.mu.Unlock()
	go h.run()
	h.spawnChild(1)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		h.mu.Lock()
		n := len(h.parts)
		h.mu.Unlock()
		if n == 0 {
			h.shutdown()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	h.shutdown()
	t.Fatal("spawned child was never reaped/dropped")
}

func TestRunChildWithoutGUIReports(t *testing.T) {
	// The default build has no GUI, so gui.Run returns the stub error.
	if err := runChild(gui.Config{Width: 8, Height: 8}, 1); err == nil {
		t.Error("expected an error from the non-GUI build stub")
	}
}

func TestLeadWiresHubAndChild(t *testing.T) {
	// A fake window runner stands in for the war-map window (no display needed);
	// `true` stands in for the child binary. lead must set up the hub, spawn the
	// child, run the window, then tear everything down cleanly.
	ran := false
	runWindow := func(cfg gui.Config) error {
		ran = true
		if cfg.Role != gui.RoleMap {
			t.Errorf("leader role = %q, want %q", cfg.Role, gui.RoleMap)
		}
		if cfg.Link == nil {
			t.Error("leader got no link")
		} else {
			cfg.Link.Out <- gui.Msg{Type: "state", Tick: 1} // exercise the forwarder
		}
		return nil
	}
	if err := lead(gui.Config{Width: 8, Height: 8}, "true", runWindow); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if !ran {
		t.Error("window runner was never called")
	}
}

func TestChildLinkDecodesStdin(t *testing.T) {
	link := childLink(strings.NewReader(`{"t":"state","tick":9}`+"\n"), io.Discard)
	select {
	case m := <-link.In:
		if m.Tick != 9 {
			t.Fatalf("decoded tick %d, want 9", m.Tick)
		}
	case <-time.After(time.Second):
		t.Fatal("childLink did not decode stdin")
	}
}

type errWriter struct{ wrote chan struct{} }

func (w errWriter) Write([]byte) (int, error) {
	select {
	case w.wrote <- struct{}{}:
	default:
	}
	return 0, io.ErrClosedPipe
}

func TestChildLinkStopsOnWriteError(t *testing.T) {
	w := errWriter{wrote: make(chan struct{}, 1)}
	link := childLink(strings.NewReader(""), w)
	link.Out <- gui.Msg{Type: "state"}
	select {
	case <-w.wrote:
	case <-time.After(time.Second):
		t.Fatal("childLink never attempted to encode")
	}
}

func TestHubSpawnChildStartFailure(t *testing.T) {
	h := newHub("/nonexistent/hegemony-does-not-exist")
	h.spawnChild(1)
	h.mu.Lock()
	n := len(h.parts)
	h.mu.Unlock()
	if n != 0 {
		t.Fatalf("a child that fails to start added %d participants, want 0", n)
	}
}

func TestChildLinkEncodesStdout(t *testing.T) {
	pr, pw := io.Pipe()
	link := childLink(strings.NewReader(""), pw)
	link.Out <- gui.Msg{Type: "state", Tick: 5}
	var got gui.Msg
	if err := json.NewDecoder(pr).Decode(&got); err != nil {
		t.Fatalf("decode encoded output: %v", err)
	}
	if got.Tick != 5 {
		t.Fatalf("encoded tick %d, want 5", got.Tick)
	}
}

func TestLeadSurfacesWindowError(t *testing.T) {
	boom := func(gui.Config) error { return errFake }
	if err := lead(gui.Config{}, "true", boom); err == nil {
		t.Error("expected lead to surface the window error")
	}
}

var errFake = errFakeType("boom")

type errFakeType string

func (e errFakeType) Error() string { return string(e) }
