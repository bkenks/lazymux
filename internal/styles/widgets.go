package styles

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/huh/v2"
)

// NewDelegate returns the default list row renderer styled for the terminal
// background, with the selected row in the accent color when one is set.
func NewDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles = list.NewDefaultItemStyles(IsDark)
	if Accent != nil {
		d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(Accent).BorderForeground(Accent)
		d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(Accent).BorderForeground(Accent)
	}
	return d
}

// NewList returns a list styled by StyleList.
func NewList(items []list.Item, delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(items, delegate, width, height)
	StyleList(&l)
	return l
}

// StyleList styles a list for the terminal background, including its filter
// input, pagination dots, and help, with the title and filter cursor in the
// accent color when one is set.
func StyleList(l *list.Model) {
	s := list.DefaultStyles(IsDark)
	if Accent != nil {
		s.Title = s.Title.Background(Accent)
		s.Filter.Cursor.Color = Accent
	}
	l.Styles = s
	l.FilterInput.SetStyles(s.Filter)
	l.Paginator.ActiveDot = s.ActivePaginationDot.String()
	l.Paginator.InactiveDot = s.InactivePaginationDot.String()
	l.Help = Help
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

// FormTheme is the huh theme for the terminal background, with its indigo and
// fuchsia accents replaced by the accent color when one is set. huh only learns
// the background from a message lazymux never requests, so this ignores its
// guess.
var FormTheme = huh.ThemeFunc(func(bool) *huh.Styles {
	t := huh.ThemeCharm(IsDark)
	if Accent != nil {
		accentFormFields(&t.Focused)
		accentFormFields(&t.Blurred)
		t.Group.Title = t.Focused.Title
	}
	return t
})

func accentFormFields(f *huh.FieldStyles) {
	f.Title = f.Title.Foreground(Accent)
	f.NoteTitle = f.NoteTitle.Foreground(Accent)
	f.Directory = f.Directory.Foreground(Accent)
	f.SelectSelector = f.SelectSelector.Foreground(Accent)
	f.MultiSelectSelector = f.MultiSelectSelector.Foreground(Accent)
	f.NextIndicator = f.NextIndicator.Foreground(Accent)
	f.PrevIndicator = f.PrevIndicator.Foreground(Accent)
	f.FocusedButton = f.FocusedButton.Background(Accent)
	f.Next = f.Next.Background(Accent)
	f.TextInput.Prompt = f.TextInput.Prompt.Foreground(Accent)
}
