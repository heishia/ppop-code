package orchestrator

type Router struct {
	routes map[TaskType]string
}

func NewRouter() *Router {
	// All requests go to Claude (sonnet) as the orchestrator
	// Claude maintains conversation context and decides when to delegate
	return &Router{
		routes: map[TaskType]string{
			TaskTypeUI:      "sonnet",
			TaskTypeDesign:  "sonnet",
			TaskTypeDebug:   "sonnet",
			TaskTypeCode:    "sonnet",
			TaskTypeGeneral: "sonnet",
		},
	}
}

func (r *Router) Route(taskType TaskType) string {
	if agent, exists := r.routes[taskType]; exists {
		return agent
	}
	return "sonnet"
}

func (r *Router) SetRoute(taskType TaskType, agentName string) {
	r.routes[taskType] = agentName
}

func (r *Router) GetRoutes() map[TaskType]string {
	result := make(map[TaskType]string)
	for k, v := range r.routes {
		result[k] = v
	}
	return result
}
