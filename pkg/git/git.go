package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
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

	if err := executeCommand(fmt.Sprintf("git clone %s %s", repoURL, tempDir)); err != nil {
		return "", getRedactedError(err, tempDir)
	}

	if err := executeCommand(fmt.Sprintf("git -C %s checkout %s", tempDir, branchName)); err != nil {
		return "", getRedactedError(err, tempDir)
	}

	newBranchName := fmt.Sprintf("%s-cherry-pick-%s", branchName, commitSHA)
	if err := executeCommand(fmt.Sprintf("git -C %s checkout -b %s", tempDir, newBranchName)); err != nil {
		return "", getRedactedError(err, tempDir)
	}

	if err := executeCommand(fmt.Sprintf("git -C %s cherry-pick %s", tempDir, commitSHA)); err != nil {
		return "", getRedactedError(err, tempDir)
	}

	if err := executeCommand(fmt.Sprintf("git -C %s push origin %s", tempDir, newBranchName)); err != nil {
		return "", getRedactedError(err, tempDir)
	}

	return newBranchName, nil
}

func getRedactedError(err error, toBeRedacted ...string) error {
	result := err.Error()
	for _, toBeRedactedString := range toBeRedacted {
		result = strings.Replace(result, toBeRedactedString, "REDACTED", -1)
	}
	return fmt.Errorf(result)
}

func executeCommand(command string) error {
	commandSlice := strings.Fields(command)
	if len(commandSlice) < 1 {
		return fmt.Errorf("command must not be empty!")
	}
	var args []string
	if len(commandSlice) > 1 {
		args = commandSlice[1:]
	}

	glog.V(6).Infof("Executing command %s", command)
	if out, err := exec.Command(commandSlice[0], args...).CombinedOutput(); err != nil {
		return fmt.Errorf("error executing command %s: out=%s, err=%v", command, string(out), err)
	}

	return nil
}
