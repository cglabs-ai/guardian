package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ScanResults holds the results of an AI project scan
type ScanResults struct {
	Language        string
	Framework       string
	SourceDir       string
	TestDir         string
	MockPatterns    []string
	SecretsFound    []SecretLocation
	Recommendations []string
	Conflicts       []string
}

// SecretLocation identifies where a potential secret was found
type SecretLocation struct {
	File string
	Line int
}

// ValidateKey validates a Gemini API key
func ValidateKey(apiKey string) (bool, error) {
	if apiKey == "" {
		return false, fmt.Errorf("API key is empty")
	}

	// Simple validation - try to make a request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to connect to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return false, fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return true, nil
}

// ScanProject uses Gemini to analyze a project
func ScanProject(apiKey string, dir string) (*ScanResults, error) {
	// First, gather project info locally
	info := gatherProjectInfo(dir)

	// Build prompt for Gemini
	prompt := buildScanPrompt(info)

	// Call Gemini API
	response, err := callGemini(apiKey, prompt)
	if err != nil {
		// Fall back to local analysis
		return localAnalysis(info), nil
	}

	// Parse Gemini response
	results := parseGeminiResponse(response, info)

	return results, nil
}

// ProjectInfo holds locally gathered project information
type ProjectInfo struct {
	Files        []string
	Directories  []string
	HasPyproject bool
	HasPackage   bool
	HasGoMod     bool
	HasComposer  bool
	Requirements []string
	PackageJSON  map[string]interface{}
	SampleCode   map[string]string
}

func gatherProjectInfo(dir string) *ProjectInfo {
	info := &ProjectInfo{
		SampleCode: make(map[string]string),
	}

	// Check for project files
	info.HasPyproject = fileExists(filepath.Join(dir, "pyproject.toml"))
	info.HasPackage = fileExists(filepath.Join(dir, "package.json"))
	info.HasGoMod = fileExists(filepath.Join(dir, "go.mod"))
	info.HasComposer = fileExists(filepath.Join(dir, "composer.json"))

	// Read requirements.txt if exists
	if data, err := os.ReadFile(filepath.Join(dir, "requirements.txt")); err == nil {
		info.Requirements = strings.Split(string(data), "\n")
	}

	// Read package.json if exists
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		json.Unmarshal(data, &info.PackageJSON)
	}

	// Walk directory to find files
	filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)

		// Skip hidden and common exclude dirs
		if fileInfo.IsDir() {
			name := fileInfo.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" ||
				name == "__pycache__" || name == "venv" || name == ".venv" {
				return filepath.SkipDir
			}
			info.Directories = append(info.Directories, relPath)
			return nil
		}

		info.Files = append(info.Files, relPath)

		// Get sample of code files
		ext := filepath.Ext(path)
		if (ext == ".py" || ext == ".js" || ext == ".ts" || ext == ".go") && len(info.SampleCode) < 5 {
			if data, err := os.ReadFile(path); err == nil {
				// Only first 100 lines
				lines := strings.Split(string(data), "\n")
				if len(lines) > 100 {
					lines = lines[:100]
				}
				info.SampleCode[relPath] = strings.Join(lines, "\n")
			}
		}

		return nil
	})

	return info
}

func buildScanPrompt(info *ProjectInfo) string {
	var sb strings.Builder

	sb.WriteString("Analyze this codebase and provide configuration recommendations.\n\n")

	sb.WriteString("Project structure:\n")
	sb.WriteString("Directories: " + strings.Join(info.Directories[:min(20, len(info.Directories))], ", ") + "\n")
	sb.WriteString("Files: " + strings.Join(info.Files[:min(30, len(info.Files))], ", ") + "\n\n")

	if info.HasPyproject {
		sb.WriteString("Has pyproject.toml (Python project)\n")
	}
	if info.HasPackage {
		sb.WriteString("Has package.json (Node/TypeScript project)\n")
		if deps, ok := info.PackageJSON["dependencies"].(map[string]interface{}); ok {
			depNames := make([]string, 0, len(deps))
			for name := range deps {
				depNames = append(depNames, name)
			}
			sb.WriteString("Dependencies: " + strings.Join(depNames, ", ") + "\n")
		}
	}
	if info.HasGoMod {
		sb.WriteString("Has go.mod (Go project)\n")
	}

	sb.WriteString("\nSample code:\n")
	for path, code := range info.SampleCode {
		sb.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", path, code))
	}

	sb.WriteString(`
Please respond in JSON format with:
{
  "language": "python|typescript|go|php",
  "framework": "fastapi|django|react|express|etc",
  "source_dir": "src/",
  "test_dir": "tests/",
  "mock_patterns": ["patterns", "found", "in", "code"],
  "recommendations": ["list", "of", "recommendations"],
  "potential_secrets": [{"file": "path", "line": 123}],
  "conflicts": ["any", "existing", "config", "conflicts"]
}
`)

	return sb.String()
}

