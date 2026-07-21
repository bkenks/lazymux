package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

// giteaPageSize and giteaMaxPages bound the `tea repos list` pagination loop
// (github.com goes through `gh repo list`, which paginates internally given
// a single --limit).
const (
	giteaPageSize = 50
	giteaMaxPages = 20
	githubListCap = 1000
)

// ListNamespaceReposCmd shells out to the forge's CLI (gh for github.com, tea
// for Gitea/Forgejo hosts) to list every repo under namespace, then hands the
// resulting clone URLs to the same StartRepoClone flow used for pasted URLs.
func ListNamespaceReposCmd(forge config.Forge, namespace string) tea.Cmd {
	protocol := cfg().Behavior.DefaultProtocol
	return func() tea.Msg {
		urls, err := listNamespaceRepos(forge, namespace, protocol)
		if err != nil {
			return events.NamespaceCloneFailed{Namespace: namespace, Err: err}
		}
		if len(urls) == 0 {
			return events.NamespaceCloneFailed{
				Namespace: namespace,
				Err:       fmt.Errorf("no repos found for %q on %s", namespace, forge.Host),
			}
		}
		return events.StartRepoClone{RepoUrls: urls}
	}
}

func listNamespaceRepos(forge config.Forge, namespace, protocol string) ([]string, error) {
	if strings.EqualFold(forge.Host, "github.com") {
		return listGitHubRepos(namespace, protocol)
	}
	return listGiteaRepos(forge, namespace, protocol)
}

func listGitHubRepos(namespace, protocol string) ([]string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI not found: install it to clone a github.com namespace")
	}
	out, err := exec.Command("gh", "repo", "list", namespace,
		"--json", "sshUrl,url", "--limit", strconv.Itoa(githubListCap)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh repo list %s: %s", namespace, firstLine(string(out)))
	}
	var repos []struct {
		SSHUrl string `json:"sshUrl"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("gh repo list %s: unexpected output: %w", namespace, err)
	}
	urls := make([]string, 0, len(repos))
	for _, r := range repos {
		if protocol == "ssh" {
			urls = append(urls, r.SSHUrl)
		} else {
			urls = append(urls, r.URL)
		}
	}
	return urls, nil
}

func listGiteaRepos(forge config.Forge, namespace, protocol string) ([]string, error) {
	if _, err := exec.LookPath("tea"); err != nil {
		return nil, fmt.Errorf("tea CLI not found: install it to clone a %s namespace", forge.Host)
	}
	login, err := teaLoginForHost(forge.Host)
	if err != nil {
		return nil, err
	}

	var urls []string
	for page := 1; page <= giteaMaxPages; page++ {
		out, err := exec.Command("tea", "repos", "list",
			"--owner", namespace,
			"--fields", "ssh,url",
			"--output", "json",
			"--login", login,
			"--page", strconv.Itoa(page),
			"--limit", strconv.Itoa(giteaPageSize),
		).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("tea repos list %s: %s", namespace, firstLine(string(out)))
		}
		var repos []struct {
			SSH string `json:"ssh"`
			URL string `json:"url"`
		}
		if err := json.Unmarshal(out, &repos); err != nil {
			return nil, fmt.Errorf("tea repos list %s: unexpected output: %w", namespace, err)
		}
		for _, r := range repos {
			if protocol == "ssh" {
				urls = append(urls, r.SSH)
			} else {
				urls = append(urls, r.URL)
			}
		}
		if len(repos) < giteaPageSize {
			break
		}
	}
	return urls, nil
}

// teaLoginForHost finds the `tea login` whose URL host matches the forge's
// host, since `tea repos list --login` takes a login name, not a host.
func teaLoginForHost(host string) (string, error) {
	out, err := exec.Command("tea", "login", "list", "-o", "json").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tea login list: %s", firstLine(string(out)))
	}
	var logins []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := json.Unmarshal(out, &logins); err != nil {
		return "", fmt.Errorf("tea login list: unexpected output: %w", err)
	}
	for _, l := range logins {
		u, err := url.Parse(l.URL)
		if err != nil {
			continue
		}
		if strings.EqualFold(u.Hostname(), host) {
			return l.Name, nil
		}
	}
	return "", fmt.Errorf("no `tea login` configured for %s; run `tea login add`", host)
}
