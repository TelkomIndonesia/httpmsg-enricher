package main

import (
	"net"
	"time"
)

type httpRecordedMessageContext struct {
	ID string `json:"id,omitempty"`

	Connection *httpRecordedMessageContextConnection `json:"connection,omitempty"`
	Durations  *httpRecordedMessageContextDurations  `json:"durations,omitempty"`
	Credential *httpRecordedMessageContextCredential `json:"credential,omitempty"`
	Domain     *httpRecordedMessageContextDomain     `json:"domain,omitempty"`
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

type httpRecordedMessageContextCredential struct {
	Username    string  `json:"username,omitempty"`
	PublicKey   *string `json:"public_key,omitempty"`
	Fingerprint *string `json:"fingerprint,omitempty"`
}

type httpRecordedMessageContextDomain struct {
	Name   string  `json:"name,omitempty"`
	Target *string `json:"target,omitempty"`
}

type httpRecordedMessageContextConnectionClient struct {
	IP net.IP `json:"ip,omitempty"`
}
