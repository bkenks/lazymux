package settings

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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

// editPaneHeight is the number of rows the inline editor borrows from the list
// while it is open: the input, its status line, and its help line.
const editPaneHeight = 3

// internal key bindings
type keyMap struct {
	Next   key.Binding
	Prev   key.Binding
	Exit   key.Binding
	Commit key.Binding
	Cancel key.Binding
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
		key.WithKeys("backspace", "delete"),
		key.WithHelp("backspace", "back"),
	),
	Commit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

var (
	editOKStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	editErrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	editHelpStyle = lipgloss.NewStyle().Faint(true)
)

// Model is the settings screen Bubble Tea model.
type Model struct {
	list      list.Model
	settings  []Setting
	width     int
	height    int
	widthPad  int
	heightPad int

	editing  bool
	editIdx  int
	input    textinput.Model
	editHint string
	editErr  error
}

func New(title string, settings []Setting, width, height, widthPad, heightPad int) Model {
	items := make([]list.Item, len(settings))
	for i, s := range settings {
		items[i] = s
	}

	l := list.New(items, list.NewDefaultDelegate(), width-widthPad, height-heightPad)
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

	input := textinput.New()
	input.Prompt = "> "
	input.CharLimit = 200

	m := Model{
		list:      l,
		settings:  settings,
		width:     width,
		height:    height,
		widthPad:  widthPad,
		heightPad: heightPad,
		input:     input,
	}
	m.applyListSize()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.applyListSize()

	case tea.KeyMsg:
		if m.editing {
			return m, m.updateEditor(msg)
		}
		idx := m.list.Index()
		current := m.settingAt(idx)
		switch {
		case key.Matches(msg, keys.Exit):
			return m, func() tea.Msg { return Exited{} }

		case key.Matches(msg, keys.Next):
			if text, ok := current.(Text); ok {
				return m, m.beginEdit(idx, text)
			}
			if current != nil {
				cmds = append(cmds, m.applySetting(idx, current.Next()))
			}

		case key.Matches(msg, keys.Prev):
			if _, ok := current.(Text); ok {
				return m, nil
			}
			if current != nil {
				cmds = append(cmds, m.applySetting(idx, current.Prev()))
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// updateEditor owns every keystroke while the inline editor is open, so the
// list underneath never sees the text the user is typing.
func (m *Model) updateEditor(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.endEdit()
		return nil

	case key.Matches(msg, keys.Commit):
		return m.commitEdit()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.revalidate()
	return cmd
}

func (m *Model) beginEdit(idx int, text Text) tea.Cmd {
	m.editing = true
	m.editIdx = idx
	m.input.SetValue(text.ValueString())
	m.revalidate()
	m.applyListSize()
	return m.input.Focus()
}

func (m *Model) endEdit() {
	m.editing = false
	m.editHint, m.editErr = "", nil
	m.input.Blur()
	m.input.SetValue("")
	m.applyListSize()
}

// commitEdit saves the typed value, or keeps the editor open when the value
// fails validation so the status line can explain why.
func (m *Model) commitEdit() tea.Cmd {
	text, ok := m.settingAt(m.editIdx).(Text)
	if !ok {
		m.endEdit()
		return nil
	}
	m.revalidate()
	if m.editErr != nil {
		return nil
	}

	cmd := m.applySetting(m.editIdx, text.WithValue(m.editedValue()))
	m.endEdit()
	return cmd
}

func (m *Model) editedValue() string { return strings.TrimSpace(m.input.Value()) }

func (m *Model) revalidate() {
	text, ok := m.settingAt(m.editIdx).(Text)
	if !ok {
		m.editHint, m.editErr = "", nil
		return
	}
	m.editHint, m.editErr = text.Validate(m.editedValue())
}

// applySetting stores updated at idx, refreshes the list row, and announces the
// change.
func (m *Model) applySetting(idx int, updated Setting) tea.Cmd {
	m.settings[idx] = updated
	m.rebuildItems(idx)
	return func() tea.Msg { return SettingChanged{Key: updated.Key(), Setting: updated} }
}

func (m *Model) settingAt(idx int) Setting {
	if idx < 0 || idx >= len(m.settings) {
		return nil
	}
	return m.settings[idx]
}

func (m *Model) applyListSize() {
	height := m.height - m.heightPad
	if m.editing {
		height -= editPaneHeight
	}
	m.list.SetSize(m.width-m.widthPad, height)
}

func (m *Model) rebuildItems(selectedIdx int) {
	items := make([]list.Item, len(m.settings))
	for i, s := range m.settings {
		items[i] = s
	}
	m.list.SetItems(items)
	m.list.Select(selectedIdx)
}

func (m *Model) View() string {
	if !m.editing {
		return m.list.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.list.View(),
		m.input.View(),
		m.statusView(),
		editHelpStyle.Render("enter save · esc cancel"),
	)
}

func (m *Model) statusView() string {
	switch {
	case m.editErr != nil:
		return editErrStyle.Render("✗ " + m.editErr.Error())
	case m.editHint != "":
		return editOKStyle.Render("✓ " + m.editHint)
	}
	return ""
}

// Settings returns the current settings slice.
func (m *Model) Settings() []Setting { return m.settings }
