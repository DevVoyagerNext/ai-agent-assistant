import { defineStore } from 'pinia'
import { getUserProgress } from '../api/userProgress'
import type { Node } from '../types/userProgress'

export const useUserProgressStore = defineStore('userProgress', {
  state: () => ({
    nodes: [] as Node[],
    loading: false,
  }),
  actions: {
    async fetchProgress() {
      this.loading = true
      try {
        const data = await getUserProgress()
        this.nodes = data
      } catch (error) {
        console.error('Failed to fetch user progress:', error)
      } finally {
        this.loading = false
      }
    }
  }
})
