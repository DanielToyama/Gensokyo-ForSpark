package sys

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

var (
	singleProcessMu   sync.Mutex
	singleProcessFile *os.File
	singleProcessPath string
)

// singleProcessLockPath 用 exe 绝对路径生成锁文件名，不同安装目录互不影响。
func singleProcessLockPath() string {
	exe, err := os.Executable()
	if err != nil {
		exe = "gensokyo"
	}
	abs, err := filepath.Abs(exe)
	if err == nil {
		exe = abs
	}
	sum := sha1.Sum([]byte(strings.ToLower(exe)))
	name := ".gensokyo-" + hex.EncodeToString(sum[:6]) + ".lock"
	// 优先放在 exe 同目录；无写权限时退回系统临时目录。
	if dir := filepath.Dir(exe); dir != "" {
		return filepath.Join(dir, name)
	}
	return filepath.Join(os.TempDir(), name)
}

// processAlive 判断锁文件里的 PID 是否仍存在。
func processAlive(pid int32) bool {
	if pid <= 0 {
		return false
	}
	ok, err := process.PidExists(pid)
	return err == nil && ok
}

// AcquireSingleProcessLock 获取单实例锁；已有实例时返回 false。
// 使用 O_EXCL 创建锁文件，崩溃后留下死 PID 会自动清理。
func AcquireSingleProcessLock() bool {
	singleProcessMu.Lock()
	defer singleProcessMu.Unlock()

	if singleProcessFile != nil {
		return true
	}

	lockPath := singleProcessLockPath()
	for attempt := 0; attempt < 10; attempt++ {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
			singleProcessFile = file
			singleProcessPath = lockPath
			return true
		}
		if !os.IsExist(err) {
			// 锁文件不可写时不阻止启动，避免只读目录导致程序完全无法运行。
			return true
		}

		data, readErr := os.ReadFile(lockPath)
		if readErr == nil {
			text := strings.TrimSpace(string(data))
			if text == "" {
				// 另一个实例正在创建锁文件，稍后重试，不要误删。
				time.Sleep(100 * time.Millisecond)
				continue
			}
			pid, _ := strconv.Atoi(text)
			if processAlive(int32(pid)) {
				return false
			}
		}
		// 死进程/空文件/被手工残留：清理后重试。
		_ = os.Remove(lockPath)
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// ReleaseSingleProcessLock 释放锁文件，供面板模式重启时让子进程接管。
func ReleaseSingleProcessLock() {
	singleProcessMu.Lock()
	defer singleProcessMu.Unlock()

	if singleProcessFile != nil {
		_ = singleProcessFile.Close()
		singleProcessFile = nil
	}
	if singleProcessPath != "" {
		_ = os.Remove(singleProcessPath)
		singleProcessPath = ""
	}
}
