//go:build !ebiten

package gui

import "errors"

func Available() bool { return false }

func Run(cfg Config) error {
	// Recording needs no window: the war map is composed in software, so demo
	// media builds anywhere, with no display.
	if cfg.Rec.Recording() {
		return Render(cfg)
	}
	return errors.New("built without the GUI; rebuild with -tags ebiten to watch a match")
}
