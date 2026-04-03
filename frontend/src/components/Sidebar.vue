<script setup lang="ts">
import { BookOpen, CheckSquare, Sparkles } from 'lucide-vue-next'

const tasks = [
  { id: 1, title: '复习单链表', type: 'note', icon: BookOpen },
  { id: 2, title: '解决循环队列错题', type: 'error', icon: CheckSquare },
  { id: 3, title: '今日推荐：动态规划基础', type: 'rag', icon: Sparkles },
]
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <h2>任务看板</h2>
    </div>
    <div class="task-list">
      <div v-for="task in tasks" :key="task.id" class="task-item">
        <div class="task-icon">
          <component :is="task.icon" :size="20" />
        </div>
        <div class="task-info">
          <p>{{ task.title }}</p>
        </div>
      </div>
    </div>
    
    <div class="recent-notes">
      <h3>最近笔记</h3>
      <ul>
        <li>链表与指针的概念 (2024-03-25)</li>
        <li>栈的数组实现 (2024-03-24)</li>
      </ul>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 300px;
  height: 100%;
  background: white;
  border-left: 1px solid #e2e8f0;
  padding: 20px;
  box-shadow: -4px 0 10px rgba(0,0,0,0.02);
  display: flex;
  flex-direction: column;
}

.sidebar-header h2 {
  font-size: 1.25rem;
  color: #1e293b;
  margin-bottom: 20px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #f1f5f9;
  cursor: pointer;
  transition: all 0.2s;
}

.task-item:hover {
  border-color: #60a5fa;
  transform: translateX(-4px);
}

.task-icon {
  color: #60a5fa;
}

.task-info p {
  margin: 0;
  font-size: 0.9rem;
  color: #334155;
}

.recent-notes {
  margin-top: 40px;
}

.recent-notes h3 {
  font-size: 1rem;
  color: #64748b;
  margin-bottom: 12px;
}

.recent-notes ul {
  list-style: none;
  padding: 0;
}

.recent-notes li {
  font-size: 0.85rem;
  color: #475569;
  padding: 8px 0;
  border-bottom: 1px solid #f1f5f9;
}
</style>
