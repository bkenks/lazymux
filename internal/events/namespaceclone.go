package events

// NamespaceCloneFailed is emitted when listing a namespace/org's repos via
// the forge's CLI (gh for github.com, tea for Gitea/Forgejo hosts) fails, or
// comes back with no repos.
type NamespaceCloneFailed struct {
	Namespace string
	Err       error
}

func (NamespaceCloneFailed) isEvent() {}
