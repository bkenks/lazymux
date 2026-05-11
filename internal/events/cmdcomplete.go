package events

type CmdComplete struct{ Err error }

func (CmdComplete) isEvent() {}
