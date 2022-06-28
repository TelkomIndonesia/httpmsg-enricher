package ecs

type URL struct {
	Domain           string `json:"domain,omitempty"`
	Extension        string `json:"extension,omitempty"`
	Fragment         string `json:"fragment,omitempty"`
	Full             string `json:"full,omitempty"`
	Original         string `json:"original,omitempty"`
	Password         string `json:"password,omitempty"`
	Path             string `json:"path,omitempty"`
	Port             uint16 `json:"port,omitempty"`
	Query            string `json:"query,omitempty"`
	RegisteredDomain string `json:"registered_domain,omitempty"`
	Scheme           string `json:"scheme,omitempty"`
	Subdomain        string `json:"subdomain,omitempty"`
	TopLevelDomain   string `json:"top_level_domain,omitempty"`
	Username         string `json:"username,omitempty"`
}
