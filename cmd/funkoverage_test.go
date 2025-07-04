package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- isELF tests ---

func TestIsELF(t *testing.T) {
	tmp := t.TempDir()

	// Create a fake ELF file
	elfFile := filepath.Join(tmp, "elf")
	if err := os.WriteFile(elfFile, []byte("\x7fELFfoobar"), 0644); err != nil {
		t.Fatal(err)
	}
	if !isELF(elfFile) {
		t.Errorf("isELF should return true for ELF magic")
	}

	// Create a shell script
	shFile := filepath.Join(tmp, "script.sh")
	if err := os.WriteFile(shFile, []byte("#!/bin/bash\necho hi\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if isELF(shFile) {
		t.Errorf("isELF should return false for shell script")
	}

	// Create an empty file
	emptyFile := filepath.Join(tmp, "empty")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if isELF(emptyFile) {
		t.Errorf("isELF should return false for empty file")
	}
}

// --- findPinTool tests ---

func TestFindPinTool(t *testing.T) {
	tmp := t.TempDir()
	// Should not find anything
	_, err := findPinTool(tmp)
	if err == nil {
		t.Error("findPinTool should fail if FuncTracer.so is not present")
	}
	// Create a dummy FuncTracer.so
	subdir := filepath.Join(tmp, "sub")
	os.Mkdir(subdir, 0755)
	target := filepath.Join(subdir, "FuncTracer.so")
	if err := os.WriteFile(target, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}
	found, err := findPinTool(tmp)
	if err != nil {
		t.Fatalf("findPinTool failed: %v", err)
	}
	if found != target {
		t.Errorf("findPinTool returned wrong path: got %s, want %s", found, target)
	}
}

// --- analyzeLogs tests ---

func TestAnalyzeLogs(t *testing.T) {
	tmp := t.TempDir()
	logFile := filepath.Join(tmp, "log.txt")
	content := `[Image:prog] [Function:foo]
[Image:prog] [Function:bar]
[Image:prog] [Called:foo]
[Image:prog] [Function:baz]
`
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	coverage, err := analyzeLogs([]string{logFile})
	if err != nil {
		t.Fatal(err)
	}
	data, ok := coverage["prog"]
	if !ok {
		t.Fatal("prog not found in coverage")
	}
	if len(data.TotalFunctions) != 3 {
		t.Errorf("expected 3 total functions, got %d", len(data.TotalFunctions))
	}
	if len(data.CalledFunctions) != 1 {
		t.Errorf("expected 1 called function, got %d", len(data.CalledFunctions))
	}
	if _, ok := data.CalledFunctions["foo"]; !ok {
		t.Error("foo should be in called functions")
	}
	if _, ok := data.TotalFunctions["baz"]; !ok {
		t.Error("baz should be in total functions")
	}
}

// --- wrap/unwrap logic (integration) ---

func TestWrapUnwrapLogic(t *testing.T) {
	tmp := t.TempDir()
	orig := filepath.Join(tmp, "origbin")
	// Write a fake ELF binary
	if err := os.WriteFile(orig, []byte("\x7fELFfoobar"), 0755); err != nil {
		t.Fatal(err)
	}
	// Set up dummy environment
	os.Setenv("PIN_ROOT", "/tmp/pin")
	os.Setenv("PIN_TOOL_SEARCH_DIR", tmp)
	os.Setenv("SAFE_BIN_DIR", tmp)
	os.Setenv("LOG_DIR", tmp)
	// Create dummy FuncTracer.so
	funcTracer := filepath.Join(tmp, "FuncTracer.so")
	if err := os.WriteFile(funcTracer, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}
	// Wrap
	if err := wrap(orig); err != nil {
		t.Fatalf("wrap failed: %v", err)
	}
	// The wrapper should now exist and be a shell script
	content, err := os.ReadFile(orig)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), wrapperIDComment) {
		t.Error("wrapper script missing ID comment")
	}
	// Unwrap
	if err := unwrap(orig); err != nil {
		t.Fatalf("unwrap failed: %v", err)
	}
	// The original ELF should be restored
	content, err = os.ReadFile(orig)
	if err != nil {
		t.Fatal(err)
	}
	if !isELF(orig) {
		t.Error("unwrap did not restore ELF binary")
	}
}
