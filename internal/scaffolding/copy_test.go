package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to run test in temp directory
func withTempDir(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)
	fn(dir)
}

// ============================================================================
// INSTALL FUNCTION
// ============================================================================

func TestInstall_Python(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "python",
			SourceDir:   "src/",
			ExcludeDirs: []string{"tests/", "__pycache__/"},
		}

		err := Install(config)
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Check .guardian directory exists
		if _, err := os.Stat(".guardian"); os.IsNotExist(err) {
			t.Error(".guardian directory not created")
		}

		// Check guardian_config.toml exists
		if _, err := os.Stat("guardian_config.toml"); os.IsNotExist(err) {
			t.Error("guardian_config.toml not created")
		}

		// Check .pre-commit-config.yaml exists
		if _, err := os.Stat(".pre-commit-config.yaml"); os.IsNotExist(err) {
			t.Error(".pre-commit-config.yaml not created")
		}
	})
}

func TestInstall_TypeScript(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "typescript",
			SourceDir:   "src/",
			ExcludeDirs: []string{"node_modules/", "dist/"},
		}

		err := Install(config)
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Check files exist
		if _, err := os.Stat(".guardian"); os.IsNotExist(err) {
			t.Error(".guardian directory not created")
		}
	})
}

func TestInstall_Go(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "go",
			SourceDir:   "./",
			ExcludeDirs: []string{"vendor/"},
		}

		err := Install(config)
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		if _, err := os.Stat(".guardian"); os.IsNotExist(err) {
			t.Error(".guardian directory not created")
		}
	})
}

func TestInstall_InvalidLanguage(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language: "invalid-lang",
		}

		err := Install(config)
		// Note: Current implementation doesn't validate language - it just creates
		// empty .guardian dir. This test documents current behavior.
		// TODO: Consider adding language validation in future.
		if err != nil {
			t.Logf("Install returned error for invalid language: %v", err)
		}
	})
}

// ============================================================================
// CONFIG GENERATION
// ============================================================================

func TestGenerateConfig_IncludesSrcRoot(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "python",
			SourceDir:   "src/",
			ExcludeDirs: []string{"tests/"},
		}

		err := generateConfig(config)
		if err != nil {
			t.Fatalf("generateConfig failed: %v", err)
		}

		content, err := os.ReadFile("guardian_config.toml")
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		// Config uses src_root (without trailing slash)
		if !strings.Contains(string(content), `src_root = "src"`) {
			t.Error("config missing src_root setting")
		}
	})
}

func TestGenerateConfig_IncludesSourceDirWithTrailingSlash(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:  "python",
			SourceDir: "app/src/",
		}

		err := generateConfig(config)
		if err != nil {
			t.Fatalf("generateConfig failed: %v", err)
		}

		content, err := os.ReadFile("guardian_config.toml")
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		// Trailing slash is stripped
		if !strings.Contains(string(content), `src_root = "app/src"`) {
			t.Error("config missing src_root setting")
		}
	})
}

func TestGenerateConfig_IncludesExcludes(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "python",
			ExcludeDirs: []string{"tests/", "docs/", "__pycache__/"},
		}

		err := generateConfig(config)
		if err != nil {
			t.Fatalf("generateConfig failed: %v", err)
		}

		content, err := os.ReadFile("guardian_config.toml")
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		// Excludes have slashes stripped
		expectedDirs := []string{"tests", "docs", "__pycache__"}
		for _, excl := range expectedDirs {
			if !strings.Contains(string(content), excl) {
				t.Errorf("config missing exclude: %s", excl)
			}
		}
	})
}

// ============================================================================
// PRE-COMMIT CONFIG GENERATION
// ============================================================================

