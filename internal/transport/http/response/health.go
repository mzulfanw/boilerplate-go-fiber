package response

type Status struct {
	Status string `json:"status"`
	App    string `json:"app"`
	Env    string `json:"env"`
	Uptime string `json:"uptime"`
	Time   string `json:"time"`
}

type DependencyStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Latency string `json:"latency,omitempty"`
}

type ReadinessStatus struct {
	Status       string             `json:"status"`
	App          string             `json:"app"`
	Env          string             `json:"env"`
	Uptime       string             `json:"uptime"`
	Time         string             `json:"time"`
	Dependencies []DependencyStatus `json:"dependencies"`
}
