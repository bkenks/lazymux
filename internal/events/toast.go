package events

type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastError
)

// Toast asks the app to show a transient status message at the bottom of the screen.
type Toast struct {
	Msg   string
	Level ToastLevel
}

func (Toast) isEvent() {}

// ToastClear hides the currently displayed toast if its sequence number still
// matches the most recent Toast. Emitted by a tea.Tick after Toast fires.
// Seq guards against an old ticker clearing a newer toast.
type ToastClear struct{ Seq int }

func (ToastClear) isEvent() {}