func callGemini(apiKey string, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.1,
			"maxOutputTokens": 2048,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Extract text from response
	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	candidate := candidates[0].(map[string]interface{})
	content := candidate["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	text := parts[0].(map[string]interface{})["text"].(string)

	return text, nil
}

func parseGeminiResponse(response string, info *ProjectInfo) *ScanResults {
	results := &ScanResults{}

	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]

		var parsed struct {
			Language         string           `json:"language"`
			Framework        string           `json:"framework"`
			SourceDir        string           `json:"source_dir"`
			TestDir          string           `json:"test_dir"`
			MockPatterns     []string         `json:"mock_patterns"`
			Recommendations  []string         `json:"recommendations"`
			PotentialSecrets []SecretLocation `json:"potential_secrets"`
			Conflicts        []string         `json:"conflicts"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
			results.Language = parsed.Language
			results.Framework = parsed.Framework
			results.SourceDir = parsed.SourceDir
			results.TestDir = parsed.TestDir
			results.MockPatterns = parsed.MockPatterns
			results.Recommendations = parsed.Recommendations
			results.SecretsFound = parsed.PotentialSecrets
			results.Conflicts = parsed.Conflicts
			return results
		}
	}

	// Fall back to local analysis
	return localAnalysis(info)
}

func localAnalysis(info *ProjectInfo) *ScanResults {
	results := &ScanResults{
		MockPatterns: []string{},
		Conflicts:    []string{},
	}

	// Detect language
	if info.HasPyproject || len(filterByExt(info.Files, ".py")) > 0 {
		results.Language = "Python"

		// Check for frameworks
		for _, req := range info.Requirements {
			if strings.Contains(strings.ToLower(req), "fastapi") {
				results.Framework = "FastAPI"
			} else if strings.Contains(strings.ToLower(req), "django") {
				results.Framework = "Django"
			} else if strings.Contains(strings.ToLower(req), "flask") {
				results.Framework = "Flask"
			}
		}
	} else if info.HasPackage || len(filterByExt(info.Files, ".ts")) > 0 || len(filterByExt(info.Files, ".js")) > 0 {
		results.Language = "TypeScript"

		if deps, ok := info.PackageJSON["dependencies"].(map[string]interface{}); ok {
			if _, ok := deps["react"]; ok {
				results.Framework = "React"
			} else if _, ok := deps["express"]; ok {
				results.Framework = "Express"
			} else if _, ok := deps["next"]; ok {
				results.Framework = "Next.js"
			}
		}
	} else if info.HasGoMod {
		results.Language = "Go"
	} else if info.HasComposer {
		results.Language = "PHP"
	}

	// Detect source directory
	for _, dir := range info.Directories {
		if dir == "src" || dir == "app" || dir == "lib" {
			results.SourceDir = dir + "/"
			break
		}
	}
	if results.SourceDir == "" {
		results.SourceDir = "./"
	}

	// Detect test directory
	for _, dir := range info.Directories {
		if dir == "tests" || dir == "test" || dir == "__tests__" {
			results.TestDir = dir + "/"
			break
		}
	}

	// Look for mock patterns in code
	mockPatterns := findMockPatterns(info.SampleCode)
	results.MockPatterns = mockPatterns

	// Look for potential secrets
	results.SecretsFound = findSecrets(info.SampleCode)

	// Generate recommendations
	results.Recommendations = generateRecommendations(results)

	// Check for conflicts
	if fileExists("guardian_config.toml") {
		results.Conflicts = append(results.Conflicts, "guardian_config.toml already exists")
	}
	if fileExists(".guardian") {
		results.Conflicts = append(results.Conflicts, ".guardian directory already exists")
	}

	return results
}

func findMockPatterns(sampleCode map[string]string) []string {
	patterns := make(map[string]bool)

	// Common patterns to look for
	matchers := []*regexp.Regexp{
		regexp.MustCompile(`(dummy_\w+)`),
		regexp.MustCompile(`(stub_\w+)`),
		regexp.MustCompile(`(fake_\w+)`),
		regexp.MustCompile(`(mock_\w+)`),
		regexp.MustCompile(`(test_\w+)`),
		regexp.MustCompile(`(PLACEHOLDER_\w+)`),
	}

	for _, code := range sampleCode {
		for _, matcher := range matchers {
			matches := matcher.FindAllString(code, -1)
			for _, match := range matches {
				patterns[match] = true
			}
		}
	}

	result := make([]string, 0, len(patterns))
	for p := range patterns {
		result = append(result, p)
	}

	return result
}

func findSecrets(sampleCode map[string]string) []SecretLocation {
	var secrets []SecretLocation

	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api_key|apikey)\s*=\s*["'][^"']+["']`),
		regexp.MustCompile(`(?i)(password|passwd)\s*=\s*["'][^"']+["']`),
		regexp.MustCompile(`(?i)(secret|private_key)\s*=\s*["'][^"']+["']`),
		regexp.MustCompile(`(?i)AWS_SECRET`),
	}

	for file, code := range sampleCode {
		lines := strings.Split(code, "\n")
		for i, line := range lines {
			for _, pattern := range secretPatterns {
				if pattern.MatchString(line) {
					secrets = append(secrets, SecretLocation{
						File: file,
						Line: i + 1,
					})
					break
				}
			}
		}
	}

	return secrets
}

func generateRecommendations(results *ScanResults) []string {
	var recs []string

	if results.Language == "Python" && results.Framework == "FastAPI" {
		recs = append(recs, "Enable SQL injection checks (SQLAlchemy integration)")
		recs = append(recs, "Enable async checks (FastAPI is async-first)")
	}

	if len(results.SecretsFound) > 0 {
		recs = append(recs, fmt.Sprintf("Review %d potential hardcoded secrets", len(results.SecretsFound)))
	}

	if len(results.MockPatterns) > 0 {
		recs = append(recs, "Custom mock patterns detected - adding to config")
	}

	return recs
}

// Helper functions
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func filterByExt(files []string, ext string) []string {
	var result []string
	for _, f := range files {
		if strings.HasSuffix(f, ext) {
			result = append(result, f)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
