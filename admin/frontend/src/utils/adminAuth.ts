import type { AuthResponse } from '../types/admin'

const SESSION_KEY = 'adminSession'
const LEGACY_TOKEN_KEY = 'adminToken'
const EXPIRES_KEY = 'adminExpiresAt'
const ADMIN_KEY = 'admin'
const DEVICE_KEY = 'adminDeviceId'

const randomID = () => {
  if (crypto?.randomUUID) return crypto.randomUUID()
  return `device-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

export const getOrCreateAdminDeviceId = () => {
  let deviceId = localStorage.getItem(DEVICE_KEY)
  if (!deviceId) {
    deviceId = randomID()
    localStorage.setItem(DEVICE_KEY, deviceId)
  }
  return deviceId
}

export const getAdminSession = () => {
  return localStorage.getItem(SESSION_KEY) || ''
}

export const clearAdminAuth = () => {
  localStorage.removeItem(SESSION_KEY)
  localStorage.removeItem(LEGACY_TOKEN_KEY)
  localStorage.removeItem(EXPIRES_KEY)
  localStorage.removeItem(ADMIN_KEY)
}

export const isAdminSessionValid = () => {
  const session = getAdminSession()
  if (!session) return false

  const expiresAt = Number(localStorage.getItem(EXPIRES_KEY) || 0)
  if (expiresAt > 0 && expiresAt <= Math.floor(Date.now() / 1000)) {
    clearAdminAuth()
    return false
  }
  return true
}

export const saveAdminAuth = (payload: AuthResponse) => {
  const session = payload.session || payload.sessionId
  localStorage.setItem(SESSION_KEY, session)
  localStorage.setItem(EXPIRES_KEY, String(payload.expiresAt))
  localStorage.setItem(ADMIN_KEY, JSON.stringify(payload.admin))
  localStorage.removeItem(LEGACY_TOKEN_KEY)
}
