package models

import "time"

// CorrelationResult summarises investigation output.
type CorrelationResult struct {
	CorrelationID    string
	IncidentID       string
	RootCause        string
	Confidence       float64
	AffectedServices []string
	RedAnchors       []RedAnchor
	Timeline         []TimelineEvent
	Recommendations  []string
	CreatedAt        time.Time
}

// RedAnchor highlights a strong anomaly linked to the root cause.
type RedAnchor struct {
	Service      string
	Selector     string
	DataType     DataType
	Timestamp    time.Time
	AnomalyScore float64
	Threshold    float64
}

// TimelineEvent records a notable progression during the incident window.
type TimelineEvent struct {
	Time         time.Time
	Event        string
	Service      string
	Severity     Severity
	AnomalyScore float64
	DataSource   DataType
}

// DataType enumerates signal categories.
type DataType string

const (
	DataTypeMetrics DataType = "metrics"
	DataTypeLogs    DataType = "logs"
	DataTypeTraces  DataType = "traces"
)

// Severity captures impact levels.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)
