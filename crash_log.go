package main

import "os"

const (
	crashLogPath    = "crash_log"
	crashLogBackup  = "crash_log.1"
	maxCrashLogSize = 1 << 20
)

func openCrashLog(perm os.FileMode) (*os.File, error) {
	if info, err := os.Stat(crashLogPath); err == nil && info.Size() > maxCrashLogSize {
		_ = os.Remove(crashLogBackup)
		if err := os.Rename(crashLogPath, crashLogBackup); err != nil {
			_ = os.Truncate(crashLogPath, 0)
		}
	}
	return os.OpenFile(crashLogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
}
