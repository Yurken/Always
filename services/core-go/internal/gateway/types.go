package gateway

import "always/core/internal/models"

type GatewayDecision = models.GatewayDecision

type OverrideResult struct {
	OriginalAction  models.Action
	FinalAction     models.Action
	GatewayDecision models.GatewayDecision
}
