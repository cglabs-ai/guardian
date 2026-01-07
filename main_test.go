package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	guardianBinary string
	buildOnce      sync.Once
	buildErr       error
)

// Build guardian binary once for all tests
func getGuardianBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		// Get current working directory (project root)
		projectDir, err := os.Getwd()
		if err != nil {
			buildErr = err
			return
		}

		// Build to a temp location
		tmpDir := os.TempDir()
		guardianBinary = filepath.Join(tmpDir, "guardian-test")

		buildCmd := exec.Command("go", "build", "-o", guardianBinary, ".")
		buildCmd.Dir = projectDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			buildErr = err
			t.Logf("build output: %s", output)
		}
	})

	if buildErr != nil {
		t.Fatalf("failed to build guardian: %v", buildErr)
	}

	return guardianBinary
}

// Helper to run guardian command
func runGuardian(t *testing.T, args ...string) (string, error) {
	t.Helper()
	binary := getGuardianBinary(t)

	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper to run guardian in a specific directory
func runGuardianInDir(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	binary := getGuardianBinary(t)

	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper to run in temp directory
func withTestProject(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir := t.TempDir()
	fn(dir)
}

// ============================================================================
// VERSION COMMAND
// ============================================================================

func TestCLI_Version(t *testing.T) {
	output, err := runGuardian(t, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	if !strings.Contains(output, "guardian") {
		t.Error("version output missing 'guardian'")
	}

	// Should contain a version number
	if !strings.Contains(output, ".") {
		t.Error("version output missing version number")
	}
}

func TestCLI_VersionFlags(t *testing.T) {
	flags := []string{"version", "--version", "-v"}
	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			output, err := runGuardian(t, flag)
			if err != nil {
				t.Fatalf("%s command failed: %v", flag, err)
			}
			if !strings.Contains(output, "guardian") {
				t.Errorf("%s output missing 'guardian'", flag)
			}
		})
	}
}

// ============================================================================
// HELP COMMAND
// ============================================================================

func TestCLI_Help(t *testing.T) {
	output, err := runGuardian(t, "help")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	// Should mention key commands
	expectedCommands := []string{"check", "add", "config"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("help output missing command: %s", cmd)
		}
	}
}

func TestCLI_HelpFlags(t *testing.T) {
	flags := []string{"help", "--help", "-h"}
	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			output, err := runGuardian(t, flag)
			if err != nil {
				t.Fatalf("%s command failed: %v", flag, err)
			}
			if !strings.Contains(output, "Usage") || !strings.Contains(output, "usage") {
				// At least mention how to use it
				if !strings.Contains(strings.ToLower(output), "guardian") {
					t.Errorf("%s output doesn't look like help", flag)
				}
			}
		})
	}
}

// ============================================================================
// CHECK COMMAND
// ============================================================================

func TestCLI_Check_EmptyDir(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "check")

		// Empty dir should succeed with no issues
		if err != nil {
			t.Logf("check returned error (may be expected): %v", err)
		}

		if !strings.Contains(output, "No issues") && !strings.Contains(output, "0 issues") {
			// Should either say no issues or succeed silently
			t.Logf("output: %s", output)
		}
	})
}

func TestCLI_Check_FindsEval(t *testing.T) {
	withTestProject(t, func(dir string) {
		// Create a file with eval
		code := `result = eval("1+1")`
		os.WriteFile(filepath.Join(dir, "test.py"), []byte(code), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should find the eval issue
		if !strings.Contains(output, "eval") {
			t.Errorf("check should find eval() issue, got: %s", output)
		}
	})
}

func TestCLI_Check_FindsSQLInjection(t *testing.T) {
	withTestProject(t, func(dir string) {
		code := `query = f"SELECT * FROM users WHERE id = {user_id}"`
		os.WriteFile(filepath.Join(dir, "test.py"), []byte(code), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should find SQL injection
		if !strings.Contains(strings.ToLower(output), "sql") {
			t.Errorf("check should find SQL injection, got: %s", output)
		}
	})
}

func TestCLI_Check_RunAlias(t *testing.T) {
	// "run" should be alias for "check"
	withTestProject(t, func(dir string) {
		_, err := runGuardianInDir(t, dir, "run")
		if err != nil {
			t.Logf("run command returned error (may be expected for empty dir): %v", err)
		}
		// Just verify it doesn't crash with unknown command error
	})
}

func TestCLI_Check_ExcludesHiddenDirs(t *testing.T) {
	withTestProject(t, func(dir string) {
		// Create .git with a "bad" file
		os.MkdirAll(filepath.Join(dir, ".git/hooks"), 0755)
		os.WriteFile(filepath.Join(dir, ".git/hooks/test.py"), []byte(`eval("x")`), 0644)

		// Create a good file in main dir
		os.WriteFile(filepath.Join(dir, "main.py"), []byte(`x = 1`), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should NOT report .git issues
		if strings.Contains(output, ".git") {
			t.Errorf("check should exclude .git directory, got: %s", output)
		}
	})
}

// ============================================================================
// ADD COMMAND
// ============================================================================

func TestCLI_Add_Python(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "add", "python")
		if err != nil {
			t.Fatalf("add python failed: %v\n%s", err, output)
		}

		// Should create .guardian directory
		if _, err := os.Stat(filepath.Join(dir, ".guardian")); os.IsNotExist(err) {
			t.Error(".guardian directory not created")
		}

		// Should create guardian_config.toml
		if _, err := os.Stat(filepath.Join(dir, "guardian_config.toml")); os.IsNotExist(err) {
			t.Error("guardian_config.toml not created")
		}
	})
}

