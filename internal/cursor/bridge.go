package cursor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Bridge struct {
	workDir   string
	timeout   time.Duration
	maxRetry  int
}

type EditRequest struct {
	Prompt     string
	TargetPath string
	Context    string
}

type EditResult struct {
	Success   bool
	Output    string
	Error     error
	Duration  time.Duration
}

func NewBridge(workDir string) *Bridge {
	return &Bridge{
		workDir:  workDir,
		timeout:  5 * time.Minute,
		maxRetry: 2,
	}
}

func (b *Bridge) CheckAvailability() error {
	cmd := b.getCursorCommand()
	if cmd == "" {
		return fmt.Errorf("cursor-agent command not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, "--version")
	if err := c.Run(); err != nil {
		return fmt.Errorf("cursor-agent not responding: %w", err)
	}

	return nil
}

func (b *Bridge) Execute(ctx context.Context, req EditRequest) *EditResult {
	start := time.Now()

	var lastErr error
	for attempt := 0; attempt <= b.maxRetry; attempt++ {
		result := b.executeOnce(ctx, req)
		if result.Success {
			result.Duration = time.Since(start)
			return result
		}
		lastErr = result.Error

		if attempt < b.maxRetry {
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}

	return &EditResult{
		Success:  false,
		Error:    fmt.Errorf("failed after %d attempts: %w", b.maxRetry+1, lastErr),
		Duration: time.Since(start),
	}
}

func (b *Bridge) executeOnce(ctx context.Context, req EditRequest) *EditResult {
	cmd := b.getCursorCommand()
	if cmd == "" {
		return &EditResult{
			Success: false,
			Error:   fmt.Errorf("cursor-agent not found"),
		}
	}

	promptFile, err := b.createPromptFile(req)
	if err != nil {
		return &EditResult{
			Success: false,
			Error:   fmt.Errorf("failed to create prompt file: %w", err),
		}
	}
	defer os.Remove(promptFile)

	args := []string{"--prompt-file", promptFile}
	if req.TargetPath != "" {
		args = append(args, "--target", req.TargetPath)
	}

	execCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	c := exec.CommandContext(execCtx, cmd, args...)
	c.Dir = b.workDir

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()

	if execCtx.Err() == context.DeadlineExceeded {
		return &EditResult{
			Success: false,
			Error:   fmt.Errorf("timeout after %v", b.timeout),
			Output:  stderr.String(),
		}
	}

	if err != nil {
		return &EditResult{
			Success: false,
			Error:   fmt.Errorf("cursor-agent error: %w\nstderr: %s", err, stderr.String()),
			Output:  stdout.String(),
		}
	}

	return &EditResult{
		Success: true,
		Output:  stdout.String(),
	}
}

func (b *Bridge) createPromptFile(req EditRequest) (string, error) {
	content := req.Prompt
	if req.Context != "" {
		content = fmt.Sprintf("Context:\n%s\n\nTask:\n%s", req.Context, req.Prompt)
	}

	tmpDir := os.TempDir()
	filename := filepath.Join(tmpDir, fmt.Sprintf("ppopcode-prompt-%d.txt", time.Now().UnixNano()))

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", err
	}

	return filename, nil
}

func (b *Bridge) getCursorCommand() string {
	if runtime.GOOS == "windows" {
		possiblePaths := []string{
			"cursor-agent",
			"cursor-agent.exe",
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "cursor", "resources", "app", "bin", "cursor-agent.exe"),
		}

		for _, p := range possiblePaths {
			if _, err := exec.LookPath(p); err == nil {
				return p
			}
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	} else {
		possiblePaths := []string{
			"cursor-agent",
			"/usr/local/bin/cursor-agent",
			filepath.Join(os.Getenv("HOME"), ".cursor", "bin", "cursor-agent"),
		}

		for _, p := range possiblePaths {
			if _, err := exec.LookPath(p); err == nil {
				return p
			}
		}
	}

	return ""
}

func (b *Bridge) ExecuteWithScript(ctx context.Context, req EditRequest) *EditResult {
	start := time.Now()

	scriptPath := filepath.Join(b.workDir, ".claude", "skills", "cursor-edit", "scripts", "apply.ps1")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return b.Execute(ctx, req)
	}

	promptFile, err := b.createPromptFile(req)
	if err != nil {
		return &EditResult{
			Success: false,
			Error:   fmt.Errorf("failed to create prompt file: %w", err),
		}
	}
	defer os.Remove(promptFile)

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath, promptFile)
	} else {
		c = exec.CommandContext(ctx, "pwsh", "-File", scriptPath, promptFile)
	}

	c.Dir = b.workDir

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()
	if err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "cursor-agent") {
			return &EditResult{
				Success: false,
				Error:   fmt.Errorf("cursor-agent not available: %s", errMsg),
			}
		}
		return &EditResult{
			Success: false,
			Error:   fmt.Errorf("script error: %w\n%s", err, errMsg),
		}
	}

	return &EditResult{
		Success:  true,
		Output:   stdout.String(),
		Duration: time.Since(start),
	}
}
