package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDeploySuccessSequenceAndLogs(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "project")
	binDir := filepath.Join(tmp, "bin")
	cmdLog := filepath.Join(tmp, "commands.log")

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}

	createStubTool(t, binDir, "git", fmt.Sprintf("echo \"git:$*\" >> %q\necho 'Already up to date.'", cmdLog))
	createStubTool(t, binDir, "mvn", fmt.Sprintf("echo \"mvn:$*\" >> %q\nprintf '\\033[1;34mINFO\\033[0m building\\n'", cmdLog))
	createStubTool(t, binDir, "docker", fmt.Sprintf("echo \"docker:$*\" >> %q\nprintf '\\033[32mdocker ok\\033[0m\\n'", cmdLog))

	svc := NewDeploymentService(Config{
		GitPath:    filepath.Join(binDir, "git"),
		MavenPath:  filepath.Join(binDir, "mvn"),
		DockerPath: filepath.Join(binDir, "docker"),
	})

	var got []string
	var mu sync.Mutex
	err := svc.Deploy(context.Background(), projectDir, func(line string) error {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, line)
		return nil
	})
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	assertContainsLine(t, got, "[step] git pull")
	assertContainsLine(t, got, "[step] mvn clean package")
	assertContainsLine(t, got, "[step] docker build")
	assertContainsLine(t, got, "[step] docker run")
	assertContainsLine(t, got, "[deploy] completed successfully")

	foundANSI := false
	for _, line := range got {
		if strings.Contains(line, "\u001b") {
			foundANSI = true
			break
		}
	}
	if !foundANSI {
		t.Fatalf("expected ANSI output line in logs, got: %v", got)
	}

	logBytes, err := os.ReadFile(cmdLog)
	if err != nil {
		t.Fatalf("read command log: %v", err)
	}
	cmdLines := strings.Split(strings.TrimSpace(string(logBytes)), "\n")
	if len(cmdLines) != 4 {
		t.Fatalf("expected 4 command invocations, got %d (%v)", len(cmdLines), cmdLines)
	}

	if !strings.Contains(cmdLines[0], "git:-C "+projectDir+" pull") {
		t.Fatalf("unexpected git invocation: %s", cmdLines[0])
	}
	if !strings.Contains(cmdLines[1], "mvn:clean package -Dstyle.color=always") {
		t.Fatalf("unexpected maven invocation: %s", cmdLines[1])
	}
	if !strings.Contains(cmdLines[2], "docker:build -t project:latest .") {
		t.Fatalf("unexpected docker build invocation: %s", cmdLines[2])
	}
	if !strings.Contains(cmdLines[3], "docker:run --rm -d --name project-runner project:latest") {
		t.Fatalf("unexpected docker run invocation: %s", cmdLines[3])
	}
}

func TestDeployBlocksConcurrentSameProject(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "project")
	binDir := filepath.Join(tmp, "bin")
	marker := filepath.Join(tmp, "mvn.started")

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}

	createStubTool(t, binDir, "git", "echo ok")
	createStubTool(t, binDir, "mvn", fmt.Sprintf("touch %q\nsleep 1\necho ok", marker))
	createStubTool(t, binDir, "docker", "echo ok")

	svc := NewDeploymentService(Config{
		GitPath:    filepath.Join(binDir, "git"),
		MavenPath:  filepath.Join(binDir, "mvn"),
		DockerPath: filepath.Join(binDir, "docker"),
	})

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- svc.Deploy(context.Background(), projectDir, func(string) error { return nil })
	}()

	if err := waitForFileOrDeployResult(marker, firstDone, 5*time.Second); err != nil {
		t.Fatalf("first deployment did not reach mvn step: %v", err)
	}

	err := svc.Deploy(context.Background(), projectDir, func(string) error { return nil })
	if !errors.Is(err, ErrDeploymentInProgress) {
		t.Fatalf("expected ErrDeploymentInProgress, got %v", err)
	}

	if err := <-firstDone; err != nil {
		t.Fatalf("first deployment failed: %v", err)
	}
}

func TestDeployNilSender(t *testing.T) {
	t.Parallel()

	svc := NewDeploymentService(Config{})
	err := svc.Deploy(context.Background(), ".", nil)
	if err == nil || !strings.Contains(err.Error(), "send callback") {
		t.Fatalf("expected send callback error, got %v", err)
	}
}

func TestIsBenignPipeClose(t *testing.T) {
	t.Parallel()

	if !isBenignPipeClose(os.ErrClosed) {
		t.Fatalf("expected os.ErrClosed to be benign")
	}
	if !isBenignPipeClose(errors.New("read |0: file already closed")) {
		t.Fatalf("expected file already closed to be benign")
	}
	if isBenignPipeClose(errors.New("other scanner error")) {
		t.Fatalf("unexpected benign classification")
	}
}

func createStubTool(t *testing.T, dir, name, body string) {
	t.Helper()

	path := filepath.Join(dir, name)
	script := fmt.Sprintf("#!/bin/sh\nset -e\n")
	script += body + "\n"

	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", name, err)
	}
}

func waitForFileOrDeployResult(path string, done <-chan error, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("deploy exited before marker: %w", err)
			}
			return fmt.Errorf("deploy exited before marker without error")
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("file not found before timeout: %s", path)
}

func assertContainsLine(t *testing.T, lines []string, want string) {
	t.Helper()
	for _, line := range lines {
		if strings.Contains(line, want) {
			return
		}
	}
	t.Fatalf("expected line containing %q, got %v", want, lines)
}
