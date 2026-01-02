package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"luma/core/internal/models"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

func (c *Client) Decide(ctx models.Context) (models.Action, string, error) {
	payload := map[string]any{"context": ctx}
	body, err := json.Marshal(payload)
	if err != nil {
		return models.Action{}, "", fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := c.http.Post(c.baseURL+"/ai/decide", "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			backoff(attempt)
			continue
		}
		if resp.StatusCode >= 400 {
			resp.Body.Close()
			lastErr = fmt.Errorf("ai status: %s", resp.Status)
			backoff(attempt)
			continue
		}
		var parsed struct {
			Action        models.Action `json:"action"`
			PolicyVersion string        `json:"policy_version"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("decode ai response: %w", err)
			backoff(attempt)
			continue
		}
		resp.Body.Close()
		if err := validateAction(parsed.Action); err != nil {
			return models.Action{}, "", err
		}
		if parsed.PolicyVersion == "" {
			parsed.PolicyVersion = "policy_v0"
		}
		return parsed.Action, parsed.PolicyVersion, nil
	}

	return models.Action{}, "", fmt.Errorf("ai decide failed: %w", lastErr)
}

func (c *Client) Feedback(reqID, feedback string) error {
	payload := map[string]any{"request_id": reqID, "feedback": feedback}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal feedback: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := c.http.Post(c.baseURL+"/ai/feedback", "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			backoff(attempt)
			continue
		}
		if resp.StatusCode >= 400 {
			resp.Body.Close()
			lastErr = fmt.Errorf("ai status: %s", resp.Status)
			backoff(attempt)
			continue
		}
		resp.Body.Close()
		return nil
	}

	return fmt.Errorf("ai feedback failed: %w", lastErr)
}

func validateAction(action models.Action) error {
	validAction := map[models.ActionType]bool{
		models.ActionDoNotDisturb:  true,
		models.ActionEncourage:     true,
		models.ActionTaskBreakdown: true,
		models.ActionRestReminder:  true,
		models.ActionReframe:       true,
	}
	if !validAction[action.ActionType] {
		return fmt.Errorf("invalid action_type: %s", action.ActionType)
	}
	validRisk := map[models.RiskLevel]bool{
		models.RiskLow:    true,
		models.RiskMedium: true,
		models.RiskHigh:   true,
	}
	if !validRisk[action.RiskLevel] {
		return fmt.Errorf("invalid risk_level: %s", action.RiskLevel)
	}
	if action.Confidence < 0 || action.Confidence > 1 {
		return errors.New("confidence out of range")
	}
	if action.Message == "" {
		return errors.New("message is required")
	}
	return nil
}

func backoff(attempt int) {
	base := 200 * time.Millisecond
	wait := time.Duration(attempt+1) * base
	time.Sleep(wait)
}
