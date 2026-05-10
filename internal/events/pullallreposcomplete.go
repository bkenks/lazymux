package events

type SkippedPull struct {
	RepoPath string
	Reason   string
}

type PullAllReposComplete struct {
	Pulled  int
	Skipped []SkippedPull
}

func (PullAllReposComplete) isEvent() {}
