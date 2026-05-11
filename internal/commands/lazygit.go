package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

func LazygitCmd(absPath string) tea.Cmd {
	return TeaCmdBuilder(cfg().Tools.Lazygit, "-p", absPath)
}
