<template>
  <div class="bg-red-50 border border-red-200 rounded-lg p-4 text-center text-red-600">
    <slot>{{ message }}</slot>
  </div>
</template>

<script setup>
defineProps({
  message: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
/* All styling via Tailwind */
</style>

