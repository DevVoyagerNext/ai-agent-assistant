export interface SubjectReviewItem {
  id: number
  creatorId: number
  creatorName: string
  creatorEmail: string
  slug: string
  name: string
  nameDraft: string
  icon: string
  iconDraft: string
  description: string
  descriptionDraft: string
  coverImageId: number
  coverImageIdDraft: number
  status: string
  auditStatus: number
  hasDraft: number
  createdAt: string
}

export interface NodeReviewItem {
  id: number
  subjectId: number
  subjectName: string
  creatorId: number
  creatorName: string
  creatorEmail: string
  parentId: number
  path: string
  name: string
  nameDraft: string
  status: string
  auditStatus: number
  hasDraft: number
  draftDescendantCount: number
  level: number
  isLeaf: number
  content: string
  contentDraft: string
  updatedAt: string
}

export interface AuditRecord {
  id: number
  targetType: 'subject' | 'node'
  targetId: number
  targetName: string
  subjectName: string
  subjectDraftName: string
  adminId: number
  adminName: string
  action: 'approve' | 'reject'
  remark: string
  createdAt: string
}
