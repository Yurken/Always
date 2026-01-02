package models

import "time"

type Mode string

const (
	ModeSilent Mode = "SILENT"
	ModeLight  Mode = "LIGHT"
	ModeActive Mode = "ACTIVE"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"
	RiskMedium RiskLevel = "MEDIUM"
	RiskHigh   RiskLevel = "HIGH"
)

type ActionType string

const (
	ActionDoNotDisturb  ActionType = "DO_NOT_DISTURB"
	ActionEncourage     ActionType = "ENCOURAGE"
	ActionTaskBreakdown ActionType = "TASK_BREAKDOWN"
	ActionRestReminder  ActionType = "REST_REMINDER"
	ActionReframe       ActionType = "REFRAME"
)

type Context struct {
	UserText       string            `json:"user_text"`
	Timestamp      int64             `json:"timestamp"`
	Mode           Mode              `json:"mode"`
	Signals        map[string]string `json:"signals"`
	HistorySummary string            `json:"history_summary"`
}

type Action struct {
	ActionType ActionType `json:"action_type"`
	Message    string     `json:"message"`
	Confidence float64    `json:"confidence"`
	Cost       float64    `json:"cost"`
	RiskLevel  RiskLevel  `json:"risk_level"`
}

type DecisionRequest struct {
	Context Context `json:"context"`
}

type DecisionResponse struct {
	RequestID     string    `json:"request_id"`
	Context       Context   `json:"context"`
	Action        Action    `json:"action"`
	PolicyVersion string    `json:"policy_version"`
	LatencyMs     int64     `json:"latency_ms"`
	CreatedAt     time.Time `json:"created_at"`
}

type FeedbackRequest struct {
	RequestID string `json:"request_id"`
	Feedback  string `json:"feedback"`
}

type LogEntry struct {
	RequestID     string    `json:"request_id"`
	ContextJSON   string    `json:"context_json"`
	ActionJSON    string    `json:"action_json"`
	PolicyVersion string    `json:"policy_version"`
	LatencyMs     int64     `json:"latency_ms"`
	UserFeedback  string    `json:"user_feedback"`
	CreatedAt     time.Time `json:"created_at"`
}
