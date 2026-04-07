<script setup lang="ts">
defineProps<{
  width?: string
  height?: string
  borderRadius?: string
  circle?: boolean
}>()
</script>

<template>
  <div 
    class="skeleton" 
    :class="{ 'skeleton-circle': circle }"
    :style="{ 
      width: width || '100%', 
      height: height || '20px', 
      borderRadius: circle ? '50%' : (borderRadius || '4px') 
    }"
  ></div>
</template>

<style scoped>
.skeleton {
  background: linear-gradient(90deg, #f0f2f5 25%, #e6e8eb 37%, #f0f2f5 63%);
  background-size: 400% 100%;
  animation: skeleton-loading 1.4s ease infinite;
}

.skeleton-circle {
  border-radius: 50%;
}

@keyframes skeleton-loading {
  0% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0 50%;
  }
}
</style>
