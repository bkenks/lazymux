// Package formtest drives a screen in tests the way the bubbletea runtime
// would, feeding every message a command returns back into the screen.
package formtest

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Press sends msg to screen and feeds every resulting message back into it,
// returning, instead of delivering, each one isEmitted reports as meant for the
// app.
func Press(screen tea.Model, msg tea.Msg, isEmitted func(tea.Msg) bool) []tea.Msg {
	var emitted []tea.Msg
	queue := []tea.Msg{msg}
	for steps := 0; len(queue) > 0 && steps < 200; steps++ {
		next := queue[0]
		queue = queue[1:]
		if isEmitted(next) {
			emitted = append(emitted, next)
			continue
		}
		_, cmd := screen.Update(next)
		queue = append(queue, runCmd(cmd)...)
	}
	return emitted
}

// runCmd runs cmd, dropping it if it hasn't returned within a short wait —
// those are timers such as the cursor blink, which the tests don't need.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	result := make(chan tea.Msg, 1)
	go func() { result <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-result:
	case <-time.After(50 * time.Millisecond):
		return nil
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var msgs []tea.Msg
	for _, inner := range batch {
		msgs = append(msgs, runCmd(inner)...)
	}
	return msgs
}

// Type sends text to screen one key press at a time.
func Type(screen tea.Model, text string) {
	for _, r := range text {
		Press(screen, tea.KeyPressMsg{Code: r, Text: string(r)}, func(tea.Msg) bool { return false })
	}
}
