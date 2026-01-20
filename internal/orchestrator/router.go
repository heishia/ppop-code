package orchestrator

type Router struct {
	routes map[TaskType]string
}

func NewRouter() *Router {
	return &Router{
		routes: map[TaskType]string{
			TaskTypeUI:      "gemini",
			TaskTypeDesign:  "gpt",
			TaskTypeDebug:   "gpt",
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
