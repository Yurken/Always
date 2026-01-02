<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount } from "vue";

const props = defineProps<{
  mode: "SILENT" | "LIGHT" | "ACTIVE";
  loading: boolean;
}>();

const emit = defineEmits<{
  (e: "click"): void;
  (e: "drag-end", x: number, y: number): void;
}>();

const orbRef = ref<HTMLElement | null>(null);
const dragging = ref(false);
const dragMoved = ref(false);
const dragStart = ref({ x: 0, y: 0 });

const handleMouseDown = (e: MouseEvent) => {
  if (e.button !== 0) return;
  dragging.value = true;
  dragMoved.value = false;
  dragStart.value = { x: e.screenX, y: e.screenY };
  
  // Tell main process to start moving window if needed, 
  // but here we just track delta for click vs drag distinction
  window.addEventListener("mousemove", handleMouseMove);
  window.addEventListener("mouseup", handleMouseUp);
};

const handleMouseMove = (e: MouseEvent) => {
  if (!dragging.value) return;
  const dx = Math.abs(e.screenX - dragStart.value.x);
  const dy = Math.abs(e.screenY - dragStart.value.y);
  if (dx > 5 || dy > 5) {
    dragMoved.value = true;
    // In a real Electron app with frameless window, 
    // we usually use -webkit-app-region: drag or IPC to move window.
    // Assuming parent handles window movement via IPC based on this element's drag.
    // For now, we just track state.
  }
};

const handleMouseUp = () => {
  dragging.value = false;
  window.removeEventListener("mousemove", handleMouseMove);
  window.removeEventListener("mouseup", handleMouseUp);
  
  if (!dragMoved.value) {
    emit("click");
  }
};
</script>

<template>
  <div
    ref="orbRef"
    class="orb"
    :class="{
      'orb-silent': mode === 'SILENT',
      'orb-light': mode === 'LIGHT',
      'orb-active': mode === 'ACTIVE',
      'orb-loading': loading,
    }"
    @mousedown="handleMouseDown"
  >
    <div class="orb-inner"></div>
    <div class="orb-ring"></div>
  </div>
</template>

<style scoped>
.orb {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  position: relative;
  cursor: grab;
  transition: transform 0.2s, filter 0.3s;
  /* Electron drag region handled by parent or CSS class if fixed */
  -webkit-app-region: drag; 
}

.orb:active {
  cursor: grabbing;
  transform: scale(0.95);
}

.orb-inner {
  position: absolute;
  inset: 4px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, #ffffff, #a0a0a0);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 2;
}

.orb-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 2px solid transparent;
  z-index: 1;
  animation: spin 10s linear infinite;
}

/* Modes */
.orb-silent .orb-inner {
  background: radial-gradient(circle at 30% 30%, #e0e0e0, #9e9e9e);
}

.orb-light .orb-inner {
  background: radial-gradient(circle at 30% 30%, #e0f7fa, #00bcd4);
  box-shadow: 0 0 15px rgba(0, 188, 212, 0.4);
}

.orb-active .orb-inner {
  background: radial-gradient(circle at 30% 30%, #fff3e0, #ff9800);
  box-shadow: 0 0 20px rgba(255, 152, 0, 0.6);
}

/* Loading */
.orb-loading .orb-ring {
  border-top-color: #ffffff;
  border-right-color: rgba(255, 255, 255, 0.5);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
