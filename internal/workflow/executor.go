package workflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/ppopcode/ppopcode/internal/orchestrator"
)

// ExecutionProgress represents progress updates during workflow execution
type ExecutionProgress struct {
	NodeID   string
	NodeName string
	NodeType string
	Status   string // "started", "output", "completed", "error", "waiting_input"
	Output   string
	Question string   // for askUserQuestion
	Options  []string // for askUserQuestion
	Done     bool
}

type NodeHandler func(ctx context.Context, node *Node, execCtx *ExecutionContext) error
type AsyncNodeHandler func(ctx context.Context, node *Node, execCtx *ExecutionContext, progress chan<- ExecutionProgress) error

type Executor struct {
	workflow      *Workflow
	orchestrator  *orchestrator.Orchestrator
	handlers      map[string]NodeHandler
	asyncHandlers map[string]AsyncNodeHandler
	execCtx       *ExecutionContext

	// For async execution with user input
	answerChan    chan string
	answerMu      sync.Mutex
	waitingNodeID string
}

func NewExecutor(workflow *Workflow, orch *orchestrator.Orchestrator) *Executor {
	e := &Executor{
		workflow:      workflow,
		orchestrator:  orch,
		handlers:      make(map[string]NodeHandler),
		asyncHandlers: make(map[string]AsyncNodeHandler),
		execCtx:       NewExecutionContext(),
		answerChan:    make(chan string, 1),
	}

	e.registerDefaultHandlers()
	e.registerAsyncHandlers()
	return e
}

func (e *Executor) registerDefaultHandlers() {
	e.handlers["start"] = e.handleStart
	e.handlers["end"] = e.handleEnd
	e.handlers["prompt"] = e.handlePrompt
	e.handlers["askUserQuestion"] = e.handleQuestion
}

func (e *Executor) registerAsyncHandlers() {
	e.asyncHandlers["start"] = e.handleStartAsync
	e.asyncHandlers["end"] = e.handleEndAsync
	e.asyncHandlers["prompt"] = e.handlePromptAsync
	e.asyncHandlers["askUserQuestion"] = e.handleQuestionAsync
}

func (e *Executor) RegisterHandler(nodeType string, handler NodeHandler) {
	e.handlers[nodeType] = handler
}

func (e *Executor) SetVariable(key string, value interface{}) {
	e.execCtx.Set(key, value)
}

func (e *Executor) Execute(ctx context.Context) error {
	startNode := e.workflow.GetStartNode()
	if startNode == nil {
		return fmt.Errorf("workflow has no start node")
	}

	return e.executeNode(ctx, startNode)
}

