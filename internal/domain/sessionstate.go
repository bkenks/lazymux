package domain

type SessionState int

const (
	StateSplash SessionState = iota
	StateMain
	StateConfirmDelete
	StateCloneRepo
	StateSettings
	StateForgeSelect   // choose forge links for repos being cloned
	StateForgeRegistry // manage the forge registry
	StateRepoForges    // change a repo's primary/scheme/links
)
