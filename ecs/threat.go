package ecs

import (
	"net"
	"time"
)

type ThreatIndicatorEmail struct {
	Address string `json:"address,omitempty"`
}
type ThreatIndicatorMarking struct {
	TLP string `json:"tlp,omitempty"`
}
type ThreatIndicator struct {
	Confidence   string                  `json:"confidence,omitempty"`
	Description  string                  `json:"description,omitempty"`
	Email        *ThreatIndicatorEmail   `json:"email,omitempty"`
	FirstSeen    *time.Time              `json:"first_seen,omitempty"`
	IP           net.IP                  `json:"ip,omitempty"`
	LastSeen     *time.Time              `json:"last_seen,omitempty"`
	Marking      *ThreatIndicatorMarking `json:"marking.tlp,omitempty"`
	ModifiedAt   *time.Time              `json:"modified_at,omitempty"`
	Port         uint16                  `json:"port,omitempty"`
	Provider     string                  `json:"provider,omitempty"`
	Reference    string                  `json:"reference,omitempty"`
	ScannerStats int                     `json:"scanner_stats,omitempty"`
	Sightings    int                     `json:"sightings,omitempty"`
	Type         string                  `json:"type,omitempty"`

	Geo *Geo `json:"geo,omitempty"`
	URL *URL `json:"url,omitempty"`
}

type ThreatEnrichmentMatch struct {
	Type     string     `json:"type,omitempty"`
	Occurred *time.Time `json:"occurred,omitempty"`
	Index    string     `json:"index,omitempty"`
	Id       string     `json:"id,omitempty"`
	Field    string     `json:"field,omitempty"`
	Atomic   string     `json:"atomic,omitempty"`
}

type ThreatEnrichments struct {
	Indicator ThreatIndicator        `json:"indicator,omitempty"`
	Match     *ThreatEnrichmentMatch `json:"match,omitempty"`
}

type Threat struct {
	Enrichments []ThreatEnrichments `json:"enrichments,omitempty"`
}