func (e *Executor) executeNode(ctx context.Context, node *Node) error {
	if node == nil {
		return nil
	}

	handler, exists := e.handlers[node.Type]
	if !exists {
		return fmt.Errorf("no handler for node type: %s", node.Type)
	}

	if err := handler(ctx, node, e.execCtx); err != nil {
		return fmt.Errorf("node %s failed: %w", node.ID, err)
	}

	if node.Type == "end" {
		return nil
	}

	nextNodes := e.workflow.GetNextNodes(node.ID)
	for _, nextNode := range nextNodes {
		if err := e.executeNode(ctx, nextNode); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) handleStart(ctx context.Context, node *Node, execCtx *ExecutionContext) error {
	return nil
}

func (e *Executor) handleEnd(ctx context.Context, node *Node, execCtx *ExecutionContext) error {
	return nil
}

func (e *Executor) handlePrompt(ctx context.Context, node *Node, execCtx *ExecutionContext) error {
	prompt := execCtx.InterpolatePrompt(node.Data.Prompt)

	task, err := e.orchestrator.Process(ctx, prompt)
	if err != nil {
		return fmt.Errorf("orchestrator failed: %w", err)
	}

	execCtx.SetResult(node.ID, task.Result)
	return nil
}

func (e *Executor) handleQuestion(ctx context.Context, node *Node, execCtx *ExecutionContext) error {
	execCtx.SetResult(node.ID, map[string]interface{}{
		"question": node.Data.QuestionText,
		"options":  node.Data.Options,
		"pending":  true,
	})
	return nil
}

func (e *Executor) GetResults() map[string]interface{} {
	return e.execCtx.Results
}

// GetWorkflow returns the workflow being executed
func (e *Executor) GetWorkflow() *Workflow {
	return e.workflow
}

// ExecuteAsync starts async execution and returns a progress channel
func (e *Executor) ExecuteAsync(ctx context.Context) <-chan ExecutionProgress {
	progress := make(chan ExecutionProgress, 100)

	go func() {
		defer close(progress)

		startNode := e.workflow.GetStartNode()
		if startNode == nil {
			progress <- ExecutionProgress{
				Status: "error",
				Output: "workflow has no start node",
				Done:   true,
			}
			return
		}

		err := e.executeNodeAsync(ctx, startNode, progress)
		if err != nil {
			progress <- ExecutionProgress{
				Status: "error",
				Output: err.Error(),
				Done:   true,
			}
			return
		}

		progress <- ExecutionProgress{
			Status: "completed",
			Output: "Workflow completed successfully",
			Done:   true,
		}
	}()

	return progress
}

func (e *Executor) executeNodeAsync(ctx context.Context, node *Node, progress chan<- ExecutionProgress) error {
	if node == nil {
		return nil
	}

	// Send started status
	progress <- ExecutionProgress{
		NodeID:   node.ID,
		NodeName: node.Data.Label,
		NodeType: node.Type,
		Status:   "started",
	}

	handler, exists := e.asyncHandlers[node.Type]
	if !exists {
		// Fall back to sync handler if no async handler exists
		syncHandler, syncExists := e.handlers[node.Type]
		if !syncExists {
			return fmt.Errorf("no handler for node type: %s", node.Type)
		}
		if err := syncHandler(ctx, node, e.execCtx); err != nil {
			progress <- ExecutionProgress{
				NodeID:   node.ID,
				NodeName: node.Data.Label,
				NodeType: node.Type,
				Status:   "error",
				Output:   err.Error(),
			}
			return fmt.Errorf("node %s failed: %w", node.ID, err)
		}
	} else {
		if err := handler(ctx, node, e.execCtx, progress); err != nil {
			progress <- ExecutionProgress{
				NodeID:   node.ID,
				NodeName: node.Data.Label,
				NodeType: node.Type,
				Status:   "error",
				Output:   err.Error(),
			}
			return fmt.Errorf("node %s failed: %w", node.ID, err)
		}
	}

	// Send completed status
	progress <- ExecutionProgress{
		NodeID:   node.ID,
		NodeName: node.Data.Label,
		NodeType: node.Type,
		Status:   "completed",
	}

	if node.Type == "end" {
		return nil
	}

	nextNodes := e.workflow.GetNextNodes(node.ID)
	for _, nextNode := range nextNodes {
		if err := e.executeNodeAsync(ctx, nextNode, progress); err != nil {
			return err
		}
	}

	return nil
}

// ProvideAnswer provides an answer for askUserQuestion node
func (e *Executor) ProvideAnswer(answer string) {
	e.answerMu.Lock()
	defer e.answerMu.Unlock()

	select {
	case e.answerChan <- answer:
	default:
		// Channel already has a value, replace it
		select {
		case <-e.answerChan:
		default:
		}
		e.answerChan <- answer
	}
}

// IsWaitingForInput returns true if executor is waiting for user input
func (e *Executor) IsWaitingForInput() bool {
	e.answerMu.Lock()
	defer e.answerMu.Unlock()
	return e.waitingNodeID != ""
}

// Async handlers

func (e *Executor) handleStartAsync(ctx context.Context, node *Node, execCtx *ExecutionContext, progress chan<- ExecutionProgress) error {
	return nil
}

func (e *Executor) handleEndAsync(ctx context.Context, node *Node, execCtx *ExecutionContext, progress chan<- ExecutionProgress) error {
	return nil
}

func (e *Executor) handlePromptAsync(ctx context.Context, node *Node, execCtx *ExecutionContext, progress chan<- ExecutionProgress) error {
	prompt := execCtx.InterpolatePrompt(node.Data.Prompt)

	if e.orchestrator == nil {
		return fmt.Errorf("orchestrator not configured")
	}

	// Use streaming API
	progressChan := e.orchestrator.ProcessStreamAsync(ctx, prompt)

	var result string
	for update := range progressChan {
		if update.Type == "output" {
			result += update.Message
			progress <- ExecutionProgress{
				NodeID:   node.ID,
				NodeName: node.Data.Label,
				NodeType: node.Type,
				Status:   "output",
				Output:   update.Message,
			}
		} else if update.Type == "status" || update.Type == "thinking" {
			progress <- ExecutionProgress{
				NodeID:   node.ID,
				NodeName: node.Data.Label,
				NodeType: node.Type,
				Status:   "output",
				Output:   "[" + update.Agent + "] " + update.Message,
			}
		} else if update.Type == "error" {
			return fmt.Errorf("orchestrator error: %s", update.Message)
		}
	}

	execCtx.SetResult(node.ID, result)
	return nil
}

func (e *Executor) handleQuestionAsync(ctx context.Context, node *Node, execCtx *ExecutionContext, progress chan<- ExecutionProgress) error {
	// Set waiting state
	e.answerMu.Lock()
	e.waitingNodeID = node.ID
	e.answerMu.Unlock()

	// Convert options to string slice
	var options []string
	for _, opt := range node.Data.Options {
		if s, ok := opt.(string); ok {
			options = append(options, s)
		} else if m, ok := opt.(map[string]interface{}); ok {
			if label, exists := m["label"]; exists {
				options = append(options, fmt.Sprintf("%v", label))
			}
		}
	}

	// Send waiting_input status
	progress <- ExecutionProgress{
		NodeID:   node.ID,
		NodeName: node.Data.Label,
		NodeType: node.Type,
		Status:   "waiting_input",
		Question: node.Data.QuestionText,
		Options:  options,
	}

	// Wait for answer
	select {
	case answer := <-e.answerChan:
		e.answerMu.Lock()
		e.waitingNodeID = ""
		e.answerMu.Unlock()

		execCtx.SetResult(node.ID, answer)
		execCtx.Set("userAnswer", answer)
		return nil
	case <-ctx.Done():
		e.answerMu.Lock()
		e.waitingNodeID = ""
		e.answerMu.Unlock()
		return ctx.Err()
	}
}
