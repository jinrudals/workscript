package runner

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type CommandInfo struct {
	Name           string
	Command        string
	CommandDir     string
	PostCommand    string
	PostCommandDir string
}

type ExecutionResult struct {
	Name        string
	RunLogPath  *string
	PostLogPath *string
}

var now = time.Now().Format("06-01-02.15:04") // like %y-%m-%d.%H:%M

func ExecuteCommand(info CommandInfo) ExecutionResult {
	runLogPath := fmt.Sprintf("logs/%s.run.%s.log", info.Name, now)
	postLogPath := fmt.Sprintf("logs/%s.post.%s.log", info.Name, now)

	var runLogFile *os.File
	var postLogFile *os.File
	var err error

	if info.Command != "" {
		runLogFile, err = os.Create(runLogPath)
		if err != nil {
			// fmt.Fprintf(os.Stderr, "Failed to create run log : %v\n", err)
			return ExecutionResult{Name: info.Name}
		}

		defer runLogFile.Close()

		cmd := exec.Command(
			"bash",
			"-c",
			fmt.Sprintf("cd %s && %s", info.CommandDir, info.Command))
		cmd.Stdout = runLogFile
		cmd.Stderr = runLogFile
		cmd.Dir = info.CommandDir // also set for safety
		cmd.Env = os.Environ()

		if err := cmd.Run(); err != nil {
			// fmt.Fprintf(os.Stderr, "Execution failed for %s (run): %v\n", info.Name, err)
			runLogPath = ""
		}
	}

	if info.PostCommand != "" {
		postLogFile, err = os.Create(postLogPath)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create post log: %v\n", err)
			return ExecutionResult{Name: info.Name}
		}
		defer postLogFile.Close()

		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", info.PostCommandDir, info.PostCommand))
		cmd.Stdout = postLogFile
		cmd.Stderr = postLogFile
		cmd.Dir = info.PostCommandDir
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			// fmt.Fprintf(os.Stderr, "Execution failed for %s (post): %v\n", info.Name, err)
			postLogPath = ""
		}
	}

	var runPathPtr, postPathPtr *string

	if info.Command != "" && runLogPath != "" {
		runPathPtr = &runLogPath
	}

	if info.PostCommand != "" && postLogPath != "" {
		postPathPtr = &postLogPath
	}

	return ExecutionResult{
		Name:        info.Name,
		RunLogPath:  runPathPtr,
		PostLogPath: postPathPtr,
	}
}
