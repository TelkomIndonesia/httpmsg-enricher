package ecs

type User struct {
	Domain   string   `json:"domain,omitempty"`
	Email    string   `json:"email,omitempty"`
	FullName string   `json:"full_name,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}
