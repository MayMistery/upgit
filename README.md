### upgit (Git Repository Updater)

**Description**:
Designed to update all the Git repositories in current directory

1. Retrieve the current remote URL for the repository.
2. Correct any malformed remote URLs that have prefixes other than `https://github.com/` to ensure they strictly begin with `https://github.com/`.
3. Execute a `git pull` operation to update the repository to the latest version from the remote.
