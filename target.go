package holler

// Target type abstracts a backend destination
type Target struct {
	URL         string `json:"url"`
	Healthy     bool   `json:"health,omitempty"`
	HealthRoute string `json:"health_route,omitempty"`
}
