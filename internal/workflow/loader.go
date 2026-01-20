package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type NodeData struct {
	Label        string                 `json:"label"`
	Prompt       string                 `json:"prompt,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	QuestionText string                 `json:"questionText,omitempty"`
	Options      []interface{}          `json:"options,omitempty"`
}

type Node struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
}

type Connection struct {
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	FromPort string `json:"fromPort"`
	ToPort   string `json:"toPort"`
}

type Workflow struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Version       string       `json:"version"`
	Nodes         []Node       `json:"nodes"`
	Connections   []Connection `json:"connections"`
	CreatedAt     string       `json:"createdAt"`
	UpdatedAt     string       `json:"updatedAt"`
	SubAgentFlows []interface{} `json:"subAgentFlows"`
}

type Loader struct {
	workflowDir string
}

func NewLoader(workflowDir string) *Loader {
	return &Loader{
		workflowDir: workflowDir,
	}
}

func (l *Loader) ListWorkflows() ([]string, error) {
	var workflows []string

	entries, err := os.ReadDir(l.workflowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			workflows = append(workflows, name)
		}
	}

	return workflows, nil
}

func (l *Loader) Load(name string) (*Workflow, error) {
	filename := name
	if !strings.HasSuffix(filename, ".json") {
		filename = name + ".json"
	}

	path := filepath.Join(l.workflowDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	return &workflow, nil
}

func (w *Workflow) GetStartNode() *Node {
	for i := range w.Nodes {
		if w.Nodes[i].Type == "start" {
			return &w.Nodes[i]
		}
	}
	return nil
}

func (w *Workflow) GetEndNode() *Node {
	for i := range w.Nodes {
		if w.Nodes[i].Type == "end" {
			return &w.Nodes[i]
		}
	}
	return nil
}

func (w *Workflow) GetNextNodes(nodeID string) []*Node {
	var nextNodes []*Node

	for _, conn := range w.Connections {
		if conn.From == nodeID {
			for i := range w.Nodes {
				if w.Nodes[i].ID == conn.To {
					nextNodes = append(nextNodes, &w.Nodes[i])
				}
			}
		}
	}

	return nextNodes
}

func (w *Workflow) GetNode(nodeID string) *Node {
	for i := range w.Nodes {
		if w.Nodes[i].ID == nodeID {
			return &w.Nodes[i]
		}
	}
	return nil
}

type ExecutionContext struct {
	Variables map[string]interface{}
	Results   map[string]interface{}
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Variables: make(map[string]interface{}),
		Results:   make(map[string]interface{}),
	}
}

func (ctx *ExecutionContext) Set(key string, value interface{}) {
	ctx.Variables[key] = value
}

func (ctx *ExecutionContext) Get(key string) interface{} {
	return ctx.Variables[key]
}

func (ctx *ExecutionContext) SetResult(nodeID string, result interface{}) {
	ctx.Results[nodeID] = result
}

func (ctx *ExecutionContext) GetResult(nodeID string) interface{} {
	return ctx.Results[nodeID]
}

func (ctx *ExecutionContext) InterpolatePrompt(prompt string) string {
	result := prompt
	for key, value := range ctx.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}
