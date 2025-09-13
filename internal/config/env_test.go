package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# This is a comment
TEST_KEY1=value1
TEST_KEY2="quoted value"
TEST_KEY3='single quoted'
TEST_KEY4=unquoted value

# Another comment
TEST_KEY5=value with spaces`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	// Clear any existing test environment variables
	testKeys := []string{"TEST_KEY1", "TEST_KEY2", "TEST_KEY3", "TEST_KEY4", "TEST_KEY5"}
	for _, key := range testKeys {
		os.Unsetenv(key)
	}

	// Load the env file
	if err := loadEnvFile(envFile); err != nil {
		t.Fatalf("Failed to load env file: %v", err)
	}

	// Verify values were set correctly
	tests := []struct {
		key      string
		expected string
	}{
		{"TEST_KEY1", "value1"},
		{"TEST_KEY2", "quoted value"},
		{"TEST_KEY3", "single quoted"},
		{"TEST_KEY4", "unquoted value"},
		{"TEST_KEY5", "value with spaces"},
	}

	for _, test := range tests {
		if got := os.Getenv(test.key); got != test.expected {
			t.Errorf("Expected %s=%s, got %s", test.key, test.expected, got)
		}
	}

	// Test that existing environment variables are not overwritten
	os.Setenv("TEST_EXISTING", "original")
	envFile2 := filepath.Join(tmpDir, ".env2")
	if err := os.WriteFile(envFile2, []byte("TEST_EXISTING=new_value"), 0644); err != nil {
		t.Fatalf("Failed to create second test env file: %v", err)
	}

	if err := loadEnvFile(envFile2); err != nil {
		t.Fatalf("Failed to load second env file: %v", err)
	}

	if got := os.Getenv("TEST_EXISTING"); got != "original" {
		t.Errorf("Expected TEST_EXISTING to remain 'original', got '%s'", got)
	}

	// Clean up
	for _, key := range testKeys {
		os.Unsetenv(key)
	}
	os.Unsetenv("TEST_EXISTING")
}

func TestLoadEnvFileErrors(t *testing.T) {
	// Test non-existent file
	if err := loadEnvFile("non_existent_file.env"); err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test invalid format
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(envFile, []byte("INVALID_LINE_NO_EQUALS"), 0644); err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	if err := loadEnvFile(envFile); err == nil {
		t.Error("Expected error for invalid line format")
	}
}
