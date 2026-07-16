package domain

import (
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
)

type BindingProvider func() []key.Binding

// FormatBindingsInline renders a one-line key hint bar through the shared
// bubbles/help renderer, so the non-list screens match the styling the list
// screens already get from help internally.
func FormatBindingsInline(get BindingProvider) string {
	return styles.Help.ShortHelpView(get())
}
