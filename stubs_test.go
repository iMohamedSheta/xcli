package xcli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestStubPublishAndOverride(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "xcli_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current working directory to restore it later
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Change working directory to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change working directory to temp dir: %v", err)
	}

	x := New()

	// 1. Verify fallback to embedded stubs when local stubs don't exist
	embeddedContent, err := stubsFS.ReadFile("stubs/model.stub")
	if err != nil {
		t.Fatalf("failed to read embedded model.stub: %v", err)
	}

	readContent, err := x.readStub("stubs/model.stub")
	if err != nil {
		t.Fatalf("readStub failed: %v", err)
	}

	if !bytes.Equal(embeddedContent, readContent) {
		t.Error("readStub did not return embedded stub content on fallback")
	}

	// 2. Publish stubs using the stub:publish command
	cmd := x.stubPublishCommand()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("stub:publish execution failed: %v", err)
	}

	// Check if local stubs folder is created and files are written
	localStubPath := filepath.Join("stubs", "model.stub")
	if _, err := os.Stat(localStubPath); os.IsNotExist(err) {
		t.Error("local stub file was not created by stub:publish")
	}

	// Verify published content matches embedded content
	publishedContent, err := os.ReadFile(localStubPath)
	if err != nil {
		t.Fatalf("failed to read published stub: %v", err)
	}
	if !bytes.Equal(embeddedContent, publishedContent) {
		t.Error("published stub content does not match embedded stub content")
	}

	// 3. Modify local stub and verify it is read instead of embedded fallback
	customContent := []byte("custom model template content")
	if err := os.WriteFile(localStubPath, customContent, 0644); err != nil {
		t.Fatalf("failed to write custom stub content: %v", err)
	}

	overriddenContent, err := x.readStub("stubs/model.stub")
	if err != nil {
		t.Fatalf("readStub failed after customization: %v", err)
	}

	if !bytes.Equal(customContent, overriddenContent) {
		t.Errorf("readStub did not return customized stub. Expected %q, got %q", string(customContent), string(overriddenContent))
	}
}

func TestIsCliOnly(t *testing.T) {
	cliOnly := []string{
		"", "help", "version", "stub:publish", "dev", "lint", "scan",
		"build:app", "build:release", "build:cli",
		"make:model", "make:vue", "make:lang", "make:handler",
		"make:action", "make:repo", "make:req", "make:notification",
		"make:task", "make:crud", "make:migration",
	}

	for _, cmd := range cliOnly {
		if !IsCliOnly(cmd) {
			t.Errorf("Expected IsCliOnly(%q) to be true, got false", cmd)
		}
	}

	needsEnv := []string{
		"migrate", "migrate:rollback", "migrate:status", "migrate:reset",
		"migrate:refresh", "migrate:next", "migrate:back", "goose",
		"unknown-command",
	}

	for _, cmd := range needsEnv {
		if IsCliOnly(cmd) {
			t.Errorf("Expected IsCliOnly(%q) to be false, got true", cmd)
		}
	}
}
