//go:build integration

package main_test // Changed from gydnc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	// Consider adding YAML parsing library, e.g., "gopkg.in/yaml.v3"
	// Consider adding JSON comparison library, e.g., "github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	baseTestDir       = "cmd_samples"
	sharedFixturesDir = "shared_fixtures"
	gydncTestBinary   = "./test_gydnc" // Relative to temp test execution dir
)

type CLITestCase struct {
	Name        string
	Path        string
	ArrangeFile string
	ActScript   string
	AssertFile  string
}

// Definitions for parsing assert.yml
type AssertionSpec struct {
	ExitCode   *int               `yaml:"exit_code"`
	Stdout     []StreamAssertion  `yaml:"stdout,omitempty"`
	Stderr     []StreamAssertion  `yaml:"stderr,omitempty"`
	Filesystem []FilesystemAssert `yaml:"filesystem,omitempty"`
}

type StreamAssertion struct {
	MatchType string `yaml:"match_type"` // EXACT, SUBSTRING, REGEX, CONTAINS_LINES, JSON_EQUALS, JSON_CONTAINS_SUBSET, etc.
	Content   string `yaml:"content"`
}

type FilesystemAssert struct {
	Path      string `yaml:"path"`
	Exists    *bool  `yaml:"exists"` // Pointer to check if explicitly set
	IsDir     bool   `yaml:"is_dir,omitempty"`
	MatchType string `yaml:"match_type,omitempty"` // For file content: EXACT, SUBSTRING, REGEX, YAML_EQUALS, JSON_EQUALS
	Content   string `yaml:"content,omitempty"`
}

type ArrangeStep struct {
	Action      string   `yaml:"action"`
	Path        string   `yaml:"path,omitempty"`
	Source      string   `yaml:"source,omitempty"`      // For copy_fixture, relative to sharedFixturesDir
	Destination string   `yaml:"destination,omitempty"` // For copy_fixture, relative to tempDir
	Content     string   `yaml:"content,omitempty"`
	Command     string   `yaml:"command,omitempty"`
	Args        []string `yaml:"args,omitempty"`
}

// buildGydncOnce builds the gydnc binary for testing.
// It includes a simple mechanism to ensure it's built only once per test suite run.
var ( // package-level variable to track build status
	gydncBuildErr error
	gydncBuilt    bool
)

func buildGydncOnce(t *testing.T) string {
	t.Helper()
	if gydncBuilt {
		// If already built, assume the caller will copy it.
		// Return the expected path *relative to the project root*.
		return "./gydnc"
	}

	// Determine project root (one level up from the current file's dir)
	// This assumes the test file is in a direct subdirectory of the project root (e.g., ./tests)
	projectRoot := ".."

	cmd := exec.Command("make", "build")
	cmd.Dir = projectRoot // Set the working directory for make
	output, err := cmd.CombinedOutput()
	if err != nil {
		gydncBuildErr = fmt.Errorf("failed to build gydnc for testing (cwd: %s): %v\nOutput:\n%s", projectRoot, err, string(output))
		t.Fatalf("%s", gydncBuildErr.Error())
	}
	gydncBuilt = true
	t.Log("gydnc binary built successfully for testing.")
	// The path returned should be relative to where the test binary expects it when copying.
	// If copyFile expects a path relative to project root, then filepath.Join(projectRoot, "gydnc") is correct.
	// The current copyFile in the harness uses the returned path directly, assuming it's accessible.
	// Let's return an absolute path or a path clearly relative to project root for clarity.
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		t.Fatalf("Failed to get absolute path for project root %s: %v", projectRoot, err)
	}
	return filepath.Join(absProjectRoot, "gydnc") // Path for the built binary
}

