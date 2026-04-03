<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import BackgroundSlider from './components/BackgroundSlider.vue'

const route = useRoute()
// 只有在登录和注册页面才显示这个全局背景
const showGlobalBg = computed(() => {
  return ['Login', 'Register'].includes(route.name as string)
})
</script>

<template>
  <div id="app-container">
    <BackgroundSlider v-if="showGlobalBg" />
    <router-view />
  </div>
</template>

<style>
#app-container {
  width: 100vw;
  height: 100vh;
  margin: 0;
  padding: 0;
  overflow: hidden;
}
</style>
