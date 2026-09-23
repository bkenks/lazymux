// Package keybind parses the key combos users type for custom keybinds (e.g.
// "ctrl + p") into the keystroke strings bubbletea reports for a key press.
package keybind

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// KeyNamesHelp lists the key names that aren't a single printable character.
const KeyNamesHelp = "Key Names: ctrl, alt, cmd, shift, tab, caps, return, esc, space, " +
	"backspace, del, up, down, left, right, home, end, pgup, pgdown, insert, f1-f12"

// modifierOrder is the order bubbletea writes modifiers in a keystroke.
var modifierOrder = []string{"ctrl", "alt", "shift", "super"}

var modifierAliases = map[string]string{
	"ctrl":    "ctrl",
	"control": "ctrl",
	"alt":     "alt",
	"opt":     "alt",
	"option":  "alt",
	"shift":   "shift",
	"cmd":     "super",
	"command": "super",
	"super":   "super",
}

var namedKeys = map[string]string{
	"tab":       "tab",
	"caps":      "capslock",
	"capslock":  "capslock",
	"return":    "enter",
	"enter":     "enter",
	"esc":       "esc",
	"escape":    "esc",
	"space":     "space",
	"backspace": "backspace",
	"del":       "delete",
	"delete":    "delete",
	"up":        "up",
	"down":      "down",
	"left":      "left",
	"right":     "right",
	"home":      "home",
	"end":       "end",
	"pgup":      "pgup",
	"pageup":    "pgup",
	"pgdown":    "pgdown",
	"pagedown":  "pgdown",
	"insert":    "insert",
}

// Parse turns a typed combo such as "Ctrl + Shift + K" into the canonical
// keystroke "ctrl+shift+k". A lone uppercase letter is read as shift+letter,
// matching how bubbletea reports it.
func Parse(input string) (string, error) {
	tokens, err := splitCombo(input)
	if err != nil {
		return "", err
	}

	modifiers := map[string]bool{}
	var keyName string
	for _, token := range tokens {
		if modifier, ok := modifierAliases[strings.ToLower(token)]; ok {
			if modifiers[modifier] {
				return "", fmt.Errorf("%q is listed twice", token)
			}
			modifiers[modifier] = true
			continue
		}
		if keyName != "" {
			return "", fmt.Errorf("only one non-modifier key allowed, got %q and %q", keyName, token)
		}
		name, shifted, err := parseKeyName(token)
		if err != nil {
			return "", err
		}
		keyName = name
		if shifted {
			modifiers["shift"] = true
		}
	}
	if keyName == "" {
		return "", errors.New("add a key to go with the modifiers, e.g. ctrl + p")
	}
	return composeKeystroke(modifiers, keyName), nil
}

// splitCombo splits on "+", treating a trailing "++" (or a lone "+") as the
// plus key itself.
func splitCombo(input string) ([]string, error) {
	compact := strings.Join(strings.Fields(input), "")
	if compact == "" {
		return nil, errors.New("keybind cannot be empty")
	}

	isPlusKey := compact == "+" || strings.HasSuffix(compact, "++")
	if isPlusKey {
		compact = strings.TrimSuffix(strings.TrimSuffix(compact, "+"), "+")
	}

	var tokens []string
	if compact != "" {
		tokens = strings.Split(compact, "+")
	}
	for _, token := range tokens {
		if token == "" {
			return nil, fmt.Errorf("%q is missing a key between or after the + signs", input)
		}
	}
	if isPlusKey {
		tokens = append(tokens, "+")
	}
	return tokens, nil
}

func parseKeyName(token string) (name string, shifted bool, err error) {
	lower := strings.ToLower(token)
	if named, ok := namedKeys[lower]; ok {
		return named, false, nil
	}
	if isFunctionKey(lower) {
		return lower, false, nil
	}
	if utf8.RuneCountInString(token) == 1 {
		if token >= "A" && token <= "Z" {
			return lower, true, nil
		}
		return token, false, nil
	}
	return "", false, fmt.Errorf("%q isn't a key name", token)
}

func isFunctionKey(name string) bool {
	for i := 1; i <= 12; i++ {
		if name == fmt.Sprintf("f%d", i) {
			return true
		}
	}
	return false
}

func composeKeystroke(modifiers map[string]bool, keyName string) string {
	var parts []string
	for _, modifier := range modifierOrder {
		if modifiers[modifier] {
			parts = append(parts, modifier)
		}
	}
	return strings.Join(append(parts, keyName), "+")
}

// FindClash reports the first entry of taken whose keystroke is the same as
// keystroke. Entries of taken are parsed the same way, so "K" and "shift+k"
// clash. Entries that don't parse are skipped.
func FindClash(keystroke string, taken []string) (string, bool) {
	for _, candidate := range taken {
		if parsed, err := Parse(candidate); err == nil && parsed == keystroke {
			return candidate, true
		}
	}
	return "", false
}