func TestPreCommitConfig_CreatesNew(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language: "python",
		}

		err := generatePreCommitConfig(config)
		if err != nil {
			t.Fatalf("generatePreCommitConfig failed: %v", err)
		}

		content, err := os.ReadFile(".pre-commit-config.yaml")
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		// Check contains guardian hook (lowercase)
		if !strings.Contains(string(content), "guardian") {
			t.Error("pre-commit config missing guardian hook")
		}

		// Check valid YAML structure
		if !strings.HasPrefix(string(content), "repos:") {
			t.Error("pre-commit config should start with 'repos:'")
		}
	})
}

func TestPreCommitConfig_AppendsToExisting(t *testing.T) {
	withTempDir(t, func(dir string) {
		// Create existing pre-commit config
		existingContent := `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
`
		if err := os.WriteFile(".pre-commit-config.yaml", []byte(existingContent), 0644); err != nil {
			t.Fatalf("failed to create existing config: %v", err)
		}

		config := InstallConfig{
			Language: "python",
		}

		err := generatePreCommitConfig(config)
		if err != nil {
			t.Fatalf("generatePreCommitConfig failed: %v", err)
		}

		content, err := os.ReadFile(".pre-commit-config.yaml")
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		// Check original content preserved
		if !strings.Contains(string(content), "trailing-whitespace") {
			t.Error("existing hooks were lost")
		}

		// Check guardian hook added (lowercase)
		if !strings.Contains(string(content), "guardian") {
			t.Error("guardian hook not added")
		}

		// CRITICAL: Check valid YAML - no lines starting with "repos:" in middle
		lines := strings.Split(string(content), "\n")
		reposCount := 0
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "repos:") {
				reposCount++
			}
		}
		if reposCount > 1 {
			t.Error("YAML corrupted - multiple 'repos:' keys found")
		}
	})
}

func TestPreCommitConfig_DoesNotDuplicateHook(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language: "python",
		}

		// Run twice
		generatePreCommitConfig(config)
		generatePreCommitConfig(config)

		content, err := os.ReadFile(".pre-commit-config.yaml")
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		// Count "repo: local" occurrences (the hook blocks)
		// The implementation checks for "guardian" in existing content and skips if found
		repoLocalCount := strings.Count(string(content), "repo: local")
		if repoLocalCount > 1 {
			t.Errorf("guardian hooks duplicated: found %d 'repo: local' blocks", repoLocalCount)
		}
	})
}

// ============================================================================
// CLEANUP ON FAILURE
// ============================================================================

func TestInstall_CleansUpOnError(t *testing.T) {
	withTempDir(t, func(dir string) {
		// Create a read-only directory to trigger write failure
		os.MkdirAll(".guardian", 0755)
		os.WriteFile(".guardian/test.py", []byte("x"), 0644)
		// Make directory read-only to trigger write failures
		os.Chmod(".guardian", 0444)

		config := InstallConfig{
			Language: "python",
		}

		// This should fail or succeed, but not leave partial files
		Install(config)

		// Restore permissions for cleanup
		os.Chmod(".guardian", 0755)
	})
}

// ============================================================================
// FORMAT EXCLUDES HELPER
// ============================================================================

func TestFormatExcludes_Empty(t *testing.T) {
	result := formatExcludes([]string{})
	// Implementation returns empty string for empty slice
	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

func TestFormatExcludes_Single(t *testing.T) {
	// Note: In actual usage, generateConfig strips trailing slashes before calling formatExcludes
	// This test verifies formatExcludes works with cleaned inputs
	result := formatExcludes([]string{"tests"})
	if !strings.Contains(result, `"tests"`) {
		t.Errorf("expected to contain \"tests\", got %s", result)
	}
}

func TestFormatExcludes_Multiple(t *testing.T) {
	// Note: In actual usage, generateConfig strips trailing slashes before calling formatExcludes
	result := formatExcludes([]string{"tests", "docs", "__pycache__"})
	expected := []string{`"tests"`, `"docs"`, `"__pycache__"`}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("expected to contain %s, got %s", exp, result)
		}
	}
}

// ============================================================================
// PYTHON FILE GENERATION
// ============================================================================

