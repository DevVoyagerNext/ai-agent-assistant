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
  score: number
}

export interface ActivityCalendarRes {
  activities: ActivityCalendarItem[]
}

export interface PublicPrivateNoteItem {
  id: number
  title: string
  updatedAt: string
  type?: 'folder' | 'markdown'
  isPublic: 0 | 1
  isShared?: boolean
}

// ---------------- 私人笔记分层 API 类型 ----------------

export interface PrivateNoteBase {
  id: number
  title: string
  updatedAt: string
  type: 'folder' | 'markdown'
  isPublic: 0 | 1
  isShared?: boolean
}

export interface PrivateFolderContent {
  type: 'folder'
  total: number
  children: PrivateNoteBase[]
}

export interface PrivateMarkdownDetail {
  id: number
  title: string
  content: string
  updatedAt?: string
  isPublic: 0 | 1
  isShared?: boolean
}

export interface PrivateMarkdownContent {
  type: 'markdown'
  content: PrivateMarkdownDetail
}

export type PrivateNoteResponse = PrivateFolderContent | PrivateMarkdownContent

export interface CreatePrivateNoteReq {
  parentId: number
  type: 'folder' | 'markdown'
  title: string
  content?: string
  isPublic: 0 | 1
}

export interface PublicPrivateNotesRes {
  total: number
  list: PublicPrivateNoteItem[]
}

export interface SharedNoteItem {
  id: number
  nodeId: number
  noteTitle: string
  nodeName?: string
  noteType: string
  shareToken: string
  shareCode: string
  viewCount: number
  isActive: boolean
  createdAt: string
  expiresAt: string
}

export interface SharedNotesRes {
  total: number
  list: SharedNoteItem[]
}

export interface SharePrivateNoteReq {
  expiresAt: string
}

export interface SharePrivateNoteRes {
  shareToken: string
  shareCode: string
  expiresAt: string
}

export interface ShareBasicInfoRes {
  authorName: string
  authorAvatar: string
  noteTitle: string
  noteType: string
  isActive: boolean
  isExpired: boolean
  noteId?: number
}

export interface ShareAccessReq {
  shareToken: string
  shareCode: string
  privateNoteId?: number
}

export interface ShareAccessRes {
  type: 'folder' | 'markdown'
  title: string
  content: string
  children: PrivateNoteBase[] | null
  parent: PrivateNoteBase | null
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

// ---------------- 新增教材相关类型 ----------------

import type { Subject } from './subject'

export interface CollectFolderRes {
  id: number
  name: string
  description: string
  isPublic: number
  createdAt: string
  updatedAt: string
}

export interface UserSubjectProgressRes extends Subject {
  status: string
  lastStudyTime: string
}

export interface RecentSubjectListRes {
  total: number
  list: UserSubjectProgressRes[]
}
