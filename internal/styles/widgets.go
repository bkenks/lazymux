package styles

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/huh/v2"
)

// NewDelegate returns the default list row renderer styled for the terminal
// background.
func NewDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles = list.NewDefaultItemStyles(IsDark)
	return d
}

// NewList returns a list styled for the terminal background, including its
// filter input, pagination dots, and help.
func NewList(items []list.Item, delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(items, delegate, width, height)
	s := list.DefaultStyles(IsDark)
	l.Styles = s
	l.FilterInput.SetStyles(s.Filter)
	l.Paginator.ActiveDot = s.ActivePaginationDot.String()
	l.Paginator.InactiveDot = s.InactivePaginationDot.String()
	l.Help.Styles = help.DefaultStyles(IsDark)
	return l
}

// NewTextInput returns a text input styled for the terminal background.
func NewTextInput() textinput.Model {
	ti := textinput.New()
	ti.SetStyles(textinput.DefaultStyles(IsDark))
	return ti
}

// NewTextArea returns a text area styled for the terminal background.
func NewTextArea() textarea.Model {
	ta := textarea.New()
	ta.SetStyles(textarea.DefaultStyles(IsDark))
	return ta
}

// FormTheme is the huh theme for the terminal background. huh only learns the
// background from a message lazymux never requests, so this ignores its guess.
var FormTheme = huh.ThemeFunc(func(bool) *huh.Styles { return huh.ThemeCharm(IsDark) })
