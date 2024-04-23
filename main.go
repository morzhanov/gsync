package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Get the current directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}

	// Check if the directory is a git repository
	if !isGitRepo(wd) {
		fmt.Println("Not a git repository.")
		return
	}

	// Check if the "gsync" remote exists
	if !remoteExists("gsync") {
		fmt.Println("Remote 'gsync' not found.")
		url := getInput("Enter the URL for the 'gsync' repository:")
		setupRemote("gsync", url)
	}

	// Prompt the user to choose action
	action := getInput("Do you want to push or pull changes? (push/pull):")

	switch action {
	case "push":
		pushChanges()
	case "pull":
		pullChanges()
	default:
		fmt.Println("Invalid action. Exiting.")
	}
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

func remoteExists(remote string) bool {
	cmd := exec.Command("git", "remote", "get-url", remote)
	err := cmd.Run()
	return err == nil
}

func getInput(prompt string) string {
	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func setupRemote(remote, url string) {
	cmd := exec.Command("git", "remote", "add", remote, url)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to set up remote:", err)
	}
}

func pushChanges() {
	// Commit changes
	execGitCmd("git", "add", "-A")
	execGitCmd("git", "commit", "-m", "gsync auto commit")

	// Push to remote with force
	execGitCmd("git", "push", "gsync", "HEAD:gsync", "--force")

	// Reset --soft to remove the commit locally
	execGitCmd("git", "reset", "--soft", "HEAD^")
}

func pullChanges() {
	// Check if there are changes in the working directory
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to check for changes:", err)
		return
	}

	if len(out) > 0 {
		// There are changes, ask if user wants to stash them
		stash := getInput("There are uncommitted changes. Do you want to stash them? (y/n):")
		if stash == "y" {
			execGitCmd("git", "stash", "push", "--include-untracked")
		} else {
			fmt.Println("Exiting without pulling changes.")
			return
		}
	}

	// Fetch from remote
	execGitCmd("git", "fetch", "gsync")

	// Reset --hard to the remote branch
	execGitCmd("git", "reset", "--hard", "gsync/gsync")

	// Reset --soft to remove the last commit but keep the changes
	execGitCmd("git", "reset", "--soft", "HEAD^")

	fmt.Printf("performed git reset --soft to keep the pulled changes without commit")
}

func execGitCmd(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to execute '%s': %v\n", command, err)
	}
}
