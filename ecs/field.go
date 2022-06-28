package ecs

type Document struct {
	Base

	ECS ECS `json:"ecs,omitempty"`

	Event Event `json:"event,omitempty"`

	Client      Endpoint `json:"client,omitempty"`
	Server      Endpoint `json:"server,omitempty"`
	Source      Endpoint `json:"source,omitempty"`
	Destination Endpoint `json:"destination,omitempty"`

	HTTP HTTP `json:"http,omitempty"`
	URL  URL  `json:"url,omitempty"`

	Threat Threat `json:"threat,omitempty"`
}