func TestCLI(t *testing.T) {
	// Build the binary once for all tests.
	// The path returned is relative to the project root.
	gydncBinaryPath := buildGydncOnce(t)

	testCases, err := discoverTestCases(baseTestDir)
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	// Check if tests were discovered
	if len(testCases) == 0 {
		t.Logf("No test cases found in %s", baseTestDir)
		return
	} else {
		t.Logf("Discovered %d test cases", len(testCases))
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel() // Run test cases in parallel

			tempDir := t.TempDir()
			t.Logf("Test %s running in tempDir: %s", tc.Name, tempDir)

			// 1. ARRANGE Phase (parse arrange.yml and execute steps)
			if err := arrangeTestData(t, tempDir, tc.ArrangeFile); err != nil {
				t.Fatalf("Arrange phase failed: %v", err)
			}

			// Copy the pre-built binary into the tempDir for this test
			localBinaryPath := filepath.Join(tempDir, "gydnc")
			if err := copyFile(gydncBinaryPath, localBinaryPath); err != nil {
				t.Fatalf("Failed to copy gydnc binary from %s to %s: %v", gydncBinaryPath, localBinaryPath, err)
			}
			if err := os.Chmod(localBinaryPath, 0755); err != nil {
				t.Fatalf("Failed to make copied gydnc binary executable: %v", err)
			}
			t.Logf("Copied test binary to %s", localBinaryPath)

			// 2. ACT Phase (run act.sh)
			actualStdout, actualStderr, actualExitCode, scriptErr := runActScript(t, tempDir, tc.ActScript)
			if scriptErr != nil && actualExitCode == -1 {
				t.Fatalf("Act phase script execution harness failed: %v", scriptErr)
			}
			if scriptErr != nil {
				t.Logf("Act script finished with non-zero exit code (%d). Error: %v", actualExitCode, scriptErr)
			}

			// Log output regardless of exit code
			t.Logf("Act script stdout:\n%s", actualStdout)
			t.Logf("Act script stderr:\n%s", actualStderr)
			t.Logf("Act script exitCode: %d", actualExitCode)

			// 3. ASSERT Phase
			if err := assertResults(t, tempDir, tc.AssertFile, actualStdout, actualStderr, actualExitCode); err != nil {
				t.Errorf("Assert phase failed: %v", err) // Use Errorf to allow other tests to run
			}
		})
	}
}

func discoverTestCases(baseDir string) ([]CLITestCase, error) {
	var cases []CLITestCase

	// Walk the baseDir to find test case directories.
	// A test case directory is expected to be a grandchild of baseDir,
	// e.g., baseDir/suiteName/testCaseName/
	// and must contain act.sh and assert.yml.

	suiteEntries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return cases, nil // No suites found, no tests
		}
		return nil, fmt.Errorf("reading base test directory %s: %w", baseDir, err)
	}

	for _, suiteEntry := range suiteEntries {
		if !suiteEntry.IsDir() {
			continue
		}
		suiteName := suiteEntry.Name()
		suitePath := filepath.Join(baseDir, suiteName)

		testCaseEntries, err := os.ReadDir(suitePath)
		if err != nil {
			// Log warning but continue, one suite being unreadable shouldn't stop others
			fmt.Fprintf(os.Stderr, "Warning: could not read test suite directory %s: %v\n", suitePath, err)
			continue
		}

		for _, testCaseEntry := range testCaseEntries {
			if !testCaseEntry.IsDir() {
				continue
			}
			testCaseName := testCaseEntry.Name()
			fullTestCaseName := filepath.Join(suiteName, testCaseName) // e.g., "core/01_list_fails_no_config"
			testCasePath := filepath.Join(suitePath, testCaseName)

			actScriptPath := filepath.Join(testCasePath, "act.sh")
			assertFilePath := filepath.Join(testCasePath, "assert.yml")

			if _, err := os.Stat(actScriptPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: Skipping test case '%s': missing act.sh in %s\n", fullTestCaseName, testCasePath)
				continue
			}
			if _, err := os.Stat(assertFilePath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: Skipping test case '%s': missing assert.yml in %s\n", fullTestCaseName, testCasePath)
				continue
			}

			cases = append(cases, CLITestCase{
				Name:        fullTestCaseName, // Use combined name for t.Run
				Path:        testCasePath,
				ArrangeFile: filepath.Join(testCasePath, "arrange.yml"),
				ActScript:   actScriptPath,
				AssertFile:  assertFilePath,
			})
		}
	}
	return cases, nil
}

