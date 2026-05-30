import axios, { AxiosHeaders } from 'axios'
import { clearAdminAuth, getAdminSession, getOrCreateAdminDeviceId } from './adminAuth'

const request = axios.create({
  baseURL: import.meta.env.VITE_ADMIN_API_BASE || 'http://localhost:8081/v1',
  timeout: 30000
})

request.interceptors.request.use(config => {
  const headers = AxiosHeaders.from(config.headers)
  headers.set('x-device-id', getOrCreateAdminDeviceId())
  const session = getAdminSession()
  if (session) {
    headers.set('x-session-id', session)
    headers.set('Authorization', `Session ${session}`)
  }
  config.headers = headers
  return config
})

request.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      clearAdminAuth()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
