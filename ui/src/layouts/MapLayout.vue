<template>
  <div class="min-h-screen grid grid-rows-[auto_1fr_auto]">
    <!-- Header -->
    <TheHeader />
    
    <!-- Map container taking full available space -->
    <main class="flex-grow overflow-hidden">
      <slot></slot>
    </main>
    
    <!-- Footer -->
    <TheFooter class="z-10" />
  </div>
</template>

<script setup>
import TheHeader from '../components/TheHeader.vue'
import TheFooter from '../components/TheFooter.vue'
</script>

<style scoped>
</style> 