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

// ToastAnim advances the toast fade state machine (fade-in → hold → fade-out)
// one frame. Emitted by a tea.Tick after Toast fires and reschedules itself
// until the toast has faded out. Seq guards against an old animation driving a
// newer toast — a stale tick (Seq != current) is ignored.
type ToastAnim struct{ Seq int }

func (ToastAnim) isEvent() {}
