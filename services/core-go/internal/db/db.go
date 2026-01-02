package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"luma/core/internal/models"
)

const schema = `
CREATE TABLE IF NOT EXISTS event_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  request_id TEXT NOT NULL UNIQUE,
  context_json TEXT NOT NULL,
  action_json TEXT NOT NULL,
  policy_version TEXT NOT NULL,
  latency_ms INTEGER NOT NULL,
  user_feedback TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS feedback_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  request_id TEXT NOT NULL,
  feedback TEXT NOT NULL,
  created_at TEXT NOT NULL
);
`

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) InsertDecision(entry models.DecisionResponse) error {
	ctxJSON, err := json.Marshal(entry.Context)
	if err != nil {
		return fmt.Errorf("marshal context: %w", err)
	}
	actionJSON, err := json.Marshal(entry.Action)
	if err != nil {
		return fmt.Errorf("marshal action: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO event_logs (request_id, context_json, action_json, policy_version, latency_ms, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		entry.RequestID,
		string(ctxJSON),
		string(actionJSON),
		entry.PolicyVersion,
		entry.LatencyMs,
		entry.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert event log: %w", err)
	}
	return nil
}

func (s *Store) RecordFeedback(reqID, feedback string) error {
	_, err := s.db.Exec(
		`UPDATE event_logs SET user_feedback = ? WHERE request_id = ?`,
		feedback,
		reqID,
	)
	if err != nil {
		return fmt.Errorf("update feedback: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO feedback_events (request_id, feedback, created_at) VALUES (?, ?, ?)`,
		reqID,
		feedback,
		time.Now().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert feedback event: %w", err)
	}
	return nil
}

func (s *Store) ListLogs(limit int) ([]models.LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT request_id, context_json, action_json, policy_version, latency_ms, COALESCE(user_feedback, ''), created_at
		 FROM event_logs ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
		var createdAt string
		if err := rows.Scan(
			&entry.RequestID,
			&entry.ContextJSON,
			&entry.ActionJSON,
			&entry.PolicyVersion,
			&entry.LatencyMs,
			&entry.UserFeedback,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan log: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			t = time.Now()
		}
		entry.CreatedAt = t
		logs = append(logs, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return logs, nil
}
