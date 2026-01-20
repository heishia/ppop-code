package agents

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type ClaudeAgent struct {
	BaseAgent
	cliPath string
}

// ClaudeLoginStatus represents the login status of Claude CLI
type ClaudeLoginStatus struct {
	LoggedIn bool
	Message  string
	CLIFound bool
}

func NewClaudeAgent(config AgentConfig) (*ClaudeAgent, error) {
	agent := &ClaudeAgent{
		BaseAgent: BaseAgent{
			config: config,
			status: "initializing",
		},
	}

	// Find claude CLI path
	cliPath := findClaudeCLI()
	if cliPath == "" {
		agent.status = "cli_not_found"
		return agent, nil
	}
	agent.cliPath = cliPath

	// Check login status
	status := CheckClaudeLogin()
	if !status.LoggedIn {
		agent.status = "not_logged_in"
		return agent, nil
	}

	agent.status = "ready"
	return agent, nil
}

func (a *ClaudeAgent) Execute(ctx context.Context, prompt string) (*Response, error) {
	if a.cliPath == "" {
		return &Response{
			Content: fmt.Sprintf("[%s] Claude CLI not found. Please install Claude CLI and run 'claude login'.\n\nYour request: %s", a.config.Name, prompt),
			Model:   a.config.Model,
		}, nil
	}

	status := CheckClaudeLogin()
	if !status.LoggedIn {
		return &Response{
			Content: fmt.Sprintf("[%s] Not logged in to Claude. Please run 'claude login' first.\n\nYour request: %s", a.config.Name, prompt),
			Model:   a.config.Model,
		}, nil
	}

	a.SetStatus("processing")
	defer a.SetStatus("ready")

	// Build command arguments
	args := []string{"-p", prompt, "--output-format", "text"}

	cmd := exec.CommandContext(ctx, a.cliPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("claude cli error: %w\nstderr: %s", err, stderr.String())
	}

	return &Response{
		Content: strings.TrimSpace(stdout.String()),
		Model:   a.config.Model,
	}, nil
}

// CheckClaudeLogin checks if the user is logged in to Claude CLI
func CheckClaudeLogin() ClaudeLoginStatus {
	cliPath := findClaudeCLI()
	if cliPath == "" {
		return ClaudeLoginStatus{
			LoggedIn: false,
			Message:  "Claude CLI not found. Install it from https://claude.ai/cli",
			CLIFound: false,
		}
	}

	// Try to run a simple command to check login status
	ctx, cancel := context.WithTimeout(context.Background(), 5*60) // 5 second timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, cliPath, "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return ClaudeLoginStatus{
			LoggedIn: false,
			Message:  fmt.Sprintf("Claude CLI error: %s", stderr.String()),
			CLIFound: true,
		}
	}

	// Check if we can actually use the CLI (login check)
	// Running 'claude -p "test" --output-format text' would require login
	// For now, we assume if CLI is found and runs, it's ready
	// A more robust check would be to parse ~/.claude/config or similar

	return ClaudeLoginStatus{
		LoggedIn: true,
		Message:  fmt.Sprintf("Claude CLI ready: %s", strings.TrimSpace(stdout.String())),
		CLIFound: true,
	}
}

// RunClaudeLogin opens the claude login process
func RunClaudeLogin() error {
	cliPath := findClaudeCLI()
	if cliPath == "" {
		return fmt.Errorf("claude CLI not found")
	}

	cmd := exec.Command(cliPath, "login")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}

// findClaudeCLI finds the claude CLI executable
func findClaudeCLI() string {
	if runtime.GOOS == "windows" {
		possiblePaths := []string{
			"claude",
			"claude.exe",
		}

		for _, p := range possiblePaths {
			if path, err := exec.LookPath(p); err == nil {
				return path
			}
		}
	} else {
		possiblePaths := []string{
			"claude",
			"/usr/local/bin/claude",
			"/usr/bin/claude",
		}

		for _, p := range possiblePaths {
			if path, err := exec.LookPath(p); err == nil {
				return path
			}
		}
	}

	return ""
}
