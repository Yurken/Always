import json
import logging
import os
import random
import time
from typing import Tuple

from fastapi import FastAPI
from fastapi.responses import JSONResponse

from models import (
    Action,
    ActionType,
    Context,
    DecideRequest,
    DecideResponse,
    FeedbackRequest,
    Mode,
    RiskLevel,
)

app = FastAPI(title="Luma AI Service")

DATA_DIR = os.path.join(os.path.dirname(__file__), "data")
LOG_PATH = os.path.join(DATA_DIR, "feedback.log")

os.makedirs(DATA_DIR, exist_ok=True)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("luma-ai")


def choose_action(context: Context) -> Tuple[Action, str]:
    user_text = context.user_text.lower()
    hour = time.localtime(context.timestamp / 1000).tm_hour

    if context.mode == Mode.SILENT and random.random() < 0.8:
        return (
            Action(
                action_type=ActionType.DO_NOT_DISTURB,
                message="保持安静。如果需要帮助，随时告诉我。",
                confidence=0.82,
                cost=0.1,
                risk_level=RiskLevel.LOW,
            ),
            "policy_v0",
        )

    if 23 <= hour or hour <= 5:
        action_type = random.choice([ActionType.REST_REMINDER, ActionType.ENCOURAGE])
        message = (
            "时间有点晚了，建议做个短暂休息或伸展一下。"
            if action_type == ActionType.REST_REMINDER
            else "深夜工作很辛苦，你已经做得很好了。"
        )
        return (
            Action(
                action_type=action_type,
                message=message,
                confidence=0.7,
                cost=0.2,
                risk_level=RiskLevel.LOW,
            ),
            "policy_v0",
        )

    keywords = ["赶", "截止", "来不及", "压力", "deadline", "stress"]
    if any(k in user_text for k in keywords):
        action_type = random.choice([ActionType.TASK_BREAKDOWN, ActionType.REFRAME])
        message = (
            "可以先列出接下来三个最小的步骤，降低压力。"
            if action_type == ActionType.TASK_BREAKDOWN
            else "这件事很重，但你以前也扛过困难。"
        )
        return (
            Action(
                action_type=action_type,
                message=message,
                confidence=0.78,
                cost=0.3,
                risk_level=RiskLevel.LOW,
            ),
            "policy_v0",
        )

    action_type = random.choice(
        [ActionType.ENCOURAGE, ActionType.DO_NOT_DISTURB, ActionType.REST_REMINDER]
    )
    message_map = {
        ActionType.ENCOURAGE: "继续加油，小进步也很重要。",
        ActionType.DO_NOT_DISTURB: "暂时不介入，需要时我会在。",
        ActionType.REST_REMINDER: "可以短暂休息一下，帮助恢复专注。",
    }
    return (
        Action(
            action_type=action_type,
            message=message_map[action_type],
            confidence=round(random.uniform(0.55, 0.85), 2),
            cost=0.2,
            risk_level=RiskLevel.LOW,
        ),
        "policy_v0",
    )


@app.post("/ai/decide", response_model=DecideResponse)
async def decide(payload: DecideRequest) -> DecideResponse:
    action, policy_version = choose_action(payload.context)
    return DecideResponse(action=action, policy_version=policy_version)


@app.post("/ai/feedback")
async def feedback(payload: FeedbackRequest) -> JSONResponse:
    entry = {
        "request_id": payload.request_id,
        "feedback": payload.feedback,
        "timestamp": int(time.time() * 1000),
    }
    logger.info("feedback: %s", json.dumps(entry, ensure_ascii=True))
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(json.dumps(entry, ensure_ascii=True) + "\n")
    return JSONResponse({"status": "ok"})
