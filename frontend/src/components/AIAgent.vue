<script setup lang="ts">
import { ref } from 'vue'
import { MessageSquare, Send } from 'lucide-vue-next'

const isOpen = ref(true)
const message = ref('')

const toggle = () => {
  isOpen.value = !isOpen.value
}
</script>

<template>
  <div class="ai-agent-container" :class="{ 'is-open': isOpen }">
    <div class="agent-trigger" @click="toggle">
      <MessageSquare :size="24" color="white" />
    </div>

    <div v-if="isOpen" class="agent-window">
      <div class="agent-header">
        <div class="agent-avatar">🤖</div>
        <div class="agent-info">
          <h4>AI 助手</h4>
          <p>在线</p>
        </div>
      </div>
      <div class="agent-messages">
        <div class="message system">
          欢迎回来！上次我们学到了『循环链表』，你还有 2 条笔记没有整理完，今天要继续吗？
        </div>
      </div>
      <div class="agent-input">
        <input v-model="message" type="text" placeholder="我想学 Go 语言的 Map..." @keyup.enter="message = ''" />
        <button @click="message = ''">
          <Send :size="18" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-agent-container {
  position: fixed;
  bottom: 24px;
  right: 24px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 16px;
  z-index: 1000;
}

.agent-trigger {
  width: 56px;
  height: 56px;
  border-radius: 28px;
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  transition: transform 0.2s;
}

.agent-trigger:hover {
  transform: scale(1.1);
}

.agent-window {
  width: 320px;
  height: 450px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 8px 30px rgba(0,0,0,0.12);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e2e8f0;
}

.agent-header {
  padding: 16px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.agent-avatar {
  font-size: 1.5rem;
}

.agent-info h4 {
  margin: 0;
  font-size: 1rem;
  color: #1e293b;
}

.agent-info p {
  margin: 0;
  font-size: 0.75rem;
  color: #22c55e;
}

.agent-messages {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message {
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 0.9rem;
  line-height: 1.5;
}

.message.system {
  background: #f1f5f9;
  color: #334155;
  border-bottom-left-radius: 2px;
  align-self: flex-start;
}

.agent-input {
  padding: 12px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  gap: 8px;
}

.agent-input input {
  flex: 1;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 0.9rem;
  outline: none;
}

.agent-input input:focus {
  border-color: #3b82f6;
}

.agent-input button {
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
