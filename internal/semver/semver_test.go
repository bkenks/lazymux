package semver

import (
	"testing"
	"testing/quick"
)

func TestParse(t *testing.T) {
	cases := []struct {
		format Format
		tag    string
		want   Version
		ok     bool
	}{
		{Format{}, "1.2.3", Version{1, 2, 3}, true},
		{Format{Prefix: "v"}, "v1.2.3", Version{1, 2, 3}, true},
		{Format{Prefix: "mypkg/v"}, "mypkg/v10.0.7", Version{10, 0, 7}, true},
		{Format{Suffix: "-mypkg"}, "1.2.3-mypkg", Version{1, 2, 3}, true},
		{Format{Prefix: "v"}, "1.2.3", Version{}, false},
		{Format{}, "v1.2.3", Version{}, false},
		{Format{Suffix: "-mypkg"}, "1.2.3", Version{}, false},
		{Format{Suffix: "-mypkg"}, "1.2.3-mypkg-rc1", Version{}, false},
		{Format{Prefix: "v"}, "v1.2", Version{}, false},
		{Format{Prefix: "v"}, "v01.2.3", Version{}, false},
		{Format{Prefix: "a.b"}, "axb1.2.3", Version{}, false},
		{Format{}, "", Version{}, false},
	}
	for _, c := range cases {
		got, ok := c.format.Parse(c.tag)
		if ok != c.ok || got != c.want {
			t.Errorf("%+v.Parse(%q) = %v, %v; want %v, %v", c.format, c.tag, got, ok, c.want, c.ok)
		}
	}
}

func TestLatestPicksHighestMatchingTag(t *testing.T) {
	tags := []string{"v1.9.0", "v1.10.0", "v2.0.0-rc", "mypkg/v9.0.0", "v1.2.3", "junk"}

	got, ok := Format{Prefix: "v"}.Latest(tags)

	if !ok || got != (Version{1, 10, 0}) {
		t.Errorf("Latest = %v, %v; want 1.10.0", got, ok)
	}
}

func TestLatestWithNoMatchingTag(t *testing.T) {
	if got, ok := (Format{Prefix: "v"}).Latest([]string{"1.0.0", "release"}); ok {
		t.Errorf("Latest = %v, want no match", got)
	}
}

func TestBump(t *testing.T) {
	v := Version{1, 2, 3}
	cases := map[Bump]Version{
		Patch: {1, 2, 4},
		Minor: {1, 3, 0},
		Major: {2, 0, 0},
	}
	for bump, want := range cases {
		if got := v.Bump(bump); got != want {
			t.Errorf("%v.Bump(%s) = %v, want %v", v, bump, got, want)
		}
	}
}

func TestTagRoundTripsThroughParse(t *testing.T) {
	formats := []Format{
		{Prefix: "mypkg/v", Suffix: "-mypkg"},
		{Prefix: "v1", Suffix: "1"},
		{Prefix: "1.", Suffix: ".1"},
	}
	for _, format := range formats {
		roundTrips := func(major, minor, patch uint16) bool {
			v := Version{int(major), int(minor), int(patch)}
			got, ok := format.Parse(format.Tag(v))
			return ok && got == v
		}
		if err := quick.Check(roundTrips, nil); err != nil {
			t.Errorf("%+v: %v", format, err)
		}
	}
}

func TestBumpAlwaysIncreases(t *testing.T) {
	increases := func(major, minor, patch uint16, part uint8) bool {
		v := Version{int(major), int(minor), int(patch)}
		return v.Compare(v.Bump(Bumps[int(part)%len(Bumps)])) < 0
	}
	if err := quick.Check(increases, nil); err != nil {
		t.Error(err)
	}
}
