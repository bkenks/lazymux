# 02-00-00 — the palette comes from three base colors

## Context

The UI had two fixed themes, `default` and `mono`, plus an optional accent color
that replaced three of the theme's eight colors. The bubbles list, the huh forms
and the progress bars kept their own built-in pinks, indigos, greens and grays,
so a new accent sat next to colors picked for the old one. The selected row, for
example, drew its title in the new accent and its description in bubbles' pink.

## Decision

**Three base colors, `main`, `accent` and `gray`, replace the themes and the
accent.** The user sets them once for dark terminals and once for light ones,
and lazymux uses the set that matches the background it detects at launch.
Each color is stretched into an 11-step scale in OkLCh: a pale tint at step 0,
the base itself at step 5 and a deep shade at step 10, all with the base's
hue. Every UI color is a named role (`Main`, `AccentMuted`, `Subdued`,
and so on) that picks a step, with one step for a dark terminal and another for
a light one. The lazymux styles, the bubbles list and row styles, the huh theme,
the progress bars and the splash gradient all read those roles. The only color
left from the libraries is huh's red for validation errors.

The base sits in the middle of its scale unchanged, so the color the user typed
is the one on the title bar or the selected row. The other steps move toward
white and black from there. A dark base makes darker derived shades, which
keeps the user's choice visible.

## Consequences

- Picking one color re-colors everything built from it, with no leftover
  default pink or indigo.
- `mono` is gone. A gray `main` and `accent` gets close to it.
- `ui.theme` and `ui.accentColor` in an existing config are ignored. The three
  colors start empty, which means the defaults, and the defaults look like the
  old `default` theme.
- A very pale or very dark base leaves little room on one side of its scale, so
  the shades on that side come out close together.
