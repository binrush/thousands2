<template>
  <div class="min-h-screen flex flex-col">
    <!-- Header -->
    <TheHeader />
    
    <!-- Content area with top padding for header -->
    <main class="flex-1 overflow-auto">
      <div class="container mx-auto px-6 py-4">
        <slot></slot>
      </div>
    </main>
    
    <!-- Footer -->
    <TheFooter />
  </div>
</template>

<script setup>
import TheHeader from '../components/TheHeader.vue'
import TheFooter from '../components/TheFooter.vue'
</script> 