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
}

// ---------------- 私人笔记分层 API 类型 ----------------

export interface PrivateNoteBase {
  id: number
  title: string
  updatedAt: string
  type: 'folder' | 'markdown'
  isPublic: 0 | 1
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

export interface SharePrivateNoteReq {
  expiresAt: string
}

export interface SharePrivateNoteRes {
  shareToken: string
  shareCode: string
  expiresAt: string
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
