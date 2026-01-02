from abc import ABC, abstractmethod
from typing import Tuple

from models import Action, Context


class Policy(ABC):
    name = "base"

    @abstractmethod
    def decide(self, context: Context) -> Tuple[Action, str, str]:
        raise NotImplementedError
