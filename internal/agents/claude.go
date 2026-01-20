package agents

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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

	// Find claude CLI path (fast operation)
	cliPath := findClaudeCLI()
	if cliPath == "" {
		agent.status = "cli_not_found"
		return agent, nil
	}
	agent.cliPath = cliPath

	// Don't check login status at startup - it's slow (5s timeout)
	// Login will be checked when Execute() is called
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

// ExecuteStream executes the prompt and streams output in real-time
func (a *ClaudeAgent) ExecuteStream(ctx context.Context, prompt string, stream chan<- StreamChunk) (*Response, error) {
	defer close(stream)

	if a.cliPath == "" {
		stream <- StreamChunk{Content: "Claude CLI not found", Type: "error", Done: true}
		return &Response{
			Content: fmt.Sprintf("[%s] Claude CLI not found.", a.config.Name),
			Model:   a.config.Model,
		}, nil
	}

	status := CheckClaudeLogin()
	if !status.LoggedIn {
		stream <- StreamChunk{Content: "Not logged in to Claude", Type: "error", Done: true}
		return &Response{
			Content: fmt.Sprintf("[%s] Not logged in to Claude.", a.config.Name),
			Model:   a.config.Model,
		}, nil
	}

	a.SetStatus("processing")
	defer a.SetStatus("ready")

	stream <- StreamChunk{Content: "Starting Claude...", Type: "status"}

	// Build command arguments - use streaming output
	args := []string{"-p", prompt, "--output-format", "stream-json"}

	cmd := exec.CommandContext(ctx, a.cliPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stream <- StreamChunk{Content: fmt.Sprintf("Failed to create pipe: %v", err), Type: "error", Done: true}
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stream <- StreamChunk{Content: fmt.Sprintf("Failed to create stderr pipe: %v", err), Type: "error", Done: true}
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stream <- StreamChunk{Content: fmt.Sprintf("Failed to start: %v", err), Type: "error", Done: true}
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}

	var fullOutput strings.Builder

	// Read stdout in real-time
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					stream <- StreamChunk{Content: fmt.Sprintf("Read error: %v", err), Type: "error"}
				}
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse streaming JSON output from Claude CLI
			chunkType, content := parseClaudeStreamLine(line)
			if content != "" {
				fullOutput.WriteString(content)
				stream <- StreamChunk{Content: content, Type: chunkType}
			}
		}
	}()

	// Read stderr for thinking/status
	go func() {
		reader := bufio.NewReader(stderr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line != "" {
				stream <- StreamChunk{Content: line, Type: "thinking"}
			}
		}
	}()

	err = cmd.Wait()
	if err != nil {
		if ctx.Err() != nil {
			stream <- StreamChunk{Content: "Cancelled", Type: "status", Done: true}
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		stream <- StreamChunk{Content: fmt.Sprintf("Error: %v", err), Type: "error", Done: true}
		return nil, fmt.Errorf("claude error: %w", err)
	}

	stream <- StreamChunk{Content: "Done", Type: "status", Done: true}

	return &Response{
		Content: strings.TrimSpace(fullOutput.String()),
		Model:   a.config.Model,
	}, nil
}

// parseClaudeStreamLine parses a line from Claude CLI stream-json output
func parseClaudeStreamLine(line string) (chunkType, content string) {
	// Claude CLI stream-json format outputs JSON objects
	// We do simple parsing here - could use encoding/json for more robust parsing
	if strings.Contains(line, `"type":"thinking"`) || strings.Contains(line, `"thinking"`) {
		// Extract thinking content
		if idx := strings.Index(line, `"content":"`); idx != -1 {
			start := idx + len(`"content":"`)
			end := strings.Index(line[start:], `"`)
			if end != -1 {
				return "thinking", line[start : start+end]
			}
		}
		return "thinking", line
	}
	if strings.Contains(line, `"type":"text"`) || strings.Contains(line, `"text"`) {
		if idx := strings.Index(line, `"content":"`); idx != -1 {
			start := idx + len(`"content":"`)
			end := strings.Index(line[start:], `"`)
			if end != -1 {
				return "output", line[start : start+end]
			}
		}
		// If it's plain text output
		return "output", line
	}
	// Default: treat as output
	return "output", line
}

// CheckClaudeLogin checks if Claude CLI is available (fast check - no API calls)
// Actual login status is verified when Execute() is called
func CheckClaudeLogin() ClaudeLoginStatus {
	cliPath := findClaudeCLI()
	if cliPath == "" {
		return ClaudeLoginStatus{
			LoggedIn: false,
			Message:  "Claude CLI not found. Install it from https://claude.ai/cli",
			CLIFound: false,
		}
	}

	// Fast check: just verify CLI exists and is executable
	// Don't check login status here - it's slow and will be checked when actually used
	return ClaudeLoginStatus{
		LoggedIn: true, // Assume logged in - will error on Execute() if not
		Message:  "Claude CLI ready",
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
