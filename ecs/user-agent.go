package ecs

type UserAgent struct {
	Name     string           `json:"name,omitempty"`
	Original string           `json:"original,omitempty"`
	Version  string           `json:"version,omitempty"`
	Device   *UserAgentDevice `json:"device,omitempty"`

	OS *OS `json:"os,omitempty"`
}

type UserAgentDevice struct {
	Name string `json:"name,omitempty"`
}
