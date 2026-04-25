import axios from 'axios'

const api = axios.create({
  baseURL: '/',
})

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('schematic2_token')
  if (token) {
    cfg.headers.Authorization = `Bearer ${token}`
  }
  return cfg
})

export default api
