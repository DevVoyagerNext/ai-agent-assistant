import { defineStore } from 'pinia'

export interface Node {
  id: string
  name: string
  status: 'unstarted' | 'learning' | 'completed'
  subject: string
  parentId?: string
  x?: number
  y?: number
}

export const useUserProgressStore = defineStore('userProgress', {
  state: () => ({
    nodes: [] as Node[],
    loading: false,
  }),
  actions: {
    async fetchProgress() {
      this.loading = true
      // TODO: Replace with real API call: GET /api/v1/user/progress
      // Mocking data for now
      setTimeout(() => {
        this.nodes = [
          { id: '1', name: '数据结构', status: 'completed', subject: 'CS', x: 400, y: 50 },
          { id: '2', name: '链表', status: 'completed', subject: 'DataStructure', parentId: '1', x: 200, y: 150 },
          { id: '3', name: '栈', status: 'learning', subject: 'DataStructure', parentId: '1', x: 400, y: 150 },
          { id: '4', name: '队列', status: 'unstarted', subject: 'DataStructure', parentId: '1', x: 600, y: 150 },
          { id: '5', name: '单链表', status: 'completed', subject: 'DataStructure', parentId: '2', x: 100, y: 250 },
          { id: '6', name: '双链表', status: 'completed', subject: 'DataStructure', parentId: '2', x: 200, y: 250 },
          { id: '7', name: '循环链表', status: 'learning', subject: 'DataStructure', parentId: '2', x: 300, y: 250 },
        ]
        this.loading = false
      }, 500)
    }
  }
})
