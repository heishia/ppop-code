package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test fixtures
func createTestWorkflow() *Workflow {
	return &Workflow{
		ID:      "test-workflow",
		Name:    "Test Workflow",
		Version: "1.0.0",
		Nodes: []Node{
			{
				ID:   "node-1",
				Type: "start",
				Data: NodeData{Label: "Start"},
			},
			{
				ID:   "node-2",
				Type: "prompt",
				Data: NodeData{
					Label:  "Prompt Node",
					Prompt: "Hello {{name}}!",
				},
			},
			{
				ID:   "node-3",
				Type: "askUserQuestion",
				Data: NodeData{
					Label:        "Question Node",
					QuestionText: "What is your choice?",
					Options:      []interface{}{"Option A", "Option B"},
				},
			},
			{
				ID:   "node-4",
				Type: "end",
				Data: NodeData{Label: "End"},
			},
		},
		Connections: []Connection{
			{ID: "conn-1", From: "node-1", To: "node-2"},
			{ID: "conn-2", From: "node-2", To: "node-3"},
			{ID: "conn-3", From: "node-3", To: "node-4"},
		},
	}
}

func createSimpleWorkflow() *Workflow {
	return &Workflow{
		ID:      "simple-workflow",
		Name:    "Simple Workflow",
		Version: "1.0.0",
		Nodes: []Node{
			{ID: "start", Type: "start", Data: NodeData{Label: "Start"}},
			{ID: "end", Type: "end", Data: NodeData{Label: "End"}},
		},
		Connections: []Connection{
			{ID: "conn-1", From: "start", To: "end"},
		},
	}
}

// ============ Workflow Tests ============

func TestWorkflow_GetStartNode(t *testing.T) {
	wf := createTestWorkflow()

	startNode := wf.GetStartNode()
	if startNode == nil {
		t.Fatal("GetStartNode() should not return nil")
	}
	if startNode.ID != "node-1" {
		t.Errorf("GetStartNode() ID = %s, want node-1", startNode.ID)
	}
	if startNode.Type != "start" {
		t.Errorf("GetStartNode() Type = %s, want start", startNode.Type)
	}
}

func TestWorkflow_GetStartNode_NotFound(t *testing.T) {
	wf := &Workflow{
		Nodes: []Node{
			{ID: "end", Type: "end"},
		},
	}

	startNode := wf.GetStartNode()
	if startNode != nil {
		t.Error("GetStartNode() should return nil when no start node exists")
	}
}

func TestWorkflow_GetEndNode(t *testing.T) {
	wf := createTestWorkflow()

	endNode := wf.GetEndNode()
	if endNode == nil {
		t.Fatal("GetEndNode() should not return nil")
	}
	if endNode.ID != "node-4" {
		t.Errorf("GetEndNode() ID = %s, want node-4", endNode.ID)
	}
	if endNode.Type != "end" {
		t.Errorf("GetEndNode() Type = %s, want end", endNode.Type)
	}
}

func TestWorkflow_GetNode(t *testing.T) {
	wf := createTestWorkflow()

	tests := []struct {
		nodeID   string
		wantNil  bool
		wantType string
	}{
		{"node-1", false, "start"},
		{"node-2", false, "prompt"},
		{"node-3", false, "askUserQuestion"},
		{"node-4", false, "end"},
		{"nonexistent", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.nodeID, func(t *testing.T) {
			node := wf.GetNode(tt.nodeID)
			if tt.wantNil {
				if node != nil {
					t.Errorf("GetNode(%s) should return nil", tt.nodeID)
				}
			} else {
				if node == nil {
					t.Fatalf("GetNode(%s) should not return nil", tt.nodeID)
				}
				if node.Type != tt.wantType {
					t.Errorf("GetNode(%s).Type = %s, want %s", tt.nodeID, node.Type, tt.wantType)
				}
			}
		})
	}
}

func TestWorkflow_GetNextNodes(t *testing.T) {
	wf := createTestWorkflow()

	tests := []struct {
		nodeID    string
		wantCount int
		wantIDs   []string
	}{
		{"node-1", 1, []string{"node-2"}},
		{"node-2", 1, []string{"node-3"}},
		{"node-3", 1, []string{"node-4"}},
		{"node-4", 0, nil},
		{"nonexistent", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.nodeID, func(t *testing.T) {
			nextNodes := wf.GetNextNodes(tt.nodeID)
			if len(nextNodes) != tt.wantCount {
				t.Errorf("GetNextNodes(%s) count = %d, want %d", tt.nodeID, len(nextNodes), tt.wantCount)
			}
			for i, wantID := range tt.wantIDs {
				if i < len(nextNodes) && nextNodes[i].ID != wantID {
					t.Errorf("GetNextNodes(%s)[%d].ID = %s, want %s", tt.nodeID, i, nextNodes[i].ID, wantID)
				}
			}
		})
	}
}

