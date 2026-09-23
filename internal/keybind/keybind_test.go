package keybind

import "testing"

func TestParseCanonicalizes(t *testing.T) {
	cases := map[string]string{
		"ctrl + p":             "ctrl+p",
		"Ctrl + Shift + K":     "ctrl+shift+k",
		"shift+ctrl+k":         "ctrl+shift+k",
		"cmd + r":              "super+r",
		"option+command+6":     "alt+super+6",
		"K":                    "shift+k",
		"shift + K":            "shift+k",
		"Ctrl + G":             "ctrl+g",
		"alt+K":                "alt+k",
		"tab":                  "tab",
		"return":               "enter",
		"del":                  "delete",
		"caps":                 "capslock",
		"ctrl + space":         "ctrl+space",
		"F5":                   "f5",
		"alt + /":              "alt+/",
		"+":                    "+",
		"ctrl + +":             "ctrl++",
		"  ctrl   +   x   ":    "ctrl+x",
		"ctrl+alt+shift+cmd+z": "ctrl+alt+shift+super+z",
	}
	for input, want := range cases {
		got, err := Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("Parse(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseRejectsInvalid(t *testing.T) {
	for _, input := range []string{
		"",
		"   ",
		"ctrl",
		"ctrl + shift",
		"ctrl +",
		"ctrl ++ p",
		"ctrl + p + q",
		"ctrl + ctrl + p",
		"ctrl + banana",
		"f13",
		"hyper + p",
	} {
		if got, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) = %q, want error", input, got)
		}
	}
}

func TestFindClash(t *testing.T) {
	taken := []string{"n", "S", "ctrl+c", "not a key"}

	if match, ok := FindClash("shift+s", taken); !ok || match != "S" {
		t.Errorf("FindClash(shift+s) = %q, %v; want S, true", match, ok)
	}
	if match, ok := FindClash("n", taken); !ok || match != "n" {
		t.Errorf("FindClash(n) = %q, %v; want n, true", match, ok)
	}
	if match, ok := FindClash("ctrl+p", taken); ok {
		t.Errorf("FindClash(ctrl+p) = %q, true; want no clash", match)
	}
}
