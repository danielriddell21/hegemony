package cli

import (
	"fmt"
	"os"

	"github.com/danielriddell21/crucible/hub"

	"github.com/danielriddell21/hegemony/internal/gui"
)

// route classifies a window message for the multi-window hub. hegemony only
// broadcasts shared match state; the leaderboard child is spawned once at
// startup rather than via messages.
func route(m gui.Msg) hub.Route {
	if m.Type == "state" {
		return hub.RouteState
	}
	return hub.RouteNone
}

func hubConfig(self string) hub.Config[gui.Msg] {
	return hub.Config[gui.Msg]{
		Self: self,
		ChildArgs: func(idx int) []string {
			return []string{"run", fmt.Sprintf("--child=%d", idx)}
		},
		Route: route,
		Quit:  gui.Msg{Type: "quit"},
	}
}

func runLeader(cfg gui.Config) error {
	if !gui.Available() {
		if err := gui.Run(cfg); err != nil {
			return fmt.Errorf("run gui: %w", err)
		}
		return nil
	}

	h := hub.New(hubConfig(os.Args[0]))
	leaderIn := make(chan gui.Msg, 64)
	leaderOut := make(chan gui.Msg, 64)
	h.AddParticipant(leaderIn, nil) // the leader is id 0
	go h.Run()
	go func() {
		for m := range leaderOut {
			h.Inject(0, m)
		}
	}()
	h.SpawnChild() // the leaderboard

	cfg.Role = gui.RoleMap
	cfg.Link = &gui.Link{In: leaderIn, Out: leaderOut}
	err := gui.Run(cfg)
	close(leaderOut)
	h.Shutdown()
	if err != nil {
		return fmt.Errorf("run war map: %w", err)
	}
	return nil
}

func runChild(cfg gui.Config, index int) error {
	cfg.Role = gui.RoleBoard
	cfg.OffsetIndex = index
	err := hub.RunChild(func(l gui.Link) error {
		cfg.Link = &l
		if err := gui.Run(cfg); err != nil {
			return fmt.Errorf("run leaderboard: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("leaderboard window: %w", err)
	}
	return nil
}
