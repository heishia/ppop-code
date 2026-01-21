package agents

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	// --continue: maintain conversation context across calls
	args := []string{"-p", prompt, "--output-format", "text", "--continue"}

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
	// Note: stream-json requires --verbose flag
	// --continue: maintain conversation context across calls
	args := []string{"-p", prompt, "--output-format", "stream-json", "--verbose", "--continue"}

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
	var wg sync.WaitGroup

	// Read stdout in real-time
	wg.Add(1)
	go func() {
		defer wg.Done()
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
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	// Wait for command to complete
	cmdErr := cmd.Wait()

	// Wait for all readers to finish BEFORE closing stream
	wg.Wait()

	if cmdErr != nil {
		if ctx.Err() != nil {
			stream <- StreamChunk{Content: "Cancelled", Type: "status", Done: true}
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		stream <- StreamChunk{Content: fmt.Sprintf("Error: %v", cmdErr), Type: "error", Done: true}
		return nil, fmt.Errorf("claude error: %w", cmdErr)
	}

	stream <- StreamChunk{Content: "Done", Type: "status", Done: true}

	return &Response{
		Content: strings.TrimSpace(fullOutput.String()),
		Model:   a.config.Model,
	}, nil
}

// claudeStreamEvent represents the JSON structure from Claude CLI stream-json output
type claudeStreamEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
	Result  string `json:"result,omitempty"`
	Message struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text,omitempty"`
		} `json:"content,omitempty"`
	} `json:"message,omitempty"`
}

// parseClaudeStreamLine parses a line from Claude CLI stream-json output
func parseClaudeStreamLine(line string) (chunkType, content string) {
	var event claudeStreamEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		// If JSON parsing fails, return the raw line
		return "output", line
	}

	switch event.Type {
	case "system":
		// Init event - show as status
		return "status", "Claude initialized"
	case "assistant":
		// Extract text from message content
		for _, c := range event.Message.Content {
			if c.Type == "text" && c.Text != "" {
				return "output", c.Text
			}
			if c.Type == "thinking" && c.Text != "" {
				return "thinking", c.Text
			}
		}
		return "", ""
	case "result":
		// Final result - could use this instead of accumulating
		if event.Result != "" {
			return "output", event.Result
		}
		return "status", "Completed"
	default:
		return "", ""
	}
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