func arrangeTestData(t *testing.T, tempDir, arrangeFile string) error {
	t.Helper()
	if _, err := os.Stat(arrangeFile); os.IsNotExist(err) {
		t.Log("arrange.yml not found, skipping arrange phase.")
		return nil // No arrange file is fine
	}

	yamlData, err := os.ReadFile(arrangeFile)
	if err != nil {
		return fmt.Errorf("reading arrange file %s: %w", arrangeFile, err)
	}

	var spec struct { // Local struct for parsing arrange steps
		Steps []ArrangeStep `yaml:"steps"`
	}
	if err := yaml.Unmarshal(yamlData, &spec); err != nil {
		return fmt.Errorf("parsing arrange YAML %s: %w", arrangeFile, err)
	}

	t.Logf("Executing %d arrange steps from %s", len(spec.Steps), arrangeFile)
	for i, step := range spec.Steps {
		t.Logf("Arrange step %d: action=%s, path=%s, source=%s, dest=%s", i+1, step.Action, step.Path, step.Source, step.Destination)
		switch step.Action {
		case "create_dir":
			if step.Path == "" {
				return fmt.Errorf("arrange step %d: create_dir missing 'path'", i+1)
			}
			dirPath := filepath.Join(tempDir, step.Path)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("arrange step %d (create_dir %s): %w", i+1, step.Path, err)
			}
		case "create_file":
			if step.Path == "" {
				return fmt.Errorf("arrange step %d: create_file missing 'path'", i+1)
			}
			filePath := filepath.Join(tempDir, step.Path)
			// Ensure parent dir exists
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return fmt.Errorf("arrange step %d (create_file %s): creating parent dir: %w", i+1, step.Path, err)
			}
			if err := os.WriteFile(filePath, []byte(step.Content), 0644); err != nil {
				return fmt.Errorf("arrange step %d (create_file %s): %w", i+1, step.Path, err)
			}
		case "copy_fixture":
			if step.Source == "" || step.Destination == "" {
				return fmt.Errorf("arrange step %d: copy_fixture missing 'source' or 'destination'", i+1)
			}
			// Path to shared_fixtures is relative to project root, so construct it carefully
			// Assuming test runs from project root for `go test ./...`
			// projectRootSharedFixtures := sharedFixturesDir
			sharedFixturePath := filepath.Join(sharedFixturesDir, step.Source) // Corrected: sharedFixturesDir is now relative to test file
			destinationPath := filepath.Join(tempDir, step.Destination)

			if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
				return fmt.Errorf("arrange step %d (copy_fixture %s -> %s): creating parent dir: %w", i+1, step.Source, step.Destination, err)
			}

			// Check if source is dir or file
			srcInfo, err := os.Stat(sharedFixturePath)
			if err != nil {
				return fmt.Errorf("arrange step %d (copy_fixture): accessing source %s: %w", i+1, sharedFixturePath, err)
			}
			if srcInfo.IsDir() {
				// TODO: Implement directory copying if needed
				return fmt.Errorf("arrange step %d: copy_fixture for directory source '%s' not implemented", i+1, step.Source)
			} else {
				if err := copyFile(sharedFixturePath, destinationPath); err != nil {
					return fmt.Errorf("arrange step %d (copy_fixture %s -> %s): %w", i+1, step.Source, step.Destination, err)
				}
			}
		case "run_command":
			if step.Command == "" {
				return fmt.Errorf("arrange step %d: run_command missing 'command'", i+1)
			}
			cmd := exec.Command(step.Command, step.Args...)
			cmd.Dir = tempDir // Run command within the temp directory
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("arrange step %d (run_command %s): %w\nOutput:\n%s", i+1, step.Command, err, string(output))
			}
			t.Logf("Arrange step %d command output:\n%s", i+1, string(output))
		default:
			return fmt.Errorf("arrange step %d: unknown action '%s'", i+1, step.Action)
		}
	}
	return nil
}

func runActScript(t *testing.T, tempDir, actScriptPath string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	if _, err := os.Stat(actScriptPath); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("act script not found: %s", actScriptPath)
	}

	// The actScriptPath is the source. It should be copied into tempDir to run.
	localActScript := filepath.Join(tempDir, "act.sh")
	if err := copyFile(actScriptPath, localActScript); err != nil {
		return "", "", -1, fmt.Errorf("copying act.sh to tempDir: %w", err)
	}
	if err := os.Chmod(localActScript, 0755); err != nil {
		return "", "", -1, fmt.Errorf("chmod act.sh in tempDir: %w", err)
	}

	// The main gydnc binary is already copied into tempDir by TestCLI function.
	// The act.sh script should use ./gydnc to refer to it.

	cmd := exec.Command("./act.sh") // act.sh should call ./gydnc
	cmd.Dir = tempDir               // Execute from within the temp directory

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	t.Logf("Executing act script: %s from %s", localActScript, tempDir)
	execErr := cmd.Run()

	stdout = outBuf.String()
	stderr = errBuf.String()

	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			return stdout, stderr, exitError.ExitCode(), execErr // Pass original error for logging
		}
		return stdout, stderr, -1, fmt.Errorf("running act.sh failed: %w", execErr)
	}
	return stdout, stderr, 0, nil
}

