package gateway

import "luma/core/internal/models"

type GatewayDecision = models.GatewayDecision

type OverrideResult struct {
	OriginalAction  models.Action
	FinalAction     models.Action
	GatewayDecision models.GatewayDecision
}
