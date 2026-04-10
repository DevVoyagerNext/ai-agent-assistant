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
