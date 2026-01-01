package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the guardian configuration
type Config struct {
	Project  ProjectConfig  `toml:"project"`
	Limits   LimitsConfig   `toml:"limits"`
	Quality  QualityConfig  `toml:"quality"`
	Security SecurityConfig `toml:"security"`
}

// ProjectConfig holds project settings
type ProjectConfig struct {
	SrcRoot     string   `toml:"src_root"`
	ExcludeDirs []string `toml:"exclude_dirs"`
}

// LimitsConfig holds size limits
type LimitsConfig struct {
	MaxFileLines       int            `toml:"max_file_lines"`
	MaxFunctionLines   int            `toml:"max_function_lines"`
	CustomFileLimits   map[string]int `toml:"custom_file_limits"`
}

// QualityConfig holds quality rules
type QualityConfig struct {
	BanPrint           bool     `toml:"ban_print"`
	BanBareExcept      bool     `toml:"ban_bare_except"`
	BanMutableDefaults bool     `toml:"ban_mutable_defaults"`
	BanStarImports     bool     `toml:"ban_star_imports"`
	BanTodoMarkers     bool     `toml:"ban_todo_markers"`
	BanMockData        bool     `toml:"ban_mock_data"`
	MockPatterns       []string `toml:"mock_patterns"`
}

// SecurityConfig holds security rules
type SecurityConfig struct {
	BanEvalExec          bool     `toml:"ban_eval_exec"`
	BanSubprocessShell   bool     `toml:"ban_subprocess_shell"`
	BanDangerousCommands bool     `toml:"ban_dangerous_commands"`
	DangerousPatterns    []string `toml:"dangerous_patterns"`
	SecretPatterns       []string `toml:"secret_patterns"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			SrcRoot:     "src",
			ExcludeDirs: []string{"tests", "__pycache__", "node_modules", ".venv", "venv"},
		},
		Limits: LimitsConfig{
			MaxFileLines:     500,
			MaxFunctionLines: 50,
			CustomFileLimits: make(map[string]int),
		},
		Quality: QualityConfig{
			BanPrint:           true,
			BanBareExcept:      true,
			BanMutableDefaults: true,
			BanStarImports:     true,
			BanTodoMarkers:     true,
			BanMockData:        true,
			MockPatterns: []string{
				"mock_", "_mock", "fake_", "_fake", "dummy_", "_dummy",
				"test_user", "test_email", "test_password",
				"example@", "@example.com", "@test.com",
				"placeholder", "sample_", "hardcoded",
				"changeme", "replace_me", "your_", "xxx",
				"lorem ipsum", "foo_bar", "asdf",
			},
		},
		Security: SecurityConfig{
			BanEvalExec:          true,
			BanSubprocessShell:   true,
			BanDangerousCommands: true,
			DangerousPatterns: []string{
				"rm -rf",
				"DROP TABLE",
				"DROP DATABASE",
				"DELETE FROM",
				"TRUNCATE TABLE",
			},
			SecretPatterns: []string{
				"api_key", "apikey", "api-key",
				"secret", "password", "passwd",
				"private_key", "privatekey",
				"access_token", "auth_token",
			},
		},
	}
}

// Load loads configuration from guardian_config.toml
func Load(dir string) (*Config, error) {
	configPath := filepath.Join(dir, "guardian_config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	config := DefaultConfig()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save saves configuration to guardian_config.toml
func Save(dir string, config *Config) error {
	configPath := filepath.Join(dir, "guardian_config.toml")

	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetConfigPath returns the path to the config file
func GetConfigPath(dir string) string {
	return filepath.Join(dir, "guardian_config.toml")
}

// Exists checks if a config file exists
func Exists(dir string) bool {
	_, err := os.Stat(GetConfigPath(dir))
	return err == nil
}
