// Package settings is the screen for editing lazymux's preferences as one huh
// form. Submitting the form saves every field; esc leaves without saving.
package settings

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
)

const title = "Settings"

type Model struct {
	form  *huh.Form
	draft *config.Config
}

// New builds the form over a copy of cfg, sized to the current window.
func New(cfg config.Config) *Model {
	m := &Model{draft: &cfg}
	m.form = m.newForm()
	m.resize()
	return m
}

func (m *Model) newForm() *huh.Form {
	d := m.draft
	formKeys := huh.NewDefaultKeyMap()
	formKeys.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Editor").
			DescriptionFunc(func() string { return describeEditor(d.Tools.Editor) }, &d.Tools.Editor).
			Value(&d.Tools.Editor).Validate(validateEditorCommand),
		huh.NewSelect[string]().Title("Default clone protocol").Inline(true).
			Options(huh.NewOptions(config.SchemeHTTPS, config.SchemeSSH)...).
			Value(&d.Behavior.DefaultProtocol),
		toggle("Confirm before deleting", &d.Behavior.ConfirmDelete),
		toggle("Show full path on rows", &d.UI.ShowFullPath),
		toggle("Show forge label on rows", &d.UI.ShowForge),
		toggle("Show git stats on rows", &d.UI.ShowStats),
		huh.NewSelect[string]().Title("Sort repos by").Inline(true).
			Options(sortOptions()...).
			Value(&d.UI.SortMode),
		huh.NewInput().Title("Accent color").
			Description("Hex value such as #7D56F4. Empty keeps the theme's color.").
			Placeholder("#7D56F4").
			Value(&d.UI.AccentColor).Validate(validateAccentColor),
	)).WithKeyMap(formKeys).WithShowHelp(true).WithTheme(styles.FormTheme)
}

func toggle(title string, value *bool) *huh.Confirm {
	return huh.NewConfirm().Title(title).Affirmative("On").Negative("Off").Value(value)
}

func sortOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(domain.SortModes))
	for _, mode := range domain.SortModes {
		options = append(options, huh.NewOption(mode.Label(), string(mode)))
	}
	return options
}

// validateEditorCommand resolves an editor command the way exec.Command will
// when a repo is opened, so a value the form accepts is a value that runs.
func validateEditorCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return errors.New("editor cannot be empty")
	}
	if strings.ContainsAny(command, " \t") {
		return errors.New("editor takes a command name only, no arguments")
	}
	if _, err := exec.LookPath(command); err != nil {
		return fmt.Errorf("%q not found on PATH", command)
	}
	return nil
}

// describeEditor shows where the editor command resolves on PATH, or what the
// field takes when it doesn't resolve.
func describeEditor(command string) string {
	path, err := exec.LookPath(strings.TrimSpace(command))
	if err != nil {
		return "Any command on your PATH."
	}
	return "✓ " + path
}

func validateAccentColor(hex string) error {
	_, err := styles.ParseAccent(strings.TrimSpace(hex))
	return err
}

func (m *Model) Init() tea.Cmd { return m.form.Init() }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.form.State != huh.StateNormal {
		return m, nil
	}
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.resize()
	}

	model, cmd := m.form.Update(msg)
	if form, ok := model.(*huh.Form); ok {
		m.form = form
	}

	switch m.form.State {
	case huh.StateAborted:
		return m, commands.SetState(domain.StateMain)
	case huh.StateCompleted:
		edited := *m.draft
		edited.Tools.Editor = strings.TrimSpace(edited.Tools.Editor)
		edited.UI.AccentColor = strings.TrimSpace(edited.UI.AccentColor)
		return m, func() tea.Msg { return events.SettingsChanged{Config: edited} }
	}
	return m, cmd
}

// resize fits the form and its help line under the title. huh cannot lay out
// at the 1×1 floor ContentSize returns before the first window size arrives,
// so it waits.
func (m *Model) resize() {
	if constants.WindowSize.Width == 0 {
		return
	}
	const helpRows = 1
	width, height := styles.ContentSize(lipgloss.Height(styles.MenuTitle.Render(title)) + helpRows)
	m.form = m.form.WithWidth(width).WithHeight(height)
}

func (m *Model) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		styles.MenuTitle.Render(title), m.form.View()))
}
