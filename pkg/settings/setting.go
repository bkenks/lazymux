package settings

import "github.com/charmbracelet/bubbles/list"

// Setting is the interface all settings must implement. It satisfies
// list.DefaultItem so the standard two-line list delegate can render it the
// same way the repo list renders repos (label on top, value below).
type Setting interface {
	list.DefaultItem
	Key() string
	Label() string
	ValueString() string
	Value() any
	Next() Setting
	Prev() Setting
}

// Toggle is a boolean on/off setting.
type Toggle struct {
	key   string
	label string
	value bool
}

func NewToggle(key, label string, value bool) Toggle {
	return Toggle{key: key, label: label, value: value}
}

func (t Toggle) Key() string         { return t.key }
func (t Toggle) Label() string       { return t.label }
func (t Toggle) FilterValue() string { return t.label }
func (t Toggle) ValueString() string {
	if t.value {
		return "on"
	}
	return "off"
}
func (t Toggle) Value() any          { return t.value }
func (t Toggle) Title() string       { return t.label }
func (t Toggle) Description() string { return t.ValueString() }
func (t Toggle) Next() Setting       { return Toggle{key: t.key, label: t.label, value: !t.value} }
func (t Toggle) Prev() Setting       { return Toggle{key: t.key, label: t.label, value: !t.value} }

// Select is a string-options setting with a cycling index.
type Select struct {
	key     string
	label   string
	options []string
	index   int
}

func NewSelect(key, label string, options []string, index int) Select {
	return Select{key: key, label: label, options: options, index: index}
}

func (s Select) Key() string         { return s.key }
func (s Select) Label() string       { return s.label }
func (s Select) FilterValue() string { return s.label }
func (s Select) ValueString() string { return s.options[s.index] }
func (s Select) Value() any          { return s.options[s.index] }
func (s Select) Title() string       { return s.label }
func (s Select) Description() string { return s.ValueString() }
func (s Select) Next() Setting {
	return Select{key: s.key, label: s.label, options: s.options, index: (s.index + 1) % len(s.options)}
}
func (s Select) Prev() Setting {
	return Select{key: s.key, label: s.label, options: s.options, index: (s.index - 1 + len(s.options)) % len(s.options)}
}
