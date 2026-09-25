package events

// OpenTagRelease opens the screen that tags and pushes the next version of the
// repo at Dir. Tags are the repo's existing tags.
type OpenTagRelease struct {
	Key, Dir string
	Tags     []string
}

func (OpenTagRelease) isEvent() {}

// TagReleased reports the outcome of creating and pushing a release tag.
type TagReleased struct {
	Tag string
	Err error
}

func (TagReleased) isEvent() {}
