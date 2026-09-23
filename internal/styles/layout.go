package styles

import "github.com/bkenks/lazymux/internal/constants"

// ContentSize is the space a screen has to draw in: the window minus
// DocStyle's frame, the app's footer, and reservedRows the screen keeps for
// itself. Both dimensions are at least 1.
func ContentSize(reservedRows int) (width, height int) {
	x, y := DocStyle.GetFrameSize()
	width = max(constants.WindowSize.Width-x, 1)
	height = max(constants.WindowSize.Height-y-constants.FooterReservedLines-reservedRows, 1)
	return width, height
}

// ProgressBarWidth fits a progress bar into width, leaving room for its label.
func ProgressBarWidth(width int) int {
	return min(max(width-24, 10), 40)
}
