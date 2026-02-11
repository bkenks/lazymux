package repolist

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetFullRepoPath(repo string) string {
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
