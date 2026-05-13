import axios from 'axios'

// __API_BASE__ is injected by vite.config.js at build time
// e.g. '/schematics2/' for reverse-proxy or '/' for direct access
const api = axios.create({
  baseURL: typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/',
})

// Store for router and auth callbacks (set by main.js or App.vue)
let errorHandlers = {
  onUnauthorized: null,
}

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('schematic2_token')
  if (token) {
    cfg.headers.Authorization = `Bearer ${token}`
  }
  return cfg
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle 401 Unauthorized (token expired or invalid)
    if (error.response?.status === 401) {
      if (errorHandlers.onUnauthorized) {
        errorHandlers.onUnauthorized()
      }
    }
    return Promise.reject(error)
  }
)

export function setApiErrorHandler(handlers) {
  Object.assign(errorHandlers, handlers)
}

export default api
