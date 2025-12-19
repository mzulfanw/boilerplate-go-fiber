package response

type Status struct {
	Status string `json:"status"`
	App    string `json:"app"`
	Env    string `json:"env"`
	Uptime string `json:"uptime"`
	Time   string `json:"time"`
}
