package constants

import (
	tea "github.com/charmbracelet/bubbletea"
)

// WindowSize tracks the latest tea.WindowSizeMsg so views that compute their
// own layout can read terminal dimensions without each one subscribing
// independently.
var WindowSize tea.WindowSizeMsg

// FooterReservedLines is the number of rows ModelManager renders below the
// active view (status footer / toast). Sub-models that compute their own
// height must subtract this so their content doesn't overflow.
const FooterReservedLines = 1
