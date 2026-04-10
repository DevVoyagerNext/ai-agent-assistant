export interface SubjectCategory {
  id: number;
  name: string;
  slug: string;
  icon: string;
  sortOrder: number;
}

export interface Subject {
  id: number;
  slug: string;
  name: string;
  icon: string;
  description: string;
  coverImageId: number;
  isLiked: boolean;
  isCollected: boolean;
  progressPercent: number;
  lastNodeId: number;
}

export interface SubjectSearchRes {
  total: number;
  list: Subject[];
}
