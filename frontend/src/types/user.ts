export interface UserInfoRes {
  username: string
  email: string
  avatarUrl: string
  signature: string
  followersCount: number
  followingCount: number
  learnedSubjectsCount: number
  sharedNotesCount: number
}

export interface ActivityCalendarItem {
  date: string
  count: number
}

export interface ActivityCalendarRes {
  activities: ActivityCalendarItem[]
}

export interface PublicPrivateNoteItem {
  id: number
  title: string
  updatedAt: string
}

export interface PublicPrivateNotesRes {
  total: number
  list: PublicPrivateNoteItem[]
}

export interface SharedNoteItem {
  id: number
  nodeId: number
  nodeName: string
  shareToken: string
  viewCount: number
  createdAt: string
  expiresAt: string
}

export interface SharedNotesRes {
  total: number
  list: SharedNoteItem[]
}

export interface LearnedSubjectItem {
  subjectId: number
  subjectName: string
  coverImage: string
  learned: number
  total: number
}

export interface LearnedSubjectsRes {
  list: LearnedSubjectItem[]
}
