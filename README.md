# Luma MVP (Local Silent Companion Agent)

Luma is a local-first, desktop companion agent that decides when to gently intervene. This MVP focuses on clean architecture, data closure, and safety gates.

## Architecture (MVP)

```
+--------------------+        HTTP        +---------------------+
| Electron + Vue UI  | <----------------> | Go Core Service     |
| apps/desktop       |                    | services/core-go    |
+--------------------+                    +----------+----------+
                                                      |
                                                      | HTTP (strict validation + retry)
                                                      v
                                            +---------------------+
                                            | Python AI Service   |
                                            | services/ai-py      |
                                            +---------------------+
                                                      |
                                                      v
                                               SQLite (local)
```

## Goals
- Local-only execution, no cloud dependency.
- AI outputs only Action; system operations are blocked by a permission gateway.
- All decisions/feedback are logged and auditable in SQLite.
- Policy versioning is explicit for rollback.

## Services & Ports
- Desktop UI: Vite dev server `http://localhost:5173`
- Core Go API: `http://127.0.0.1:8081`
- AI Service: `http://127.0.0.1:8788`

## Quick Start

### 1) Start all services
```
./scripts/dev.sh
```

### 2) Open the desktop UI
Electron launches automatically via `npm run dev`.

## API

### POST /v1/decision
Request:
```json
{
  "context": {
    "user_text": "I am rushing a paper and feel stressed",
    "timestamp": 1710000000000,
    "mode": "LIGHT",
    "signals": {
      "hour_of_day": "21",
      "session_minutes": "40"
    },
    "history_summary": ""
  }
}
```

Response (example):
```json
{
  "request_id": "b0f2c78e-1aa5-4d4c-9c77-3d7b41b3e8bd",
  "context": {
    "user_text": "I am rushing a paper and feel stressed",
    "timestamp": 1710000000000,
    "mode": "LIGHT",
    "signals": {
      "hour_of_day": "21",
      "session_minutes": "40"
    },
    "history_summary": ""
  },
  "action": {
    "action_type": "TASK_BREAKDOWN",
    "message": "Try listing the next three smallest steps to reduce pressure.",
    "confidence": 0.78,
    "cost": 0.3,
    "risk_level": "LOW"
  },
  "policy_version": "policy_v0",
  "latency_ms": 34,
  "created_at": "2024-05-10T12:00:00Z"
}
```

### POST /v1/feedback
```json
{
  "request_id": "b0f2c78e-1aa5-4d4c-9c77-3d7b41b3e8bd",
  "feedback": "LIKE"
}
```

### GET /v1/logs?limit=50
Returns the latest decision logs.

## SQLite Logs
- DB path: `./data/luma.db`
- Tables: `event_logs`, `feedback_events`
- Example query:
```
sqlite3 ./data/luma.db "select request_id, policy_version, user_feedback, created_at from event_logs order by id desc limit 5;"
```

## Safety & Extensibility
- AI service only outputs Action; it never executes system operations.
- Any HIGH risk action is blocked by the Go permission gateway.
- `policy_version` enables rollback and A/B experiments.
- The `policy/` module in `services/ai-py` is reserved for future contextual bandit and preference learning.

## Project Structure
```
apps/desktop        Electron + Vue
services/core-go    Go HTTP API + SQLite
services/ai-py      FastAPI AI service
proto               gRPC definitions (future use)
scripts             Dev scripts
```

## gRPC Protocol
The file `proto/luma.proto` defines Context/Action/Feedback for future gRPC communication. Current MVP uses HTTP with strict validation and retry in the Go client.
