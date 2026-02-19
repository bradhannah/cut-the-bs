package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultAppDirName is the directory name inside the user's application
	// support directory.
	DefaultAppDirName = "cut-the-bs"

	// ConfigFileName is the name of the configuration file.
	ConfigFileName = "config.json"

	// DefaultDBName is the default SQLite database filename.
	DefaultDBName = "cut-the-bs.db"

	// DefaultRollingBackupCount is the default number of rolling backup
	// copies to retain.
	DefaultRollingBackupCount = 5
)

// Config holds persistent application configuration.
type Config struct {
	// DataDirectory is the path where the SQLite database and backups
	// are stored. If empty, the default platform-specific application
	// support directory is used.
	DataDirectory string `json:"data_directory"`

	// Backup holds backup-related settings.
	Backup BackupConfig `json:"backup"`
}

// BackupConfig holds backup-related settings.
type BackupConfig struct {
	// RollingBackupCount is the number of rolling backup copies to retain.
	RollingBackupCount int `json:"rolling_backup_count"`
}

// DefaultDataDir returns the default data directory for the current
// platform. On macOS this is ~/Library/Application Support/cut-the-bs/.
func DefaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: unable to determine home directory: %w", err)
	}
	return filepath.Join(home, "Library", "Application Support", DefaultAppDirName), nil
}

// DefaultConfig returns a Config with all default values.
func DefaultConfig() Config {
	return Config{
		Backup: BackupConfig{
			RollingBackupCount: DefaultRollingBackupCount,
		},
	}
}

// ResolveDataDir returns the effective data directory. If the config
// has a custom DataDirectory set, that is used. Otherwise the
// platform default is returned.
func (c *Config) ResolveDataDir() (string, error) {
	if c.DataDirectory != "" {
		return c.DataDirectory, nil
	}
	return DefaultDataDir()
}

// DBPath returns the full path to the SQLite database file.
func (c *Config) DBPath() (string, error) {
	dir, err := c.ResolveDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DefaultDBName), nil
}

// ExportDir returns the full path to the PDF export directory,
// creating it if needed.
func (c *Config) ExportDir() (string, error) {
	dir, err := c.ResolveDataDir()
	if err != nil {
		return "", err
	}
	exportDir := filepath.Join(dir, "exports")
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", fmt.Errorf("config: unable to create export directory: %w", err)
	}
	return exportDir, nil
}

// EnsureDataDir creates the data directory if it does not exist.
func (c *Config) EnsureDataDir() error {
	dir, err := c.ResolveDataDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// configFilePath returns the path to the config JSON file inside the
// data directory.
func (c *Config) configFilePath() (string, error) {
	dir, err := c.ResolveDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// Load reads the configuration from the JSON file in the data
// directory. If the file does not exist, a default config is returned
// without error (first-run scenario).
func Load() (Config, error) {
	cfg := DefaultConfig()

	// First resolve the default data dir to find the config file.
	defaultDir, err := DefaultDataDir()
	if err != nil {
		return cfg, err
	}

	path := filepath.Join(defaultDir, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: unable to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("config: unable to parse config file: %w", err)
	}

	// Apply defaults for any zero values that should have defaults.
	if cfg.Backup.RollingBackupCount <= 0 {
		cfg.Backup.RollingBackupCount = DefaultRollingBackupCount
	}

	return cfg, nil
}

// Save writes the configuration to the JSON file in the data
// directory. It creates the directory if needed.
func (c *Config) Save() error {
	if err := c.EnsureDataDir(); err != nil {
		return fmt.Errorf("config: unable to create data directory: %w", err)
	}

	path, err := c.configFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config: unable to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("config: unable to write config file: %w", err)
	}

	return nil
}

// LoadFrom reads a configuration from a specific file path. Useful
// for testing or loading from a non-default location.
func LoadFrom(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: unable to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("config: unable to parse config file: %w", err)
	}

	if cfg.Backup.RollingBackupCount <= 0 {
		cfg.Backup.RollingBackupCount = DefaultRollingBackupCount
	}

	return cfg, nil
}

// SaveTo writes the configuration to a specific file path.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("config: unable to create directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config: unable to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("config: unable to write config file: %w", err)
	}

	return nil
}
