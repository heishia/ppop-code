package workflow

import (
	"context"
	"fmt"

	"github.com/ppopcode/ppopcode/internal/orchestrator"
)

type NodeHandler func(ctx context.Context, node *Node, execCtx *ExecutionContext) error

type Executor struct {
	workflow     *Workflow
	orchestrator *orchestrator.Orchestrator
	handlers     map[string]NodeHandler
	execCtx      *ExecutionContext
}

func NewExecutor(workflow *Workflow, orch *orchestrator.Orchestrator) *Executor {
	e := &Executor{
		workflow:     workflow,
		orchestrator: orch,
		handlers:     make(map[string]NodeHandler),
		execCtx:      NewExecutionContext(),
	}

	e.registerDefaultHandlers()
	return e
}

func (e *Executor) registerDefaultHandlers() {
	e.handlers["start"] = e.handleStart
	e.handlers["end"] = e.handleEnd
	e.handlers["prompt"] = e.handlePrompt
	e.handlers["askUserQuestion"] = e.handleQuestion
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
