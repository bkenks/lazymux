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
	cmd := exec.Command("ghq", "list", "--full-path", repo)
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error getting repo path:", repo)
		os.Exit(1)
	}

	/////////////////////////////////////

	path := strings.TrimSpace(string(out))
	return path
}

func ConvertToRepoType(i list.Item) domain.Repo {
	if domainRepo, ok := i.(domain.Repo); ok {
		return domainRepo
	}

	return domain.Repo{}
}

func AbsRepoPath(i list.Item) string {
	domainRepo := ConvertToRepoType(i)
	absRepoPath := GetAbsRepoPath(domainRepo.Path)

	return absRepoPath
}

func SizeBuffer() (width, height int) {
	x, y := styles.DocStyle.GetFrameSize()
	widthBuffer := constants.WindowSize.Width - x
	heightBuffer := constants.WindowSize.Height - y
	return widthBuffer, heightBuffer
}
