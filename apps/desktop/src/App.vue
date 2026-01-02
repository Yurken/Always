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
  const mapping: Record<Mode, string> = {
    SILENT: "静默",
    LIGHT: "轻度",
    ACTIVE: "积极",
  };
  return mapping[currentMode.value];
});

const apiBase = "http://127.0.0.1:8081";
const panelOpen = ref(false);
const settingsOpen = ref(false);

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

const togglePanel = () => {
  panelOpen.value = !panelOpen.value;
  if (!panelOpen.value) {
    settingsOpen.value = false;
  }
};

const openSettings = () => {
  panelOpen.value = true;
  settingsOpen.value = true;
};
</script>

<template>
  <div class="floating-shell">
    <button
      class="orb"
      title="Luma"
      @click="togglePanel"
      @contextmenu.prevent="openSettings"
    >
      L
    </button>

    <div v-if="panelOpen" class="panel">
      <div class="header">
        <div>
          <h1>Luma 陪伴助手</h1>
          <p>当前模式：{{ formattedMode }}</p>
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
        placeholder="描述你当前的状态或任务..."
      />

      <div class="actions">
        <button class="primary" :disabled="loading" @click="requestSuggestion">
          {{ loading ? "请求中..." : "请求建议" }}
        </button>
        <button class="secondary" @click="userText = ''">清空</button>
      </div>

      <div v-if="error" class="result">
        <h3>请求失败</h3>
        <p>{{ error }}</p>
      </div>

      <div v-if="result" class="result">
        <h3>建议卡片</h3>
        <p>{{ result.action.message }}</p>
        <p>
          类型：{{ result.action.action_type }} | 置信度：
          {{ result.action.confidence }} | 风险：{{ result.action.risk_level }}
        </p>
        <div class="feedback">
          <button class="secondary" @click="sendFeedback('LIKE')">赞同</button>
          <button class="secondary" @click="sendFeedback('DISLIKE')">不赞同</button>
        </div>
      </div>

      <div v-if="settingsOpen" class="settings">
        <strong>设置（占位）</strong>
        <p>这里将加入介入频率、敏感度、时间段等设置。</p>
      </div>
    </div>
  </div>
</template>
