package main

import (
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

func updateRepo(path string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull repo %v: %v\n%s", path, err, out)
	}
	fmt.Printf("Updated repo at %s\n", path)
	return nil
}

func main() {
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

	for _, file := range files {
		if file.IsDir() {
			repoPath := filepath.Join(currentDir, file.Name())
			if isGitRepo(repoPath) {
				// 获取远程URL并更新（如果需要）
				remoteURL, err := getRemoteURL(repoPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting remote URL for repo %s: %v\n", repoPath, err)
					continue
				}

				// 查找目标前缀，并剪切URL
				if idx := strings.Index(remoteURL, targetPrefix); idx != -1 {
					newURL := remoteURL[idx:]
					if newURL != remoteURL {
						err := updateRemoteURL(repoPath, newURL)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error updating remote URL for repo %s: %v\n", repoPath, err)
							continue
						}
						fmt.Printf("Updated remote URL for repo at %s\n", repoPath)
					}
				}

				// 拉取更新
				err = updateRepo(repoPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error updating repo %s: %v\n", repoPath, err)
				}
			}
		}
	}
}
