package orchestrator

import (
	"testing"
)

func TestAnalyzeTask(t *testing.T) {
	o := &Orchestrator{}

	tests := []struct {
		name     string
		input    string
		expected TaskType
	}{
		{
			name:     "any input returns general",
			input:    "create a button component",
			expected: TaskTypeGeneral,
		},
		{
			name:     "debug keyword returns general (no keyword routing)",
			input:    "fix this bug",
			expected: TaskTypeGeneral,
		},
		{
			name:     "empty input returns general",
			input:    "",
			expected: TaskTypeGeneral,
		},
		{
			name:     "korean input returns general",
			input:    "버그 수정해줘",
			expected: TaskTypeGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.analyzeTask(tt.input)
			if result != tt.expected {
				t.Errorf("analyzeTask(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTaskTypes(t *testing.T) {
	// Verify all task types are defined
	types := []TaskType{
		TaskTypeGeneral,
		TaskTypeUI,
		TaskTypeDesign,
		TaskTypeDebug,
		TaskTypeCode,
	}

	for _, taskType := range types {
		if taskType == "" {
			t.Error("TaskType should not be empty")
		}
	}
}

func TestOrchestratorNew(t *testing.T) {
	// Test with empty config
	o := New(nil)
	if o == nil {
		t.Error("New() should not return nil")
	}
	if o.router == nil {
		t.Error("Orchestrator should have a router")
	}
	if o.agents == nil {
		t.Error("Orchestrator should have agents map initialized")
	}
}

func TestGetCurrentTask(t *testing.T) {
	o := New(nil)

	// Initially should be nil
	if o.GetCurrentTask() != nil {
		t.Error("GetCurrentTask() should return nil initially")
	}
}

func TestGetAgentStatus(t *testing.T) {
	o := New(nil)

	status := o.GetAgentStatus()
	if status == nil {
		t.Error("GetAgentStatus() should not return nil")
	}
}