// ============ ExecutionContext Tests ============

func TestExecutionContext_SetGet(t *testing.T) {
	ctx := NewExecutionContext()

	ctx.Set("name", "John")
	ctx.Set("age", 30)

	if ctx.Get("name") != "John" {
		t.Errorf("Get(name) = %v, want John", ctx.Get("name"))
	}
	if ctx.Get("age") != 30 {
		t.Errorf("Get(age) = %v, want 30", ctx.Get("age"))
	}
	if ctx.Get("nonexistent") != nil {
		t.Error("Get(nonexistent) should return nil")
	}
}

func TestExecutionContext_SetResultGetResult(t *testing.T) {
	ctx := NewExecutionContext()

	ctx.SetResult("node-1", "result1")
	ctx.SetResult("node-2", map[string]string{"key": "value"})

	if ctx.GetResult("node-1") != "result1" {
		t.Errorf("GetResult(node-1) = %v, want result1", ctx.GetResult("node-1"))
	}
	result2 := ctx.GetResult("node-2").(map[string]string)
	if result2["key"] != "value" {
		t.Errorf("GetResult(node-2)[key] = %v, want value", result2["key"])
	}
	if ctx.GetResult("nonexistent") != nil {
		t.Error("GetResult(nonexistent) should return nil")
	}
}

func TestExecutionContext_InterpolatePrompt(t *testing.T) {
	ctx := NewExecutionContext()
	ctx.Set("name", "Alice")
	ctx.Set("project", "ppopcode")

	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "single variable",
			prompt:   "Hello {{name}}!",
			expected: "Hello Alice!",
		},
		{
			name:     "multiple variables",
			prompt:   "{{name}} is working on {{project}}",
			expected: "Alice is working on ppopcode",
		},
		{
			name:     "no variables",
			prompt:   "Hello World!",
			expected: "Hello World!",
		},
		{
			name:     "unknown variable",
			prompt:   "Hello {{unknown}}!",
			expected: "Hello {{unknown}}!",
		},
		{
			name:     "empty prompt",
			prompt:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.InterpolatePrompt(tt.prompt)
			if result != tt.expected {
				t.Errorf("InterpolatePrompt(%q) = %q, want %q", tt.prompt, result, tt.expected)
			}
		})
	}
}

// ============ Executor Tests ============

func TestNewExecutor(t *testing.T) {
	wf := createTestWorkflow()
	executor := NewExecutor(wf, nil)

	if executor == nil {
		t.Fatal("NewExecutor() should not return nil")
	}
	if executor.workflow != wf {
		t.Error("Executor should have the correct workflow")
	}
	if executor.handlers == nil {
		t.Error("Executor should have handlers map initialized")
	}
	if executor.asyncHandlers == nil {
		t.Error("Executor should have asyncHandlers map initialized")
	}
	if executor.execCtx == nil {
		t.Error("Executor should have execution context initialized")
	}
}

func TestExecutor_GetWorkflow(t *testing.T) {
	wf := createTestWorkflow()
	executor := NewExecutor(wf, nil)

	if executor.GetWorkflow() != wf {
		t.Error("GetWorkflow() should return the workflow")
	}
}

func TestExecutor_SetVariable(t *testing.T) {
	wf := createTestWorkflow()
	executor := NewExecutor(wf, nil)

	executor.SetVariable("test", "value")
	if executor.execCtx.Get("test") != "value" {
		t.Error("SetVariable() should set variable in execution context")
	}
}

func TestExecutor_GetResults(t *testing.T) {
	wf := createTestWorkflow()
	executor := NewExecutor(wf, nil)

	results := executor.GetResults()
	if results == nil {
		t.Error("GetResults() should not return nil")
	}
}

