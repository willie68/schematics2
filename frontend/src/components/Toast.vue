<template>
  <div class="toast-container">
    <!-- Toasts -->
    <div
      v-for="toast in displayedToasts"
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
import { computed } from 'vue'

const { toasts, removeToast } = useToast()

// Use computed to ensure Vue reactivity
const displayedToasts = computed(() => toasts.value)
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  z-index: 9999;
  pointer-events: none;
}

.toast {
  min-width: 300px;
  max-width: 400px;
  padding: 1rem;
  border-radius: 0.375rem;
  background-color: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  animation: slideIn 0.3s ease-in-out;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toast-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 1rem;
  word-break: break-word;
}

.toast-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: inherit;
  opacity: 0.7;
  padding: 0;
  min-width: 2.5rem;
  text-align: center;
  flex-shrink: 0;
  line-height: 1;
}

.toast-close:hover {
  opacity: 1;
}

.toast-info {
  background-color: #dbeafe;
  color: #1e40af;
  border-left: 4px solid #3b82f6;
}

.toast-success {
  background-color: #dcfce7;
  color: #166534;
  border-left: 4px solid #10b981;
}

.toast-warning {
  background-color: #fef3c7;
  color: #92400e;
  border-left: 4px solid #f59e0b;
}

.toast-error {
  background-color: #fee2e2;
  color: #991b1b;
  border-left: 4px solid #ef4444;
}

@keyframes slideIn {
  from {
    transform: translateX(450px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.toast-debug {
  color: red;
  padding: 1rem;
  background: yellow;
  position: fixed;
  bottom: 10rem;
  right: 1.5rem;
  width: 300px;
  z-index: 10000;
  border: 2px solid red;
  font-weight: bold;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>

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
