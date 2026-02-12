package events

type OpenInVSCodeComplete struct{ Err error }

func (OpenInVSCodeComplete) isEvent() {}
