//go:build windows
// +build windows

package sys

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// WindowsRestarter implements the Restarter interface for Windows systems.
type WindowsRestarter struct{}

// NewRestarter creates a new Restarter appropriate for Windows systems.
func NewRestarter() *WindowsRestarter {
	return &WindowsRestarter{}
}

func (r *WindowsRestarter) Restart(executablePath string) error {
	// 面板模式：直接在同一个控制台/管道内启动子进程，不再用 `start` 新开窗口。
	// 第三方面板只捕获父进程 stdout 时，旧实现会让日志“消失”在新窗口里。
	if PanelMode() {
		ReleaseSingleProcessLock()
		exePath, err := os.Executable()
		if err != nil {
			exePath = executablePath
		}
		cmd := exec.Command(exePath, "-faststart")
		cmd.Dir = filepath.Dir(exePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		if err := cmd.Start(); err != nil {
			return err
		}
		os.Exit(0)
		return nil
	}

	// Separate the directory and the executable name
	execDir, execName := filepath.Split(executablePath)

	// Including -faststart parameter in the script that starts the executable
	scriptContent := "@echo off\n" +
		"pushd " + strconv.Quote(execDir) + "\n" +
		// Add the -faststart parameter here
		"start \"\" " + strconv.Quote(execName) + " -faststart\n" +
		"popd\n"

	scriptName := "restart.bat"
	if err := os.WriteFile(scriptName, []byte(scriptContent), 0755); err != nil {
		return err
	}

	cmd := exec.Command("cmd.exe", "/C", scriptName)

	if err := cmd.Start(); err != nil {
		return err
	}

	// The current process can now exit
	os.Exit(0)

	// This return statement will never be reached
	return nil
}
