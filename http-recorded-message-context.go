package main

import (
	"net"
	"time"
)

type httpRecordedMessageContext struct {
	ID string `json:"id"`

	Connection httpRecordedMessageContextConnection `json:"connection"`
	Durations  httpRecordedMessageContextDurations  `json:"durations"`
	Credential httpRecordedMessageContextCredential `json:"credential"`
	Domain     httpRecordedMessageContextDomain     `json:"domain"`
}

type httpRecordedMessageContextConnection struct {
	Client   httpRecordedMessageContextConnectionClient `json:"client"`
	Protocol string                                     `json:"protocol"`
}

type httpRecordedMessageContextDurations struct {
	Proxy *httpRecordedMessageContextDuration `json:"proxy"`
	Total httpRecordedMessageContextDuration  `json:"total"`
}
type httpRecordedMessageContextDuration struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end"`
}

type httpRecordedMessageContextCredential struct {
	Username    string `json:"username"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
}

type httpRecordedMessageContextDomain struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type httpRecordedMessageContextConnectionClient struct {
	IP *net.IP `json:"ip"`
}
