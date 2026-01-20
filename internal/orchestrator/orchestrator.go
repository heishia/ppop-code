package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/ppopcode/ppopcode/internal/agents"
)

// ProgressUpdate represents a progress update during processing
type ProgressUpdate struct {
	Stage   string // "routing", "processing", "streaming"
	Message string
	Agent   string
	Type    string // "status", "thinking", "output", "error"
	Done    bool
}

type TaskType string

const (
	TaskTypeGeneral TaskType = "general"
	TaskTypeUI      TaskType = "ui"
	TaskTypeDesign  TaskType = "design"
	TaskTypeDebug   TaskType = "debug"
	TaskTypeCode    TaskType = "code"
)

type Task struct {
	ID         string
	Content    string
	Type       TaskType
	AssignedTo string
	Status     string
	Result     string
}

type Orchestrator struct {
	router      *Router
	agents      map[string]agents.Agent
	currentTask *Task
}

func New(agentConfigs map[string]agents.AgentConfig) *Orchestrator {
	o := &Orchestrator{
		router: NewRouter(),
		agents: make(map[string]agents.Agent),
	}

	for name, config := range agentConfigs {
		agent, err := agents.NewAgent(config)
		if err != nil {
			fmt.Printf("Warning: Failed to create agent %s: %v\n", name, err)
			continue
		}
		o.agents[name] = agent
	}

	return o
}

func (o *Orchestrator) Process(ctx context.Context, input string) (*Task, error) {
	taskType := o.analyzeTask(input)

	agentName := o.router.Route(taskType)

	task := &Task{
		ID:         fmt.Sprintf("task-%d", len(input)),
		Content:    input,
		Type:       taskType,
		AssignedTo: agentName,
		Status:     "processing",
	}
	o.currentTask = task

	agent, exists := o.agents[agentName]
	if !exists {
		task.Status = "error"
		task.Result = fmt.Sprintf("Agent %s not found", agentName)
		return task, fmt.Errorf("agent %s not found", agentName)
	}

	response, err := agent.Execute(ctx, input)
	if err != nil {
		task.Status = "error"
		task.Result = err.Error()
		return task, err
	}

	task.Status = "completed"
	task.Result = response.Content
	return task, nil
}

func (o *Orchestrator) analyzeTask(input string) TaskType {
	lower := strings.ToLower(input)

	uiKeywords := []string{"ui", "ux", "frontend", "component", "button", "form", "layout", "style", "css", "design", "화면", "디자인", "컴포넌트", "프론트"}
	for _, kw := range uiKeywords {
		if strings.Contains(lower, kw) {
			return TaskTypeUI
		}
	}

	designKeywords := []string{"architecture", "structure", "refactor", "pattern", "설계", "아키텍처", "구조", "리팩토링"}
	for _, kw := range designKeywords {
		if strings.Contains(lower, kw) {
			return TaskTypeDesign
		}
	}

	debugKeywords := []string{"bug", "error", "fix", "debug", "issue", "problem", "버그", "에러", "오류", "수정", "디버그"}
	for _, kw := range debugKeywords {
		if strings.Contains(lower, kw) {
			return TaskTypeDebug
		}
	}

	return TaskTypeGeneral
}

func (o *Orchestrator) GetCurrentTask() *Task {
	return o.currentTask
}

// ProcessStream processes the input with real-time progress updates
func (o *Orchestrator) ProcessStream(ctx context.Context, input string, progress chan<- ProgressUpdate) (*Task, error) {
	defer close(progress)

	// Send routing status
	progress <- ProgressUpdate{Stage: "routing", Message: "Analyzing task...", Type: "status"}

	taskType := o.analyzeTask(input)
	agentName := o.router.Route(taskType)

	progress <- ProgressUpdate{
		Stage:   "routing",
		Message: fmt.Sprintf("Routing to %s agent", agentName),
		Agent:   agentName,
		Type:    "status",
	}

	task := &Task{
		ID:         fmt.Sprintf("task-%d", len(input)),
		Content:    input,
		Type:       taskType,
		AssignedTo: agentName,
		Status:     "processing",
	}
	o.currentTask = task

	agent, exists := o.agents[agentName]
	if !exists {
		task.Status = "error"
		task.Result = fmt.Sprintf("Agent %s not found", agentName)
		progress <- ProgressUpdate{Stage: "error", Message: task.Result, Type: "error", Done: true}
		return task, fmt.Errorf("agent %s not found", agentName)
	}

	progress <- ProgressUpdate{
		Stage:   "processing",
		Message: fmt.Sprintf("Starting %s...", agentName),
		Agent:   agentName,
		Type:    "status",
	}

	// Create stream channel for agent
	agentStream := make(chan agents.StreamChunk, 100)

	// Start streaming execution in goroutine
	var response *agents.Response
	var execErr error

	go func() {
		response, execErr = agent.ExecuteStream(ctx, input, agentStream)
	}()

	// Forward agent stream chunks to progress channel
	for chunk := range agentStream {
		progress <- ProgressUpdate{
			Stage:   "streaming",
			Message: chunk.Content,
			Agent:   agentName,
			Type:    chunk.Type,
			Done:    chunk.Done,
		}
	}

	if execErr != nil {
		task.Status = "error"
		task.Result = execErr.Error()
		progress <- ProgressUpdate{Stage: "error", Message: execErr.Error(), Type: "error", Done: true}
		return task, execErr
	}

	task.Status = "completed"
	task.Result = response.Content
	progress <- ProgressUpdate{Stage: "completed", Message: "Done", Agent: agentName, Type: "status", Done: true}

	return task, nil
}

func (o *Orchestrator) GetAgentStatus() map[string]string {
	status := make(map[string]string)
	for name, agent := range o.agents {
		status[name] = agent.Status()
	}
	return status
}
