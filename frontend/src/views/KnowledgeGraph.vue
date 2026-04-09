<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'
import KnowledgeTree from '../components/KnowledgeTree.vue'
import Sidebar from '../components/Sidebar.vue'
import AIAgent from '../components/AIAgent.vue'
import { useUserProgressStore } from '../store/useUserProgressStore'

const router = useRouter()
const store = useUserProgressStore()

onMounted(() => {
  store.fetchProgress()
})
</script>

<template>
  <div class="knowledge-graph-page">
    <div class="main-content">
      <header class="page-header">
        <div class="header-left">
          <button class="back-btn" @click="router.push('/')">
            <ArrowLeft :size="20" />
            返回大厅
          </button>
          <h1>知识全景图</h1>
        </div>
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

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px solid #e2e8f0;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  color: #475569;
  font-size: 14px;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
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
