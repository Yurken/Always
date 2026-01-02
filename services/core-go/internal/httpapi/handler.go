package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"luma/core/internal/ai"
	"luma/core/internal/db"
	"luma/core/internal/models"
)

type Handler struct {
	store  *db.Store
	ai     *ai.Client
	logger zerolog.Logger
}

func NewHandler(store *db.Store, aiClient *ai.Client, logger zerolog.Logger) *Handler {
	return &Handler{store: store, ai: aiClient, logger: logger}
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Post("/v1/decision", h.handleDecision)
	r.Post("/v1/feedback", h.handleFeedback)
	r.Get("/v1/logs", h.handleLogs)
	return r
}

func (h *Handler) handleDecision(w http.ResponseWriter, r *http.Request) {
	var req models.DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateContext(req.Context); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Context.Timestamp == 0 {
		req.Context.Timestamp = time.Now().UnixMilli()
	}
	if req.Context.Signals == nil {
		req.Context.Signals = map[string]string{}
	}

	start := time.Now()
	action, policyVersion, err := h.ai.Decide(req.Context)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		h.logger.Error().Err(err).Msg("ai decide failed")
		respondError(w, http.StatusBadGateway, "ai service unavailable")
		return
	}

	if action.RiskLevel == models.RiskHigh {
		action = models.Action{
			ActionType: models.ActionDoNotDisturb,
			Message:    "高风险动作已被权限网关拦截。",
			Confidence: 1,
			Cost:       0,
			RiskLevel:  models.RiskLow,
		}
		policyVersion = policyVersion + ":gateway_block"
	}

	resp := models.DecisionResponse{
		RequestID:     uuid.NewString(),
		Context:       req.Context,
		Action:        action,
		PolicyVersion: policyVersion,
		LatencyMs:     latency,
		CreatedAt:     time.Now(),
	}
	if err := h.store.InsertDecision(resp); err != nil {
		h.logger.Error().Err(err).Msg("insert decision failed")
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleFeedback(w http.ResponseWriter, r *http.Request) {
	var req models.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.RequestID == "" || req.Feedback == "" {
		respondError(w, http.StatusBadRequest, "request_id and feedback required")
		return
	}

	if err := h.store.RecordFeedback(req.RequestID, req.Feedback); err != nil {
		h.logger.Error().Err(err).Msg("record feedback failed")
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if err := h.ai.Feedback(req.RequestID, req.Feedback); err != nil {
		h.logger.Error().Err(err).Msg("forward feedback failed")
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil {
			limit = parsed
		}
	}
	logs, err := h.store.ListLogs(limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("list logs failed")
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parseInt(val string) (int, error) {
	var n int
	_, err := fmt.Sscanf(val, "%d", &n)
	return n, err
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validateContext(ctx models.Context) error {
	validModes := map[models.Mode]bool{
		models.ModeSilent: true,
		models.ModeLight:  true,
		models.ModeActive: true,
	}
	if ctx.UserText == "" {
		return fmt.Errorf("user_text required")
	}
	if !validModes[ctx.Mode] {
		return fmt.Errorf("invalid mode")
	}
	return nil
}
