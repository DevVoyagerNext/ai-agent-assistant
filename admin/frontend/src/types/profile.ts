export interface AdminInfo {
  id: number
  username: string
  email: string
  role: 'admin'
  signature: string
  avatarUrl: string
  lastLoginAt: string | null
  createdAt: string
}

export interface DashboardStats {
  pending: number
  pendingSubject: number
  pendingNode: number
  approved: number
  rejected: number
  todayReview: number
}

export interface MeResponse {
  admin: AdminInfo
  stats: DashboardStats
}
