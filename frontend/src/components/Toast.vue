<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { CheckCircle, AlertCircle, X } from 'lucide-vue-next'

const props = defineProps<{
  message: string
  type: 'success' | 'error'
  duration?: number
}>()

const emit = defineEmits(['close'])
const visible = ref(true)

const close = () => {
  visible.value = false
  setTimeout(() => emit('close'), 300)
}

onMounted(() => {
  if (props.duration !== 0) {
    setTimeout(close, props.duration || 3000)
  }
})
</script>

<template>
  <Transition name="toast-fade">
    <div v-if="visible" class="toast-wrapper" :class="type">
      <div class="toast-content">
        <CheckCircle v-if="type === 'success'" :size="20" class="icon" />
        <AlertCircle v-else :size="20" class="icon" />
        <span class="message">{{ message }}</span>
        <button class="close-btn" @click="close">
          <X :size="16" />
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.toast-wrapper {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 9999;
  min-width: 300px;
  padding: 16px 24px;
  border-radius: 16px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
}

.toast-wrapper.success {
  background: rgba(74, 222, 128, 0.9);
  color: #064e3b;
}

.toast-wrapper.error {
  background: rgba(248, 113, 113, 0.9);
  color: #7f1d1d;
}

.toast-content {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 12px;
}

.icon {
  flex-shrink: 0;
}

.message {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
}

.close-btn {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 4px;
  display: flex;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.close-btn:hover {
  opacity: 1;
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.3s ease;
}

.toast-fade-enter-from {
  opacity: 0;
  transform: translate(-50%, -40%);
}

.toast-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -60%);
}
</style>
