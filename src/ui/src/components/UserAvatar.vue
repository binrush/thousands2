<template>
  <component 
    :is="to ? 'RouterLink' : 'div'" 
    :class="wrapperClasses"
  >
    <img 
      v-if="imageUrl" 
      :src="getImageUrl(imageUrl)" 
      :alt="altText"
      class="h-12 w-12 rounded-full object-cover"
    >
    <img 
      v-else 
      src="/climber_no_photo.svg" 
      :alt="altText" 
      class="h-12 w-12 rounded-full"
    >
  </component>
</template>

<script setup>
import { computed } from 'vue'
import { getImageUrl } from '../utils/images'

const props = defineProps({
  imageUrl: {
    type: String,
    default: null
  },
  altText: {
    type: String,
    required: true
  },
  to: {
    type: [String, Object],
    default: null
  }
})

const wrapperClasses = computed(() => {
  const base = 'flex-shrink-0'
  return props.to ? `${base} inline-block` : base
})
</script>

<style scoped>
/* No additional styles needed - using Tailwind */
</style>

