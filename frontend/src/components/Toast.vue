<template>
  <div class="toast-container">
    <div v-if="toasts.value.length > 0" style="color: red; padding: 1rem; background: yellow; position: fixed; bottom: 10rem; right: 1.5rem; width: 300px;">
      DEBUG: {{ toasts.value.length }} toasts - {{ toasts.value.map(t => t.message).join(', ') }}
    </div>
    <div
      v-for="toast in toasts.value"
      :key="toast.id"
      :class="['toast', `toast-${toast.type}`]"
    >
      <div class="toast-content">
        <span>{{ toast.message }}</span>
        <button class="toast-close" @click="removeToast(toast.id)">✕</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useToast } from '../composables/useToast'

const { toasts, removeToast } = useToast()
console.log('Toast component mounted, toasts:', toasts)
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  pointer-events: none;
}

.toast {
  background: white;
  border-radius: 0.375rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  padding: 1rem;
  min-width: 300px;
  max-width: 500px;
  pointer-events: auto;
  border-left: 4px solid;
}

.toast-info {
  border-left-color: #0ea5e9;
  background: #f0f9ff;
}

.toast-success {
  border-left-color: #10b981;
  background: #f0fdf4;
}

.toast-warning {
  border-left-color: #f59e0b;
  background: #fffbeb;
}

.toast-error {
  border-left-color: #ef4444;
  background: #fef2f2;
}

.toast-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  color: #1f2937;
  font-size: 0.875rem;
}

.toast-close {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  font-size: 1.125rem;
  padding: 0;
  flex-shrink: 0;
  transition: color 0.2s;
}

.toast-close:hover {
  color: #1f2937;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
