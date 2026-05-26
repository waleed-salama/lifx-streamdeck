package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenCrashLogRotatesLargeFile(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	defer func() {
		require.NoError(t, os.Chdir(wd))
	}()

	data := make([]byte, maxCrashLogSize+1)
	require.NoError(t, os.WriteFile(crashLogPath, data, 0600))

	f, err := openCrashLog(0600)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	active, err := os.Stat(crashLogPath)
	require.NoError(t, err)
	require.Equal(t, int64(0), active.Size())

	backup, err := os.Stat(crashLogBackup)
	require.NoError(t, err)
	require.Equal(t, int64(maxCrashLogSize+1), backup.Size())
}
