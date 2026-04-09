<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, BookOpen, Edit, FlaskConical } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const nodeId = route.params.node_id
</script>

<template>
  <div class="study-room">
    <header class="study-header">
      <button class="back-btn" @click="router.push('/graph')">
        <ArrowLeft :size="20" />
        返回图谱
      </button>
      <h2>正在学习：{{ nodeId }}</h2>
    </header>

    <div class="study-layout">
      <!-- 三位一体界面: 对话学习 + 笔记编辑 + 代码实验室 -->
      <section class="ai-chat">
        <div class="section-title">
          <BookOpen :size="18" />
          对话学习
        </div>
        <div class="chat-placeholder">
          <p>AI 正在为您准备教学素材...</p>
        </div>
      </section>

      <section class="note-editor">
        <div class="section-title">
          <Edit :size="18" />
          笔记编辑
        </div>
        <textarea placeholder="在此记录您的心得..."></textarea>
      </section>

      <section class="code-lab">
        <div class="section-title">
          <FlaskConical :size="18" />
          代码实验室
        </div>
        <div class="code-placeholder">
          <code>// 在此编写实验代码</code>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.study-room {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f1f5f9;
}

.study-header {
  padding: 16px 24px;
  background: white;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  gap: 24px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: none;
  border: 1px solid #e2e8f0;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  color: #475569;
}

.back-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.study-layout {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 16px;
  padding: 16px;
  height: calc(100vh - 73px);
}

section {
  background: white;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.section-title {
  padding: 12px 16px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  color: #1e293b;
}

.chat-placeholder, .code-placeholder {
  flex: 1;
  padding: 20px;
  color: #64748b;
  font-size: 0.9rem;
}

.note-editor textarea {
  flex: 1;
  padding: 20px;
  border: none;
  outline: none;
  font-size: 1rem;
  line-height: 1.6;
  resize: none;
  color: #334155;
}

.code-placeholder {
  background: #1e293b;
  color: #94a3b8;
  font-family: monospace;
}
</style>
