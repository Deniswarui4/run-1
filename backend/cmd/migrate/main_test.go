//go:build !integration

package main_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once into a temp dir.
	dir, err := os.MkdirTemp("", "migrate-test-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	binaryPath = dir + "/migrate"

	out, err := exec.Command("go", "build", "-o", binaryPath, ".").CombinedOutput()
	if err != nil {
		panic("failed to build migrate binary: " + err.Error() + "\n" + string(out))
	}

	code := m.Run()

	os.RemoveAll(dir)
	os.Exit(code)
}

// run executes the migrate binary with the given args and returns exit code + stderr.
func run(args ...string) (exitCode int, stderr string) {
	cmd := exec.Command(binaryPath, args...)
	// Use an unreachable DB so valid subcommands fail at connection, not routing.
	cmd.Env = append(os.Environ(),
		"DB_HOST=127.0.0.1",
		"DB_PORT=19999", // nothing listening here
		"DB_USER=nobody",
		"DB_PASSWORD=nobody",
		"DB_NAME=nobody",
		"DB_SSLMODE=disable",
	)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return exitCode, stderrBuf.String()
}

func TestNoArgs_ExitsOneWithUsage(t *testing.T) {
	code, stderr := run()
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "Usage") {
		t.Fatalf("expected stderr to contain 'Usage', got: %q", stderr)
	}
}

func TestUnknownCommand_foo(t *testing.T) {
	code, stderr := run("foo")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected stderr to contain 'unknown command', got: %q", stderr)
	}
}

func TestUnknownCommand_bar(t *testing.T) {
	code, stderr := run("bar")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected stderr to contain 'unknown command', got: %q", stderr)
	}
}

func TestValidSubcommand_up(t *testing.T) {
	code, stderr := run("up")
	if code != 1 {
		t.Fatalf("expected exit code 1 (DB error), got %d", code)
	}
	if strings.Contains(stderr, "unknown command") {
		t.Fatalf("'up' was not routed correctly; got 'unknown command' in stderr: %q", stderr)
	}
	// Should contain a DB-related error message.
	if !strings.Contains(stderr, "error") && !strings.Contains(stderr, "database") && !strings.Contains(stderr, "connect") && !strings.Contains(stderr, "reach") {
		t.Fatalf("expected a DB-related error in stderr, got: %q", stderr)
	}
}

func TestValidSubcommand_down(t *testing.T) {
	code, stderr := run("down")
	if code != 1 {
		t.Fatalf("expected exit code 1 (DB error), got %d", code)
	}
	if strings.Contains(stderr, "unknown command") {
		t.Fatalf("'down' was not routed correctly; got 'unknown command' in stderr: %q", stderr)
	}
	if !strings.Contains(stderr, "error") && !strings.Contains(stderr, "database") && !strings.Contains(stderr, "connect") && !strings.Contains(stderr, "reach") {
		t.Fatalf("expected a DB-related error in stderr, got: %q", stderr)
	}
}

func TestValidSubcommand_status(t *testing.T) {
	code, stderr := run("status")
	if code != 1 {
		t.Fatalf("expected exit code 1 (DB error), got %d", code)
	}
	if strings.Contains(stderr, "unknown command") {
		t.Fatalf("'status' was not routed correctly; got 'unknown command' in stderr: %q", stderr)
	}
	if !strings.Contains(stderr, "error") && !strings.Contains(stderr, "database") && !strings.Contains(stderr, "connect") && !strings.Contains(stderr, "reach") {
		t.Fatalf("expected a DB-related error in stderr, got: %q", stderr)
	}
}
