package domain

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

var KeyStyle = lipgloss.NewStyle().Foreground(styles.SubduedColor)
var HelpStyle = lipgloss.NewStyle().Foreground(styles.VerySubduedColor)

type BindingProvider func() []key.Binding

func FormatBindingsInline(get BindingProvider) string {
	bindings := get()
	parts := make([]string, 0, len(bindings))

	for _, b := range bindings {
		if b.Enabled() { // assuming exported fields
			parts = append(parts,
				fmt.Sprintf("%s %s",
					KeyStyle.Render(b.Help().Key),
					HelpStyle.Render(b.Help().Desc),
				),
			)
		}
	}

	return strings.Join(parts, HelpStyle.Render(" • "))
}
