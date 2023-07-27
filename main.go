package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func getRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func updateRemoteURL(path, newURL string) error {
	cmd := exec.Command("git", "remote", "set-url", "origin", newURL)
	cmd.Dir = path
	return cmd.Run()
}

func getChangedFiles(path string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var changedFiles []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 1 {
			changedFiles = append(changedFiles, parts[1])
		}
	}
	return changedFiles, nil
}

func resetAndPullRepo(path string) error {
	cmd := exec.Command("git", "reset", "--hard", "origin/HEAD")
	cmd.Dir = path
	cmd.Run()

	cmdClean := exec.Command("git", "clean", "-fd")
	cmdClean.Dir = path
	cmdClean.Run()

	cmdPull := exec.Command("git", "pull")
	cmdPull.Dir = path
	return cmdPull.Run()
}

func correctURL(remoteURL string) string {
	githubPrefix := "https://github.com/"
	gitlabPrefix := "https://gitlab.com/"

	if strings.Contains(remoteURL, githubPrefix) {
		return remoteURL[strings.Index(remoteURL, githubPrefix):]
	}
	if strings.Contains(remoteURL, gitlabPrefix) {
		return remoteURL[strings.Index(remoteURL, gitlabPrefix):]
	}
	return remoteURL
}

func main() {
	excludeDirs := flag.String("x", "", "Comma-separated list of directories to exclude")
	flag.Parse()

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		return
	}

	files, err := os.ReadDir(currentDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading current directory: %v\n", err)
		return
	}

	const targetPrefix = "https://github.com/"
	var failedRepos []string
	reader := bufio.NewReader(os.Stdin)

	for _, file := range files {
		if file.IsDir() && !strings.Contains(*excludeDirs, file.Name()) {
			repoPath := filepath.Join(currentDir, file.Name())
			if isGitRepo(repoPath) {
				changedFiles, _ := getChangedFiles(repoPath)
				if len(changedFiles) > 0 {
					fmt.Printf("Changed files in repo %s:\n", repoPath)
					for _, f := range changedFiles {
						fmt.Println("  ", f)
					}
				} else {
					fmt.Printf("No changes in repo %s\n", repoPath)
				}

				remoteURL, err := getRemoteURL(repoPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting remote URL for repo %s: %v\n", repoPath, err)
					continue
				}

				correctedURL := correctURL(remoteURL)
				if correctedURL != remoteURL {
					err := updateRemoteURL(repoPath, correctedURL)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error updating remote URL for repo %s: %v\n", repoPath, err)
						continue
					}
					fmt.Printf("Updated remote URL for repo at %s\n", repoPath)
				}

				err = resetAndPullRepo(repoPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error updating repo %s due to local changes: %v\n", repoPath, err)
					fmt.Printf("Do you want to reset local changes and pull the latest updates for %s? (y/n): ", repoPath)
					response, _ := reader.ReadString('\n')
					response = strings.TrimSpace(response)

					if response == "y" {
						err = resetAndPullRepo(repoPath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error after reset and pull for repo %s: %v\n", repoPath, err)
						}
					} else {
						failedRepos = append(failedRepos, repoPath)
					}
				}
			}
		}
	}

	if len(failedRepos) > 0 {
		fmt.Println("\nRepositories not updated due to local changes:")
		for _, repo := range failedRepos {
			fmt.Println(repo)
		}
	}
}
