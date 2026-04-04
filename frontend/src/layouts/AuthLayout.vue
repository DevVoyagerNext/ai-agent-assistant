<script setup lang="ts">
import BackgroundSlider from '../components/BackgroundSlider.vue'
</script>

<template>
  <div class="auth-layout">
    <BackgroundSlider />
    <router-view />
  </div>
</template>

<style scoped>
.auth-layout {
  width: 100vw;
  height: 100vh;
  position: relative;
  overflow: hidden;
}
</style>
