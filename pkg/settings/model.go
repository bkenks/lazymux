package settings

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingChanged is emitted when a setting value changes.
type SettingChanged struct {
	Key     string
	Setting Setting
}

// Exited is emitted when the user presses esc to leave the settings screen.
type Exited struct{}

// internal key bindings
type keyMap struct {
	Next key.Binding
	Prev key.Binding
	Exit key.Binding
}

var keys = keyMap{
	Next: key.NewBinding(
		key.WithKeys("right", "l", "enter", " "),
		key.WithHelp("→/l/enter", "next"),
	),
	Prev: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev"),
	),
	Exit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

// settingDelegate renders each setting row.
type settingDelegate struct{}

func (d settingDelegate) Height() int                              { return 1 }
func (d settingDelegate) Spacing() int                            { return 0 }
func (d settingDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d settingDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	s, ok := item.(Setting)
	if !ok {
		return
	}

	label := s.Label()
	value := fmt.Sprintf("[%s]", s.ValueString())

	width := m.Width() - 4
	needed := len(label) + len(value) + 1
	if width < needed {
		width = needed
	}
	spacer := strings.Repeat(" ", width-len(label)-len(value))
	line := fmt.Sprintf("  %s%s%s", label, spacer, value)

	if index == m.Index() {
		line = lipgloss.NewStyle().Bold(true).Render(line)
	}

	fmt.Fprint(w, line)
}

// Model is the settings screen Bubble Tea model.
type Model struct {
	list      list.Model
	settings  []Setting
	widthPad  int
	heightPad int
}

func New(title string, settings []Setting, width, height, widthPad, heightPad int) Model {
	items := make([]list.Item, len(settings))
	for i, s := range settings {
		items[i] = s
	}

	l := list.New(items, settingDelegate{}, width-widthPad, height-heightPad)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	// Disable pagination keys that conflict with value cycling
	l.KeyMap.PrevPage = key.NewBinding()
	l.KeyMap.NextPage = key.NewBinding()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Prev, keys.Next, keys.Exit}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Prev, keys.Next, keys.Exit}
	}

	return Model{
		list:      l,
		settings:  settings,
		widthPad:  widthPad,
		heightPad: heightPad,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-m.widthPad, msg.Height-m.heightPad)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Exit):
			return m, func() tea.Msg { return Exited{} }

		case key.Matches(msg, keys.Next):
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.settings) {
				newSetting := m.settings[idx].Next()
				m.settings[idx] = newSetting
				m.rebuildItems(idx)
				k, s := newSetting.Key(), newSetting
				cmds = append(cmds, func() tea.Msg { return SettingChanged{Key: k, Setting: s} })
			}

		case key.Matches(msg, keys.Prev):
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.settings) {
				newSetting := m.settings[idx].Prev()
				m.settings[idx] = newSetting
				m.rebuildItems(idx)
				k, s := newSetting.Key(), newSetting
				cmds = append(cmds, func() tea.Msg { return SettingChanged{Key: k, Setting: s} })
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) rebuildItems(selectedIdx int) {
	items := make([]list.Item, len(m.settings))
	for i, s := range m.settings {
		items[i] = s
	}
	m.list.SetItems(items)
	m.list.Select(selectedIdx)
}

func (m *Model) View() string { return m.list.View() }

// Settings returns the current settings slice.
func (m *Model) Settings() []Setting { return m.settings }
