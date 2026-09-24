package clonerepos

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/styles"
)

func newSizedModel(t *testing.T, width, height int) *Model {
	t.Helper()
	previous := constants.WindowSize
	t.Cleanup(func() { constants.WindowSize = previous })
	constants.WindowSize = tea.WindowSizeMsg{Width: width, Height: height}
	m := New(config.Config{})
	m.Update(constants.WindowSize)
	return m
}

func TestPasteFillsURLTextarea(t *testing.T) {
	m := newSizedModel(t, 80, 30)

	m.Update(tea.PasteMsg{Content: "git@a:x/one.git\ngit@a:x/two.git"})

	if got := m.textarea.Value(); got != "git@a:x/one.git\ngit@a:x/two.git" {
		t.Errorf("textarea value = %q, want the pasted URLs", got)
	}
}

func TestPasteKeepsMoreURLsThanTheTextareaShows(t *testing.T) {
	m := newSizedModel(t, 80, 15)
	urls := make([]string, 30)
	for i := range urls {
		urls[i] = "git@a:x/repo.git"
	}

	m.Update(tea.PasteMsg{Content: strings.Join(urls, "\n")})

	if got := len(strings.Split(m.textarea.Value(), "\n")); got != len(urls) {
		t.Errorf("textarea holds %d lines, want all %d pasted", got, len(urls))
	}
}

func TestPasteFillsNamespaceInput(t *testing.T) {
	m := newSizedModel(t, 80, 30)
	m.toggleMode()

	m.Update(tea.PasteMsg{Content: "my-org"})

	if got := m.nsInput.Value(); got != "my-org" {
		t.Errorf("namespace input = %q, want %q", got, "my-org")
	}
	if got := m.textarea.Value(); got != "" {
		t.Errorf("textarea = %q, want it untouched while namespace mode is on", got)
	}
}

func TestViewStartsWithTitleAndFillsHeight(t *testing.T) {
	m := newSizedModel(t, 80, 30)
	_, height := sizeBuffer()
	_, contentHeight := styles.ContentSize(0)

	content := m.View().Content

	titleRow := styles.MenuTitle.GetMarginTop()
	if line := strings.Split(content, "\n")[titleRow]; !strings.Contains(line, "Repository Clone") {
		t.Errorf("row %d = %q, want the title with no spacer rows above it", titleRow, line)
	}
	if got := lipgloss.Height(content); got != contentHeight {
		t.Errorf("view is %d rows, want it to fill the %d-row content area", got, contentHeight)
	}
	if got := m.textarea.Height(); got != height {
		t.Errorf("textarea is %d rows, want the %d rows left under header and footer", got, height)
	}
}
