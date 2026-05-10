import { ref } from 'vue'

// Global state for toasts
const toastsState = ref([])
let nextId = 0

function showToast(message, type = 'info', duration = 5000) {
  const id = nextId++
  const toast = { id, message, type }
  
  // Force reactivity by creating a new array
  toastsState.value = [...toastsState.value, toast]

  if (duration > 0) {
    setTimeout(() => {
      removeToast(id)
    }, duration)
  }

  return id
}

function removeToast(id) {
  toastsState.value = toastsState.value.filter(t => t.id !== id)
}

function success(message, duration = 5000) {
  return showToast(message, 'success', duration)
}

function error(message, duration = 5000) {
  return showToast(message, 'error', duration)
}

function info(message, duration = 5000) {
  return showToast(message, 'info', duration)
}

function warning(message, duration = 5000) {
  return showToast(message, 'warning', duration)
}

export function useToast() {
  return {
    toasts: toastsState,
    showToast,
    removeToast,
    success,
    error,
    info,
    warning,
  }
}
