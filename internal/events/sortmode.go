package events

import "github.com/bkenks/lazymux/internal/domain"

// SortModeChanged is emitted when the repo list cycles its sort order, so the
// choice is persisted to config. The list has already reordered itself.
type SortModeChanged struct{ Mode domain.SortMode }

func (SortModeChanged) isEvent() {}
