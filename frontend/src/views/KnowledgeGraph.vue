<script setup lang="ts">
import { onMounted } from 'vue'
import KnowledgeTree from '../components/KnowledgeTree.vue'
import Sidebar from '../components/Sidebar.vue'
import AIAgent from '../components/AIAgent.vue'
import { useUserProgressStore } from '../store/userProgress'

const store = useUserProgressStore()

onMounted(() => {
  store.fetchProgress()
})
</script>

<template>
  <div class="knowledge-graph-page">
    <div class="main-content">
      <header class="page-header">
        <h1>知识全景图</h1>
        <div class="user-stats">
          <div class="stat-item">
            <span class="label">已掌握节点:</span>
            <span class="value">{{ store.nodes.filter(n => n.status === 'completed').length }}</span>
          </div>
          <div class="stat-item">
            <span class="label">正在学习:</span>
            <span class="value">{{ store.nodes.filter(n => n.status === 'learning').length }}</span>
          </div>
        </div>
      </header>
      
      <div class="graph-viewport">
        <KnowledgeTree :nodes="store.nodes" />
      </div>
    </div>
    
    <Sidebar />
    <AIAgent />
  </div>
</template>

<style scoped>
.knowledge-graph-page {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #f8fafc;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.page-header {
  padding: 24px 32px;
  background: white;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h1 {
  font-size: 1.5rem;
  color: #0f172a;
  margin: 0;
}

.user-stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  display: flex;
  gap: 8px;
  align-items: baseline;
}

.stat-item .label {
  font-size: 0.875rem;
  color: #64748b;
}

.stat-item .value {
  font-size: 1.125rem;
  font-weight: 600;
  color: #3b82f6;
}

.graph-viewport {
  flex: 1;
  position: relative;
  overflow: hidden;
}
</style>
