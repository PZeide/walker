package ui

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/abenz1267/walker/modules"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

type Handler struct {
	receiver chan []modules.Entry
	entries  []modules.Entry
	ctx      context.Context
	mut      sync.Mutex
}

func (h *Handler) handle() {
	for {
		select {
		case entries := <-h.receiver:
			if len(entries) == 0 {
				continue
			}

			h.mut.Lock()
			h.entries = append(h.entries, entries...)

			sortEntries(h.entries)

			if len(h.entries) > 0 {
				glib.IdleAdd(func() {
					ui.items.Splice(0, ui.items.NItems(), h.entries...)
					ui.selection.SetSelected(0)
				})
			}

			h.mut.Unlock()
		case <-h.ctx.Done():
			return
		default:
		}
	}
}

func sortEntries(entries []modules.Entry) {
	slices.SortFunc(entries, func(a, b modules.Entry) int {
		if a.Matching == modules.AlwaysTop && b.Matching != modules.AlwaysTop {
			return -1
		}

		if b.Matching == modules.AlwaysTop && a.Matching != modules.AlwaysTop {
			return 1
		}

		if a.Matching == modules.AlwaysBottom && b.Matching != modules.AlwaysBottom {
			return 1
		}

		if b.Matching == modules.AlwaysBottom && a.Matching != modules.AlwaysBottom {
			return -1
		}

		if cfg.Hyprland.ContextAwareHistory && cfg.IsService {
			if !a.LastUsed.IsZero() && !b.LastUsed.IsZero() {
				if a.OpenWindows > b.OpenWindows {
					return 1
				}

				if a.OpenWindows < b.OpenWindows {
					return -1
				}
			}
		}

		if a.ScoreFinal == b.ScoreFinal {
			if !a.LastUsed.IsZero() && !b.LastUsed.IsZero() {
				return b.LastUsed.Compare(a.LastUsed)
			}

			return strings.Compare(a.Label, b.Label)
		}

		if a.ScoreFinal > b.ScoreFinal {
			return -1
		}

		if a.ScoreFinal < b.ScoreFinal {
			return 1
		}

		return 0
	})
}
