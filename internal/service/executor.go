package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var ErrDeploymentInProgress = errors.New("deployment already in progress for this project")

// Config defines executable paths for external tools.
type Config struct {
	GitPath    string
	MavenPath  string
	DockerPath string
}

type logSender func(line string) error

type DeploymentService struct {
	cfg    Config
	mu     sync.Mutex
	active map[string]context.CancelFunc
}

type commandStep struct {
	name string
	bin  string
	args []string
	dir  string
	env  []string
}

func NewDeploymentService(cfg Config) *DeploymentService {
	return &DeploymentService{
		cfg:    cfg,
		active: make(map[string]context.CancelFunc),
	}
}

func (s *DeploymentService) Deploy(ctx context.Context, projectPath string, send logSender) error {
	if send == nil {
		return errors.New("send callback cannot be nil")
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	if err := s.registerActive(absPath, cancel); err != nil {
		return err
	}
	defer func() {
		cancel()
		s.unregisterActive(absPath)
	}()

	imageTag := sanitizeName(filepath.Base(absPath)) + ":latest"
	containerName := sanitizeName(filepath.Base(absPath)) + "-runner"

	steps := []commandStep{
		{
			name: "git pull",
			bin:  s.cfg.GitPath,
			args: []string{"-C", absPath, "pull"},
		},
		{
			name: "mvn clean package",
			bin:  s.cfg.MavenPath,
			args: []string{"clean", "package", "-Dstyle.color=always"},
			dir:  absPath,
			env:  []string{"TERM=xterm-256color", "MAVEN_OPTS=-Djansi.force=true"},
		},
		{
			name: "docker build",
			bin:  s.cfg.DockerPath,
			args: []string{"build", "-t", imageTag, "."},
			dir:  absPath,
			env:  []string{"TERM=xterm-256color", "FORCE_COLOR=1"},
		},
		{
			name: "docker run",
			bin:  s.cfg.DockerPath,
			args: []string{"run", "--rm", "-d", "--name", containerName, imageTag},
			env:  []string{"TERM=xterm-256color", "FORCE_COLOR=1"},
		},
	}

	if err := send("[deploy] start: " + absPath); err != nil {
		return err
	}

	for _, step := range steps {
		if err := runCtx.Err(); err != nil {
			return err
		}

		if err := send(fmt.Sprintf("[step] %s", step.name)); err != nil {
			return err
		}

		if err := s.runCommand(runCtx, step, send); err != nil {
			msg := fmt.Sprintf("[error] %s failed: %v", step.name, err)
			_ = send(msg)
			return err
		}
	}

	if err := send("[deploy] completed successfully"); err != nil {
		return err
	}

	return nil
}

func (s *DeploymentService) registerActive(projectPath string, cancel context.CancelFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.active[projectPath]; exists {
		return ErrDeploymentInProgress
	}

	s.active[projectPath] = cancel
	return nil
}

func (s *DeploymentService) unregisterActive(projectPath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.active, projectPath)
}

func (s *DeploymentService) runCommand(ctx context.Context, step commandStep, send logSender) error {
	cmd := exec.CommandContext(ctx, step.bin, step.args...)
	if step.dir != "" {
		cmd.Dir = step.dir
	}
	if len(step.env) > 0 {
		cmd.Env = append(cmd.Environ(), step.env...)
	}

	// Create a process group so all child processes can be terminated on cancel.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command %q: %w", step.name, err)
	}

	childDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
		case <-childDone:
		}
	}()

	var (
		wg      sync.WaitGroup
		scanErr error
		once    sync.Once
	)

	scanPipe := func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if err := send(line); err != nil {
				once.Do(func() {
					scanErr = err
				})
				if cmd.Process != nil {
					_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			if isBenignPipeClose(err) {
				return
			}
			once.Do(func() {
				scanErr = err
			})
		}
	}

	wg.Add(2)
	go scanPipe(stdout)
	go scanPipe(stderr)

	waitErr := cmd.Wait()
	close(childDone)
	wg.Wait()

	if scanErr != nil {
		return fmt.Errorf("stream output: %w", scanErr)
	}

	if waitErr != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("command exited with error: %w", waitErr)
	}

	return nil
}

func sanitizeName(input string) string {
	lower := strings.ToLower(input)
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ".", "-")
	safe := replacer.Replace(lower)
	safe = strings.Trim(safe, "-")
	if safe == "" {
		return "app"
	}
	return safe
}

func isBenignPipeClose(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrClosed) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "file already closed") || strings.Contains(msg, "use of closed file")
}
