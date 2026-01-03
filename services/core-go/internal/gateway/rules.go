package gateway

import "always/core/internal/models"

const (
	ReasonInvalidActionType  = "invalid_action_type"
	ReasonInvalidRiskLevel   = "invalid_risk_level"
	ReasonInvalidConfidence  = "invalid_confidence"
	ReasonModeSilentOverride = "mode_silent_override"
	ReasonLowQualityAction   = "low_quality_action"
	ReasonHighRiskBlocked    = "high_risk_blocked"
)

func ruleInvalidAction(action models.Action) (string, bool) {
	if !isValidActionType(action.ActionType) {
		return ReasonInvalidActionType, true
	}
	if !isValidRiskLevel(action.RiskLevel) {
		return ReasonInvalidRiskLevel, true
	}
	if action.Confidence < 0 || action.Confidence > 1 {
		return ReasonInvalidConfidence, true
	}
	return "", false
}

func ruleHighRisk(action models.Action) bool {
	return action.RiskLevel == models.RiskHigh
}

func ruleLowQuality(action models.Action) bool {
	return action.Message == "" || action.Confidence < 0.5
}

func ruleSilentOverride(ctx models.Context, action models.Action) bool {
	return ctx.Mode == models.ModeSilent && action.ActionType != models.ActionDoNotDisturb
}

func isValidActionType(actionType models.ActionType) bool {
	switch actionType {
	case models.ActionDoNotDisturb,
		models.ActionEncourage,
		models.ActionTaskBreakdown,
		models.ActionRestReminder,
		models.ActionReframe:
		return true
	default:
		return false
	}
}

func isValidRiskLevel(level models.RiskLevel) bool {
	switch level {
	case models.RiskLow, models.RiskMedium, models.RiskHigh:
		return true
	default:
		return false
	}
}