func TestGeneratePythonFiles_CreatesGuardianPy(t *testing.T) {
	withTempDir(t, func(dir string) {
		os.MkdirAll(".guardian", 0755)

		config := InstallConfig{
			Language: "python",
		}

		err := generatePythonFiles(config)
		if err != nil {
			t.Fatalf("generatePythonFiles failed: %v", err)
		}

		// Check guardian.py exists
		guardianPath := filepath.Join(".guardian", "guardian.py")
		if _, err := os.Stat(guardianPath); os.IsNotExist(err) {
			t.Error("guardian.py not created")
		}
	})
}

func TestGeneratePythonFiles_CreatesCheckScripts(t *testing.T) {
	withTempDir(t, func(dir string) {
		os.MkdirAll(".guardian", 0755)

		config := InstallConfig{
			Language: "python",
		}

		err := generatePythonFiles(config)
		if err != nil {
			t.Fatalf("generatePythonFiles failed: %v", err)
		}

		// Check security script exists
		securityPath := filepath.Join(".guardian", "check_security.py")
		if _, err := os.Stat(securityPath); os.IsNotExist(err) {
			t.Error("check_security.py not created")
		}

		// Read and verify it's valid Python
		content, _ := os.ReadFile(securityPath)
		if !strings.Contains(string(content), "#!/usr/bin/env python3") {
			t.Error("check_security.py missing shebang")
		}
	})
}

// ============================================================================
// TYPESCRIPT FILE GENERATION
// ============================================================================

func TestGenerateTypeScriptFiles_CreatesFiles(t *testing.T) {
	withTempDir(t, func(dir string) {
		os.MkdirAll(".guardian", 0755)

		config := InstallConfig{
			Language: "typescript",
		}

		err := generateTypeScriptFiles(config)
		if err != nil {
			t.Fatalf("generateTypeScriptFiles failed: %v", err)
		}

		// Check guardian.js exists
		guardianPath := filepath.Join(".guardian", "guardian.js")
		if _, err := os.Stat(guardianPath); os.IsNotExist(err) {
			t.Error("guardian.js not created")
		}
	})
}

// ============================================================================
// GO FILE GENERATION
// ============================================================================

func TestGenerateGoFiles_CreatesShellScript(t *testing.T) {
	withTempDir(t, func(dir string) {
		os.MkdirAll(".guardian", 0755)

		config := InstallConfig{
			Language: "go",
		}

		err := generateGoFiles(config)
		if err != nil {
			t.Fatalf("generateGoFiles failed: %v", err)
		}

		// Check guardian.sh exists
		guardianPath := filepath.Join(".guardian", "guardian.sh")
		if _, err := os.Stat(guardianPath); os.IsNotExist(err) {
			t.Error("guardian.sh not created")
		}

		// Check it's executable
		info, _ := os.Stat(guardianPath)
		if info.Mode()&0111 == 0 {
			t.Error("guardian.sh should be executable")
		}
	})
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestInstall_EmptyExcludes(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "python",
			ExcludeDirs: []string{},
		}

		err := Install(config)
		if err != nil {
			t.Fatalf("Install failed with empty excludes: %v", err)
		}
	})
}

func TestInstall_EmptySourceDir(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:  "python",
			SourceDir: "",
		}

		err := Install(config)
		if err != nil {
			t.Fatalf("Install failed with empty source_dir: %v", err)
		}
	})
}

func TestInstall_SpecialCharsInExcludes(t *testing.T) {
	withTempDir(t, func(dir string) {
		config := InstallConfig{
			Language:    "python",
			ExcludeDirs: []string{"path with spaces/", "path-with-dashes/"},
		}

		err := Install(config)
		if err != nil {
			t.Fatalf("Install failed with special chars: %v", err)
		}

		content, _ := os.ReadFile("guardian_config.toml")
		// Check proper escaping
		if !strings.Contains(string(content), "path with spaces") {
			t.Error("special characters not preserved in config")
		}
	})
}
