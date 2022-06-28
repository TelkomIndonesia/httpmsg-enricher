package ecs

type HTTPMessageBody struct {
	Bytes   int64  `json:"bytes,omitempty"`
	Content string `json:"content,omitempty"`
}
type HTTPMessage struct {
	Bytes    int64           `json:"bytes,omitempty"`
	MimeType string          `json:"mime_type,omitempty"`
	Body     HTTPMessageBody `json:"body,omitempty"`
}
type HTTPRequest struct {
	HTTPMessage

	ID       string `json:"id,omitempty"`
	Method   string `json:"method,omitempty"`
	Referrer string `json:"referrer,omitempty"`
}
type HTTPResponse struct {
	HTTPMessage

	StatusCode int `json:"status_code,omitmepty"`
}
type HTTP struct {
	Version  string       `json:"version,omitempty"`
	Request  HTTPRequest  `json:"request,omitempty"`
	Response HTTPResponse `json:"response,omitempty"`
}
