package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	TempDirPrefix = "cherry-pick-bot"
)

func PushCherryPick(repoURL, branchName, commitSHA string) (string, error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), TempDirPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	if out, err := exec.Command("git", "clone", repoURL, tempDir).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to clone: output=%s, err=%v", string(out), err)
	}

	if out, err := exec.Command("git", "-C", tempDir, "checkout", branchName).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to checkout branch %s: output=%s, err=%v", branchName, string(out), err)
	}

	newBranchName := fmt.Sprintf("%s-cherry-pick-%s", branchName, commitSHA)
	if out, err := exec.Command("git", "-C", tempDir, "checkout", "-b", newBranchName).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create branch %s: out=%s, err=%v", newBranchName, string(out), err)
	}

	if out, err := exec.Command("git", "-C", tempDir, "cherry-pick", commitSHA).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to cherry-pick commit %s: out=%s, err=%v", commitSHA, string(out), err)
	}

	if out, err := exec.Command("git", "-C", tempDir, "push", "origin", newBranchName).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to push to branch %s; out=%s, err=%v", newBranchName, string(out), err)
	}

	return newBranchName, nil
}
