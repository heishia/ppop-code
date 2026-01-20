package agents

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*1000000000) // 5 second timeout (5 * 1 billion nanoseconds)
	defer cancel()

	// First check version to see if CLI works
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

	// Check actual login status by trying a minimal prompt
	// If not logged in, this will fail with an auth error
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*1000000000) // 5 seconds
	defer cancel2()

	testCmd := exec.CommandContext(ctx2, cliPath, "-p", "test", "--output-format", "text")
	var testStdout, testStderr bytes.Buffer
	testCmd.Stdout = &testStdout
	testCmd.Stderr = &testStderr

	testErr := testCmd.Run()
	if testErr != nil {
		stderrStr := testStderr.String()
		// Check for common auth error messages
		if strings.Contains(stderrStr, "login") || strings.Contains(stderrStr, "Invalid API key") ||
			strings.Contains(stderrStr, "authentication") || strings.Contains(stderrStr, "Please run /login") {
			return ClaudeLoginStatus{
				LoggedIn: false,
				Message:  "Not logged in. Please run 'claude login'",
				CLIFound: true,
			}
		}
		// Other errors might not be auth-related
		return ClaudeLoginStatus{
			LoggedIn: false,
			Message:  fmt.Sprintf("Claude CLI error: %s", stderrStr),
			CLIFound: true,
		}
	}

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

	// On Windows, we need to open a new terminal window for interactive login
	switch runtime.GOOS {
	case "windows":
		// Use cmd.exe to open a new window and run claude with /login command
		// Note: "start" requires empty string "" as window title when the command has arguments
		// /login is an internal Claude Code command (not a CLI flag)
		cmd := exec.Command("cmd", "/c", "start", "", "cmd", "/k", cliPath, "/login")
		return cmd.Start()
	case "darwin":
		// macOS: open in Terminal.app
		script := fmt.Sprintf(`tell application "Terminal" to do script "%s /login"`, cliPath)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Start()
	default:
		// Linux: try various terminal emulators
		terminals := []string{"gnome-terminal", "konsole", "xterm"}
		for _, term := range terminals {
			if _, err := exec.LookPath(term); err == nil {
				var cmd *exec.Cmd
				switch term {
				case "gnome-terminal":
					cmd = exec.Command(term, "--", cliPath, "/login")
				case "konsole":
					cmd = exec.Command(term, "-e", cliPath, "/login")
				default:
					cmd = exec.Command(term, "-e", cliPath, "/login")
				}
				return cmd.Start()
			}
		}
		return fmt.Errorf("no suitable terminal emulator found")
	}
}

// findClaudeCLI finds the claude CLI executable
func findClaudeCLI() string {
	var possiblePaths []string

	switch runtime.GOOS {
	case "windows":
		// Check npm global install location first
		npmPath := filepath.Join(os.Getenv("APPDATA"), "npm", "claude.cmd")
		possiblePaths = []string{
			npmPath,
			"claude.cmd",
			"claude",
			"claude.exe",
		}
	default:
		possiblePaths = []string{
			"claude",
			"/usr/local/bin/claude",
			"/usr/bin/claude",
		}
	}

	for _, p := range possiblePaths {
		// Check if file exists directly
		if _, err := os.Stat(p); err == nil {
			return p
		}
		// Check in PATH
		if path, err := exec.LookPath(p); err == nil {
			return path
		}
	}

	return ""
}
