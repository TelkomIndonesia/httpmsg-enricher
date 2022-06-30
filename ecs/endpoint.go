package ecs

import "net"

type EndpointNAT struct {
	IP   net.IP `json:"ip,omitempty"`
	Port uint16 `json:"port,omitempty"`
}

type Endpoint struct {
	Address          string       `json:"address,omitempty"`
	Bytes            int64        `json:"bytes,omitempty"`
	Domain           string       `json:"domain,omitempty"`
	IP               net.IP       `json:"ip,omitempty"`
	NAT              *EndpointNAT `json:"nat,omitempty"`
	Mac              string       `json:"mac,omitempty"`
	Packets          int64        `json:"packets,omitempty"`
	Port             uint16       `json:"port,omitempty"`
	RegisteredDomain string       `json:"registered_domain,omitempty"`
	Subdomain        string       `json:"subdomain,omitempty"`
	TopLevelDomain   string       `json:"top_level_domain,omitempty"`

	Geo  Geo  `json:"geo,omitempty"`
	User User `json:"user,omitempty"`
}
