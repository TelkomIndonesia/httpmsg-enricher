package ecs

import "time"

type Base struct {
	Timestamp time.Time         `json:"@timestamp,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Message   string            `json:"message,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
}
