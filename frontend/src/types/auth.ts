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
