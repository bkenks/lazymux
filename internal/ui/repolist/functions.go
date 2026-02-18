package repolist

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/list"
)

func GetAbsRepoPath(repo string) string {
	// Get ALL repos (no filtering argument)
	cmd := exec.Command("ghq", "list", "--full-path")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting repo list")
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	var matches []string
	for _, line := range lines {
		// Exact suffix match
		if strings.HasSuffix(line, repo) {
			matches = append(matches, line)
		}
	}

	switch len(matches) {
	case 0:
		fmt.Println("No repo found:", repo)
		os.Exit(1)
	case 1:
		return matches[0]
	default:
		fmt.Println("Ambiguous repo name:", repo)
		for _, m := range matches {
			fmt.Println(" -", m)
		}
		os.Exit(1)
	}

	return ""
}

func ConvertToRepoType(i list.Item) domain.Repo {
	if domainRepo, ok := i.(domain.Repo); ok {
		return domainRepo
	}
	return domain.Repo{}
}

func AbsRepoPath(i list.Item) string {
	domainRepo := ConvertToRepoType(i)
	return GetAbsRepoPath(domainRepo.Path)
}

func SizeBuffer() (width, height int) {
	x, y := styles.DocStyle.GetFrameSize()
	widthBuffer := constants.WindowSize.Width - x
	heightBuffer := constants.WindowSize.Height - y
	return widthBuffer, heightBuffer
}
