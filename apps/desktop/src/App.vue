<script lang="ts" setup>
import { computed, ref } from "vue";

type Mode = "SILENT" | "LIGHT" | "ACTIVE";

type Action = {
  action_type: string;
  message: string;
  confidence: number;
  cost: number;
  risk_level: string;
};

type DecisionResponse = {
  request_id: string;
  context: {
    user_text: string;
    timestamp: number;
    mode: Mode;
    signals: Record<string, string>;
    history_summary: string;
  };
  action: Action;
  policy_version: string;
  latency_ms: number;
  created_at: string;
};

const modes: Mode[] = ["SILENT", "LIGHT", "ACTIVE"];
const currentMode = ref<Mode>("LIGHT");
const userText = ref("");
const result = ref<DecisionResponse | null>(null);
const loading = ref(false);
const error = ref("");

const formattedMode = computed(() => {
  return currentMode.value.charAt(0) + currentMode.value.slice(1).toLowerCase();
});

const apiBase = "http://127.0.0.1:8081";

const requestSuggestion = async () => {
  error.value = "";
  loading.value = true;
  const payload = {
    context: {
      user_text: userText.value,
      timestamp: Date.now(),
      mode: currentMode.value,
      signals: {
        hour_of_day: new Date().getHours().toString(),
        session_minutes: "0",
      },
      history_summary: "",
    },
  };

  try {
    const res = await fetch(`${apiBase}/v1/decision`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      throw new Error("Decision request failed");
    }
    result.value = (await res.json()) as DecisionResponse;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Unknown error";
  } finally {
    loading.value = false;
  }
};

const sendFeedback = async (feedback: "LIKE" | "DISLIKE") => {
  if (!result.value) {
    return;
  }
  await fetch(`${apiBase}/v1/feedback`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      request_id: result.value.request_id,
      feedback,
    }),
  });
};
</script>

<template>
  <div class="card">
    <div class="header">
      <div>
        <h1>Luma Companion</h1>
        <p>Current mode: {{ formattedMode }}</p>
      </div>
      <div class="mode">
        <button
          v-for="mode in modes"
          :key="mode"
          :class="{ active: mode === currentMode }"
          @click="currentMode = mode"
        >
          {{ mode }}
        </button>
      </div>
    </div>

    <textarea
      v-model="userText"
      placeholder="Describe what you are working on..."
    />

    <div class="actions">
      <button class="primary" :disabled="loading" @click="requestSuggestion">
        {{ loading ? "Requesting..." : "Request suggestion" }}
      </button>
      <button class="secondary" @click="userText = ''">Clear</button>
    </div>

    <div v-if="error" class="result">
      <h3>Request error</h3>
      <p>{{ error }}</p>
    </div>

    <div v-if="result" class="result">
      <h3>Suggestion</h3>
      <p>{{ result.action.message }}</p>
      <p>
        Type: {{ result.action.action_type }} | Confidence:
        {{ result.action.confidence }} | Risk: {{ result.action.risk_level }}
      </p>
      <div class="feedback">
        <button class="secondary" @click="sendFeedback('LIKE')">Like</button>
        <button class="secondary" @click="sendFeedback('DISLIKE')">Dislike</button>
      </div>
    </div>

    <div class="settings">
      <strong>Settings (placeholder)</strong>
      <p>Intervention frequency, sensitivity, and schedule will appear here.</p>
    </div>
  </div>
</template>
