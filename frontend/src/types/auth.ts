export interface AuthUser {
  username: string
  email: string
  avatarUrl: string
  signature: string
}

export interface AuthData {
  token: string
  refreshToken: string
  expiresAt: number
  user: AuthUser
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  code: string
  signature?: string
}

export interface SendCodeRequest {
  email: string
}
