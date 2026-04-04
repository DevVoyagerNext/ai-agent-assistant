export interface Node {
  id: string
  name: string
  status: 'unstarted' | 'learning' | 'completed'
  subject: string
  parentId?: string
  x?: number
  y?: number
}
