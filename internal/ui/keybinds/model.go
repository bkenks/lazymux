// Package keybinds is the screen for creating, editing and deleting custom
// keybinds — key combos on the repo list that run a shell command in the
// selected repo.
package keybinds

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/gitkeeper/internal/commands"
	"github.com/bkenks/gitkeeper/internal/config"
	"github.com/bkenks/gitkeeper/internal/constants"
	"github.com/bkenks/gitkeeper/internal/domain"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/keybind"
	"github.com/bkenks/gitkeeper/internal/styles"
)

type keyMap struct {
	New, Edit, Delete, Exit key.Binding
}

var keys = keyMap{
	New:    key.NewBinding(key.WithKeys("n", "N"), key.WithHelp("n", "new")),
	Edit:   key.NewBinding(key.WithKeys("e", "E"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("ctrl+\\"), key.WithHelp("ctrl+\\", "delete")),
	Exit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

func helpKeys() []key.Binding {
	return []key.Binding{keys.New, keys.Edit, keys.Delete, keys.Exit}
}

type keybindItem struct{ config.Keybind }

func (i keybindItem) Title() string       { return i.Name }
func (i keybindItem) Description() string { return i.Keys + "  →  " + i.Command }
func (i keybindItem) FilterValue() string { return i.Name }

type formPurpose int

const (
	purposeEdit formPurpose = iota
	purposeDelete
)

type Model struct {
	list     list.Model
	keybinds []config.Keybind
	reserved []string

	form              *huh.Form
	formFields        []huh.Field
	purpose           formPurpose
	editIndex         int // -1 = new keybind
	draft             *config.Keybind
	isDeleteConfirmed bool
}

// New builds the screen over a copy of cfg's keybinds. reserved lists the keys
// the repo list already uses, which custom keybinds may not take.
func New(cfg config.Config, reserved []string) *Model {
	w, h := styles.ContentSize(0)
	l := styles.NewList(nil, styles.NewDelegate(), w, h)
	l.Title = "Keybinds"
	l.KeyMap.Quit = constants.ListQuit
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = helpKeys
	l.AdditionalFullHelpKeys = helpKeys

	m := &Model{
		list:     l,
		keybinds: append([]config.Keybind(nil), cfg.Keybinds...),
		reserved: reserved,
	}
	m.refresh()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) refresh() {
	items := make([]list.Item, len(m.keybinds))
	for i, bind := range m.keybinds {
		items[i] = keybindItem{bind}
	}
	m.list.SetItems(items)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.list.SetSize(styles.ContentSize(0))
	}
	if m.form != nil {
		return m.updateForm(msg)
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(keyMsg, constants.GlobalKeyMap.Quit):
		return m, tea.Quit
	case key.Matches(keyMsg, keys.Exit):
		return m, commands.SetState(domain.StateMain)
	case key.Matches(keyMsg, keys.New):
		return m, m.startEdit(-1)
	case key.Matches(keyMsg, keys.Edit):
		if len(m.keybinds) > 0 {
			return m, m.startEdit(m.list.Index())
		}
	case key.Matches(keyMsg, keys.Delete):
		if len(m.keybinds) > 0 {
			return m, m.startDelete(m.list.Index())
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) startEdit(index int) tea.Cmd {
	m.draft = &config.Keybind{}
	if index >= 0 {
		*m.draft = m.keybinds[index]
	}
	m.editIndex = index
	m.purpose = purposeEdit
	formKeys := cancelableKeyMap()
	formKeys.Confirm.Accept.SetEnabled(false)
	formKeys.Confirm.Reject.SetEnabled(false)
	m.form = m.newForm(formKeys,
		huh.NewInput().Title("Name").Placeholder("Git log").
			Value(&m.draft.Name).Validate(requireText("name")),
		huh.NewInput().Title("Keybind").Placeholder("ctrl + l").
			Value(&m.draft.Keys).Validate(m.validateKeys),
		huh.NewInput().Title("Command").Placeholder("git log --oneline --graph").
			Value(&m.draft.Command).Validate(requireText("command")),
		huh.NewConfirm().Title("Return to gitkeeper on command end").
			Affirmative("Yes").Negative("No").
			Value(&m.draft.ReturnOnExit),
	)
	return m.form.Init()
}

func (m *Model) startDelete(index int) tea.Cmd {
	m.editIndex = index
	m.purpose = purposeDelete
	m.isDeleteConfirmed = false
	name := lipgloss.NewStyle().Bold(true).Render(m.keybinds[index].Name)
	m.form = m.newForm(cancelableKeyMap(),
		huh.NewConfirm().
			Title(fmt.Sprintf("Are you sure you'd like to delete %s?", name)).
			Affirmative("Yes").
			Negative("No").
			Value(&m.isDeleteConfirmed),
	)
	return m.form.Init()
}

// cancelableKeyMap is huh's default key map with esc cancelling the form, like
// every other gitkeeper screen.
func cancelableKeyMap() *huh.KeyMap {
	formKeys := huh.NewDefaultKeyMap()
	formKeys.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	return formKeys
}

func (m *Model) newForm(formKeys *huh.KeyMap, fields ...huh.Field) *huh.Form {
	m.formFields = fields
	return huh.NewForm(huh.NewGroup(fields...)).
		WithKeyMap(formKeys).WithShowHelp(false).WithTheme(styles.FormTheme)
}

func (m *Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = styles.UpdateForm(m.form, m.formFields, msg)

	switch m.form.State {
	case huh.StateAborted:
		m.form = nil
		return m, nil
	case huh.StateCompleted:
		m.form = nil
		return m, m.applyForm()
	}
	return m, cmd
}

// applyForm saves the finished form's result and emits the new keybind list.
func (m *Model) applyForm() tea.Cmd {
	switch m.purpose {
	case purposeDelete:
		if !m.isDeleteConfirmed {
			return nil
		}
		m.keybinds = append(m.keybinds[:m.editIndex], m.keybinds[m.editIndex+1:]...)
	case purposeEdit:
		saved := *m.draft
		keys, err := keybind.Parse(saved.Keys)
		if err != nil {
			return func() tea.Msg {
				return events.Toast{Level: events.ToastError, Msg: "keybind not saved: " + err.Error()}
			}
		}
		saved.Keys = keys
		if m.editIndex < 0 {
			m.keybinds = append(m.keybinds, saved)
		} else {
			m.keybinds[m.editIndex] = saved
		}
	}
	m.refresh()

	keybinds := append([]config.Keybind(nil), m.keybinds...)
	return func() tea.Msg { return events.KeybindsChanged{Keybinds: keybinds} }
}

// validateKeys accepts a combo that parses and isn't already taken by the repo
// list or another custom keybind.
func (m *Model) validateKeys(input string) error {
	keystroke, err := keybind.Parse(input)
	if err != nil {
		return err
	}
	if used, ok := keybind.FindClash(keystroke, m.reserved); ok {
		return fmt.Errorf("%s is already a gitkeeper key", used)
	}
	for i, other := range m.keybinds {
		if i != m.editIndex && other.Keys == keystroke {
			return fmt.Errorf("%s is already bound to %q", keystroke, other.Name)
		}
	}
	return nil
}

func requireText(field string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return errors.New(field + " cannot be empty")
		}
		return nil
	}
}

func (m *Model) View() tea.View {
	if m.form == nil {
		return tea.NewView(m.list.View())
	}

	title := "Delete Keybind"
	if m.purpose == purposeEdit {
		title = "New Keybind"
		if m.editIndex >= 0 {
			title = "Edit Keybind"
		}
	}
	var formKeys []key.Binding
	if m.purpose == purposeEdit {
		formKeys = append(formKeys, styles.SaveKey)
	}
	rows := []string{styles.RenderFormScreen(title, m.form, formKeys...)}
	if m.purpose == purposeEdit {
		width, _ := styles.ContentSize(0)
		keyNamesHelp := styles.Subtle(keybind.KeyNamesHelp)
		rows = append(rows, lipgloss.NewStyle().Width(width).Render(keyNamesHelp))
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
}
