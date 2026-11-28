package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A Softnet that `brew upgrade` replaced while the controller runs has no
// setuid bit, so an isolated VM can't get a network: serve says so, with the
// command that fixes it, before it touches Tart.
func TestServeChecksSoftnetBeforeEachIsolatedVM(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "softnet"), []byte("#!/bin/sh\necho softnet 0.26.0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	called := filepath.Join(dir, "tart-called")
	tart := filepath.Join(dir, "tart")
	if err := os.WriteFile(tart, []byte("#!/bin/sh\n: >"+called+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	c := &poolController{
		pool: Pool{Network: "isolated", MemoryMB: 1024, CPU: 1},
		host: NewHost(8192, 4, 2),
		tart: &Tart{Bin: tart},
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	err := c.serve(context.Background(), c.log, &runner{name: "eph-test"})
	if err == nil || !strings.Contains(err.Error(), "setuid") || !strings.Contains(err.Error(), "sudo chmod u+s") {
		t.Fatalf("serve = %v, want the setuid error with its fix", err)
	}
	if _, err := os.Stat(called); err == nil {
		t.Error("serve ran tart before checking Softnet")
	}
}
