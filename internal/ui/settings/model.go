// Package settings is the screen for editing lazymux's preferences as one huh
// form. Submitting the form saves every field; esc leaves without saving.
package settings

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	colorful "github.com/lucasb-eyer/go-colorful"
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
	fields := []huh.Field{
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
	}
	fields = append(fields, colorInputs("Dark", &d.UI.Colors.Dark, styles.IsDark)...)
	fields = append(fields, colorInputs("Light", &d.UI.Colors.Light, !styles.IsDark)...)
	return huh.NewForm(huh.NewGroup(fields...)).
		WithKeyMap(styles.FormKeyMap()).WithShowHelp(true).WithTheme(styles.FormTheme)
}

func toggle(title string, value *bool) *huh.Confirm {
	return huh.NewConfirm().Title(title).Affirmative("On").Negative("Off").Value(value)
}

// colorInputs edit the three base colors for one terminal background, noting
// which background this terminal has.
func colorInputs(mode string, colors *config.Colors, isActive bool) []huh.Field {
	note := ""
	if isActive {
		note = " This terminal uses these."
	}
	return []huh.Field{
		colorInput(mode+" mode main color", "Title bars and buttons."+note,
			styles.DefaultPalette.Main, &colors.Main),
		colorInput(mode+" mode accent color", "The selected row and highlights."+note,
			styles.DefaultPalette.Accent, &colors.Accent),
		colorInput(mode+" mode gray color", "Text, borders and hints."+note,
			styles.DefaultPalette.Gray, &colors.Gray),
	}
}

// colorInput edits one base color of the palette. Empty keeps the default,
// shown as the placeholder.
func colorInput(title, use string, fallback colorful.Color, value *string) *huh.Input {
	return huh.NewInput().Title(title).
		Description(use + " Hex value; empty keeps the default.").
		Placeholder(fallback.Hex()).
		Value(value).Validate(validateColor)
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

func validateColor(hex string) error {
	hex = strings.TrimSpace(hex)
	if hex == "" {
		return nil
	}
	_, err := styles.ParseColor(hex)
	return err
}

func trimColors(colors *config.Colors) {
	colors.Main = strings.TrimSpace(colors.Main)
	colors.Accent = strings.TrimSpace(colors.Accent)
	colors.Gray = strings.TrimSpace(colors.Gray)
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
		trimColors(&edited.UI.Colors.Dark)
		trimColors(&edited.UI.Colors.Light)
		return m, func() tea.Msg { return events.SettingsChanged{Config: edited} }
	}
	return m, cmd
}

func (m *Model) resize() { m.form = styles.FitForm(m.form, title) }

func (m *Model) View() tea.View { return tea.NewView(styles.RenderFormScreen(title, m.form)) }
