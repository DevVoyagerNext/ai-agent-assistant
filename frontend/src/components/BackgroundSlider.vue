<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const assets = [
  // 这里目前只有一个视频，如果有更多图片或视频可以继续添加
  { type: 'video', url: new URL('../assets/first/bg-video.mp4', import.meta.url).href },
]

const currentIndex = ref(0)
const intervalId = ref<number | null>(null)

const nextAsset = () => {
  if (assets.length <= 1) return
  currentIndex.value = (currentIndex.value + 1) % assets.length
}

onMounted(() => {
  if (assets.length > 1) {
    intervalId.value = window.setInterval(nextAsset, 20000)
  }
})

onUnmounted(() => {
  if (intervalId.value) clearInterval(intervalId.value)
})
</script>

<template>
  <div class="bg-slider">
    <transition name="fade" mode="out-in">
      <div 
        :key="currentIndex"
        class="bg-item"
      >
        <video 
          v-if="assets[currentIndex].type === 'video'" 
          :src="assets[currentIndex].url" 
          autoplay 
          muted 
          loop 
          playsinline 
          class="media-content"
        ></video>
        <div 
          v-else 
          class="media-content img-content" 
          :style="{ backgroundImage: `url(${assets[currentIndex].url})` }"
        ></div>
      </div>
    </transition>
    <div class="overlay"></div>
  </div>
</template>

<style scoped>
.bg-slider {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: -1;
  overflow: hidden;
}

.bg-item {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.media-content {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.img-content {
  background-size: cover;
  background-position: center;
}

.overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.4);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 1.5s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
