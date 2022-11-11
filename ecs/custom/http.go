package ecsx

import "github.com/telkomindonesia/httpmsg-enrichment/ecs"

type HTTPRequest struct {
	ecs.HTTPRequest

	Headers map[string][]string `json:"_headers"`
}
type HTTPResponse struct {
	ecs.HTTPResponse

	Headers map[string][]string `json:"_headers"`
}
type HTTP struct {
	ecs.HTTP

	Request  *HTTPRequest  `json:"request,omitempty"`
	Response *HTTPResponse `json:"response,omitempty"`
}
