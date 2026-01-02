package ai

import (
	"bytes"
	"encoding/json"
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
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Decide(ctx models.Context, requestID string) (models.Action, string, string, error) {
	payload := map[string]any{"context": ctx}
	if requestID != "" {
		payload["request_id"] = requestID
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return models.Action{}, "", "", fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+"/ai/decide", bytes.NewReader(body))
		if err != nil {
			return models.Action{}, "", "", fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}
		resp, err := c.http.Do(req)
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
			ModelVersion  string        `json:"model_version"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("decode ai response: %w", err)
			backoff(attempt)
			continue
		}
		resp.Body.Close()
		if parsed.PolicyVersion == "" {
			parsed.PolicyVersion = "policy_v0"
		}
		if parsed.ModelVersion == "" {
			parsed.ModelVersion = "stub"
		}
		return parsed.Action, parsed.PolicyVersion, parsed.ModelVersion, nil
	}

	return models.Action{}, "", "", fmt.Errorf("ai decide failed: %w", lastErr)
}

func (c *Client) Feedback(reqID, feedback string) error {
	payload := map[string]any{"request_id": reqID, "feedback": feedback}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal feedback: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+"/ai/feedback", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create feedback request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if reqID != "" {
			req.Header.Set("X-Request-ID", reqID)
		}
		resp, err := c.http.Do(req)
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

func backoff(attempt int) {
	base := 200 * time.Millisecond
	wait := time.Duration(attempt+1) * base
	time.Sleep(wait)
}
