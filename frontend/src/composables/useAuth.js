import { ref } from 'vue'

const TOKEN_KEY = 'schematic2_token'

// module-level shared state – same ref across all components
const isLoggedIn = ref(!!localStorage.getItem(TOKEN_KEY))

export function useAuth() {
  function setLoggedIn(value) {
    isLoggedIn.value = value
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    isLoggedIn.value = false
  }

  return { isLoggedIn, setLoggedIn, logout }
}
