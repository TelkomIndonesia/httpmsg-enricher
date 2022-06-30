package ecs

import "time"

type Event struct {
	Action        string         `json:"action,omitempty"`
	AgentIDStatus string         `json:"agent_id_status,omitempty"`
	Category      []string       `json:"category,omitempty"`
	Code          string         `json:"code,omitempty"`
	Created       *time.Time     `json:"created,omitempty"`
	Dataset       string         `json:"dataset,omitempty"`
	Duration      *time.Duration `json:"duration,omitempty"`
	End           *time.Time     `json:"end,omitempty"`
	Hash          string         `json:"hash,omitempty"`
	Id            string         `json:"id,omitempty"`
	Ingested      *time.Time     `json:"ingested,omitempty"`
	Kind          string         `json:"kind,omitempty"`
	Module        string         `json:"module,omitempty"`
	Original      string         `json:"original,omitempty"`
	Outcome       string         `json:"outcome,omitempty"`
	Provider      string         `json:"provider,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	Reference     string         `json:"reference,omitempty"`
	RiskScore     float32        `json:"risk_score,omitempty"`
	RiskScoreNorm float32        `json:"risk_score_norm,omitempty"`
	Sequence      int            `json:"sequence,omitempty"`
	Severity      int            `json:"severity,omitempty"`
	Start         *time.Time     `json:"start,omitempty"`
	Timezone      string         `json:"timezone,omitempty"`
	Type          []string       `json:"type,omitempty"`
	Url           string         `json:"url,omitempty"`
}
