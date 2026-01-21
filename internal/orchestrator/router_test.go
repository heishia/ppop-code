package orchestrator

import (
	"testing"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter()

	if r == nil {
		t.Fatal("NewRouter() should not return nil")
	}

	if r.routes == nil {
		t.Error("Router routes should be initialized")
	}
}

func TestRouterAlwaysReturnsSonnet(t *testing.T) {
	r := NewRouter()

	// All task types should route to sonnet (Claude)
	taskTypes := []TaskType{
		TaskTypeGeneral,
		TaskTypeUI,
		TaskTypeDesign,
		TaskTypeDebug,
		TaskTypeCode,
	}

	for _, taskType := range taskTypes {
		result := r.Route(taskType)
		if result != "sonnet" {
			t.Errorf("Route(%v) = %q, want %q", taskType, result, "sonnet")
		}
	}
}

func TestRouterUnknownTaskType(t *testing.T) {
	r := NewRouter()

	// Unknown task type should default to sonnet
	result := r.Route(TaskType("unknown"))
	if result != "sonnet" {
		t.Errorf("Route(unknown) = %q, want %q", result, "sonnet")
	}
}

func TestSetRoute(t *testing.T) {
	r := NewRouter()

	// Test setting a custom route
	r.SetRoute(TaskTypeUI, "custom-agent")

	result := r.Route(TaskTypeUI)
	if result != "custom-agent" {
		t.Errorf("Route(TaskTypeUI) after SetRoute = %q, want %q", result, "custom-agent")
	}
}

func TestGetRoutes(t *testing.T) {
	r := NewRouter()

	routes := r.GetRoutes()
	if routes == nil {
		t.Fatal("GetRoutes() should not return nil")
	}

	// Should have entries for all task types
	expectedTypes := []TaskType{
		TaskTypeUI,
		TaskTypeDesign,
		TaskTypeDebug,
		TaskTypeCode,
		TaskTypeGeneral,
	}

	for _, taskType := range expectedTypes {
		if _, exists := routes[taskType]; !exists {
			t.Errorf("GetRoutes() missing entry for %v", taskType)
		}
	}

	// All routes should be "sonnet"
	for taskType, agent := range routes {
		if agent != "sonnet" {
			t.Errorf("routes[%v] = %q, want %q", taskType, agent, "sonnet")
		}
	}
}

func TestGetRoutesReturnsCopy(t *testing.T) {
	r := NewRouter()

	routes1 := r.GetRoutes()
	routes1[TaskTypeUI] = "modified"

	routes2 := r.GetRoutes()
	if routes2[TaskTypeUI] == "modified" {
		t.Error("GetRoutes() should return a copy, not the original map")
	}
}