func TestExecutor_Execute_SimpleWorkflow(t *testing.T) {
	wf := createSimpleWorkflow()
	executor := NewExecutor(wf, nil)

	ctx := context.Background()
	err := executor.Execute(ctx)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestExecutor_Execute_NoStartNode(t *testing.T) {
	wf := &Workflow{
		Nodes: []Node{
			{ID: "end", Type: "end"},
		},
	}
	executor := NewExecutor(wf, nil)

	ctx := context.Background()
	err := executor.Execute(ctx)
	if err == nil {
		t.Error("Execute() should return error when no start node exists")
	}
}

func TestExecutor_ExecuteAsync_SimpleWorkflow(t *testing.T) {
	wf := createSimpleWorkflow()
	executor := NewExecutor(wf, nil)

	ctx := context.Background()
	progressChan := executor.ExecuteAsync(ctx)

	var progressUpdates []ExecutionProgress
	for progress := range progressChan {
		progressUpdates = append(progressUpdates, progress)
	}

	if len(progressUpdates) == 0 {
		t.Error("ExecuteAsync() should produce progress updates")
	}

	// Check that we got completion
	lastUpdate := progressUpdates[len(progressUpdates)-1]
	if !lastUpdate.Done {
		t.Error("Last progress update should have Done=true")
	}
}

func TestExecutor_ExecuteAsync_NoStartNode(t *testing.T) {
	wf := &Workflow{
		Nodes: []Node{
			{ID: "end", Type: "end"},
		},
	}
	executor := NewExecutor(wf, nil)

	ctx := context.Background()
	progressChan := executor.ExecuteAsync(ctx)

	var lastUpdate ExecutionProgress
	for progress := range progressChan {
		lastUpdate = progress
	}

	if lastUpdate.Status != "error" {
		t.Errorf("ExecuteAsync() with no start node should produce error status, got %s", lastUpdate.Status)
	}
	if !lastUpdate.Done {
		t.Error("Error update should have Done=true")
	}
}

func TestExecutor_ExecuteAsync_ProgressOrder(t *testing.T) {
	wf := createSimpleWorkflow()
	executor := NewExecutor(wf, nil)

	ctx := context.Background()
	progressChan := executor.ExecuteAsync(ctx)

	var statuses []string
	for progress := range progressChan {
		if progress.NodeID != "" {
			statuses = append(statuses, progress.NodeID+":"+progress.Status)
		}
	}

	// Should have started and completed for both start and end nodes
	expectedStatuses := []string{
		"start:started",
		"start:completed",
		"end:started",
		"end:completed",
	}

	if len(statuses) != len(expectedStatuses) {
		t.Errorf("Expected %d status updates, got %d: %v", len(expectedStatuses), len(statuses), statuses)
	}

	for i, expected := range expectedStatuses {
		if i < len(statuses) && statuses[i] != expected {
			t.Errorf("Status[%d] = %s, want %s", i, statuses[i], expected)
		}
	}
}

func TestExecutor_ProvideAnswer(t *testing.T) {
	wf := createTestWorkflow()
	executor := NewExecutor(wf, nil)

	// Should not panic
	executor.ProvideAnswer("test answer")

	// Should replace previous answer
	executor.ProvideAnswer("new answer")
}

func TestExecutor_IsWaitingForInput(t *testing.T) {
	wf := createTestWorkflow()
	executor := NewExecutor(wf, nil)

	if executor.IsWaitingForInput() {
		t.Error("IsWaitingForInput() should return false initially")
	}
}

func TestExecutor_ExecuteAsync_WithQuestion(t *testing.T) {
	// Create a workflow with just start -> question -> end
	wf := &Workflow{
		ID:   "question-workflow",
		Name: "Question Workflow",
		Nodes: []Node{
			{ID: "start", Type: "start", Data: NodeData{Label: "Start"}},
			{
				ID:   "question",
				Type: "askUserQuestion",
				Data: NodeData{
					Label:        "Question",
					QuestionText: "What is your name?",
					Options:      []interface{}{"Alice", "Bob"},
				},
			},
			{ID: "end", Type: "end", Data: NodeData{Label: "End"}},
		},
		Connections: []Connection{
			{ID: "conn-1", From: "start", To: "question"},
			{ID: "conn-2", From: "question", To: "end"},
		},
	}

	executor := NewExecutor(wf, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	progressChan := executor.ExecuteAsync(ctx)

	// Collect progress until we get waiting_input
	var gotWaitingInput bool
	var questionText string
	var options []string

	go func() {
		// Simulate user providing answer after a short delay
		time.Sleep(100 * time.Millisecond)
		executor.ProvideAnswer("Alice")
	}()

	for progress := range progressChan {
		if progress.Status == "waiting_input" {
			gotWaitingInput = true
			questionText = progress.Question
			options = progress.Options
		}
	}

	if !gotWaitingInput {
		t.Error("Should have received waiting_input status")
	}
	if questionText != "What is your name?" {
		t.Errorf("Question = %q, want %q", questionText, "What is your name?")
	}
	if len(options) != 2 {
		t.Errorf("Options count = %d, want 2", len(options))
	}
}

func TestExecutor_ExecuteAsync_CancelContext(t *testing.T) {
	// Create a workflow with a question that won't be answered
	wf := &Workflow{
		ID:   "cancel-workflow",
		Name: "Cancel Workflow",
		Nodes: []Node{
			{ID: "start", Type: "start", Data: NodeData{Label: "Start"}},
			{
				ID:   "question",
				Type: "askUserQuestion",
				Data: NodeData{
					Label:        "Question",
					QuestionText: "Waiting forever?",
				},
			},
			{ID: "end", Type: "end", Data: NodeData{Label: "End"}},
		},
		Connections: []Connection{
			{ID: "conn-1", From: "start", To: "question"},
			{ID: "conn-2", From: "question", To: "end"},
		},
	}

	executor := NewExecutor(wf, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is always called
	progressChan := executor.ExecuteAsync(ctx)

	// Wait for waiting_input then cancel
	for progress := range progressChan {
		if progress.Status == "waiting_input" {
			cancel()
			break
		}
	}

	// Drain remaining progress
	for range progressChan {
	}

	// Should not be waiting anymore after cancel
	if executor.IsWaitingForInput() {
		t.Error("Should not be waiting for input after context cancel")
	}
}

// ============ Loader Tests ============

func TestLoader_Load(t *testing.T) {
	// Create a temporary directory with a test workflow
	tmpDir := t.TempDir()

	workflowJSON := `{
		"id": "test-wf",
		"name": "Test Workflow",
		"version": "1.0.0",
		"nodes": [
			{"id": "start", "type": "start", "data": {"label": "Start"}},
			{"id": "end", "type": "end", "data": {"label": "End"}}
		],
		"connections": [
			{"id": "conn-1", "from": "start", "to": "end"}
		]
	}`

	err := os.WriteFile(filepath.Join(tmpDir, "test.json"), []byte(workflowJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test workflow file: %v", err)
	}

	loader := NewLoader(tmpDir)

	// Test loading with .json extension
	wf, err := loader.Load("test.json")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if wf.Name != "Test Workflow" {
		t.Errorf("Load() Name = %s, want Test Workflow", wf.Name)
	}

	// Test loading without .json extension
	wf2, err := loader.Load("test")
	if err != nil {
		t.Fatalf("Load() without extension error = %v", err)
	}
	if wf2.Name != "Test Workflow" {
		t.Errorf("Load() without extension Name = %s, want Test Workflow", wf2.Name)
	}
}

func TestLoader_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	_, err := loader.Load("nonexistent.json")
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}

func TestLoader_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("not json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader(tmpDir)

	_, err = loader.Load("invalid.json")
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func TestLoader_ListWorkflows(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some workflow files
	os.WriteFile(filepath.Join(tmpDir, "workflow1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "workflow2.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "notjson.txt"), []byte(""), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	loader := NewLoader(tmpDir)

	workflows, err := loader.ListWorkflows()
	if err != nil {
		t.Fatalf("ListWorkflows() error = %v", err)
	}

	if len(workflows) != 2 {
		t.Errorf("ListWorkflows() count = %d, want 2", len(workflows))
	}

	// Should not include .json extension
	for _, name := range workflows {
		if filepath.Ext(name) == ".json" {
			t.Errorf("ListWorkflows() should not include .json extension: %s", name)
		}
	}
}

func TestLoader_ListWorkflows_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	workflows, err := loader.ListWorkflows()
	if err != nil {
		t.Fatalf("ListWorkflows() error = %v", err)
	}

	if len(workflows) != 0 {
		t.Errorf("ListWorkflows() count = %d, want 0", len(workflows))
	}
}

func TestLoader_ListWorkflows_NonexistentDir(t *testing.T) {
	loader := NewLoader("/nonexistent/path")

	_, err := loader.ListWorkflows()
	if err == nil {
		t.Error("ListWorkflows() should return error for nonexistent directory")
	}
}

// ============ ExecutionProgress Tests ============

func TestExecutionProgress_Fields(t *testing.T) {
	progress := ExecutionProgress{
		NodeID:   "node-1",
		NodeName: "Test Node",
		NodeType: "prompt",
		Status:   "started",
		Output:   "test output",
		Question: "test question",
		Options:  []string{"a", "b"},
		Done:     false,
	}

	if progress.NodeID != "node-1" {
		t.Errorf("NodeID = %s, want node-1", progress.NodeID)
	}
	if progress.NodeName != "Test Node" {
		t.Errorf("NodeName = %s, want Test Node", progress.NodeName)
	}
	if progress.NodeType != "prompt" {
		t.Errorf("NodeType = %s, want prompt", progress.NodeType)
	}
	if progress.Status != "started" {
		t.Errorf("Status = %s, want started", progress.Status)
	}
	if progress.Output != "test output" {
		t.Errorf("Output = %s, want test output", progress.Output)
	}
	if progress.Question != "test question" {
		t.Errorf("Question = %s, want test question", progress.Question)
	}
	if len(progress.Options) != 2 {
		t.Errorf("Options count = %d, want 2", len(progress.Options))
	}
	if progress.Done {
		t.Error("Done should be false")
	}
}
