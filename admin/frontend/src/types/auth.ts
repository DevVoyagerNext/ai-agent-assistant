import type { AdminInfo } from './profile'

export interface AuthResponse {
  session: string
  sessionId: string
  expiresAt: number
  admin: AdminInfo
  device?: {
    deviceId: string
    userAgent: string
    ip: string
    loginAt: number
  }
}

export interface LoginPayload {
  account: string
  password: string
}

export interface RegisterPayload {
  username: string
  email: string
  password: string
  inviteCode: string
  signature: string
}
