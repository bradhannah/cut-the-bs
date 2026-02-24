package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadPersistsCustomDataDirectoryFromStableConfigPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	customDataDir := filepath.Join(tmpHome, "sample-sets", "screenshots")
	cfg := DefaultConfig()
	cfg.DataDirectory = customDataDir
	cfg.Backup.RollingBackupCount = 9

	require.NoError(t, cfg.Save())

	defaultDir := filepath.Join(tmpHome, "Library", "Application Support", DefaultAppDirName)
	configPath := filepath.Join(defaultDir, ConfigFileName)
	require.FileExists(t, configPath)
	require.DirExists(t, customDataDir)

	legacyPath := filepath.Join(customDataDir, ConfigFileName)
	_, legacyErr := os.Stat(legacyPath)
	require.ErrorIs(t, legacyErr, os.ErrNotExist)

	loaded, err := Load()
	require.NoError(t, err)
	require.Equal(t, customDataDir, loaded.DataDirectory)
	require.Equal(t, 9, loaded.Backup.RollingBackupCount)
}
