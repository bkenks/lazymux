package domain

import (
	"charm.land/bubbles/v2/key"
	"github.com/bkenks/gitkeeper/internal/styles"
)

type BindingProvider func() []key.Binding

// FormatBindingsInline renders a one-line key hint bar through the shared
// bubbles/help renderer, so the non-list screens match the styling the list
// screens already get from help internally.
func FormatBindingsInline(get BindingProvider) string {
	return styles.Help.ShortHelpView(get())
}
