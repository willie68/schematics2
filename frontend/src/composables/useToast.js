import { ref } from 'vue'

// Global state for toasts
const toastsState = ref([])
let nextId = 0

export function useToast() {
  function showToast(message, type = 'info', duration = 5000) {
    const id = nextId++
    const toast = { id, message, type }
    console.log('showToast called with:', message)
    console.log('toastsState.value:', toastsState.value)
    toastsState.value.push(toast)
    console.log('toastsState.value after push:', toastsState.value)

    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }

  function removeToast(id) {
    console.log('removeToast called with id:', id)
    const index = toastsState.value.findIndex((t) => t.id === id)
    if (index !== -1) {
      toastsState.value.splice(index, 1)
    }
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
