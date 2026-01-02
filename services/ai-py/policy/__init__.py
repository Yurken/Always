from policy.base import Policy
from policy.ollama import OllamaPolicy
from policy.unavailable import UnavailablePolicy

_POLICIES = {
    "ollama": OllamaPolicy(),
}


def get_policy(name: str) -> Policy:
    key = (name or "").strip().lower()
    if not key:
        return _POLICIES["ollama"]
    return _POLICIES.get(key, UnavailablePolicy())
