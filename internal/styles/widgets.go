package styles

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/constants"
)

// NewDelegate returns the default list row renderer in the palette's colors.
func NewDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	s := list.NewDefaultItemStyles(IsDark)
	s.NormalTitle = s.NormalTitle.Foreground(Text)
	s.NormalDesc = s.NormalDesc.Foreground(Muted)
	s.SelectedTitle = s.SelectedTitle.Foreground(Accent).BorderForeground(AccentMuted)
	s.SelectedDesc = s.SelectedDesc.Foreground(AccentMuted).BorderForeground(AccentMuted)
	s.DimmedTitle = s.DimmedTitle.Foreground(Muted)
	s.DimmedDesc = s.DimmedDesc.Foreground(Subdued)
	d.Styles = s
	return d
}

// NewList returns a list styled by StyleList.
func NewList(items []list.Item, delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(items, delegate, width, height)
	StyleList(&l)
	return l
}

// StyleList styles a list, including its filter input, pagination dots, and
// help, in the palette's colors.
func StyleList(l *list.Model) {
	s := list.DefaultStyles(IsDark)
	s.Title = s.Title.Background(Main).Foreground(OnMain)
	s.Spinner = s.Spinner.Foreground(Muted)
	s.Filter.Cursor.Color = Accent
	s.Filter.Focused.Prompt = s.Filter.Focused.Prompt.Foreground(Accent)
	s.Filter.Blurred.Prompt = s.Filter.Blurred.Prompt.Foreground(Accent)
	s.StatusBar = s.StatusBar.Foreground(Muted)
	s.StatusEmpty = s.StatusEmpty.Foreground(Subdued)
	s.StatusBarActiveFilter = s.StatusBarActiveFilter.Foreground(Text)
	s.StatusBarFilterCount = s.StatusBarFilterCount.Foreground(Faint)
	s.NoItems = s.NoItems.Foreground(Muted)
	s.ArabicPagination = s.ArabicPagination.Foreground(Subdued)
	s.ActivePaginationDot = s.ActivePaginationDot.Foreground(Muted)
	s.InactivePaginationDot = s.InactivePaginationDot.Foreground(Faint)
	s.DividerDot = s.DividerDot.Foreground(Faint)
	l.Styles = s
	l.FilterInput.SetStyles(s.Filter)
	l.Paginator.ActiveDot = s.ActivePaginationDot.String()
	l.Paginator.InactiveDot = s.InactivePaginationDot.String()
	l.Help = Help
}

// NewProgress returns a progress bar that blends from the main color to the
// accent.
func NewProgress() progress.Model {
	return progress.New(progress.WithColors(Main, Accent), progress.WithoutPercentage())
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

// FormTheme is the huh theme in the palette's colors, keeping huh's red for
// errors. huh only learns the background from a message lazymux never
// requests, so this ignores its guess.
var FormTheme = huh.ThemeFunc(func(bool) *huh.Styles {
	t := huh.ThemeCharm(IsDark)
	colorFormFields(&t.Focused)
	colorFormFields(&t.Blurred)
	t.Focused.Base = t.Focused.Base.BorderForeground(Faint)
	t.Focused.Card = t.Focused.Base
	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	t.Help = Help.Styles
	return t
})

func colorFormFields(f *huh.FieldStyles) {
	f.Title = f.Title.Foreground(MainText)
	f.NoteTitle = f.NoteTitle.Foreground(MainText)
	f.Directory = f.Directory.Foreground(MainText)
	f.Description = f.Description.Foreground(Muted)
	f.SelectSelector = f.SelectSelector.Foreground(Accent)
	f.MultiSelectSelector = f.MultiSelectSelector.Foreground(Accent)
	f.NextIndicator = f.NextIndicator.Foreground(Accent)
	f.PrevIndicator = f.PrevIndicator.Foreground(Accent)
	f.Option = f.Option.Foreground(Text)
	f.UnselectedOption = f.UnselectedOption.Foreground(Text)
	f.SelectedOption = f.SelectedOption.Foreground(Accent)
	f.SelectedPrefix = f.SelectedPrefix.Foreground(Accent)
	f.UnselectedPrefix = f.UnselectedPrefix.Foreground(Muted)
	f.FocusedButton = f.FocusedButton.Foreground(OnMain).Background(Main)
	f.Next = f.Next.Foreground(OnMain).Background(Main)
	f.BlurredButton = f.BlurredButton.Foreground(OnSurface).Background(Surface)
	f.TextInput.Cursor = f.TextInput.Cursor.Foreground(Accent)
	f.TextInput.Placeholder = f.TextInput.Placeholder.Foreground(Subdued)
	f.TextInput.Prompt = f.TextInput.Prompt.Foreground(Accent)
}

// FormKeyMap is huh's default key map with esc cancelling the form.
func FormKeyMap() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	keys.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	return keys
}

// SaveKey saves a form from whichever field has focus.
var SaveKey = key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save"))

// UpdateForm passes msg to form, whose fields are fields, saving the form
// instead when msg is SaveKey.
func UpdateForm(form *huh.Form, fields []huh.Field, msg tea.Msg) (*huh.Form, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, SaveKey) {
		return form, saveForm(form, fields)
	}
	model, cmd := form.Update(msg)
	if updated, ok := model.(*huh.Form); ok {
		form = updated
	}
	return form, cmd
}

// saveForm submits form once every one of its fields is valid. huh validates a
// field only as focus leaves it, so each field is checked here; when any
// fails, the fields show their errors and focus stays where it was.
func saveForm(form *huh.Form, fields []huh.Field) tea.Cmd {
	focused := form.GetFocusedField()
	for _, field := range fields {
		field.Blur()
	}
	if len(form.Errors()) > 0 {
		return focused.Focus()
	}
	return form.NextGroup()
}

// FitForm sizes form to the content area under a screen title and above its
// help line. huh cannot lay out at the 1×1 floor ContentSize returns before
// the first window size arrives, so until then form is returned unchanged.
func FitForm(form *huh.Form, title string) *huh.Form {
	if constants.WindowSize.Width == 0 {
		return form
	}
	const helpRows = 2
	width, height := ContentSize(lipgloss.Height(MenuTitle.Render(title)) + helpRows)
	return form.WithWidth(width).WithHeight(height)
}

// RenderFormScreen draws form, built without huh's help, under its screen
// title and above a help line of the focused field's keys and extraKeys. The
// help line gives way to the form's errors, as huh's own does.
func RenderFormScreen(title string, form *huh.Form, extraKeys ...key.Binding) string {
	rows := []string{MenuTitle.Render(title), form.View()}
	if len(form.Errors()) == 0 {
		rows = append(rows, "", Help.ShortHelpView(append(form.KeyBinds(), extraKeys...)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
