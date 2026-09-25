// Package semver reads and bumps MAJOR.MINOR.PATCH versions stored in git
// tags wrapped in a fixed prefix and suffix, such as v1.2.3 or mypkg/v1.2.3.
package semver

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"
)

// Version is a MAJOR.MINOR.PATCH version.
type Version struct {
	Major, Minor, Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare returns -1, 0 or +1 as v is lower than, equal to or higher than o.
func (v Version) Compare(o Version) int {
	if c := cmp.Compare(v.Major, o.Major); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Minor, o.Minor); c != 0 {
		return c
	}
	return cmp.Compare(v.Patch, o.Patch)
}

// Bump names the version part a release increments.
type Bump int

const (
	Patch Bump = iota
	Minor
	Major
)

// Bumps lists every bump, smallest first.
var Bumps = []Bump{Patch, Minor, Major}

func (b Bump) String() string {
	switch b {
	case Major:
		return "major"
	case Minor:
		return "minor"
	default:
		return "patch"
	}
}

// Bump returns the next version for part, zeroing the parts below it.
func (v Version) Bump(part Bump) Version {
	switch part {
	case Major:
		return Version{Major: v.Major + 1}
	case Minor:
		return Version{Major: v.Major, Minor: v.Minor + 1}
	default:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	}
}

// Format is the text a repo's version tags wrap around the version:
// <Prefix>MAJOR.MINOR.PATCH<Suffix>.
type Format struct {
	Prefix, Suffix string
}

// Tag returns the tag name for v.
func (f Format) Tag(v Version) string {
	return f.Prefix + v.String() + f.Suffix
}

// Parse reads the version out of a tag in this format. A tag with another
// prefix or suffix, a missing part or a leading zero doesn't match.
func (f Format) Parse(tag string) (Version, bool) {
	match := f.pattern().FindStringSubmatch(tag)
	if match == nil {
		return Version{}, false
	}
	parts := make([]int, 3)
	for i := range parts {
		n, err := strconv.Atoi(match[i+1])
		if err != nil {
			return Version{}, false
		}
		parts[i] = n
	}
	return Version{Major: parts[0], Minor: parts[1], Patch: parts[2]}, true
}

// Latest returns the highest version among the tags in this format, or false
// when none of them are.
func (f Format) Latest(tags []string) (Version, bool) {
	var latest Version
	found := false
	for _, tag := range tags {
		v, ok := f.Parse(tag)
		if ok && (!found || v.Compare(latest) > 0) {
			latest, found = v, true
		}
	}
	return latest, found
}

func (f Format) pattern() *regexp.Regexp {
	const part = `(0|[1-9][0-9]*)`
	return regexp.MustCompile(`^` + regexp.QuoteMeta(f.Prefix) +
		part + `\.` + part + `\.` + part + regexp.QuoteMeta(f.Suffix) + `$`)
}