func assertResults(t *testing.T, tempDir, assertFile, actualStdout, actualStderr string, actualExitCode int) error {
	t.Helper()
	yamlData, err := os.ReadFile(assertFile)
	if err != nil {
		return fmt.Errorf("reading assert file %s: %w", assertFile, err)
	}

	var spec AssertionSpec
	if err := yaml.Unmarshal(yamlData, &spec); err != nil {
		return fmt.Errorf("parsing assert YAML %s: %w", assertFile, err)
	}
	t.Logf("Loaded assertions from %s", assertFile)

	var errors []string

	if spec.ExitCode != nil {
		if actualExitCode != *spec.ExitCode {
			errors = append(errors, fmt.Sprintf("exit code mismatch: expected %d, got %d", *spec.ExitCode, actualExitCode))
		}
	} else if actualExitCode != 0 { // Default to expecting 0 if not specified
		errors = append(errors, fmt.Sprintf("exit code mismatch: expected 0 (default), got %d", actualExitCode))
	}

	for i, sAssert := range spec.Stdout {
		if err := compareStreamOutput(sAssert.MatchType, sAssert.Content, actualStdout, fmt.Sprintf("stdout[%d]", i)); err != nil {
			errors = append(errors, err.Error())
		}
	}

	for i, sAssert := range spec.Stderr {
		if err := compareStreamOutput(sAssert.MatchType, sAssert.Content, actualStderr, fmt.Sprintf("stderr[%d]", i)); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(spec.Filesystem) > 0 {
		if err := compareFileSystem(t, tempDir, spec.Filesystem); err != nil {
			errors = append(errors, fmt.Sprintf("filesystem state mismatch: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("assertion(s) failed:\n- %s", strings.Join(errors, "\n- "))
	}

	t.Log("All assertions passed.")
	return nil
}

func compareStreamOutput(matchType, expectedContent, actualOutput, streamName string) error {
	if matchType == "" {
		matchType = "EXACT"
	}

	// For EXACT, SUBSTRING, REGEX, trimming is fine.
	// For JSON and YAML, we generally want to compare the raw string.
	// However, if the harness always provides trimmed actualOutput,
	// and expectedContent in YAML is also effectively trimmed by the parser,
	// this might be okay. Let's be mindful.
	// For now, keep trim for non-JSON/YAML, and use raw for JSON/YAML.

	switch strings.ToUpper(matchType) {
	case "EXACT":
		normActual := strings.TrimSpace(actualOutput)
		normExpected := strings.TrimSpace(expectedContent)
		if normActual != normExpected {
			return fmt.Errorf("%s exact match failed.\nExpected:\n```\n%s\n```\nGot:\n```\n%s\n```", streamName, normExpected, normActual)
		}
	case "SUBSTRING":
		// Substring doesn't usually need trimming of expected content.
		// actualOutput might be trimmed if it makes sense generally.
		if !strings.Contains(actualOutput, expectedContent) {
			return fmt.Errorf("%s substring match failed. Expected to find:\n```\n%s\n```\nIn output:\n```\n%s\n```", streamName, expectedContent, actualOutput)
		}
	case "REGEX":
		// Regex operates on the raw string.
		matched, err := regexp.MatchString(expectedContent, actualOutput)
		if err != nil {
			return fmt.Errorf("invalid regex in %s assertion: %w", streamName, err)
		}
		if !matched {
			return fmt.Errorf("%s regex match failed. Pattern:\n```\n%s\n```\nOutput:\n```\n%s\n```", streamName, expectedContent, actualOutput)
		}
	case "JSON":
		var expectedJSON, actualJSON interface{}

		// Unmarshal expected JSON
		if err := json.Unmarshal([]byte(expectedContent), &expectedJSON); err != nil {
			return fmt.Errorf("%s: failed to unmarshal expected JSON content: %w\nExpected JSON string:\n```\n%s\n```", streamName, err, expectedContent)
		}

		// Unmarshal actual JSON output
		if err := json.Unmarshal([]byte(actualOutput), &actualJSON); err != nil {
			// Try to trim whitespace and retry for actual output, as it might have leading/trailing newlines from CLI output
			if errRetry := json.Unmarshal([]byte(strings.TrimSpace(actualOutput)), &actualJSON); errRetry != nil {
				return fmt.Errorf("%s: failed to unmarshal actual output as JSON: %w (original error: %s)\nActual output string:\n```\n%s\n```", streamName, errRetry, err.Error(), actualOutput)
			}
		}

		if !reflect.DeepEqual(expectedJSON, actualJSON) {
			// For better diffs, marshal them back to string (pretty printed)
			prettyExpected, _ := json.MarshalIndent(expectedJSON, "", "  ")
			prettyActual, _ := json.MarshalIndent(actualJSON, "", "  ")
			return fmt.Errorf("%s JSON content mismatch.\nExpected:\n```json\n%s\n```\nGot:\n```json\n%s\n```\n(Raw Expected:\n%s\nRaw Actual:\n%s)", streamName, string(prettyExpected), string(prettyActual), expectedContent, actualOutput)
		}
	case "YAML":
		var expectedYAML, actualYAML interface{}

		// Unmarshal expected YAML
		if err := yaml.Unmarshal([]byte(expectedContent), &expectedYAML); err != nil {
			return fmt.Errorf("%s: failed to unmarshal expected YAML content: %w\nExpected YAML string:\n```\n%s\n```", streamName, err, expectedContent)
		}

		// Unmarshal actual YAML output
		if err := yaml.Unmarshal([]byte(actualOutput), &actualYAML); err != nil {
			// Try to trim whitespace and retry for actual output
			if errRetry := yaml.Unmarshal([]byte(strings.TrimSpace(actualOutput)), &actualYAML); errRetry != nil {
				return fmt.Errorf("%s: failed to unmarshal actual output as YAML: %w (original error: %s)\nActual output string:\n```\n%s\n```", streamName, errRetry, err.Error(), actualOutput)
			}
		}

		if !reflect.DeepEqual(expectedYAML, actualYAML) {
			// For better diffs, marshal them back to string (pretty printed if possible, though yaml.Marshal is standard)
			prettyExpected, _ := yaml.Marshal(expectedYAML)
			prettyActual, _ := yaml.Marshal(actualYAML)
			return fmt.Errorf("%s YAML content mismatch.\nExpected:\n```yaml\n%s\n```\nGot:\n```yaml\n%s\n```\n(Raw Expected:\n%s\nRaw Actual:\n%s)", streamName, string(prettyExpected), string(prettyActual), expectedContent, actualOutput)
		}
	default:
		return fmt.Errorf("unknown match_type '%s' for %s assertion. Supported: EXACT, SUBSTRING, REGEX, JSON, YAML", matchType, streamName)
	}
	return nil
}

func compareFileSystem(t *testing.T, tempDirRoot string, asserts []FilesystemAssert) error {
	t.Helper()
	var fsErrors []string
	for _, assert := range asserts {
		targetPath := filepath.Join(tempDirRoot, assert.Path)
		stat, err := os.Stat(targetPath)
		shouldExist := true
		if assert.Exists != nil {
			shouldExist = *assert.Exists
		}
		if err != nil {
			if os.IsNotExist(err) {
				if shouldExist {
					fsErrors = append(fsErrors, fmt.Sprintf("path '%s': expected to exist, but does not", assert.Path))
				}
			} else {
				fsErrors = append(fsErrors, fmt.Sprintf("path '%s': error checking existence: %v", assert.Path, err))
			}
			continue
		}
		if !shouldExist {
			fsErrors = append(fsErrors, fmt.Sprintf("path '%s': expected not to exist, but it does", assert.Path))
			continue
		}
		if assert.IsDir {
			if !stat.IsDir() {
				fsErrors = append(fsErrors, fmt.Sprintf("path '%s': expected to be a directory, but is not", assert.Path))
			}
			continue
		} else if stat.IsDir() {
			fsErrors = append(fsErrors, fmt.Sprintf("path '%s': expected to be a file, but is a directory", assert.Path))
			continue
		}
		if assert.Content != "" || assert.MatchType != "" {
			if stat.IsDir() {
				fsErrors = append(fsErrors, fmt.Sprintf("path '%s': cannot check content/match_type on a directory", assert.Path))
				continue
			}
			actualContentBytes, err := os.ReadFile(targetPath)
			if err != nil {
				fsErrors = append(fsErrors, fmt.Sprintf("path '%s': failed to read actual file content: %v", assert.Path, err))
				continue
			}
			actualContent := string(actualContentBytes)
			matchType := assert.MatchType
			if matchType == "" {
				matchType = "EXACT"
			}
			if err := compareStreamOutput(matchType, assert.Content, actualContent, fmt.Sprintf("file content (%s)", assert.Path)); err != nil {
				fsErrors = append(fsErrors, err.Error())
			}
		}
	}
	if len(fsErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(fsErrors, "\n"))
	}
	return nil
}

// copyFile utility
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("making dir for %s: %w", dst, err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading src %s: %w", src, err)
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src %s: %w", src, err)
	}
	err = os.WriteFile(dst, data, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("writing dst %s: %w", dst, err)
	}
	return nil
}

// TODO: Need a robust way to ensure the 'gydnc' binary used by act.sh is the one
//       built for the test, and it's correctly placed/named in the tempDir for act.sh to find.
//       The buildGydncOnce tries to address the build, but placement in tempDir is also key.
