package main

import (
	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/types/variables"
)

type baseScores struct {
	CriticalAnomalyScore int `json:"critical_anomaly_score"`
	ErrorAnomalyScore    int `json:"error_anomaly_score"`
	WarningAnomalyScore  int `json:"warning_anomaly_score"`
	NoticeAnomalyScore   int `json:"notice_anomaly_score"`
}

type scores struct {
	Base baseScores `json:"base_scores"`

	BlockingInboundAnomalyScore   int `json:"blocking_inbound_anomaly_score"`
	DetectionInboundAnomalyScore  int `json:"detection_inbound_anomaly_score"`
	BlockingOutboundAnomalyScore  int `json:"blocking_outbound_anomaly_score"`
	DetectionOutboundAnomalyScore int `json:"detection_outbound_anomaly_score"`
	InboundAnomalyScorePL1        int `json:"inbound_anomaly_score_pl1"`
	InboundAnomalyScorePL2        int `json:"inbound_anomaly_score_pl2"`
	InboundAnomalyScorePL3        int `json:"inbound_anomaly_score_pl3"`
	InboundAnomalyScorePL4        int `json:"inbound_anomaly_score_pl4"`
	OutboundAnomalyScorePL1       int `json:"outbound_anomaly_score_pl1"`
	OutboundAnomalyScorePL2       int `json:"outbound_anomaly_score_pl2"`
	OutboundAnomalyScorePL3       int `json:"outbound_anomaly_score_pl3"`
	OutboundAnomalyScorePL4       int `json:"outbound_anomaly_score_pl4"`

	SqlInjectionScore    int `json:"sql_injection_score"`
	XssScore             int `json:"xss_score"`
	RfiScore             int `json:"rfi_score"`
	LfiScore             int `json:"lfi_score"`
	RceScore             int `json:"rce_score"`
	PhpInjectionScore    int `json:"php_injection_score"`
	HttpViolationScore   int `json:"http_violation_score"`
	SessionFixationScore int `json:"session_fixation_score"`

	SqlErrorMatch int `json:"sql_error_match"`
}

func newScores(tx *coraza.Transaction) (s *scores) {
	txData := tx.GetCollection(variables.TX)
	s = &scores{
		Base: baseScores{
			CriticalAnomalyScore: txData.GetFirstInt("critical_anomaly_score"),
			ErrorAnomalyScore:    txData.GetFirstInt("error_anomaly_score"),
			WarningAnomalyScore:  txData.GetFirstInt("warning_anomaly_score"),
			NoticeAnomalyScore:   txData.GetFirstInt("notice_anomaly_score"),
		},

		BlockingInboundAnomalyScore:  txData.GetFirstInt("blocking_inbound_anomaly_score"),
		DetectionInboundAnomalyScore: txData.GetFirstInt("detection_inbound_anomaly_score"),
		InboundAnomalyScorePL1:       txData.GetFirstInt("inbound_anomaly_score_pl1"),
		InboundAnomalyScorePL2:       txData.GetFirstInt("inbound_anomaly_score_pl2"),
		InboundAnomalyScorePL3:       txData.GetFirstInt("inbound_anomaly_score_pl3"),
		InboundAnomalyScorePL4:       txData.GetFirstInt("inbound_anomaly_score_pl4"),
		OutboundAnomalyScorePL1:      txData.GetFirstInt("outbound_anomaly_score_pl1"),
		OutboundAnomalyScorePL2:      txData.GetFirstInt("outbound_anomaly_score_pl2"),
		OutboundAnomalyScorePL3:      txData.GetFirstInt("outbound_anomaly_score_pl3"),
		OutboundAnomalyScorePL4:      txData.GetFirstInt("outbound_anomaly_score_pl4"),

		SqlInjectionScore:             txData.GetFirstInt("sql_injection_score"),
		XssScore:                      txData.GetFirstInt("xss_score"),
		RfiScore:                      txData.GetFirstInt("rfi_score"),
		LfiScore:                      txData.GetFirstInt("lfi_score"),
		RceScore:                      txData.GetFirstInt("rce_score"),
		PhpInjectionScore:             txData.GetFirstInt("php_injection_score"),
		HttpViolationScore:            txData.GetFirstInt("http_violation_score"),
		SessionFixationScore:          txData.GetFirstInt("session_fixation_score"),
		BlockingOutboundAnomalyScore:  txData.GetFirstInt("blocking_outbound_anomaly_score"),
		DetectionOutboundAnomalyScore: txData.GetFirstInt("detection_outbound_anomaly_score"),
		SqlErrorMatch:                 txData.GetFirstInt("sql_error_match"),
	}
	return
}
