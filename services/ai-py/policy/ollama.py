import json
import logging
import os
import requests
from typing import Tuple

from models import Action, Context
from .base import Policy

logger = logging.getLogger("luma-ai")

class OllamaPolicy(Policy):
    name = "ollama_v0"

    def __init__(self):
        self.model = os.getenv("OLLAMA_MODEL", "llama3")
        self.api_url = os.getenv("OLLAMA_URL", "http://localhost:11434/api/generate")

    def decide(self, context: Context) -> Tuple[Action, str, str]:
        prompt = self._build_prompt(context)
        
        try:
            response = requests.post(
                self.api_url,
                json={
                    "model": self.model,
                    "prompt": prompt,
                    "stream": False,
                    "format": "json"
                },
                timeout=10
            )
            response.raise_for_status()
            data = response.json()
            content = data.get("response", "")
            
            action_data = json.loads(content)
            
            action = Action(
                action_type=action_data.get("action_type", "DO_NOT_DISTURB"),
                message=action_data.get("message", "无法生成建议"),
                confidence=float(action_data.get("confidence", 0.5)),
                cost=float(action_data.get("cost", 0.0)),
                risk_level=action_data.get("risk_level", "LOW")
            )
            return action, self.name, self.model
            
        except Exception as e:
            logger.error(f"Ollama call failed: {e}")
            return Action(
                action_type="DO_NOT_DISTURB",
                message="AI 服务暂时不可用",
                confidence=1.0,
                cost=0.0,
                risk_level="LOW"
            ), self.name, "error"

    def _build_prompt(self, context: Context) -> str:
        app_name = context.signals.get("focus_app", "Unknown")
        window_title = context.signals.get("focus_window_title", "")
        focus_minutes = context.signals.get("focus_minutes", "0")
        user_text = context.user_text
        mode = context.mode
        
        profile_section = ""
        if context.profile_summary:
            profile_section = f"\nUser Profile (Preferences & Traits):\n{context.profile_summary}\n"
            
        memory_section = ""
        if context.memory_summary:
            memory_section = f"\nRecent Memory Events:\n{context.memory_summary}\n"

        return f"""
You are Luma, an intelligent desktop companion.
Your goal is to help the user stay focused, healthy, and productive.
{profile_section}{memory_section}
Current Context:
- Mode: {mode} (SILENT: minimize disturbance, LIGHT: gentle reminders, ACTIVE: proactive)
- Current App: {app_name}
- Window Title: {window_title}
- Focus Duration: {focus_minutes} minutes
- User Input: "{user_text}" (If empty, infer from screen context)

Task:
Analyze the context and decide on the best action.
If the user is working on a productive task for a long time, maybe suggest a break.
If the user is distracted (e.g. social media) during work hours, suggest getting back to work.
If the user asks for help, provide it.
Consider the User Profile and Recent Memories to personalize your advice.

Output Format (JSON only):
{{
  "action_type": "DO_NOT_DISTURB" | "ENCOURAGE" | "TASK_BREAKDOWN" | "REST_REMINDER" | "REFRAME",
  "message": "A short, friendly message to the user (in Chinese)",
  "confidence": 0.0 to 1.0,
  "cost": 0.0 to 1.0 (interruption cost),
  "risk_level": "LOW" | "MEDIUM" | "HIGH"
}}
"""
