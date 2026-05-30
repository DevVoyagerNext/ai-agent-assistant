export interface SubjectNode {
  id: number;
  subjectId: number;
  parentId: number;
  path: string;
  name: string;
  level: number;
  isLeaf: number;
  sortOrder: number;
  imageId: number;
  easyCount: number;
  mediumCount: number;
  hardCount: number;
  userProgressStatus: 'unstarted' | 'learning' | 'completed';
}

export interface SubjectNodeDetail extends SubjectNode {
  content: string;
}

export interface NodeNote {
  id: number;
  nodeId: number;
  noteContent: string;
  isImportant: number;
  updatedAt: string;
}

export interface AuthorNode {
  id: number;
  subjectId: number;
  parentId: number;
  name: string;
  nameDraft: string;
  status: 'draft' | 'published' | 'archived';
  auditStatus: number;
  hasDraft: number;
  draftDescendantCount: number;
  path: string;
  isLeaf: number;
}

export interface AuthorNodeContent {
  content: string;
  contentDraft: string;
  auditStatus: number;
  hasDraft: number;
  isLeaf: number;
}

export interface AuthorInitRes {
  lastNodeId: number;
  nodeList: AuthorNode[];
  subjectAuditStatus?: number;
  subjectStatus?: 'draft' | 'published' | 'archived' | string;
}

export interface CreateKnowledgeNodePayload {
  subjectId: number;
  parentId: number;
  nameDraft: string;
}

export interface UpdateKnowledgeNodeDraftPayload {
  subjectId: number;
  nameDraft: string;
}

export interface UpsertKnowledgeContentPayload {
  contentDraft: string;
}

export interface PublishKnowledgeNodePayload {
  includeDescendants?: boolean;
}

export interface PublishKnowledgeNodeResult {
  submittedNodeIds: number[];
  submittedCount: number;
  includeDescendants: boolean;
}
