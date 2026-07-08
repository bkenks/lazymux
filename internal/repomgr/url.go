// Package repomgr is lazymux's native replacement for ghq: it parses clone
// URLs, decides on-disk locations (<baseDir>/<namespace>/<repo>), and manages
// each repo's placeholder origin + primary-forge insteadOf rewrite.
package repomgr

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	SchemeHTTPS = "https"
	SchemeSSH   = "ssh"
)

// RepoURL is a parsed git remote, split into the parts lazymux cares about.
type RepoURL struct {
	Scheme    string // SchemeHTTPS | SchemeSSH
	Host      string // e.g. github.com
	Namespace string // e.g. bkenks, or group/subgroup for nested namespaces
	Name      string // repo name, without .git
}

// Key is the namespace-qualified identifier used as the map key in config and
// as the on-disk path suffix: "<namespace>/<name>".
func (u RepoURL) Key() string {
	if u.Namespace == "" {
		return u.Name
	}
	return u.Namespace + "/" + u.Name
}

var (
	// scp-like syntax: git@host:namespace/repo(.git)
	scpRe = regexp.MustCompile(`^(?:([^@]+)@)?([^:/]+):(.+?)(?:\.git)?/?$`)
	// scheme URL: https://host/ns/repo  or  ssh://git@host[:port]/ns/repo
	urlRe = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*)://(?:[^@/]+@)?([^:/]+)(?::\d+)?/(.+?)(?:\.git)?/?$`)
)

// ParseRepoURL parses https, http, ssh:// and scp-like (git@host:ns/repo) forms.
func ParseRepoURL(raw string) (RepoURL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RepoURL{}, fmt.Errorf("empty url")
	}

	if m := urlRe.FindStringSubmatch(raw); m != nil {
		scheme := SchemeHTTPS
		if strings.EqualFold(m[1], "ssh") || strings.EqualFold(m[1], "git") {
			scheme = SchemeSSH
		}
		ns, name, err := splitPath(m[3])
		if err != nil {
			return RepoURL{}, err
		}
		return RepoURL{Scheme: scheme, Host: m[2], Namespace: ns, Name: name}, nil
	}

	if m := scpRe.FindStringSubmatch(raw); m != nil {
		ns, name, err := splitPath(m[3])
		if err != nil {
			return RepoURL{}, err
		}
		return RepoURL{Scheme: SchemeSSH, Host: m[2], Namespace: ns, Name: name}, nil
	}

	return RepoURL{}, fmt.Errorf("unrecognized git url: %q", raw)
}

// splitPath separates "group/subgroup/repo" into namespace + repo name.
func splitPath(p string) (namespace, name string, err error) {
	p = strings.TrimPrefix(strings.TrimSuffix(p, "/"), "/")
	p = strings.TrimSuffix(p, ".git")
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		if p == "" {
			return "", "", fmt.Errorf("url has no repo path")
		}
		return "", p, nil // no namespace
	}
	return p[:idx], p[idx+1:], nil
}

// normalizeScheme coerces an arbitrary scheme string to one we support.
func normalizeScheme(scheme string) string {
	if strings.EqualFold(scheme, SchemeSSH) {
		return SchemeSSH
	}
	return SchemeHTTPS
}

// hostBase returns the URL prefix for a host under a given scheme — the value
// used on both sides of an insteadOf rule and as the origin URL stem.
//
//	https -> "https://host/"
//	ssh   -> "git@host:"
func hostBase(scheme, host string) string {
	if normalizeScheme(scheme) == SchemeSSH {
		return "git@" + host + ":"
	}
	return "https://" + host + "/"
}

// RemoteURL builds a clone/remote URL for a host, scheme and repo key.
func RemoteURL(scheme, host, key string) string {
	return hostBase(scheme, host) + key + ".git"
}