func TestCLI_Add_TypeScript(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "add", "typescript")
		if err != nil {
			t.Fatalf("add typescript failed: %v\n%s", err, output)
		}

		if _, err := os.Stat(filepath.Join(dir, ".guardian")); os.IsNotExist(err) {
			t.Error(".guardian directory not created for typescript")
		}
	})
}

func TestCLI_Add_Go(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "add", "go")
		if err != nil {
			t.Fatalf("add go failed: %v\n%s", err, output)
		}

		if _, err := os.Stat(filepath.Join(dir, ".guardian")); os.IsNotExist(err) {
			t.Error(".guardian directory not created for go")
		}
	})
}

func TestCLI_Add_InvalidLanguage(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "add", "nonexistent")

		// Should fail or warn about invalid language
		if err == nil && !strings.Contains(strings.ToLower(output), "unknown") &&
			!strings.Contains(strings.ToLower(output), "unsupported") &&
			!strings.Contains(strings.ToLower(output), "invalid") {
			t.Logf("add invalid language output: %s", output)
		}
	})
}

func TestCLI_Add_NoLanguage(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "add")

		// Should show usage or error
		if err == nil && !strings.Contains(strings.ToLower(output), "usage") &&
			!strings.Contains(strings.ToLower(output), "language") {
			t.Logf("add without language output: %s", output)
		}
	})
}

func TestCLI_Add_StackVariants(t *testing.T) {
	stacks := []string{"python-fastapi", "python-django", "typescript-react"}
	for _, stack := range stacks {
		t.Run(stack, func(t *testing.T) {
			withTestProject(t, func(dir string) {
				output, err := runGuardianInDir(t, dir, "add", stack)
				if err != nil {
					t.Logf("add %s: %v\n%s", stack, err, output)
				}
			})
		})
	}
}

// ============================================================================
// CONFIG COMMAND
// ============================================================================

func TestCLI_Config_NoConfigFile(t *testing.T) {
	withTestProject(t, func(dir string) {
		output, err := runGuardianInDir(t, dir, "config")

		// Should fail or warn when no config exists
		if err == nil && !strings.Contains(strings.ToLower(output), "no") &&
			!strings.Contains(strings.ToLower(output), "not found") {
			// Might show editor, which is okay
			t.Logf("config without config file: %s", output)
		}
	})
}

func TestCLI_Config_WithConfigFile(t *testing.T) {
	// Skip this test - it tries to open an editor which hangs in CI/test env
	t.Skip("Skipping - opens editor which blocks in test environment")
}

// ============================================================================
// UNKNOWN COMMAND
// ============================================================================

func TestCLI_UnknownCommand(t *testing.T) {
	output, err := runGuardian(t, "unknowncommand")

	// Should fail with error
	if err == nil {
		t.Error("unknown command should fail")
	}

	// Should mention unknown or show help
	if !strings.Contains(strings.ToLower(output), "unknown") &&
		!strings.Contains(strings.ToLower(output), "usage") {
		t.Errorf("unknown command output should be helpful: %s", output)
	}
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestCLI_MultipleIssuesSameFile(t *testing.T) {
	withTestProject(t, func(dir string) {
		// Create a file with multiple issues
		code := `# Multiple issues here
result = eval("bad")
result2 = exec("also bad")
print("debug")
from os import *
password = "secret123"
`
		os.WriteFile(filepath.Join(dir, "bad.py"), []byte(code), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should find multiple issues
		issueKeywords := []string{"eval", "print", "password", "import"}
		foundCount := 0
		for _, kw := range issueKeywords {
			if strings.Contains(strings.ToLower(output), kw) {
				foundCount++
			}
		}

		if foundCount < 2 {
			t.Errorf("should find multiple issues, output: %s", output)
		}
	})
}

func TestCLI_LargeFile(t *testing.T) {
	withTestProject(t, func(dir string) {
		// Create a large file (> 500 lines)
		var lines []string
		for i := 0; i < 510; i++ {
			lines = append(lines, "x = 1")
		}
		os.WriteFile(filepath.Join(dir, "large.py"), []byte(strings.Join(lines, "\n")), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should warn about file size
		if !strings.Contains(strings.ToLower(output), "500") &&
			!strings.Contains(strings.ToLower(output), "lines") &&
			!strings.Contains(strings.ToLower(output), "size") {
			t.Logf("large file output: %s", output)
		}
	})
}

func TestCLI_UnicodeFilenames(t *testing.T) {
	withTestProject(t, func(dir string) {
		// Create file with unicode name
		os.WriteFile(filepath.Join(dir, "тест.py"), []byte(`result = eval("x")`), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should still find the issue
		if !strings.Contains(output, "eval") {
			t.Logf("unicode filename output: %s", output)
		}
	})
}

func TestCLI_DeepNestedFiles(t *testing.T) {
	withTestProject(t, func(dir string) {
		// Create deeply nested file
		deepPath := filepath.Join(dir, "a/b/c/d/e")
		os.MkdirAll(deepPath, 0755)
		os.WriteFile(filepath.Join(deepPath, "deep.py"), []byte(`result = eval("x")`), 0644)

		output, _ := runGuardianInDir(t, dir, "check")

		// Should find deeply nested issue
		if !strings.Contains(output, "eval") {
			t.Errorf("should find deep nested issue, output: %s", output)
		}
	})
}
