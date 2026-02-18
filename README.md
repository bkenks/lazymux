# lazymux
A Terminal User Interface application that combines multiple applications, for a full GitHub client application in the terminal!

## Requirements
-  [lazygit](https://github.com/jesseduffield/lazygit) by [jesseduffield](https://github.com/jesseduffield)
	- `lazygit` is a TUI (Terminal User Interface) for handling most `git` operations like commits, PRs, pulls, pushes, etc.
- [ghq](https://github.com/x-motemen/ghq) by [x-motemen](https://github.com/x-motemen)
	- `ghq` is a repository manager that handles `git` functionality such as cloning, deleting, and listing repositories while maintaining a clean folder structure to logically store your repositories. This makes it possible for `lazymux` to provide a UI for repository management since it manages repositories in a predictable folder structure.


## Installation

### Standard Installation (Binary File)
1. Follow the links in [Requirements](#requirements) to install `ghq` and `lazygit`
2. Run the installer
```bash
curl -fsSL https://raw.githubusercontent.com/bkenks/lazymux/main/installer.sh | sh
```
_Note: On Mac, you may be prompted saying the program is not verified or signed, that is because I am not paying for the Apple Developer License. To get around it, close the prompt, go to Settings -> Privacy & Security then all the way at the bottom there will be a new item saying something along the lines of "Allow Lazymux", click that, run the program again, and then continue through the prompts it gives. It will only prompt you once about this. **If you'd like to get around this entirely, use the "[Go Installation](#go%20installation)" below!_**

### Go Installation
1. Follow the links in [Requirements](#requirements) to install `ghq` and `lazygit`
2. Run the following to install with `Go`:
	```bash
	go install github.com/bkenks/lazymux@latest
	```