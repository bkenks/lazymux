// Package forgepick renders the registry forges as checkbox rows for the
// screens that choose a repo's upstreams and origin.
package forgepick

import (
	"charm.land/bubbles/v2/list"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/styles"
)

// item is a registry forge as shown in a selection list.
type item struct {
	name, host      string
	checked, origin bool
}

func (i item) Title() string {
	box := styles.GlyphCheckOff
	if i.checked {
		box = styles.GlyphCheckOn
	}
	t := box + " " + i.name
	if i.origin {
		t += " " + styles.GlyphOrigin
	}
	return t
}

func (i item) Description() string {
	if i.origin {
		return i.host + "  · origin (fetch)"
	}
	return i.host
}

func (i item) FilterValue() string { return i.name }

// Items builds one row per forge, checked when link pushes to it and marked
// when it is link's origin.
func Items(forges []config.Forge, link config.RepoLink) []list.Item {
	items := make([]list.Item, len(forges))
	for i, f := range forges {
		items[i] = item{name: f.Name, host: f.Host, checked: link.HasUpstream(f.Name), origin: link.Origin == f.Name}
	}
	return items
}

// IndexOf returns the position of the named forge, or -1.
func IndexOf(forges []config.Forge, name string) int {
	for i, f := range forges {
		if f.Name == name {
			return i
		}
	}
	return -1
}
