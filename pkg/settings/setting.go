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

// Validator checks a candidate Text value. It returns a short hint describing
// the accepted value (an editor's resolved path, say) or an error explaining
// why the value cannot be used. A nil Validator accepts everything.
type Validator func(candidate string) (hint string, err error)

// Text is a free-form string setting. It has no options to cycle: the settings
// screen opens an inline text input when the row is activated, and validate
// gates what the user is allowed to commit.
type Text struct {
	key      string
	label    string
	value    string
	validate Validator
}

func NewText(key, label, value string, validate Validator) Text {
	return Text{key: key, label: label, value: value, validate: validate}
}

func (t Text) Key() string         { return t.key }
func (t Text) Label() string       { return t.label }
func (t Text) FilterValue() string { return t.label }
func (t Text) ValueString() string { return t.value }
func (t Text) Value() any          { return t.value }
func (t Text) Title() string       { return t.label }
func (t Text) Description() string {
	if t.value == "" {
		return "(unset)"
	}
	return t.value
}

// Next and Prev leave a Text unchanged: there is nothing to cycle through, and
// the settings screen opens the inline editor instead of calling them.
func (t Text) Next() Setting { return t }
func (t Text) Prev() Setting { return t }

func (t Text) Validate(candidate string) (string, error) {
	if t.validate == nil {
		return "", nil
	}
	return t.validate(candidate)
}

// WithValue returns a copy holding candidate, keeping key, label and validator.
func (t Text) WithValue(candidate string) Text {
	t.value = candidate
	return t
}
