package main

import (
	"net"
	"time"
)

// TODO: This is not generic enough since it is currently only used for httpsig-proxy log
type httpRecordedMessageContext struct {
	ID string `json:"id,omitempty"`

	Connection *httpRecordedMessageContextConnection `json:"connection,omitempty"`
	Durations  *httpRecordedMessageContextDurations  `json:"durations,omitempty"`
	User       *httpRecordedMessageContextUser       `json:"user,omitempty"`
	Host       *httpRecordedMessageContextHost       `json:"host,omitempty"`
}

type httpRecordedMessageContextConnection struct {
	Client   httpRecordedMessageContextConnectionClient `json:"client,omitempty"`
	Protocol string                                     `json:"protocol,omitempty"`
}

type httpRecordedMessageContextDurations struct {
	Proxy *httpRecordedMessageContextDuration `json:"proxy,omitempty"`
	Total httpRecordedMessageContextDuration  `json:"total,omitempty"`
}
type httpRecordedMessageContextDuration struct {
	Start time.Time  `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

type httpRecordedMessageContextUser struct {
	Username    string  `json:"username,omitempty"`
	PublicKey   *string `json:"public_key,omitempty"`
	Fingerprint *string `json:"public_key_fingerprint,omitempty"`
}

type httpRecordedMessageContextHost struct {
	Name   string  `json:"name,omitempty"`
	Target *string `json:"target,omitempty"`
}

type httpRecordedMessageContextConnectionClient struct {
	IP net.IP `json:"ip,omitempty"`
}
